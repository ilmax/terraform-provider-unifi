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

// dpiApplicationDataSource exposes /v1/dpi/applications, a GET-only,
// non-site-scoped reference-data endpoint in the real API, as a data
// source.
type dpiApplicationDataSource struct {
	client *unifi.Client
}

type dpiApplicationDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	ApplicationID types.Int64  `tfsdk:"application_id"`
	Name          types.String `tfsdk:"name"`
}

func NewDPIApplicationDataSource() datasource.DataSource {
	return &dpiApplicationDataSource{}
}

func (d *dpiApplicationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dpi_application"
}

func (d *dpiApplicationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A Deep Packet Inspection application identifier, e.g. for referencing in traffic rules. " +
			"This is global reference data (not scoped to a site) and GET-only in the real API, so it's a data " +
			"source rather than a resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"application_id": schema.Int64Attribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *dpiApplicationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *dpiApplicationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config dpiApplicationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	applicationID := config.ApplicationID.ValueInt64()

	var result dpi.ListApplicationsResponse
	if err := d.client.Get(ctx, "/v1/dpi/applications", &result); err != nil {
		resp.Diagnostics.AddError("Unable to list DPI applications", err.Error())
		return
	}

	for _, item := range result.Data {
		if item.Id == applicationID {
			state := dpiApplicationDataSourceModel{
				ID:            types.StringValue(strconv.FormatInt(item.Id, 10)),
				ApplicationID: types.Int64Value(item.Id),
				Name:          types.StringValue(item.Name),
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}

	resp.Diagnostics.AddError("DPI application not found", fmt.Sprintf("DPI application %d was not found.", applicationID))
}
