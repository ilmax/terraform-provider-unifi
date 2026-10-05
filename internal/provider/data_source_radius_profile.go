package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/sdkcompat/radius"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
)

// radiusProfileDataSource exposes /v1/sites/{siteId}/radius/profiles, a
// GET-only endpoint in the real API (no create/update/delete), as a data
// source.
type radiusProfileDataSource struct {
	client *unifi.Client
	siteID string
}

type radiusProfileDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	SiteID          types.String `tfsdk:"site_id"`
	RadiusProfileID types.String `tfsdk:"radius_profile_id"`
	Name            types.String `tfsdk:"name"`
	Origin          types.String `tfsdk:"origin"`
}

func NewRadiusProfileDataSource() datasource.DataSource {
	return &radiusProfileDataSource{}
}

func (d *radiusProfileDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_radius_profile"
}

func (d *radiusProfileDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A configured RADIUS profile. This endpoint is GET-only in the real API, so it's a data source rather than a resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"site_id": schema.StringAttribute{
				Optional: true,
			},
			"radius_profile_id": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"origin": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *radiusProfileDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *radiusProfileDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config radiusProfileDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(config.SiteID, d.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	radiusProfileID := config.RadiusProfileID.ValueString()
	if radiusProfileID == "" {
		resp.Diagnostics.AddAttributeError(path.Root("radius_profile_id"), "Missing radius_profile_id", "radius_profile_id must be provided.")
		return
	}

	var result radius.ListProfilesResponse
	apiPath := fmt.Sprintf("/v1/sites/%s/radius/profiles", siteID)
	if err := d.client.Get(ctx, apiPath, &result); err != nil {
		resp.Diagnostics.AddError("Unable to list RADIUS profiles", err.Error())
		return
	}

	for _, item := range result.Data {
		if item.Id == radiusProfileID {
			state := radiusProfileDataSourceModel{
				ID:              types.StringValue(item.Id),
				SiteID:          types.StringValue(siteID),
				RadiusProfileID: types.StringValue(item.Id),
				Name:            types.StringValue(item.Name),
				Origin:          stringValueOrNull(item.Metadata.Origin),
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}

	resp.Diagnostics.AddError("RADIUS profile not found", fmt.Sprintf("RADIUS profile %q was not found in site %q.", radiusProfileID, siteID))
}
