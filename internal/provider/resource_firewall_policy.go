package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/sdkcompat/firewallpolicies"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
	"github.com/ilmax/unifi-client-go/pkg/errors"
)

type firewallPolicyResource struct {
	client *unifi.Client
	siteID string
}

type firewallPolicyResourceModel struct {
	ID                           types.String `tfsdk:"id"`
	SiteID                       types.String `tfsdk:"site_id"`
	Name                         types.String `tfsdk:"name"`
	Description                  types.String `tfsdk:"description"`
	Action                       types.String `tfsdk:"action"`
	AllowReturnTraffic           types.Bool   `tfsdk:"allow_return_traffic"`
	Enabled                      types.Bool   `tfsdk:"enabled"`
	LoggingEnabled               types.Bool   `tfsdk:"logging_enabled"`
	IPProtocolScope              types.String `tfsdk:"ip_protocol_scope"`
	IpsecFilter                  types.String `tfsdk:"ipsec_filter"`
	ConnectionStateFilter        types.Set    `tfsdk:"connection_state_filter"`
	SourceZoneID                 types.String `tfsdk:"source_zone_id"`
	DestinationZoneID            types.String `tfsdk:"destination_zone_id"`
	SourceTrafficFilterJSON      types.String `tfsdk:"source_traffic_filter_json"`
	DestinationTrafficFilterJSON types.String `tfsdk:"destination_traffic_filter_json"`
	Index                        types.Int64  `tfsdk:"index"`
	Origin                       types.String `tfsdk:"origin"`
}

func NewFirewallPolicyResource() resource.Resource {
	return &firewallPolicyResource{}
}

func (r *firewallPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_policy"
}

func (r *firewallPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A Zone Based Firewall policy: an allow/block/reject rule between two firewall zones. " +
			"Requires unifi_firewall_zone for source_zone_id/destination_zone_id, and (like real UniFi hardware) " +
			"a gateway model that actually supports Zone Based Firewall. " +
			"The traffic filter (matching specific networks/IPs/ports/etc. within a zone, rather than the whole zone) " +
			"is a large discriminated union in the real API; source_traffic_filter_json/destination_traffic_filter_json " +
			"take it as a raw JSON string — the same escape hatch unifi_firewall_rule uses for its filters — rather " +
			"than modeling every filter variant as native attributes. Schedules are not yet supported: an unscheduled " +
			"policy is always active, which is the common case.",
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
			"name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"action": schema.StringAttribute{
				Required:    true,
				Description: "ALLOW, BLOCK, or REJECT.",
			},
			"allow_return_traffic": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Only valid when action = ALLOW: also creates a derived policy on the mirrored zone pair to allow return traffic. Omitting it sends false.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"logging_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"ip_protocol_scope": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("IPV4_AND_IPV6"),
				Description: "IPV4, IPV6, or IPV4_AND_IPV6. The real API nests this under ipProtocolScope.ipVersion " +
					"alongside an optional protocolFilter (named protocols/presets/protocol numbers, itself a deeply " +
					"nested discriminated union); protocolFilter isn't modeled here, so this always matches all protocols.",
			},
			"ipsec_filter": schema.StringAttribute{
				Optional:    true,
				Description: "MATCH_ENCRYPTED or MATCH_NOT_ENCRYPTED. Omitted: matches all traffic regardless of IPsec encryption.",
			},
			"connection_state_filter": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Any of NEW, INVALID, ESTABLISHED, RELATED. Omitted: matches all connection states. " +
					"A set, not a list: the real API stores this unordered and doesn't echo back the order it was sent in.",
			},
			"source_zone_id": schema.StringAttribute{
				Required: true,
			},
			"destination_zone_id": schema.StringAttribute{
				Required: true,
			},
			"source_traffic_filter_json": schema.StringAttribute{
				Optional:    true,
				Description: "Raw JSON matching the API's source traffic filter shape. Omitted: matches all traffic from the source zone.",
			},
			"destination_traffic_filter_json": schema.StringAttribute{
				Optional:    true,
				Description: "Raw JSON matching the API's destination traffic filter shape. Omitted: matches all traffic to the destination zone.",
			},
			"index": schema.Int64Attribute{
				Computed:    true,
				Description: "Priority position among this site's firewall policies (lower = higher priority). Manage with unifi_firewall_policy_ordering, not here.",
			},
			"origin": schema.StringAttribute{
				Computed:    true,
				Description: "USER_DEFINED, SYSTEM_DEFINED, or DERIVED (auto-created by action.allow_return_traffic on another policy).",
			},
		},
	}
}

func (r *firewallPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *firewallPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan firewallPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(plan.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	payload, diags := buildFirewallPolicyPayload(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result firewallpolicies.Policy
	apiPath := fmt.Sprintf("/v1/sites/%s/firewall/policies", siteID)
	if err := r.client.Post(ctx, apiPath, payload, &result); err != nil {
		resp.Diagnostics.AddError("Unable to create firewall policy", err.Error())
		return
	}

	state := firewallPolicyStateFromAPI(siteID, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *firewallPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state firewallPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(state.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	if state.ID.IsNull() || state.ID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("id"), "Missing firewall policy ID", "The firewall policy ID is required to read the resource.")
		return
	}

	var result firewallpolicies.Policy
	apiPath := fmt.Sprintf("/v1/sites/%s/firewall/policies/%s", siteID, state.ID.ValueString())
	if err := r.client.Get(ctx, apiPath, &result); err != nil {
		if errors.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read firewall policy", err.Error())
		return
	}

	updated := firewallPolicyStateFromAPI(siteID, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &updated)...)
}

func (r *firewallPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan firewallPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(plan.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	if plan.ID.IsNull() || plan.ID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("id"), "Missing firewall policy ID", "The firewall policy ID is required to update the resource.")
		return
	}

	payload, diags := buildFirewallPolicyPayload(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result firewallpolicies.Policy
	apiPath := fmt.Sprintf("/v1/sites/%s/firewall/policies/%s", siteID, plan.ID.ValueString())
	if err := r.client.Put(ctx, apiPath, payload, &result); err != nil {
		resp.Diagnostics.AddError("Unable to update firewall policy", err.Error())
		return
	}

	state := firewallPolicyStateFromAPI(siteID, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *firewallPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state firewallPolicyResourceModel
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

	apiPath := fmt.Sprintf("/v1/sites/%s/firewall/policies/%s", siteID, state.ID.ValueString())
	if err := r.client.Delete(ctx, apiPath, nil); err != nil && !errors.IsNotFoundError(err) {
		resp.Diagnostics.AddError("Unable to delete firewall policy", err.Error())
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *firewallPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	siteID, policyID, err := splitImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import identifier", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site_id"), siteID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), policyID)...)
}

func buildFirewallPolicyPayload(plan firewallPolicyResourceModel) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics

	actionType := strings.ToUpper(plan.Action.ValueString())
	action := map[string]any{"type": actionType}
	switch actionType {
	case "ALLOW":
		action["allowReturnTraffic"] = boolOrDefault(plan.AllowReturnTraffic, false)
	case "BLOCK", "REJECT":
		if !plan.AllowReturnTraffic.IsNull() && !plan.AllowReturnTraffic.IsUnknown() {
			diags.AddAttributeError(path.Root("allow_return_traffic"), "Unsupported field", "allow_return_traffic is only supported for action = ALLOW.")
			return nil, diags
		}
	default:
		diags.AddAttributeError(path.Root("action"), "Invalid action", "Supported values are ALLOW, BLOCK, or REJECT.")
		return nil, diags
	}

	source := map[string]any{"zoneId": plan.SourceZoneID.ValueString()}
	sourceFilter, sourceDiags := decodeJSONRaw(plan.SourceTrafficFilterJSON, path.Root("source_traffic_filter_json"))
	diags.Append(sourceDiags...)
	if sourceFilter != nil {
		source["trafficFilter"] = sourceFilter
	}

	destination := map[string]any{"zoneId": plan.DestinationZoneID.ValueString()}
	destFilter, destDiags := decodeJSONRaw(plan.DestinationTrafficFilterJSON, path.Root("destination_traffic_filter_json"))
	diags.Append(destDiags...)
	if destFilter != nil {
		destination["trafficFilter"] = destFilter
	}

	if diags.HasError() {
		return nil, diags
	}

	payload := map[string]any{
		"name":            plan.Name.ValueString(),
		"action":          action,
		"enabled":         boolOrDefault(plan.Enabled, true),
		"loggingEnabled":  boolOrDefault(plan.LoggingEnabled, false),
		"ipProtocolScope": map[string]any{"ipVersion": stringOrDefault(plan.IPProtocolScope, "IPV4_AND_IPV6")},
		"source":          source,
		"destination":     destination,
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		payload["description"] = plan.Description.ValueString()
	}
	if !plan.IpsecFilter.IsNull() && !plan.IpsecFilter.IsUnknown() {
		payload["ipsecFilter"] = plan.IpsecFilter.ValueString()
	}
	if !plan.ConnectionStateFilter.IsNull() && !plan.ConnectionStateFilter.IsUnknown() {
		values, csDiags := setToStrings(plan.ConnectionStateFilter, path.Root("connection_state_filter"))
		diags.Append(csDiags...)
		if len(values) > 0 {
			payload["connectionStateFilter"] = values
		}
	}

	return payload, diags
}

func stringOrDefault(v types.String, def string) string {
	if v.IsNull() || v.IsUnknown() {
		return def
	}
	return v.ValueString()
}

func firewallPolicyStateFromAPI(siteID string, p firewallpolicies.Policy) firewallPolicyResourceModel {
	state := firewallPolicyResourceModel{
		ID:                types.StringValue(p.Id),
		SiteID:            types.StringValue(siteID),
		Name:              types.StringValue(p.Name),
		Description:       stringValueOrNull(p.Description),
		Action:            types.StringValue(p.Action.Type),
		Enabled:           types.BoolValue(p.Enabled),
		LoggingEnabled:    types.BoolValue(p.LoggingEnabled),
		IPProtocolScope:   types.StringValue(p.IpProtocolScope.IpVersion),
		IpsecFilter:       stringValueOrNull(p.IpsecFilter),
		SourceZoneID:      types.StringValue(p.Source.ZoneId),
		DestinationZoneID: types.StringValue(p.Destination.ZoneId),
		Index:             types.Int64Value(p.Index),
		Origin:            types.StringNull(),
	}

	if p.Action.AllowReturnTraffic != nil {
		state.AllowReturnTraffic = types.BoolValue(*p.Action.AllowReturnTraffic)
	} else {
		state.AllowReturnTraffic = types.BoolNull()
	}

	state.SourceTrafficFilterJSON = normalizeJSONString(p.Source.TrafficFilter)
	state.DestinationTrafficFilterJSON = normalizeJSONString(p.Destination.TrafficFilter)
	state.ConnectionStateFilter = stringsToSet(p.ConnectionStateFilter)

	if p.Metadata != nil {
		state.Origin = stringValueOrNull(p.Metadata.Origin)
	}

	return state
}

func stringsToList(values []string) types.List {
	if len(values) == 0 {
		return types.ListNull(types.StringType)
	}
	list, _ := types.ListValueFrom(context.Background(), types.StringType, values)
	return list
}

// setToStrings mirrors listToStrings (resource_wifi.go), but for a
// types.Set: connection_state_filter is a set, not a list, because the real
// API stores it unordered and doesn't echo back the order it was sent in —
// a types.List there would trip Terraform's "provider produced inconsistent
// result" check the moment the API's own order differs from the config's.
func setToStrings(set types.Set, attrPath path.Path) ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if set.IsNull() || set.IsUnknown() {
		return nil, diags
	}

	var values []string
	if elemDiags := set.ElementsAs(context.Background(), &values, false); elemDiags.HasError() {
		diags.Append(elemDiags...)
		return nil, diags
	}

	return values, diags
}

// stringsToSet mirrors stringsToList above, for the same reason
// setToStrings mirrors listToStrings.
func stringsToSet(values []string) types.Set {
	if len(values) == 0 {
		return types.SetNull(types.StringType)
	}
	set, _ := types.SetValueFrom(context.Background(), types.StringType, values)
	return set
}
