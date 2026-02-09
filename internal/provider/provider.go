package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
	"github.com/ilmax/unifi-client-go/pkg/sitemanager"
)

const (
	providerTypeName = "unifi"
)

type unifiProvider struct {
	version string
}

type unifiProviderModel struct {
	APIKey        types.String `tfsdk:"api_key"`
	APIURL        types.String `tfsdk:"api_url"`
	SiteID        types.String `tfsdk:"site_id"`
	UserAgent     types.String `tfsdk:"user_agent"`
	AllowInsecure types.Bool   `tfsdk:"allow_insecure"`
}

type providerData struct {
	client *unifi.Client
	siteID string
}

// New returns a new provider instance.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &unifiProvider{version: version}
	}
}

func (p *unifiProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = providerTypeName
	resp.Version = p.version
}

func (p *unifiProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
			"api_url": schema.StringAttribute{
				Optional: true,
			},
			"site_id": schema.StringAttribute{
				Optional: true,
			},
			"user_agent": schema.StringAttribute{
				Optional: true,
			},
			"allow_insecure": schema.BoolAttribute{
				Optional: true,
			},
		},
	}
}

func (p *unifiProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config unifiProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.APIKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown API Key",
			"The provider cannot create the UniFi API client because the API key is unknown.",
		)
		return
	}

	apiKey := config.APIKey.ValueString()
	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing API Key",
			"The provider cannot create the UniFi API client because the API key is missing.",
		)
		return
	}

	apiURL := config.APIURL.ValueString()
	if apiURL == "" {
		apiURL = sitemanager.DefaultBaseURL
	}

	allowInsecure := false
	if !config.AllowInsecure.IsNull() && !config.AllowInsecure.IsUnknown() {
		allowInsecure = config.AllowInsecure.ValueBool()
	}

	userAgent := config.UserAgent.ValueString()
	if userAgent == "" {
		userAgent = fmt.Sprintf("terraform-provider-unifi/%s", p.version)
	}

	tflog.Info(ctx, "Configuring UniFi provider client", map[string]any{
		"api_url":        apiURL,
		"allow_insecure": allowInsecure,
		"site_id_set":    !config.SiteID.IsNull() && !config.SiteID.IsUnknown(),
		"user_agent":     userAgent,
	})

	client, err := unifi.NewClient(unifi.Config{
		APIKey:        apiKey,
		BaseURL:       apiURL,
		UserAgent:     userAgent,
		AllowInsecure: allowInsecure,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create UniFi API client",
			err.Error(),
		)
		return
	}

	data := &providerData{
		client: client,
	}
	if !config.SiteID.IsNull() && !config.SiteID.IsUnknown() {
		data.siteID = config.SiteID.ValueString()
	}

	resp.ResourceData = data
	resp.DataSourceData = data
}

func (p *unifiProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewFirewallZoneResource,
		NewNetworkResource,
		NewWifiResource,
		NewFirewallResource,
	}
}

func (p *unifiProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDeviceDataSource,
		NewWanDataSource,
	}
}
