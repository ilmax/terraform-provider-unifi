package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/sdkcompat/vpn"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
)

// vpnServerDataSource exposes /v1/sites/{siteId}/vpn/servers, a GET-only
// endpoint in the real API (no create/update/delete), as a data source.
type vpnServerDataSource struct {
	client *unifi.Client
	siteID string
}

type vpnServerDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	SiteID      types.String `tfsdk:"site_id"`
	VPNServerID types.String `tfsdk:"vpn_server_id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Origin      types.String `tfsdk:"origin"`
}

func NewVPNServerDataSource() datasource.DataSource {
	return &vpnServerDataSource{}
}

func (d *vpnServerDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpn_server"
}

func (d *vpnServerDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A configured VPN server. This endpoint is GET-only in the real API, so it's a data source rather than a resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"site_id": schema.StringAttribute{
				Optional: true,
			},
			"vpn_server_id": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"type": schema.StringAttribute{
				Computed: true,
			},
			"enabled": schema.BoolAttribute{
				Computed: true,
			},
			"origin": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *vpnServerDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *vpnServerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config vpnServerDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(config.SiteID, d.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	vpnServerID := config.VPNServerID.ValueString()
	if vpnServerID == "" {
		resp.Diagnostics.AddAttributeError(path.Root("vpn_server_id"), "Missing vpn_server_id", "vpn_server_id must be provided.")
		return
	}

	var result vpn.ListServersResponse
	apiPath := fmt.Sprintf("/v1/sites/%s/vpn/servers", siteID)
	if err := d.client.Get(ctx, apiPath, &result); err != nil {
		resp.Diagnostics.AddError("Unable to list VPN servers", err.Error())
		return
	}

	for _, item := range result.Data {
		if item.Id == vpnServerID {
			state := vpnServerDataSourceModel{
				ID:          types.StringValue(item.Id),
				SiteID:      types.StringValue(siteID),
				VPNServerID: types.StringValue(item.Id),
				Name:        types.StringValue(item.Name),
				Type:        types.StringValue(item.Type),
				Enabled:     types.BoolValue(item.Enabled),
				Origin:      stringValueOrNull(item.Metadata.Origin),
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}

	resp.Diagnostics.AddError("VPN server not found", fmt.Sprintf("VPN server %q was not found in site %q.", vpnServerID, siteID))
}
