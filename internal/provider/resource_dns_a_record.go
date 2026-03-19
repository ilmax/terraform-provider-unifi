package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	networkapi "github.com/ilmax/unifi-client-go/pkg/network"
)

type dnsARecordResource struct {
	dnsPolicyResourceBase
}

type dnsARecordResourceModel struct {
	ID          types.String `tfsdk:"id"`
	SiteID      types.String `tfsdk:"site_id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Domain      types.String `tfsdk:"domain"`
	Origin      types.String `tfsdk:"origin"`
	IPv4Address types.String `tfsdk:"ipv4_address"`
	TTLSeconds  types.Int64  `tfsdk:"ttl_seconds"`
}

func NewDNSARecordResource() resource.Resource {
	return &dnsARecordResource{}
}

func (r *dnsARecordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_a_record"
}

func (r *dnsARecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":           dnsPolicyIDAttribute(),
			"site_id":      dnsPolicySiteIDAttribute(),
			"enabled":      dnsPolicyEnabledAttribute(),
			"domain":       dnsPolicyDomainAttribute(true),
			"origin":       dnsPolicyOriginAttribute(),
			"ipv4_address": schema.StringAttribute{Required: true},
			"ttl_seconds":  schema.Int64Attribute{Optional: true},
		},
	}
}

func (r *dnsARecordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}

func (r *dnsARecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	createTypedDNSPolicyResource(ctx, req, resp, r.client, r.siteID, buildDNSARecordPayload, dnsARecordStateFromAPI)
}

func (r *dnsARecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readTypedDNSPolicyResource(ctx, req, resp, r.client, r.siteID, dnsARecordStateFromAPI)
}

func (r *dnsARecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateTypedDNSPolicyResource(ctx, req, resp, r.client, r.siteID, buildDNSARecordPayload, dnsARecordStateFromAPI)
}

func (r *dnsARecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteTypedDNSPolicyResource[dnsARecordResourceModel](ctx, req, resp, r.client, r.siteID)
}

func (r *dnsARecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importTypedDNSPolicyResource(ctx, req, resp)
}

func (m dnsARecordResourceModel) dnsPolicySiteID() types.String { return m.SiteID }
func (m dnsARecordResourceModel) dnsPolicyID() types.String     { return m.ID }

func buildDNSARecordPayload(plan dnsARecordResourceModel) (networkapi.CreateOrUpdateDNSPolicy, error) {
	ttlSeconds, diags := terraformInt32Pointer(plan.TTLSeconds, path.Root("ttl_seconds"))
	if diags.HasError() {
		return networkapi.CreateOrUpdateDNSPolicy{}, fmt.Errorf("%s", diags[0].Summary())
	}

	dto := networkapi.IntegrationDnsARecordCreateUpdateDto{
		Domain:      terraformStringPointer(plan.Domain),
		Enabled:     plan.Enabled.ValueBool(),
		Ipv4Address: terraformStringPointer(plan.IPv4Address),
		TtlSeconds:  ttlSeconds,
	}

	var payload networkapi.CreateOrUpdateDNSPolicy
	if err := payload.FromIntegrationDnsARecordCreateUpdateDto(dto); err != nil {
		return networkapi.CreateOrUpdateDNSPolicy{}, fmt.Errorf("encode A record payload: %w", err)
	}

	return payload, nil
}

func dnsARecordStateFromAPI(siteID string, policy networkapi.DNSPolicy) (dnsARecordResourceModel, error) {
	if err := expectDNSPolicyType(policy, "A_RECORD"); err != nil {
		return dnsARecordResourceModel{}, err
	}

	record, err := policy.AsIntegrationDnsARecordDto()
	if err != nil {
		return dnsARecordResourceModel{}, fmt.Errorf("decode A record policy: %w", err)
	}

	return dnsARecordResourceModel{
		ID:          types.StringValue(record.Id.String()),
		SiteID:      types.StringValue(siteID),
		Enabled:     types.BoolValue(record.Enabled),
		Domain:      stringPointerValueOrNull(record.Domain),
		Origin:      stringValueOrNull(record.Metadata.Origin),
		IPv4Address: stringPointerValueOrNull(record.Ipv4Address),
		TTLSeconds:  int32PointerValueOrNull(record.TtlSeconds),
	}, nil
}
