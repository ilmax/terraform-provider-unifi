package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/sdkcompat/zones"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
	"github.com/ilmax/unifi-client-go/pkg/errors"
)

type firewallZoneDataSource struct {
	client *unifi.Client
	siteID string
}

type firewallZoneDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	SiteID     types.String `tfsdk:"site_id"`
	ZoneID     types.String `tfsdk:"zone_id"`
	Name       types.String `tfsdk:"name"`
	NetworkIDs types.List   `tfsdk:"network_ids"`
	Origin     types.String `tfsdk:"origin"`
}

func NewFirewallZoneDataSource() datasource.DataSource {
	return &firewallZoneDataSource{}
}

func (d *firewallZoneDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_zone"
}

func (d *firewallZoneDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"site_id": schema.StringAttribute{
				Optional: true,
			},
			"zone_id": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"network_ids": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			"origin": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *firewallZoneDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*providerData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", "Expected provider data to be of type *providerData.")
		return
	}
	d.client = data.client
	d.siteID = data.siteID
}

func (d *firewallZoneDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config firewallZoneDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(config.SiteID, d.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	zoneID := config.ZoneID.ValueString()
	if zoneID == "" {
		resp.Diagnostics.AddAttributeError(path.Root("zone_id"), "Missing zone_id", "zone_id must be provided.")
		return
	}

	apiPath := fmt.Sprintf("/v1/sites/%s/firewall/zones/%s", siteID, zoneID)
	var result zones.GetFirewallZoneResponse
	if err := d.client.Get(ctx, apiPath, &result); err != nil {
		if errors.IsNotFoundError(err) {
			resp.Diagnostics.AddError("Firewall zone not found", fmt.Sprintf("Firewall zone %q was not found in site %q.", zoneID, siteID))
			return
		}
		resp.Diagnostics.AddError("Unable to read firewall zone", err.Error())
		return
	}

	state := firewallZoneDataSourceModel{
		ID:         types.StringValue(result.Id),
		SiteID:     types.StringValue(siteID),
		ZoneID:     types.StringValue(result.Id),
		Name:       types.StringValue(result.Name),
		NetworkIDs: rawMessagesToStringList(result.NetworkIds),
		Origin:     types.StringNull(),
	}
	if result.Metadata != nil {
		state.Origin = stringValueOrNull(result.Metadata.Origin)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
