package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	networkapi "github.com/ilmax/terraform-provider-unifi/internal/sdkcompat/dns"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
	"github.com/ilmax/unifi-client-go/pkg/errors"
)

type dnsPolicyDataSource struct {
	client *unifi.Client
	siteID string
}

type dnsPolicyDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	SiteID   types.String `tfsdk:"site_id"`
	PolicyID types.String `tfsdk:"policy_id"`
	Type     types.String `tfsdk:"type"`
	Enabled  types.Bool   `tfsdk:"enabled"`
	Domain   types.String `tfsdk:"domain"`
	Origin   types.String `tfsdk:"origin"`
}

func NewDNSPolicyDataSource() datasource.DataSource {
	return &dnsPolicyDataSource{}
}

func (d *dnsPolicyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_policy"
}

func (d *dnsPolicyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"site_id": schema.StringAttribute{
				Optional: true,
			},
			"policy_id": schema.StringAttribute{
				Required: true,
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
	}
}

func (d *dnsPolicyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *dnsPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config dnsPolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(config.SiteID, d.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	policyID := config.PolicyID.ValueString()
	if policyID == "" {
		resp.Diagnostics.AddAttributeError(path.Root("policy_id"), "Missing policy_id", "policy_id must be provided.")
		return
	}

	apiPath := fmt.Sprintf("/v1/sites/%s/dns-policies/%s", siteID, policyID)
	var result networkapi.DNSPolicy
	if err := d.client.Get(ctx, apiPath, &result); err != nil {
		if errors.IsNotFoundError(err) {
			resp.Diagnostics.AddError("DNS policy not found", fmt.Sprintf("DNS policy %q was not found in site %q.", policyID, siteID))
			return
		}
		resp.Diagnostics.AddError("Unable to read DNS policy", err.Error())
		return
	}

	state, err := dnsPolicyDataSourceStateFromAPI(siteID, result)
	if err != nil {
		resp.Diagnostics.AddError("Unable to decode DNS policy", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func dnsPolicyDataSourceStateFromAPI(siteID string, policy networkapi.DNSPolicy) (dnsPolicyDataSourceModel, error) {
	base, err := dnsPolicyBaseFromAPI(policy)
	if err != nil {
		return dnsPolicyDataSourceModel{}, err
	}

	domain := types.StringNull()
	if base.Domain != nil {
		domain = stringValueOrNull(*base.Domain)
	}

	return dnsPolicyDataSourceModel{
		ID:       types.StringValue(base.Id.String()),
		SiteID:   types.StringValue(siteID),
		PolicyID: types.StringValue(base.Id.String()),
		Type:     stringValueOrNull(base.Type),
		Enabled:  types.BoolValue(base.Enabled),
		Domain:   domain,
		Origin:   stringValueOrNull(base.Metadata.Origin),
	}, nil
}
