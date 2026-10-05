package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/sdkcompat/acl_rules"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
	"github.com/ilmax/unifi-client-go/pkg/errors"
)

type firewallResource struct {
	client *unifi.Client
	siteID string
}

type firewallResourceModel struct {
	ID                types.String         `tfsdk:"id"`
	SiteID            types.String         `tfsdk:"site_id"`
	Type              types.String         `tfsdk:"type"`
	Name              types.String         `tfsdk:"name"`
	Description       types.String         `tfsdk:"description"`
	Action            types.String         `tfsdk:"action"`
	Enabled           types.Bool           `tfsdk:"enabled"`
	Index             types.Int64          `tfsdk:"index"`
	SourceFilter      *firewallFilterModel `tfsdk:"source_filter"`
	DestinationFilter *firewallFilterModel `tfsdk:"destination_filter"`
	ProtocolFilter    types.Set            `tfsdk:"protocol_filter"`
	NetworkIDFilter   types.String         `tfsdk:"network_id_filter"`
}

// firewallFilterModel is the native shape shared by source_filter and
// destination_filter. The real API's filter is a discriminated union keyed
// on `type` (IP_ADDRESSES_OR_SUBNETS/NETWORKS/PORTS for an IPV4 rule,
// MAC_ADDRESSES for a MAC rule); rather than expose that as an opaque JSON
// string, every variant's fields live here as ordinary optional attributes,
// and buildFirewallFilterPayload validates that only the fields matching
// `type` are set.
type firewallFilterModel struct {
	Type                 types.String `tfsdk:"type"`
	IPAddressesOrSubnets types.List   `tfsdk:"ip_addresses_or_subnets"`
	NetworkIDs           types.List   `tfsdk:"network_ids"`
	PortFilter           types.List   `tfsdk:"port_filter"`
	MacAddresses         types.List   `tfsdk:"mac_addresses"`
	PrefixLength         types.Int64  `tfsdk:"prefix_length"`
}

func NewFirewallRuleResource() resource.Resource {
	return &firewallResource{}
}

func (r *firewallResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_rule"
}

func (r *firewallResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an ACL firewall rule. source_filter/destination_filter model the real API's " +
			"discriminated union of match criteria as native attributes (one shared shape covering every " +
			"variant, validated per the chosen type) rather than a raw JSON string.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"site_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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
			"source_filter": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Traffic source filter. Omit to match all traffic.",
				Attributes:  firewallFilterAttributes(),
			},
			"destination_filter": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Traffic destination filter. Omit to match all traffic.",
				Attributes:  firewallFilterAttributes(),
			},
			"protocol_filter": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Protocols this rule applies to, e.g. TCP, UDP. A set, not a list: the real API " +
					"stores this unordered and doesn't echo back the order it was sent in.",
			},
			"network_id_filter": schema.StringAttribute{
				Optional:    true,
				Description: "Network ID this ACL rule applies to. Required by the API when type is MAC; unused for IPV4.",
			},
		},
	}
}

// firewallFilterAttributes is shared by source_filter and destination_filter:
// one flat set of attributes covering every variant of the real API's
// discriminated union, since which ones are actually used is determined by
// `type` at apply time (validated in buildFirewallFilterPayload), not by the
// schema itself.
func firewallFilterAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"type": schema.StringAttribute{
			Required:    true,
			Description: "IP_ADDRESSES_OR_SUBNETS, NETWORKS, or PORTS for an IPV4 rule; MAC_ADDRESSES for a MAC rule.",
		},
		"ip_addresses_or_subnets": schema.ListAttribute{
			Optional:    true,
			ElementType: types.StringType,
			Description: "IP addresses or CIDR subnets to match. Required when type = IP_ADDRESSES_OR_SUBNETS; must be omitted otherwise.",
		},
		"network_ids": schema.ListAttribute{
			Optional:    true,
			ElementType: types.StringType,
			Description: "Network IDs to match. Required when type = NETWORKS; must be omitted otherwise.",
		},
		"port_filter": schema.ListAttribute{
			Optional:    true,
			ElementType: types.Int64Type,
			Description: "Ports (1-65535) to match. Required when type = PORTS; an optional extra restriction when type = IP_ADDRESSES_OR_SUBNETS or NETWORKS (the rule's protocol_filter must also be set); must be omitted when type = MAC_ADDRESSES.",
		},
		"mac_addresses": schema.ListAttribute{
			Optional:    true,
			ElementType: types.StringType,
			Description: "MAC addresses to match. Required when type = MAC_ADDRESSES; must be omitted otherwise.",
		},
		"prefix_length": schema.Int64Attribute{
			Optional:    true,
			Description: "MAC address prefix length (1-48). Only valid when type = MAC_ADDRESSES; omit for a full-address match.",
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

	payload, diags := buildFirewallPayload(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var raw json.RawMessage
	path := fmt.Sprintf("/v1/sites/%s/acl-rules", siteID)
	if err := r.client.Post(ctx, path, payload, &raw); err != nil {
		resp.Diagnostics.AddError("Unable to create ACL rule", err.Error())
		return
	}

	decoded, err := acl_rules.DecodeCreateACLRuleResponse(raw)
	if err != nil {
		resp.Diagnostics.AddError("Unable to decode ACL rule", err.Error())
		return
	}

	state, stateDiags := firewallStateFromResponse(siteID, decoded)
	resp.Diagnostics.Append(stateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

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

	var raw json.RawMessage
	path := fmt.Sprintf("/v1/sites/%s/acl-rules/%s", siteID, state.ID.ValueString())
	if err := r.client.Get(ctx, path, &raw); err != nil {
		if errors.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read ACL rule", err.Error())
		return
	}

	decoded, err := acl_rules.DecodeGetACLRuleResponse(raw)
	if err != nil {
		resp.Diagnostics.AddError("Unable to decode ACL rule", err.Error())
		return
	}

	state, stateDiags := firewallStateFromResponse(siteID, decoded)
	resp.Diagnostics.Append(stateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

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

	payload, diags := buildFirewallPayload(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var raw json.RawMessage
	path := fmt.Sprintf("/v1/sites/%s/acl-rules/%s", siteID, plan.ID.ValueString())
	if err := r.client.Put(ctx, path, payload, &raw); err != nil {
		resp.Diagnostics.AddError("Unable to update ACL rule", err.Error())
		return
	}

	decoded, err := acl_rules.DecodeUpdateACLRuleResponse(raw)
	if err != nil {
		resp.Diagnostics.AddError("Unable to decode ACL rule", err.Error())
		return
	}

	state, stateDiags := firewallStateFromResponse(siteID, decoded)
	resp.Diagnostics.Append(stateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

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

func buildFirewallPayload(plan firewallResourceModel) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	payload := map[string]any{
		"type":    plan.Type.ValueString(),
		"enabled": plan.Enabled.ValueBool(),
		"name":    plan.Name.ValueString(),
		"action":  plan.Action.ValueString(),
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		payload["description"] = plan.Description.ValueString()
	}

	if !plan.Index.IsNull() && !plan.Index.IsUnknown() {
		payload["index"] = plan.Index.ValueInt64()
	}

	sourceFilter, sourceDiags := buildFirewallFilterPayload(plan.SourceFilter, path.Root("source_filter"))
	diags.Append(sourceDiags...)
	if sourceFilter != nil {
		payload["sourceFilter"] = sourceFilter
	}

	destinationFilter, destDiags := buildFirewallFilterPayload(plan.DestinationFilter, path.Root("destination_filter"))
	diags.Append(destDiags...)
	if destinationFilter != nil {
		payload["destinationFilter"] = destinationFilter
	}

	protocolFilter, protoDiags := decodeProtocolFilter(plan.ProtocolFilter)
	diags.Append(protoDiags...)
	if protocolFilter != nil {
		payload["protocolFilter"] = protocolFilter
	}

	if !plan.NetworkIDFilter.IsNull() && !plan.NetworkIDFilter.IsUnknown() {
		payload["networkIdFilter"] = plan.NetworkIDFilter.ValueString()
	}

	return payload, diags
}

// buildFirewallFilterPayload validates and builds the wire payload for one
// source_filter/destination_filter, matching the real API's discriminated
// union: only the fields relevant to filter.Type may be set, and the
// field(s) that type requires must be non-empty.
func buildFirewallFilterPayload(filter *firewallFilterModel, attrPath path.Path) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	if filter == nil {
		return nil, diags
	}

	filterType := filter.Type.ValueString()
	payload := map[string]any{"type": filterType}

	switch filterType {
	case "IP_ADDRESSES_OR_SUBNETS":
		values, valueDiags := listToStrings(filter.IPAddressesOrSubnets, attrPath.AtName("ip_addresses_or_subnets"))
		diags.Append(valueDiags...)
		if len(values) == 0 {
			diags.AddAttributeError(attrPath.AtName("ip_addresses_or_subnets"), "Missing ip_addresses_or_subnets", "ip_addresses_or_subnets is required when type = IP_ADDRESSES_OR_SUBNETS.")
		} else {
			payload["ipAddressesOrSubnets"] = values
		}
		if ports, portDiags := listToInt64s(filter.PortFilter, attrPath.AtName("port_filter")); len(ports) > 0 {
			payload["portFilter"] = ports
		} else {
			diags.Append(portDiags...)
		}
	case "NETWORKS":
		values, valueDiags := listToStrings(filter.NetworkIDs, attrPath.AtName("network_ids"))
		diags.Append(valueDiags...)
		if len(values) == 0 {
			diags.AddAttributeError(attrPath.AtName("network_ids"), "Missing network_ids", "network_ids is required when type = NETWORKS.")
		} else {
			payload["networkIds"] = values
		}
		if ports, portDiags := listToInt64s(filter.PortFilter, attrPath.AtName("port_filter")); len(ports) > 0 {
			payload["portFilter"] = ports
		} else {
			diags.Append(portDiags...)
		}
	case "PORTS":
		ports, portDiags := listToInt64s(filter.PortFilter, attrPath.AtName("port_filter"))
		diags.Append(portDiags...)
		if len(ports) == 0 {
			diags.AddAttributeError(attrPath.AtName("port_filter"), "Missing port_filter", "port_filter is required when type = PORTS.")
		} else {
			payload["portFilter"] = ports
		}
	case "MAC_ADDRESSES":
		values, valueDiags := listToStrings(filter.MacAddresses, attrPath.AtName("mac_addresses"))
		diags.Append(valueDiags...)
		if len(values) == 0 {
			diags.AddAttributeError(attrPath.AtName("mac_addresses"), "Missing mac_addresses", "mac_addresses is required when type = MAC_ADDRESSES.")
		} else {
			payload["macAddresses"] = values
		}
		if !filter.PrefixLength.IsNull() && !filter.PrefixLength.IsUnknown() {
			payload["prefixLength"] = filter.PrefixLength.ValueInt64()
		}
	default:
		diags.AddAttributeError(attrPath.AtName("type"), "Invalid filter type", "Supported values are IP_ADDRESSES_OR_SUBNETS, NETWORKS, PORTS, or MAC_ADDRESSES.")
		return nil, diags
	}

	return payload, diags
}

func listToInt64s(list types.List, attrPath path.Path) ([]int64, diag.Diagnostics) {
	var diags diag.Diagnostics
	if list.IsNull() || list.IsUnknown() {
		return nil, diags
	}

	var values []int64
	if elemDiags := list.ElementsAs(context.Background(), &values, false); elemDiags.HasError() {
		diags.Append(elemDiags...)
		return nil, diags
	}

	return values, diags
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

func decodeProtocolFilter(protocols types.Set) ([]json.RawMessage, diag.Diagnostics) {
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

func firewallStateFromResponse(siteID string, response any) (firewallResourceModel, diag.Diagnostics) {
	switch result := response.(type) {
	case *acl_rules.CreateACLRuleResponseIpv4:
		return firewallStateFromFields(siteID, result.Id, result.Type, result.Name, result.Description, result.Action, result.Enabled, result.Index, result.SourceFilter, result.DestinationFilter, result.ProtocolFilter, "")
	case *acl_rules.CreateACLRuleResponseMac:
		return firewallStateFromFields(siteID, result.Id, result.Type, result.Name, result.Description, result.Action, result.Enabled, result.Index, result.SourceFilter, result.DestinationFilter, result.ProtocolFilter, result.NetworkIdFilter)
	case *acl_rules.UpdateACLRuleResponseIpv4:
		return firewallStateFromFields(siteID, result.Id, result.Type, result.Name, result.Description, result.Action, result.Enabled, result.Index, result.SourceFilter, result.DestinationFilter, result.ProtocolFilter, "")
	case *acl_rules.UpdateACLRuleResponseMac:
		return firewallStateFromFields(siteID, result.Id, result.Type, result.Name, result.Description, result.Action, result.Enabled, result.Index, result.SourceFilter, result.DestinationFilter, result.ProtocolFilter, result.NetworkIdFilter)
	case *acl_rules.GetACLRuleResponseIpv4:
		return firewallStateFromFields(siteID, result.Id, result.Type, result.Name, result.Description, result.Action, result.Enabled, result.Index, result.SourceFilter, result.DestinationFilter, result.ProtocolFilter, "")
	case *acl_rules.GetACLRuleResponseMac:
		return firewallStateFromFields(siteID, result.Id, result.Type, result.Name, result.Description, result.Action, result.Enabled, result.Index, result.SourceFilter, result.DestinationFilter, nil, result.NetworkIdFilter)
	default:
		var diags diag.Diagnostics
		diags.AddError("Unsupported ACL rule response", fmt.Sprintf("unexpected response type %T", response))
		return firewallResourceModel{}, diags
	}
}

func firewallStateFromFields(siteID, id, ruleType, name, description, action string, enabled bool, index int64, sourceFilter, destinationFilter json.RawMessage, protocolFilter []json.RawMessage, networkIDFilter string) (firewallResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	state := firewallResourceModel{
		ID:          types.StringValue(id),
		SiteID:      types.StringValue(siteID),
		Type:        types.StringValue(ruleType),
		Name:        types.StringValue(name),
		Description: stringValueOrNull(description),
		Action:      types.StringValue(action),
		Enabled:     types.BoolValue(enabled),
		Index:       types.Int64Value(index),
	}

	source, sourceDiags := firewallFilterFromJSON(sourceFilter)
	diags.Append(sourceDiags...)
	state.SourceFilter = source

	destination, destDiags := firewallFilterFromJSON(destinationFilter)
	diags.Append(destDiags...)
	state.DestinationFilter = destination

	state.ProtocolFilter = rawMessagesToSet(protocolFilter)
	state.NetworkIDFilter = stringValueOrNull(networkIDFilter)
	return state, diags
}

// firewallFilterWire is the wire shape shared by every source_filter/
// destination_filter variant: a plain struct with every field any variant
// might carry, decoded generically and then split into firewallFilterModel.
// Only the fields matching the response's own `type` are actually populated
// by the API.
type firewallFilterWire struct {
	Type                 string   `json:"type"`
	IPAddressesOrSubnets []string `json:"ipAddressesOrSubnets,omitempty"`
	NetworkIds           []string `json:"networkIds,omitempty"`
	PortFilter           []int64  `json:"portFilter,omitempty"`
	MacAddresses         []string `json:"macAddresses,omitempty"`
	PrefixLength         *int64   `json:"prefixLength,omitempty"`
}

func firewallFilterFromJSON(raw json.RawMessage) (*firewallFilterModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(raw) == 0 || string(raw) == "null" {
		return nil, diags
	}

	var wire firewallFilterWire
	if err := json.Unmarshal(raw, &wire); err != nil {
		diags.AddError("Unable to decode ACL rule filter", err.Error())
		return nil, diags
	}

	model := &firewallFilterModel{
		Type:                 types.StringValue(wire.Type),
		IPAddressesOrSubnets: stringsToList(wire.IPAddressesOrSubnets),
		NetworkIDs:           stringsToList(wire.NetworkIds),
		PortFilter:           int64sToList(wire.PortFilter),
		MacAddresses:         stringsToList(wire.MacAddresses),
		PrefixLength:         types.Int64Null(),
	}
	if wire.PrefixLength != nil {
		model.PrefixLength = types.Int64Value(*wire.PrefixLength)
	}
	return model, diags
}

// normalizeJSONString re-encodes raw into JSON with Go's canonical map key
// ordering (alphabetical), which is what HCL's jsonencode() also produces.
// Without this, an opaque-JSON-string attribute (e.g. unifi_firewall_policy's
// source_traffic_filter_json/destination_traffic_filter_json, or
// unifi_traffic_matching_list's items_json) would show a perpetual diff: the
// API doesn't echo back JSON keys in the same order the request used, so a
// plain byte-for-byte passthrough of the response never matches a config
// built with jsonencode(), even when the content is identical.
func normalizeJSONString(raw json.RawMessage) types.String {
	if len(raw) == 0 || string(raw) == "null" {
		return types.StringNull()
	}
	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return types.StringValue(string(raw))
	}
	normalized, err := json.Marshal(decoded)
	if err != nil {
		return types.StringValue(string(raw))
	}
	return types.StringValue(string(normalized))
}

func rawMessagesToList(raw []json.RawMessage) types.List {
	if len(raw) == 0 {
		return types.ListNull(types.StringType)
	}
	values := make([]string, 0, len(raw))
	for _, item := range raw {
		trimmed := strings.TrimSpace(string(item))
		if trimmed == "" || trimmed == "null" {
			continue
		}
		var decoded string
		if err := json.Unmarshal(item, &decoded); err == nil {
			values = append(values, decoded)
			continue
		}
		values = append(values, string(item))
	}
	if len(values) == 0 {
		return types.ListNull(types.StringType)
	}
	list, _ := types.ListValueFrom(context.Background(), types.StringType, values)
	return list
}

// rawMessagesToSet mirrors rawMessagesToList, but for protocol_filter: the
// real API stores it unordered and doesn't echo back the order it was sent
// in, so a types.List there would trip Terraform's "provider produced
// inconsistent result" check the moment the API's own order differs from
// the config's. A types.Set has no such ordering requirement.
func rawMessagesToSet(raw []json.RawMessage) types.Set {
	if len(raw) == 0 {
		return types.SetNull(types.StringType)
	}
	values := make([]string, 0, len(raw))
	for _, item := range raw {
		trimmed := strings.TrimSpace(string(item))
		if trimmed == "" || trimmed == "null" {
			continue
		}
		var decoded string
		if err := json.Unmarshal(item, &decoded); err == nil {
			values = append(values, decoded)
			continue
		}
		values = append(values, string(item))
	}
	if len(values) == 0 {
		return types.SetNull(types.StringType)
	}
	set, _ := types.SetValueFrom(context.Background(), types.StringType, values)
	return set
}
