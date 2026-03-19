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

type dnsCNAMERecordResource struct {
	dnsPolicyResourceBase
}

type dnsCNAMERecordResourceModel struct {
	ID           types.String `tfsdk:"id"`
	SiteID       types.String `tfsdk:"site_id"`
	Enabled      types.Bool   `tfsdk:"enabled"`
	Domain       types.String `tfsdk:"domain"`
	Origin       types.String `tfsdk:"origin"`
	TargetDomain types.String `tfsdk:"target_domain"`
	TTLSeconds   types.Int64  `tfsdk:"ttl_seconds"`
}

func NewDNSCNAMERecordResource() resource.Resource {
	return &dnsCNAMERecordResource{}
}

func (r *dnsCNAMERecordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_cname_record"
}

func (r *dnsCNAMERecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":            dnsPolicyIDAttribute(),
			"site_id":       dnsPolicySiteIDAttribute(),
			"enabled":       dnsPolicyEnabledAttribute(),
			"domain":        dnsPolicyDomainAttribute(true),
			"origin":        dnsPolicyOriginAttribute(),
			"target_domain": schema.StringAttribute{Required: true},
			"ttl_seconds":   schema.Int64Attribute{Optional: true},
		},
	}
}

func (r *dnsCNAMERecordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}

func (r *dnsCNAMERecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	createTypedDNSPolicyResource(ctx, req, resp, r.client, r.siteID, buildDNSCNAMERecordPayload, dnsCNAMERecordStateFromAPI)
}

func (r *dnsCNAMERecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readTypedDNSPolicyResource(ctx, req, resp, r.client, r.siteID, dnsCNAMERecordStateFromAPI)
}

func (r *dnsCNAMERecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateTypedDNSPolicyResource(ctx, req, resp, r.client, r.siteID, buildDNSCNAMERecordPayload, dnsCNAMERecordStateFromAPI)
}

func (r *dnsCNAMERecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteTypedDNSPolicyResource[dnsCNAMERecordResourceModel](ctx, req, resp, r.client, r.siteID)
}

func (r *dnsCNAMERecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importTypedDNSPolicyResource(ctx, req, resp)
}

func (m dnsCNAMERecordResourceModel) dnsPolicySiteID() types.String { return m.SiteID }
func (m dnsCNAMERecordResourceModel) dnsPolicyID() types.String     { return m.ID }

func buildDNSCNAMERecordPayload(plan dnsCNAMERecordResourceModel) (networkapi.CreateOrUpdateDNSPolicy, error) {
	ttlSeconds, diags := terraformInt32Pointer(plan.TTLSeconds, path.Root("ttl_seconds"))
	if diags.HasError() {
		return networkapi.CreateOrUpdateDNSPolicy{}, fmt.Errorf("%s", diags[0].Summary())
	}

	dto := networkapi.IntegrationDnsCnameRecordCreateUpdateDto{
		Domain:       terraformStringPointer(plan.Domain),
		Enabled:      plan.Enabled.ValueBool(),
		TargetDomain: terraformStringPointer(plan.TargetDomain),
		TtlSeconds:   ttlSeconds,
	}

	var payload networkapi.CreateOrUpdateDNSPolicy
	if err := payload.FromIntegrationDnsCnameRecordCreateUpdateDto(dto); err != nil {
		return networkapi.CreateOrUpdateDNSPolicy{}, fmt.Errorf("encode CNAME record payload: %w", err)
	}

	return payload, nil
}

func dnsCNAMERecordStateFromAPI(siteID string, policy networkapi.DNSPolicy) (dnsCNAMERecordResourceModel, error) {
	if err := expectDNSPolicyType(policy, "CNAME_RECORD"); err != nil {
		return dnsCNAMERecordResourceModel{}, err
	}

	record, err := policy.AsIntegrationDnsCnameRecordDto()
	if err != nil {
		return dnsCNAMERecordResourceModel{}, fmt.Errorf("decode CNAME record policy: %w", err)
	}

	return dnsCNAMERecordResourceModel{
		ID:           types.StringValue(record.Id.String()),
		SiteID:       types.StringValue(siteID),
		Enabled:      types.BoolValue(record.Enabled),
		Domain:       stringPointerValueOrNull(record.Domain),
		Origin:       stringValueOrNull(record.Metadata.Origin),
		TargetDomain: stringPointerValueOrNull(record.TargetDomain),
		TTLSeconds:   int32PointerValueOrNull(record.TtlSeconds),
	}, nil
}
