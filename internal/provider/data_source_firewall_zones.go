package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/sdkcompat/zones"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
)

type firewallZonesDataSource struct {
	client *unifi.Client
	siteID string
}

type firewallZonesDataSourceModel struct {
	ID     types.String                       `tfsdk:"id"`
	SiteID types.String                       `tfsdk:"site_id"`
	Name   types.String                       `tfsdk:"name"`
	Zones  []firewallZonesDataSourceZoneModel `tfsdk:"zones"`
}

type firewallZonesDataSourceZoneModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	NetworkIDs types.List   `tfsdk:"network_ids"`
	Origin     types.String `tfsdk:"origin"`
}

func NewFirewallZonesDataSource() datasource.DataSource {
	return &firewallZonesDataSource{}
}

func (d *firewallZonesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_zones"
}

func (d *firewallZonesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"site_id": schema.StringAttribute{
				Optional: true,
			},
			"name": schema.StringAttribute{
				Optional: true,
			},
			"zones": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
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
				},
			},
		},
	}
}

func (d *firewallZonesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *firewallZonesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config firewallZonesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(config.SiteID, d.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	apiPath := fmt.Sprintf("/v1/sites/%s/firewall/zones", siteID)
	var result zones.ListFirewallZonesResponse
	if err := d.client.Get(ctx, apiPath, &result); err != nil {
		resp.Diagnostics.AddError("Unable to list firewall zones", err.Error())
		return
	}

	nameFilter := strings.TrimSpace(config.Name.ValueString())
	list := make([]firewallZonesDataSourceZoneModel, 0, len(result.Data))
	for _, zone := range result.Data {
		if nameFilter != "" && !strings.EqualFold(zone.Name, nameFilter) {
			continue
		}

		item := firewallZonesDataSourceZoneModel{
			ID:         types.StringValue(zone.Id),
			Name:       types.StringValue(zone.Name),
			NetworkIDs: rawMessagesToStringList(zone.NetworkIds),
			Origin:     types.StringNull(),
		}
		if zone.Metadata != nil {
			item.Origin = stringValueOrNull(zone.Metadata.Origin)
		}
		list = append(list, item)
	}

	state := firewallZonesDataSourceModel{
		ID:     types.StringValue(siteID),
		SiteID: types.StringValue(siteID),
		Name:   config.Name,
		Zones:  list,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
