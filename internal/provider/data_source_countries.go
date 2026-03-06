package provider

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
	networkapi "github.com/ilmax/unifi-client-go/pkg/network"
)

type countriesDataSource struct {
	client *unifi.Client
}

type countriesDataSourceModel struct {
	ID        types.String                   `tfsdk:"id"`
	Countries []countriesDataSourceItemModel `tfsdk:"countries"`
}

type countriesDataSourceItemModel struct {
	Code types.String `tfsdk:"code"`
	Name types.String `tfsdk:"name"`
}

func NewCountriesDataSource() datasource.DataSource {
	return &countriesDataSource{}
}

func (d *countriesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_countries"
}

func (d *countriesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"countries": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"code": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *countriesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *countriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state countriesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result networkapi.CountryDefinitionPage
	if err := d.client.Get(ctx, "/v1/countries", &result); err != nil {
		resp.Diagnostics.AddError("Unable to list countries", err.Error())
		return
	}

	state.ID = types.StringValue("countries")
	state.Countries = countriesFromAPI(result.Data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func countriesFromAPI(items []networkapi.CountryDefinition) []countriesDataSourceItemModel {
	countries := make([]countriesDataSourceItemModel, 0, len(items))
	for _, item := range items {
		countries = append(countries, countriesDataSourceItemModel{
			Code: stringValueOrNull(item.Code),
			Name: stringValueOrNull(item.Name),
		})
	}

	// Keep a deterministic order for stable plans.
	sort.SliceStable(countries, func(i, j int) bool {
		return countries[i].Name.ValueString() < countries[j].Name.ValueString()
	})

	return countries
}
