package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
	networkapi "github.com/ilmax/unifi-client-go/pkg/network"
)

type dnsPoliciesDataSource struct {
	client *unifi.Client
	siteID string
}

type dnsPoliciesDataSourceModel struct {
	ID       types.String                    `tfsdk:"id"`
	SiteID   types.String                    `tfsdk:"site_id"`
	Type     types.String                    `tfsdk:"type"`
	Domain   types.String                    `tfsdk:"domain"`
	Policies []dnsPoliciesDataSourceItemModel `tfsdk:"policies"`
}

type dnsPoliciesDataSourceItemModel struct {
	ID      types.String `tfsdk:"id"`
	Type    types.String `tfsdk:"type"`
	Enabled types.Bool   `tfsdk:"enabled"`
	Domain  types.String `tfsdk:"domain"`
	Origin  types.String `tfsdk:"origin"`
}

func NewDNSPoliciesDataSource() datasource.DataSource {
	return &dnsPoliciesDataSource{}
}

func (d *dnsPoliciesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_policies"
}

func (d *dnsPoliciesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"site_id": schema.StringAttribute{
				Optional: true,
			},
			"type": schema.StringAttribute{
				Optional: true,
			},
			"domain": schema.StringAttribute{
				Optional: true,
			},
			"policies": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"type": schema.StringAttribute{
							Computed: true,
						},
						"enabled": schema.BoolAttribute{
							Computed: true,
						},
						"domain": schema.StringAttribute{
							Computed: true,
						},
						"origin": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *dnsPoliciesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *dnsPoliciesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config dnsPoliciesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(config.SiteID, d.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	apiPath := fmt.Sprintf("/v1/sites/%s/dns-policies", siteID)
	var result networkapi.IntegrationDnsPolicyPageDto
	if err := d.client.Get(ctx, apiPath, &result); err != nil {
		resp.Diagnostics.AddError("Unable to list DNS policies", err.Error())
		return
	}

	typeFilter := strings.TrimSpace(config.Type.ValueString())
	domainFilter := strings.TrimSpace(config.Domain.ValueString())

	items := make([]dnsPoliciesDataSourceItemModel, 0, len(result.Data))
	for _, policy := range result.Data {
		if !dnsPolicyMatchesFilter(policy, typeFilter, domainFilter) {
			continue
		}

		domain := types.StringNull()
		if policy.Domain != nil {
			domain = stringValueOrNull(*policy.Domain)
		}

		items = append(items, dnsPoliciesDataSourceItemModel{
			ID:      types.StringValue(policy.Id.String()),
			Type:    stringValueOrNull(policy.Type),
			Enabled: types.BoolValue(policy.Enabled),
			Domain:  domain,
			Origin:  stringValueOrNull(policy.Metadata.Origin),
		})
	}

	state := dnsPoliciesDataSourceModel{
		ID:       types.StringValue(siteID),
		SiteID:   types.StringValue(siteID),
		Type:     config.Type,
		Domain:   config.Domain,
		Policies: items,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func dnsPolicyMatchesFilter(policy networkapi.DNSPolicy, typeFilter, domainFilter string) bool {
	if typeFilter != "" && !strings.EqualFold(policy.Type, typeFilter) {
		return false
	}

	if domainFilter == "" {
		return true
	}
	if policy.Domain == nil {
		return false
	}
	return strings.EqualFold(*policy.Domain, domainFilter)
}
