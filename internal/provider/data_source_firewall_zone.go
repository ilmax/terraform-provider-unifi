package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/sdkcompat/zones"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
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
			"name": schema.StringAttribute{
				Required: true,
			},
			"zone_id": schema.StringAttribute{
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

	name := strings.TrimSpace(config.Name.ValueString())
	if name == "" {
		resp.Diagnostics.AddAttributeError(path.Root("name"), "Missing name", "name must be provided.")
		return
	}

	apiPath := fmt.Sprintf("/v1/sites/%s/firewall/zones", siteID)
	var result zones.ListFirewallZonesResponse
	if err := d.client.Get(ctx, apiPath, &result); err != nil {
		resp.Diagnostics.AddError("Unable to read firewall zone", err.Error())
		return
	}

	matches := matchFirewallZonesByName(result.Data, name)
	if len(matches) == 0 {
		resp.Diagnostics.AddError("Firewall zone not found", fmt.Sprintf("Firewall zone with name %q was not found in site %q.", name, siteID))
		return
	}
	if len(matches) > 1 {
		resp.Diagnostics.AddError("Duplicate firewall zone names", fmt.Sprintf("Found %d firewall zones named %q in site %q. Use a unique zone name.", len(matches), name, siteID))
		return
	}

	zone := matches[0]
	state := firewallZoneDataSourceModel{
		ID:         types.StringValue(zone.Id),
		SiteID:     types.StringValue(siteID),
		ZoneID:     types.StringValue(zone.Id),
		Name:       types.StringValue(zone.Name),
		NetworkIDs: rawMessagesToStringList(zone.NetworkIds),
		Origin:     types.StringNull(),
	}
	if zone.Metadata != nil {
		state.Origin = stringValueOrNull(zone.Metadata.Origin)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func matchFirewallZonesByName(items []zones.ListFirewallZonesData, name string) []zones.ListFirewallZonesData {
	normalized := strings.TrimSpace(name)
	matches := make([]zones.ListFirewallZonesData, 0)
	for _, item := range items {
		if strings.EqualFold(strings.TrimSpace(item.Name), normalized) {
			matches = append(matches, item)
		}
	}
	return matches
}
