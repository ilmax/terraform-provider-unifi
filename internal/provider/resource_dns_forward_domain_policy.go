package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	networkapi "github.com/ilmax/unifi-client-go/pkg/network"
)

type dnsForwardDomainPolicyResource struct {
	dnsPolicyResourceBase
}

type dnsForwardDomainPolicyResourceModel struct {
	ID        types.String `tfsdk:"id"`
	SiteID    types.String `tfsdk:"site_id"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	Domain    types.String `tfsdk:"domain"`
	Origin    types.String `tfsdk:"origin"`
	IPAddress types.String `tfsdk:"ip_address"`
}

func NewDNSForwardDomainPolicyResource() resource.Resource {
	return &dnsForwardDomainPolicyResource{}
}

func (r *dnsForwardDomainPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_forward_domain_policy"
}

func (r *dnsForwardDomainPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":         dnsPolicyIDAttribute(),
			"site_id":    dnsPolicySiteIDAttribute(),
			"enabled":    dnsPolicyEnabledAttribute(),
			"domain":     dnsPolicyDomainAttribute(true),
			"origin":     dnsPolicyOriginAttribute(),
			"ip_address": schema.StringAttribute{Required: true},
		},
	}
}

func (r *dnsForwardDomainPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}

func (r *dnsForwardDomainPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	createTypedDNSPolicyResource(ctx, req, resp, r.client, r.siteID, buildDNSForwardDomainPolicyPayload, dnsForwardDomainPolicyStateFromAPI)
}

func (r *dnsForwardDomainPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readTypedDNSPolicyResource(ctx, req, resp, r.client, r.siteID, dnsForwardDomainPolicyStateFromAPI)
}

func (r *dnsForwardDomainPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateTypedDNSPolicyResource(ctx, req, resp, r.client, r.siteID, buildDNSForwardDomainPolicyPayload, dnsForwardDomainPolicyStateFromAPI)
}

func (r *dnsForwardDomainPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteTypedDNSPolicyResource[dnsForwardDomainPolicyResourceModel](ctx, req, resp, r.client, r.siteID)
}

func (r *dnsForwardDomainPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importTypedDNSPolicyResource(ctx, req, resp)
}

func (m dnsForwardDomainPolicyResourceModel) dnsPolicySiteID() types.String { return m.SiteID }
func (m dnsForwardDomainPolicyResourceModel) dnsPolicyID() types.String     { return m.ID }

func buildDNSForwardDomainPolicyPayload(plan dnsForwardDomainPolicyResourceModel) (networkapi.CreateOrUpdateDNSPolicy, error) {
	dto := networkapi.IntegrationDnsForwardDomainPolicyCreateUpdateDto{
		Domain:    terraformStringPointer(plan.Domain),
		Enabled:   plan.Enabled.ValueBool(),
		IpAddress: terraformStringPointer(plan.IPAddress),
	}

	var payload networkapi.CreateOrUpdateDNSPolicy
	if err := payload.FromIntegrationDnsForwardDomainPolicyCreateUpdateDto(dto); err != nil {
		return networkapi.CreateOrUpdateDNSPolicy{}, fmt.Errorf("encode forward domain policy payload: %w", err)
	}

	return payload, nil
}

func dnsForwardDomainPolicyStateFromAPI(siteID string, policy networkapi.DNSPolicy) (dnsForwardDomainPolicyResourceModel, error) {
	if err := expectDNSPolicyType(policy, "FORWARD_DOMAIN"); err != nil {
		return dnsForwardDomainPolicyResourceModel{}, err
	}

	record, err := policy.AsIntegrationDnsForwardDomainPolicyDto()
	if err != nil {
		return dnsForwardDomainPolicyResourceModel{}, fmt.Errorf("decode forward domain policy: %w", err)
	}

	return dnsForwardDomainPolicyResourceModel{
		ID:        types.StringValue(record.Id.String()),
		SiteID:    types.StringValue(siteID),
		Enabled:   types.BoolValue(record.Enabled),
		Domain:    stringPointerValueOrNull(record.Domain),
		Origin:    stringValueOrNull(record.Metadata.Origin),
		IPAddress: stringPointerValueOrNull(record.IpAddress),
	}, nil
}
