// Package legacyprovider is a temporary Terraform provider for settings that
// only the controller's legacy API exposes. Resources here move to the main
// provider (and get deleted from this one) as the official Integrations API
// gains them.
package legacyprovider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/legacy"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
)

// providerTypeName differs from the main provider's "unifi" so both can be
// configured side by side.
const providerTypeName = "unifilegacy"

type legacyProvider struct{ version string }

type providerModel struct {
	APIKey        types.String `tfsdk:"api_key"`
	APIURL        types.String `tfsdk:"api_url"`
	SiteID        types.String `tfsdk:"site_id"`
	AllowInsecure types.Bool   `tfsdk:"allow_insecure"`
}

type providerData struct {
	client *legacy.Client
	siteID string
}

// New returns a new provider instance.
func New(version string) func() provider.Provider {
	return func() provider.Provider { return &legacyProvider{version: version} }
}

func (p *legacyProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = providerTypeName
	resp.Version = p.version
}

func (p *legacyProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Temporary provider for UniFi Network settings that only the legacy controller API exposes. " +
			"It authenticates with the same API key as the `unifi` provider and is retired feature by feature as the official Integrations API covers them.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "API key for the console.",
				Required:            true,
				Sensitive:           true,
			},
			"api_url": schema.StringAttribute{
				MarkdownDescription: "Console URL, e.g. `https://192.168.1.1`. The `unifi` provider's Integrations API URL is accepted too.",
				Required:            true,
			},
			"site_id": schema.StringAttribute{
				MarkdownDescription: "Default site identifier (the Integrations API site UUID) used by site-scoped resources and data sources. A local `site_id` overrides this provider value.",
				Optional:            true,
			},
			"allow_insecure": schema.BoolAttribute{
				MarkdownDescription: "Skip TLS certificate verification.",
				Optional:            true,
			},
		},
	}
}

func (p *legacyProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var cfg providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if cfg.APIKey.IsUnknown() || cfg.APIURL.IsUnknown() {
		resp.Diagnostics.AddError("Unknown configuration", "api_key and api_url must be known to configure the provider.")
		return
	}

	client, err := legacy.NewClient(unifi.Config{
		APIKey:        cfg.APIKey.ValueString(),
		BaseURL:       cfg.APIURL.ValueString(),
		UserAgent:     fmt.Sprintf("terraform-provider-unifi-legacy/%s", p.version),
		AllowInsecure: cfg.AllowInsecure.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to create legacy UniFi API client", err.Error())
		return
	}

	data := &providerData{client: client, siteID: cfg.SiteID.ValueString()}
	resp.ResourceData = data
	resp.DataSourceData = data
}

func (p *legacyProvider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{NewIPReservationResource, NewTeleportResource}
}

func (p *legacyProvider) DataSources(context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{NewIPReservationDataSource, NewTeleportDataSource}
}
