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

type dnsSRVRecordResource struct {
	dnsPolicyResourceBase
}

type dnsSRVRecordResourceModel struct {
	ID           types.String `tfsdk:"id"`
	SiteID       types.String `tfsdk:"site_id"`
	Enabled      types.Bool   `tfsdk:"enabled"`
	Domain       types.String `tfsdk:"domain"`
	Origin       types.String `tfsdk:"origin"`
	Service      types.String `tfsdk:"service"`
	Protocol     types.String `tfsdk:"protocol"`
	ServerDomain types.String `tfsdk:"server_domain"`
	Port         types.Int64  `tfsdk:"port"`
	Priority     types.Int64  `tfsdk:"priority"`
	Weight       types.Int64  `tfsdk:"weight"`
}

func NewDNSSRVRecordResource() resource.Resource {
	return &dnsSRVRecordResource{}
}

func (r *dnsSRVRecordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_srv_record"
}

func (r *dnsSRVRecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":            dnsPolicyIDAttribute(),
			"site_id":       dnsPolicySiteIDAttribute(),
			"enabled":       dnsPolicyEnabledAttribute(),
			"domain":        dnsPolicyDomainAttribute(true),
			"origin":        dnsPolicyOriginAttribute(),
			"service":       schema.StringAttribute{Required: true},
			"protocol":      schema.StringAttribute{Required: true},
			"server_domain": schema.StringAttribute{Required: true},
			"port":          schema.Int64Attribute{Required: true},
			"priority":      schema.Int64Attribute{Optional: true},
			"weight":        schema.Int64Attribute{Optional: true},
		},
	}
}

func (r *dnsSRVRecordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}

func (r *dnsSRVRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	createTypedDNSPolicyResource(ctx, req, resp, r.client, r.siteID, buildDNSSRVRecordPayload, dnsSRVRecordStateFromAPI)
}

func (r *dnsSRVRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readTypedDNSPolicyResource(ctx, req, resp, r.client, r.siteID, dnsSRVRecordStateFromAPI)
}

func (r *dnsSRVRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateTypedDNSPolicyResource(ctx, req, resp, r.client, r.siteID, buildDNSSRVRecordPayload, dnsSRVRecordStateFromAPI)
}

func (r *dnsSRVRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteTypedDNSPolicyResource[dnsSRVRecordResourceModel](ctx, req, resp, r.client, r.siteID)
}

func (r *dnsSRVRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importTypedDNSPolicyResource(ctx, req, resp)
}

func (m dnsSRVRecordResourceModel) dnsPolicySiteID() types.String { return m.SiteID }
func (m dnsSRVRecordResourceModel) dnsPolicyID() types.String     { return m.ID }

func buildDNSSRVRecordPayload(plan dnsSRVRecordResourceModel) (networkapi.CreateOrUpdateDNSPolicy, error) {
	port, portDiags := terraformInt32Pointer(plan.Port, path.Root("port"))
	if portDiags.HasError() {
		return networkapi.CreateOrUpdateDNSPolicy{}, diagnosticsError(portDiags)
	}
	priority, priorityDiags := terraformInt32Pointer(plan.Priority, path.Root("priority"))
	if priorityDiags.HasError() {
		return networkapi.CreateOrUpdateDNSPolicy{}, diagnosticsError(priorityDiags)
	}
	weight, weightDiags := terraformInt32Pointer(plan.Weight, path.Root("weight"))
	if weightDiags.HasError() {
		return networkapi.CreateOrUpdateDNSPolicy{}, diagnosticsError(weightDiags)
	}

	dto := networkapi.IntegrationDnsSrvRecordCreateUpdateDto{
		Domain:       terraformStringPointer(plan.Domain),
		Enabled:      plan.Enabled.ValueBool(),
		Port:         port,
		Priority:     priority,
		Protocol:     terraformStringPointer(plan.Protocol),
		ServerDomain: terraformStringPointer(plan.ServerDomain),
		Service:      terraformStringPointer(plan.Service),
		Weight:       weight,
	}

	var payload networkapi.CreateOrUpdateDNSPolicy
	if err := payload.FromIntegrationDnsSrvRecordCreateUpdateDto(dto); err != nil {
		return networkapi.CreateOrUpdateDNSPolicy{}, fmt.Errorf("encode SRV record payload: %w", err)
	}

	return payload, nil
}

func dnsSRVRecordStateFromAPI(siteID string, policy networkapi.DNSPolicy) (dnsSRVRecordResourceModel, error) {
	if err := expectDNSPolicyType(policy, "SRV_RECORD"); err != nil {
		return dnsSRVRecordResourceModel{}, err
	}

	record, err := policy.AsIntegrationDnsSrvRecordDto()
	if err != nil {
		return dnsSRVRecordResourceModel{}, fmt.Errorf("decode SRV record policy: %w", err)
	}

	return dnsSRVRecordResourceModel{
		ID:           types.StringValue(record.Id.String()),
		SiteID:       types.StringValue(siteID),
		Enabled:      types.BoolValue(record.Enabled),
		Domain:       stringPointerValueOrNull(record.Domain),
		Origin:       stringValueOrNull(record.Metadata.Origin),
		Service:      stringPointerValueOrNull(record.Service),
		Protocol:     stringPointerValueOrNull(record.Protocol),
		ServerDomain: stringPointerValueOrNull(record.ServerDomain),
		Port:         int32PointerValueOrNull(record.Port),
		Priority:     int32PointerValueOrNull(record.Priority),
		Weight:       int32PointerValueOrNull(record.Weight),
	}, nil
}
