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

// vpnSiteToSiteTunnelDataSource exposes
// /v1/sites/{siteId}/vpn/site-to-site-tunnels, a GET-only endpoint in the
// real API (no create/update/delete), as a data source.
type vpnSiteToSiteTunnelDataSource struct {
	client *unifi.Client
	siteID string
}

type vpnSiteToSiteTunnelDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	SiteID   types.String `tfsdk:"site_id"`
	TunnelID types.String `tfsdk:"tunnel_id"`
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Origin   types.String `tfsdk:"origin"`
}

func NewVPNSiteToSiteTunnelDataSource() datasource.DataSource {
	return &vpnSiteToSiteTunnelDataSource{}
}

func (d *vpnSiteToSiteTunnelDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpn_site_to_site_tunnel"
}

func (d *vpnSiteToSiteTunnelDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A configured VPN site-to-site tunnel. This endpoint is GET-only in the real API, so it's a data source rather than a resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"site_id": schema.StringAttribute{
				Optional: true,
			},
			"tunnel_id": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"type": schema.StringAttribute{
				Computed: true,
			},
			"origin": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *vpnSiteToSiteTunnelDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *vpnSiteToSiteTunnelDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config vpnSiteToSiteTunnelDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(config.SiteID, d.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	tunnelID := config.TunnelID.ValueString()
	if tunnelID == "" {
		resp.Diagnostics.AddAttributeError(path.Root("tunnel_id"), "Missing tunnel_id", "tunnel_id must be provided.")
		return
	}

	var result vpn.ListSiteToSiteTunnelsResponse
	apiPath := fmt.Sprintf("/v1/sites/%s/vpn/site-to-site-tunnels", siteID)
	if err := d.client.Get(ctx, apiPath, &result); err != nil {
		resp.Diagnostics.AddError("Unable to list VPN site-to-site tunnels", err.Error())
		return
	}

	for _, item := range result.Data {
		if item.Id == tunnelID {
			state := vpnSiteToSiteTunnelDataSourceModel{
				ID:       types.StringValue(item.Id),
				SiteID:   types.StringValue(siteID),
				TunnelID: types.StringValue(item.Id),
				Name:     types.StringValue(item.Name),
				Type:     types.StringValue(item.Type),
				Origin:   stringValueOrNull(item.Metadata.Origin),
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}

	resp.Diagnostics.AddError("VPN site-to-site tunnel not found", fmt.Sprintf("VPN site-to-site tunnel %q was not found in site %q.", tunnelID, siteID))
}
