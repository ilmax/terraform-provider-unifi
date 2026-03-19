package provider

import (
	"context"
	"fmt"
	"math"

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
	ID        types.String                            `tfsdk:"id"`
	Countries map[string]countriesDataSourceItemModel `tfsdk:"countries"`
}

type countriesDataSourceItemModel struct {
	Code types.String `tfsdk:"code"`
}

const (
	countriesPageSize = int32(200)
	countriesMaxPages = 1000
)

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
			"countries": schema.MapNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"code": schema.StringAttribute{
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

	allCountries, err := d.listAllCountries(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list countries", err.Error())
		return
	}

	state.ID = types.StringValue("countries")
	state.Countries = countriesFromAPI(allCountries)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *countriesDataSource) listAllCountries(ctx context.Context) ([]networkapi.CountryDefinition, error) {
	return collectCountriesPages(countriesPageSize, countriesMaxPages, func(offset, limit int32) (networkapi.CountryDefinitionPage, error) {
		apiPath := fmt.Sprintf("/v1/countries?offset=%d&limit=%d", offset, limit)
		var page networkapi.CountryDefinitionPage
		if err := d.client.Get(ctx, apiPath, &page); err != nil {
			return networkapi.CountryDefinitionPage{}, err
		}
		return page, nil
	})
}

func collectCountriesPages(pageSize int32, maxPages int, fetch func(offset, limit int32) (networkapi.CountryDefinitionPage, error)) ([]networkapi.CountryDefinition, error) {
	offset := int32(0)
	allCountries := make([]networkapi.CountryDefinition, 0)

	for page := 0; page < maxPages; page++ {
		result, err := fetch(offset, pageSize)
		if err != nil {
			return nil, err
		}

		pageCount := len(result.Data)
		if pageCount == 0 {
			return allCountries, nil
		}

		allCountries = append(allCountries, result.Data...)
		nextOffset := int64(offset) + int64(pageCount)
		if nextOffset > math.MaxInt32 {
			return nil, fmt.Errorf("countries pagination offset exceeded int32: %d", nextOffset)
		}

		if result.TotalCount > 0 && nextOffset >= result.TotalCount {
			return allCountries, nil
		}
		if int32(pageCount) < pageSize {
			return allCountries, nil
		}

		offset = int32(nextOffset)
	}

	return nil, fmt.Errorf("countries pagination exceeded %d pages", maxPages)
}

func countriesFromAPI(items []networkapi.CountryDefinition) map[string]countriesDataSourceItemModel {
	countries := make(map[string]countriesDataSourceItemModel, len(items))
	for _, item := range items {
		if item.Name == "" {
			continue
		}
		countries[item.Name] = countriesDataSourceItemModel{
			Code: stringValueOrNull(item.Code),
		}
	}
	return countries
}
