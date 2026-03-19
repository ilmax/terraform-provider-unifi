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

type dnsAAAARecordResource struct {
	dnsPolicyResourceBase
}

type dnsAAAARecordResourceModel struct {
	ID          types.String `tfsdk:"id"`
	SiteID      types.String `tfsdk:"site_id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Domain      types.String `tfsdk:"domain"`
	Origin      types.String `tfsdk:"origin"`
	IPv6Address types.String `tfsdk:"ipv6_address"`
	TTLSeconds  types.Int64  `tfsdk:"ttl_seconds"`
}

func NewDNSAAAARecordResource() resource.Resource {
	return &dnsAAAARecordResource{}
}

func (r *dnsAAAARecordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_aaaa_record"
}

func (r *dnsAAAARecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":           dnsPolicyIDAttribute(),
			"site_id":      dnsPolicySiteIDAttribute(),
			"enabled":      dnsPolicyEnabledAttribute(),
			"domain":       dnsPolicyDomainAttribute(true),
			"origin":       dnsPolicyOriginAttribute(),
			"ipv6_address": schema.StringAttribute{Required: true},
			"ttl_seconds":  schema.Int64Attribute{Optional: true},
		},
	}
}

func (r *dnsAAAARecordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}

func (r *dnsAAAARecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	createTypedDNSPolicyResource(ctx, req, resp, r.client, r.siteID, buildDNSAAAARecordPayload, dnsAAAARecordStateFromAPI)
}

func (r *dnsAAAARecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readTypedDNSPolicyResource(ctx, req, resp, r.client, r.siteID, dnsAAAARecordStateFromAPI)
}

func (r *dnsAAAARecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateTypedDNSPolicyResource(ctx, req, resp, r.client, r.siteID, buildDNSAAAARecordPayload, dnsAAAARecordStateFromAPI)
}

func (r *dnsAAAARecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteTypedDNSPolicyResource[dnsAAAARecordResourceModel](ctx, req, resp, r.client, r.siteID)
}

func (r *dnsAAAARecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importTypedDNSPolicyResource(ctx, req, resp)
}

func (m dnsAAAARecordResourceModel) dnsPolicySiteID() types.String { return m.SiteID }
func (m dnsAAAARecordResourceModel) dnsPolicyID() types.String     { return m.ID }

func buildDNSAAAARecordPayload(plan dnsAAAARecordResourceModel) (networkapi.CreateOrUpdateDNSPolicy, error) {
	ttlSeconds, diags := terraformInt32Pointer(plan.TTLSeconds, path.Root("ttl_seconds"))
	if diags.HasError() {
		return networkapi.CreateOrUpdateDNSPolicy{}, fmt.Errorf("%s", diags[0].Summary())
	}

	dto := networkapi.IntegrationDnsAaaaRecordCreateUpdateDto{
		Domain:      terraformStringPointer(plan.Domain),
		Enabled:     plan.Enabled.ValueBool(),
		Ipv6Address: terraformStringPointer(plan.IPv6Address),
		TtlSeconds:  ttlSeconds,
	}

	var payload networkapi.CreateOrUpdateDNSPolicy
	if err := payload.FromIntegrationDnsAaaaRecordCreateUpdateDto(dto); err != nil {
		return networkapi.CreateOrUpdateDNSPolicy{}, fmt.Errorf("encode AAAA record payload: %w", err)
	}

	return payload, nil
}

func dnsAAAARecordStateFromAPI(siteID string, policy networkapi.DNSPolicy) (dnsAAAARecordResourceModel, error) {
	if err := expectDNSPolicyType(policy, "AAAA_RECORD"); err != nil {
		return dnsAAAARecordResourceModel{}, err
	}

	record, err := policy.AsIntegrationDnsAaaaRecordDto()
	if err != nil {
		return dnsAAAARecordResourceModel{}, fmt.Errorf("decode AAAA record policy: %w", err)
	}

	return dnsAAAARecordResourceModel{
		ID:          types.StringValue(record.Id.String()),
		SiteID:      types.StringValue(siteID),
		Enabled:     types.BoolValue(record.Enabled),
		Domain:      stringPointerValueOrNull(record.Domain),
		Origin:      stringValueOrNull(record.Metadata.Origin),
		IPv6Address: stringPointerValueOrNull(record.Ipv6Address),
		TTLSeconds:  int32PointerValueOrNull(record.TtlSeconds),
	}, nil
}
