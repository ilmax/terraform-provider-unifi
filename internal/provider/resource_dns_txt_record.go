package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	networkapi "github.com/ilmax/unifi-client-go/pkg/network"
)

type dnsTXTRecordResource struct {
	dnsPolicyResourceBase
}

type dnsTXTRecordResourceModel struct {
	ID      types.String `tfsdk:"id"`
	SiteID  types.String `tfsdk:"site_id"`
	Enabled types.Bool   `tfsdk:"enabled"`
	Domain  types.String `tfsdk:"domain"`
	Origin  types.String `tfsdk:"origin"`
	Text    types.String `tfsdk:"text"`
}

func NewDNSTXTRecordResource() resource.Resource {
	return &dnsTXTRecordResource{}
}

func (r *dnsTXTRecordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_txt_record"
}

func (r *dnsTXTRecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":      dnsPolicyIDAttribute(),
			"site_id": dnsPolicySiteIDAttribute(),
			"enabled": dnsPolicyEnabledAttribute(),
			"domain":  dnsPolicyDomainAttribute(true),
			"origin":  dnsPolicyOriginAttribute(),
			"text":    schema.StringAttribute{Required: true},
		},
	}
}

func (r *dnsTXTRecordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}

func (r *dnsTXTRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	createTypedDNSPolicyResource(ctx, req, resp, r.client, r.siteID, buildDNSTXTRecordPayload, dnsTXTRecordStateFromAPI)
}

func (r *dnsTXTRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readTypedDNSPolicyResource(ctx, req, resp, r.client, r.siteID, dnsTXTRecordStateFromAPI)
}

func (r *dnsTXTRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateTypedDNSPolicyResource(ctx, req, resp, r.client, r.siteID, buildDNSTXTRecordPayload, dnsTXTRecordStateFromAPI)
}

func (r *dnsTXTRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteTypedDNSPolicyResource[dnsTXTRecordResourceModel](ctx, req, resp, r.client, r.siteID)
}

func (r *dnsTXTRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importTypedDNSPolicyResource(ctx, req, resp)
}

func (m dnsTXTRecordResourceModel) dnsPolicySiteID() types.String { return m.SiteID }
func (m dnsTXTRecordResourceModel) dnsPolicyID() types.String     { return m.ID }

func buildDNSTXTRecordPayload(plan dnsTXTRecordResourceModel) (networkapi.CreateOrUpdateDNSPolicy, error) {
	dto := networkapi.IntegrationDnsTxtRecordCreateUpdateDto{
		Domain:  terraformStringPointer(plan.Domain),
		Enabled: plan.Enabled.ValueBool(),
		Text:    terraformStringPointer(plan.Text),
	}

	var payload networkapi.CreateOrUpdateDNSPolicy
	if err := payload.FromIntegrationDnsTxtRecordCreateUpdateDto(dto); err != nil {
		return networkapi.CreateOrUpdateDNSPolicy{}, fmt.Errorf("encode TXT record payload: %w", err)
	}

	return payload, nil
}

func dnsTXTRecordStateFromAPI(siteID string, policy networkapi.DNSPolicy) (dnsTXTRecordResourceModel, error) {
	if err := expectDNSPolicyType(policy, "TXT_RECORD"); err != nil {
		return dnsTXTRecordResourceModel{}, err
	}

	record, err := policy.AsIntegrationDnsTxtRecordDto()
	if err != nil {
		return dnsTXTRecordResourceModel{}, fmt.Errorf("decode TXT record policy: %w", err)
	}

	return dnsTXTRecordResourceModel{
		ID:      types.StringValue(record.Id.String()),
		SiteID:  types.StringValue(siteID),
		Enabled: types.BoolValue(record.Enabled),
		Domain:  stringPointerValueOrNull(record.Domain),
		Origin:  stringValueOrNull(record.Metadata.Origin),
		Text:    stringPointerValueOrNull(record.Text),
	}, nil
}
