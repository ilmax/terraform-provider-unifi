package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
	networkapi "github.com/ilmax/unifi-client-go/pkg/network"
)

const (
	sitePageSize = int32(200)
	siteMaxPages = 1000
)

type siteDataSource struct {
	client *unifi.Client
}

type siteDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	SiteID            types.String `tfsdk:"site_id"`
	InternalReference types.String `tfsdk:"internal_reference"`
}

func NewSiteDataSource() datasource.DataSource {
	return &siteDataSource{}
}

func (d *siteDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_site"
}

func (d *siteDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"site_id": schema.StringAttribute{
				Computed: true,
			},
			"internal_reference": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *siteDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *siteDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config siteDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := strings.TrimSpace(config.Name.ValueString())
	if name == "" {
		resp.Diagnostics.AddAttributeError(path.Root("name"), "Missing site name", "name must be provided.")
		return
	}

	sites, err := d.listAllSites(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list sites", err.Error())
		return
	}

	matches := matchSitesByName(sites, name)
	if len(matches) == 0 {
		resp.Diagnostics.AddError("Site not found", fmt.Sprintf("No site with name %q was found.", name))
		return
	}
	if len(matches) > 1 {
		resp.Diagnostics.AddError("Duplicate site names", fmt.Sprintf("Found %d sites named %q. Use a unique site name.", len(matches), name))
		return
	}

	state := siteStateFromAPI(matches[0])
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *siteDataSource) listAllSites(ctx context.Context) ([]networkapi.SiteOverview, error) {
	allSites := make([]networkapi.SiteOverview, 0)
	offset := int32(0)

	for page := 0; page < siteMaxPages; page++ {
		path := buildSiteListPath(sitePageSize, offset)
		var response networkapi.SiteOverviewPage
		if err := d.client.Get(ctx, path, &response); err != nil {
			return nil, err
		}

		allSites = append(allSites, response.Data...)
		pageCount := int32(len(response.Data))
		if pageCount == 0 {
			return allSites, nil
		}
		nextOffset := offset + pageCount
		if response.TotalCount > 0 && int64(nextOffset) >= response.TotalCount {
			return allSites, nil
		}
		if pageCount < sitePageSize {
			return allSites, nil
		}
		offset = nextOffset
	}

	return nil, fmt.Errorf("site pagination exceeded %d pages", siteMaxPages)
}

func buildSiteListPath(pageSize, offset int32) string {
	return fmt.Sprintf("/v1/sites?limit=%d&offset=%d", pageSize, offset)
}

func matchSitesByName(sites []networkapi.SiteOverview, name string) []networkapi.SiteOverview {
	normalizedName := strings.TrimSpace(name)
	matches := make([]networkapi.SiteOverview, 0)
	for _, site := range sites {
		if strings.EqualFold(strings.TrimSpace(site.Name), normalizedName) {
			matches = append(matches, site)
		}
	}
	return matches
}

func siteStateFromAPI(site networkapi.SiteOverview) siteDataSourceModel {
	return siteDataSourceModel{
		ID:                stringValueOrNull(site.Id.String()),
		Name:              stringValueOrNull(site.Name),
		SiteID:            stringValueOrNull(site.Id.String()),
		InternalReference: stringValueOrNull(site.InternalReference),
	}
}
