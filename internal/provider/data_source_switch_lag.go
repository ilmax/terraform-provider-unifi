package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/sdkcompat/switching"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
	"github.com/ilmax/unifi-client-go/pkg/errors"
)

// switchLagDataSource exposes /v1/sites/{siteId}/switching/lags/{lagId}, a
// GET-only endpoint in the real API (no create/update/delete), as a data
// source. Requires switch hardware to return anything on a real console.
type switchLagDataSource struct {
	client *unifi.Client
	siteID string
}

type switchLagDataSourceModel struct {
	ID     types.String `tfsdk:"id"`
	SiteID types.String `tfsdk:"site_id"`
	LagID  types.String `tfsdk:"lag_id"`
	Type   types.String `tfsdk:"type"`
	Origin types.String `tfsdk:"origin"`
}

func NewSwitchLagDataSource() datasource.DataSource {
	return &switchLagDataSource{}
}

func (d *switchLagDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_switch_lag"
}

func (d *switchLagDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A switch port LAG (link aggregation group). Requires an adopted switch to have anything to " +
			"read. This endpoint is GET-only in the real API, so it's a data source rather than a resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"site_id": schema.StringAttribute{
				Optional: true,
			},
			"lag_id": schema.StringAttribute{
				Required: true,
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

func (d *switchLagDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *switchLagDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config switchLagDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(config.SiteID, d.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	lagID := config.LagID.ValueString()
	if lagID == "" {
		resp.Diagnostics.AddAttributeError(path.Root("lag_id"), "Missing lag_id", "lag_id must be provided.")
		return
	}

	var result switching.Lag
	apiPath := fmt.Sprintf("/v1/sites/%s/switching/lags/%s", siteID, lagID)
	if err := d.client.Get(ctx, apiPath, &result); err != nil {
		if errors.IsNotFoundError(err) {
			resp.Diagnostics.AddError("Switch LAG not found", fmt.Sprintf("Switch LAG %q was not found in site %q.", lagID, siteID))
			return
		}
		resp.Diagnostics.AddError("Unable to read switch LAG", err.Error())
		return
	}

	state := switchLagDataSourceModel{
		ID:     types.StringValue(result.Id),
		SiteID: types.StringValue(siteID),
		LagID:  types.StringValue(result.Id),
		Type:   types.StringValue(result.Type),
		Origin: stringValueOrNull(result.Metadata.Origin),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
