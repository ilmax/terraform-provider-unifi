package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/sdkcompat/dpi"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
)

// dpiCategoryDataSource exposes /v1/dpi/categories, a GET-only,
// non-site-scoped reference-data endpoint in the real API, as a data
// source.
type dpiCategoryDataSource struct {
	client *unifi.Client
}

type dpiCategoryDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	CategoryID types.Int64  `tfsdk:"category_id"`
	Name       types.String `tfsdk:"name"`
}

func NewDPICategoryDataSource() datasource.DataSource {
	return &dpiCategoryDataSource{}
}

func (d *dpiCategoryDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dpi_category"
}

func (d *dpiCategoryDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A Deep Packet Inspection category identifier, e.g. for referencing in traffic rules. " +
			"This is global reference data (not scoped to a site) and GET-only in the real API, so it's a data " +
			"source rather than a resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"category_id": schema.Int64Attribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *dpiCategoryDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*providerData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", "Expected provider data to be of type *providerData.")
		return
	}
	d.client = data.client
}

func (d *dpiCategoryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config dpiCategoryDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	categoryID := config.CategoryID.ValueInt64()

	var result dpi.ListCategoriesResponse
	if err := d.client.Get(ctx, "/v1/dpi/categories", &result); err != nil {
		resp.Diagnostics.AddError("Unable to list DPI categories", err.Error())
		return
	}

	for _, item := range result.Data {
		if item.Id == categoryID {
			state := dpiCategoryDataSourceModel{
				ID:         types.StringValue(strconv.FormatInt(item.Id, 10)),
				CategoryID: types.Int64Value(item.Id),
				Name:       types.StringValue(item.Name),
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}

	resp.Diagnostics.AddError("DPI category not found", fmt.Sprintf("DPI category %d was not found.", categoryID))
}
