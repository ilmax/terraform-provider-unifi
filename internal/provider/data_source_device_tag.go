package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/sdkcompat/devicetags"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
)

// deviceTagDataSource exposes /v1/sites/{siteId}/device-tags, a GET-only
// endpoint in the real API (no create/update/delete), as a data source.
type deviceTagDataSource struct {
	client *unifi.Client
	siteID string
}

type deviceTagDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	SiteID      types.String `tfsdk:"site_id"`
	DeviceTagID types.String `tfsdk:"device_tag_id"`
	Name        types.String `tfsdk:"name"`
	DeviceIDs   types.List   `tfsdk:"device_ids"`
	Origin      types.String `tfsdk:"origin"`
}

func NewDeviceTagDataSource() datasource.DataSource {
	return &deviceTagDataSource{}
}

func (d *deviceTagDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_tag"
}

func (d *deviceTagDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A device tag grouping adopted devices under a shared label. This endpoint is GET-only in the real API, so it's a data source rather than a resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"site_id": schema.StringAttribute{
				Optional: true,
			},
			"device_tag_id": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"device_ids": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			"origin": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *deviceTagDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *deviceTagDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config deviceTagDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(config.SiteID, d.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	deviceTagID := config.DeviceTagID.ValueString()
	if deviceTagID == "" {
		resp.Diagnostics.AddAttributeError(path.Root("device_tag_id"), "Missing device_tag_id", "device_tag_id must be provided.")
		return
	}

	var result devicetags.ListTagsResponse
	apiPath := fmt.Sprintf("/v1/sites/%s/device-tags", siteID)
	if err := d.client.Get(ctx, apiPath, &result); err != nil {
		resp.Diagnostics.AddError("Unable to list device tags", err.Error())
		return
	}

	for _, item := range result.Data {
		if item.Id == deviceTagID {
			state := deviceTagDataSourceModel{
				ID:          types.StringValue(item.Id),
				SiteID:      types.StringValue(siteID),
				DeviceTagID: types.StringValue(item.Id),
				Name:        types.StringValue(item.Name),
				DeviceIDs:   stringsToList(item.DeviceIds),
				Origin:      stringValueOrNull(item.Metadata.Origin),
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}

	resp.Diagnostics.AddError("Device tag not found", fmt.Sprintf("Device tag %q was not found in site %q.", deviceTagID, siteID))
}
