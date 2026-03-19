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

type dnsMXRecordResource struct {
	dnsPolicyResourceBase
}

type dnsMXRecordResourceModel struct {
	ID               types.String `tfsdk:"id"`
	SiteID           types.String `tfsdk:"site_id"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	Domain           types.String `tfsdk:"domain"`
	Origin           types.String `tfsdk:"origin"`
	MailServerDomain types.String `tfsdk:"mail_server_domain"`
	Priority         types.Int64  `tfsdk:"priority"`
}

func NewDNSMXRecordResource() resource.Resource {
	return &dnsMXRecordResource{}
}

func (r *dnsMXRecordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_mx_record"
}

func (r *dnsMXRecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":                 dnsPolicyIDAttribute(),
			"site_id":            dnsPolicySiteIDAttribute(),
			"enabled":            dnsPolicyEnabledAttribute(),
			"domain":             dnsPolicyDomainAttribute(true),
			"origin":             dnsPolicyOriginAttribute(),
			"mail_server_domain": schema.StringAttribute{Required: true},
			"priority":           schema.Int64Attribute{Optional: true},
		},
	}
}

func (r *dnsMXRecordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}

func (r *dnsMXRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	createTypedDNSPolicyResource(ctx, req, resp, r.client, r.siteID, buildDNSMXRecordPayload, dnsMXRecordStateFromAPI)
}

func (r *dnsMXRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readTypedDNSPolicyResource(ctx, req, resp, r.client, r.siteID, dnsMXRecordStateFromAPI)
}

func (r *dnsMXRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateTypedDNSPolicyResource(ctx, req, resp, r.client, r.siteID, buildDNSMXRecordPayload, dnsMXRecordStateFromAPI)
}

func (r *dnsMXRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteTypedDNSPolicyResource[dnsMXRecordResourceModel](ctx, req, resp, r.client, r.siteID)
}

func (r *dnsMXRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importTypedDNSPolicyResource(ctx, req, resp)
}

func (m dnsMXRecordResourceModel) dnsPolicySiteID() types.String { return m.SiteID }
func (m dnsMXRecordResourceModel) dnsPolicyID() types.String     { return m.ID }

func buildDNSMXRecordPayload(plan dnsMXRecordResourceModel) (networkapi.CreateOrUpdateDNSPolicy, error) {
	priority, diags := terraformInt32Pointer(plan.Priority, path.Root("priority"))
	if diags.HasError() {
		return networkapi.CreateOrUpdateDNSPolicy{}, diagnosticsError(diags)
	}

	dto := networkapi.IntegrationDnsMxRecordCreateUpdateDto{
		Domain:           terraformStringPointer(plan.Domain),
		Enabled:          plan.Enabled.ValueBool(),
		MailServerDomain: terraformStringPointer(plan.MailServerDomain),
		Priority:         priority,
	}

	var payload networkapi.CreateOrUpdateDNSPolicy
	if err := payload.FromIntegrationDnsMxRecordCreateUpdateDto(dto); err != nil {
		return networkapi.CreateOrUpdateDNSPolicy{}, fmt.Errorf("encode MX record payload: %w", err)
	}

	return payload, nil
}

func dnsMXRecordStateFromAPI(siteID string, policy networkapi.DNSPolicy) (dnsMXRecordResourceModel, error) {
	if err := expectDNSPolicyType(policy, "MX_RECORD"); err != nil {
		return dnsMXRecordResourceModel{}, err
	}

	record, err := policy.AsIntegrationDnsMxRecordDto()
	if err != nil {
		return dnsMXRecordResourceModel{}, fmt.Errorf("decode MX record policy: %w", err)
	}

	return dnsMXRecordResourceModel{
		ID:               types.StringValue(record.Id.String()),
		SiteID:           types.StringValue(siteID),
		Enabled:          types.BoolValue(record.Enabled),
		Domain:           stringPointerValueOrNull(record.Domain),
		Origin:           stringValueOrNull(record.Metadata.Origin),
		MailServerDomain: stringPointerValueOrNull(record.MailServerDomain),
		Priority:         int32PointerValueOrNull(record.Priority),
	}, nil
}
