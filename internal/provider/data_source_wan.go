package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
	"github.com/ilmax/unifi-client-go/pkg/errors"
	"github.com/ilmax/unifi-client-go/pkg/wans"
)

type wanDataSource struct {
	client *unifi.Client
	siteID string
}

type wanDataSourceModel struct {
	ID     types.String `tfsdk:"id"`
	SiteID types.String `tfsdk:"site_id"`
	WanID  types.String `tfsdk:"wan_id"`
	Name   types.String `tfsdk:"name"`
}

func NewWanDataSource() datasource.DataSource {
	return &wanDataSource{}
}

func (d *wanDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wan"
}

func (d *wanDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"site_id": schema.StringAttribute{
				Optional: true,
			},
			"wan_id": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *wanDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *wanDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config wanDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(config.SiteID, d.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	wanID := config.WanID.ValueString()
	if wanID == "" {
		resp.Diagnostics.AddAttributeError(path.Root("wan_id"), "Missing wan_id", "wan_id must be provided.")
		return
	}

	state, found, err := d.readWan(ctx, siteID, wanID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list WAN interfaces", err.Error())
		return
	}
	if !found {
		resp.Diagnostics.AddError("WAN interface not found", fmt.Sprintf("WAN interface %q was not found in site %q.", wanID, siteID))
		return
	}

	state.SiteID = types.StringValue(siteID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *wanDataSource) readWan(ctx context.Context, siteID, wanID string) (wanDataSourceModel, bool, error) {
	var result wans.ListWANInterfacesResponse
	path := fmt.Sprintf("/v1/sites/%s/wans", siteID)
	if err := d.client.Get(ctx, path, &result); err != nil {
		if errors.IsNotFoundError(err) {
			return wanDataSourceModel{}, false, nil
		}
		return wanDataSourceModel{}, false, err
	}

	for _, item := range result.Data {
		if item.Id == wanID {
			state := wanDataSourceModel{
				ID:    types.StringValue(item.Id),
				WanID: types.StringValue(item.Id),
				Name:  types.StringValue(item.Name),
			}
			return state, true, nil
		}
	}

	return wanDataSourceModel{}, false, nil
}
