package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
	"github.com/ilmax/unifi-client-go/pkg/acl_rules"
	"github.com/ilmax/unifi-client-go/pkg/errors"
)

type firewallResource struct {
	client *unifi.Client
	siteID string
}

type firewallResourceModel struct {
	ID                    types.String `tfsdk:"id"`
	SiteID                types.String `tfsdk:"site_id"`
	Type                  types.String `tfsdk:"type"`
	Name                  types.String `tfsdk:"name"`
	Description           types.String `tfsdk:"description"`
	Action                types.String `tfsdk:"action"`
	Enabled               types.Bool   `tfsdk:"enabled"`
	Index                 types.Int64  `tfsdk:"index"`
	SourceFilterJSON      types.String `tfsdk:"source_filter_json"`
	DestinationFilterJSON types.String `tfsdk:"destination_filter_json"`
	ProtocolFilter        types.List   `tfsdk:"protocol_filter"`
}

func NewFirewallResource() resource.Resource {
	return &firewallResource{}
}

func (r *firewallResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall"
}

func (r *firewallResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"site_id": schema.StringAttribute{
				Optional: true,
			},
			"type": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"action": schema.StringAttribute{
				Required: true,
			},
			"enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"index": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"source_filter_json": schema.StringAttribute{
				Optional: true,
			},
			"destination_filter_json": schema.StringAttribute{
				Optional: true,
			},
			"protocol_filter": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *firewallResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*providerData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", "Expected provider data to be of type *providerData.")
		return
	}
	r.client = data.client
	r.siteID = data.siteID
}

func (r *firewallResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan firewallResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(plan.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	reqBody, diags := buildFirewallRequest(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result acl_rules.CreateACLRuleResponse
	path := fmt.Sprintf("/v1/sites/%s/acl-rules", siteID)
	if err := r.client.Post(ctx, path, reqBody, &result); err != nil {
		resp.Diagnostics.AddError("Unable to create ACL rule", err.Error())
		return
	}

	state := firewallStateFromCreate(siteID, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *firewallResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state firewallResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(state.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	if state.ID.IsNull() || state.ID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("id"), "Missing ACL rule ID", "The ACL rule ID is required to read the resource.")
		return
	}

	var result acl_rules.GetACLRuleResponse
	path := fmt.Sprintf("/v1/sites/%s/acl-rules/%s", siteID, state.ID.ValueString())
	if err := r.client.Get(ctx, path, &result); err != nil {
		if errors.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read ACL rule", err.Error())
		return
	}

	state = firewallStateFromGet(siteID, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *firewallResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan firewallResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(plan.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	if plan.ID.IsNull() || plan.ID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("id"), "Missing ACL rule ID", "The ACL rule ID is required to update the resource.")
		return
	}

	reqBody, diags := buildFirewallUpdateRequest(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result acl_rules.UpdateACLRuleResponse
	path := fmt.Sprintf("/v1/sites/%s/acl-rules/%s", siteID, plan.ID.ValueString())
	if err := r.client.Put(ctx, path, reqBody, &result); err != nil {
		resp.Diagnostics.AddError("Unable to update ACL rule", err.Error())
		return
	}

	state := firewallStateFromUpdate(siteID, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *firewallResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state firewallResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(state.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	if state.ID.IsNull() || state.ID.IsUnknown() {
		return
	}

	path := fmt.Sprintf("/v1/sites/%s/acl-rules/%s", siteID, state.ID.ValueString())
	if err := r.client.Delete(ctx, path, nil); err != nil && !errors.IsNotFoundError(err) {
		resp.Diagnostics.AddError("Unable to delete ACL rule", err.Error())
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *firewallResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	siteID, aclRuleID, err := splitImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import identifier", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site_id"), siteID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), aclRuleID)...)
}

func buildFirewallRequest(plan firewallResourceModel) (*acl_rules.CreateACLRuleRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	request := &acl_rules.CreateACLRuleRequest{
		Type:        plan.Type.ValueString(),
		Enabled:     plan.Enabled.ValueBool(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Action:      plan.Action.ValueString(),
		Index:       plan.Index.ValueInt64(),
	}

	sourceFilter, sourceDiags := decodeACLSourceFilterCreate(plan.SourceFilterJSON)
	diags.Append(sourceDiags...)
	request.SourceFilter = sourceFilter

	destinationFilter, destDiags := decodeJSONRaw(plan.DestinationFilterJSON, path.Root("destination_filter_json"))
	diags.Append(destDiags...)
	request.DestinationFilter = destinationFilter

	protocolFilter, protoDiags := decodeProtocolFilter(plan.ProtocolFilter)
	diags.Append(protoDiags...)
	request.ProtocolFilter = protocolFilter

	return request, diags
}

func buildFirewallUpdateRequest(plan firewallResourceModel) (*acl_rules.UpdateACLRuleRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	request := &acl_rules.UpdateACLRuleRequest{
		Type:        plan.Type.ValueString(),
		Enabled:     plan.Enabled.ValueBool(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Action:      plan.Action.ValueString(),
		Index:       plan.Index.ValueInt64(),
	}

	sourceFilter, sourceDiags := decodeACLSourceFilterUpdate(plan.SourceFilterJSON)
	diags.Append(sourceDiags...)
	request.SourceFilter = sourceFilter

	destinationFilter, destDiags := decodeJSONRaw(plan.DestinationFilterJSON, path.Root("destination_filter_json"))
	diags.Append(destDiags...)
	request.DestinationFilter = destinationFilter

	protocolFilter, protoDiags := decodeProtocolFilter(plan.ProtocolFilter)
	diags.Append(protoDiags...)
	request.ProtocolFilter = protocolFilter

	return request, diags
}

func decodeACLSourceFilterCreate(raw types.String) (*acl_rules.CreateACLRuleSourceFilter, diag.Diagnostics) {
	var diags diag.Diagnostics
	if raw.IsNull() || raw.IsUnknown() {
		return nil, diags
	}

	if err := requireValidJSON(raw.ValueString()); err != nil {
		diags.AddAttributeError(path.Root("source_filter_json"), "Invalid source_filter_json", err.Error())
		return nil, diags
	}

	var result acl_rules.CreateACLRuleSourceFilter
	if err := json.Unmarshal([]byte(raw.ValueString()), &result); err != nil {
		diags.AddAttributeError(path.Root("source_filter_json"), "Invalid source_filter_json", err.Error())
		return nil, diags
	}

	return &result, diags
}

func decodeACLSourceFilterUpdate(raw types.String) (*acl_rules.UpdateACLRuleSourceFilter, diag.Diagnostics) {
	var diags diag.Diagnostics
	if raw.IsNull() || raw.IsUnknown() {
		return nil, diags
	}

	if err := requireValidJSON(raw.ValueString()); err != nil {
		diags.AddAttributeError(path.Root("source_filter_json"), "Invalid source_filter_json", err.Error())
		return nil, diags
	}

	var result acl_rules.UpdateACLRuleSourceFilter
	if err := json.Unmarshal([]byte(raw.ValueString()), &result); err != nil {
		diags.AddAttributeError(path.Root("source_filter_json"), "Invalid source_filter_json", err.Error())
		return nil, diags
	}

	return &result, diags
}

func decodeJSONRaw(raw types.String, attrPath path.Path) (json.RawMessage, diag.Diagnostics) {
	var diags diag.Diagnostics
	if raw.IsNull() || raw.IsUnknown() {
		return nil, diags
	}

	if err := requireValidJSON(raw.ValueString()); err != nil {
		diags.AddAttributeError(attrPath, "Invalid JSON", err.Error())
		return nil, diags
	}

	return json.RawMessage(raw.ValueString()), diags
}

func decodeProtocolFilter(protocols types.List) ([]json.RawMessage, diag.Diagnostics) {
	var diags diag.Diagnostics
	if protocols.IsNull() || protocols.IsUnknown() {
		return nil, diags
	}

	var values []string
	if diag := protocols.ElementsAs(context.Background(), &values, false); diag.HasError() {
		diags.Append(diag...)
		return nil, diags
	}

	result := make([]json.RawMessage, 0, len(values))
	for _, value := range values {
		encoded, err := json.Marshal(value)
		if err != nil {
			diags.AddAttributeError(path.Root("protocol_filter"), "Invalid protocol_filter", err.Error())
			return nil, diags
		}
		result = append(result, json.RawMessage(encoded))
	}

	return result, diags
}

func firewallStateFromCreate(siteID string, result acl_rules.CreateACLRuleResponse) firewallResourceModel {
	state := firewallResourceModel{
		ID:          types.StringValue(result.Id),
		SiteID:      types.StringValue(siteID),
		Type:        types.StringValue(result.Type),
		Name:        types.StringValue(result.Name),
		Description: types.StringValue(result.Description),
		Action:      types.StringValue(result.Action),
		Enabled:     types.BoolValue(result.Enabled),
		Index:       types.Int64Value(result.Index),
	}
	state.SourceFilterJSON = rawMessageToString(result.SourceFilter)
	state.DestinationFilterJSON = rawMessageToString(result.DestinationFilter)
	state.ProtocolFilter = rawMessagesToList(result.ProtocolFilter)
	return state
}

func firewallStateFromUpdate(siteID string, result acl_rules.UpdateACLRuleResponse) firewallResourceModel {
	state := firewallResourceModel{
		ID:          types.StringValue(result.Id),
		SiteID:      types.StringValue(siteID),
		Type:        types.StringValue(result.Type),
		Name:        types.StringValue(result.Name),
		Description: types.StringValue(result.Description),
		Action:      types.StringValue(result.Action),
		Enabled:     types.BoolValue(result.Enabled),
		Index:       types.Int64Value(result.Index),
	}
	state.SourceFilterJSON = rawMessageToString(result.SourceFilter)
	state.DestinationFilterJSON = rawMessageToString(result.DestinationFilter)
	state.ProtocolFilter = rawMessagesToList(result.ProtocolFilter)
	return state
}

func firewallStateFromGet(siteID string, result acl_rules.GetACLRuleResponse) firewallResourceModel {
	state := firewallResourceModel{
		ID:          types.StringValue(result.Id),
		SiteID:      types.StringValue(siteID),
		Type:        types.StringValue(result.Type),
		Name:        types.StringValue(result.Name),
		Description: types.StringValue(result.Description),
		Action:      types.StringValue(result.Action),
		Enabled:     types.BoolValue(result.Enabled),
		Index:       types.Int64Value(result.Index),
	}
	state.SourceFilterJSON = rawMessageToString(result.SourceFilter)
	state.DestinationFilterJSON = rawMessageToString(result.DestinationFilter)
	state.ProtocolFilter = rawMessagesToList(result.ProtocolFilter)
	return state
}

func rawMessageToString(raw json.RawMessage) types.String {
	if len(raw) == 0 {
		return types.StringNull()
	}
	return types.StringValue(string(raw))
}

func rawMessagesToList(raw []json.RawMessage) types.List {
	if len(raw) == 0 {
		return types.ListNull(types.StringType)
	}
	values := make([]string, 0, len(raw))
	for _, item := range raw {
		values = append(values, string(item))
	}
	list, _ := types.ListValueFrom(context.Background(), types.StringType, values)
	return list
}
