package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	networkapi "github.com/ilmax/terraform-provider-unifi/internal/sdkcompat/dns"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
)

type dnsPoliciesDataSource struct {
	client *unifi.Client
	siteID string
}

type dnsPoliciesDataSourceModel struct {
	ID       types.String                     `tfsdk:"id"`
	SiteID   types.String                     `tfsdk:"site_id"`
	Type     types.String                     `tfsdk:"type"`
	Domain   types.String                     `tfsdk:"domain"`
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
		MarkdownDescription: "Lists DNS policies for a site, with optional filtering by type and domain.\n" +
			"\n" +
			"## DNS Policy Types\n" +
			"\n" +
			"The `type` filter and each returned policy `type` can currently be one of:\n" +
			"\n" +
			"- `A_RECORD`\n" +
			"- `AAAA_RECORD`\n" +
			"- `CNAME_RECORD`\n" +
			"- `FORWARD_DOMAIN`\n" +
			"- `MX_RECORD`\n" +
			"- `SRV_RECORD`\n" +
			"- `TXT_RECORD`",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier (site ID).",
				Computed:            true,
			},
			"site_id": schema.StringAttribute{
				MarkdownDescription: "Site identifier. Defaults to provider `site_id` when omitted. A local `site_id` overrides the provider setting.",
				Optional:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Case-insensitive DNS policy type filter. See supported values above.",
				Optional:            true,
			},
			"domain": schema.StringAttribute{
				MarkdownDescription: "Case-insensitive DNS policy domain filter.",
				Optional:            true,
			},
			"policies": schema.ListNestedAttribute{
				MarkdownDescription: "Matching DNS policies.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "DNS policy identifier.",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "DNS policy type. See supported values above.",
							Computed:            true,
						},
						"enabled": schema.BoolAttribute{
							MarkdownDescription: "DNS policy enabled state.",
							Computed:            true,
						},
						"domain": schema.StringAttribute{
							MarkdownDescription: "DNS policy domain (when present).",
							Computed:            true,
						},
						"origin": schema.StringAttribute{
							MarkdownDescription: "DNS policy origin metadata.",
							Computed:            true,
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

	apiPath := fmt.Sprintf("/v1/sites/%s/dns/policies", siteID)
	var result networkapi.IntegrationDnsPolicyPageDto
	if err := d.client.Get(ctx, apiPath, &result); err != nil {
		resp.Diagnostics.AddError("Unable to list DNS policies", err.Error())
		return
	}

	typeFilter := strings.TrimSpace(config.Type.ValueString())
	domainFilter := strings.TrimSpace(config.Domain.ValueString())

	items := make([]dnsPoliciesDataSourceItemModel, 0, len(result.Data))
	for _, policy := range result.Data {
		base, err := dnsPolicyBaseFromAPI(policy)
		if err != nil {
			resp.Diagnostics.AddError("Unable to decode DNS policy", err.Error())
			return
		}

		if !dnsPolicyMatchesFilter(base, typeFilter, domainFilter) {
			continue
		}

		domain := types.StringNull()
		if base.Domain != nil {
			domain = stringValueOrNull(*base.Domain)
		}

		items = append(items, dnsPoliciesDataSourceItemModel{
			ID:      types.StringValue(base.Id.String()),
			Type:    stringValueOrNull(base.Type),
			Enabled: types.BoolValue(base.Enabled),
			Domain:  domain,
			Origin:  stringValueOrNull(base.Metadata.Origin),
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

func dnsPolicyMatchesFilter(policy networkapi.DNSPolicyBase, typeFilter, domainFilter string) bool {
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
