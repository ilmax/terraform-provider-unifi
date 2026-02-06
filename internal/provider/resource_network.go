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
	"github.com/ilmax/unifi-client-go/pkg/errors"
	"github.com/ilmax/unifi-client-go/pkg/networks"
)

type networkResource struct {
	client *unifi.Client
	siteID string
}

type networkResourceModel struct {
	ID                    types.String `tfsdk:"id"`
	SiteID                types.String `tfsdk:"site_id"`
	Name                  types.String `tfsdk:"name"`
	Management            types.String `tfsdk:"management"`
	Enabled               types.Bool   `tfsdk:"enabled"`
	VlanID                types.Int64  `tfsdk:"vlan_id"`
	IPv4ConfigurationJSON types.String `tfsdk:"ipv4_configuration_json"`
	IPv6ConfigurationJSON types.String `tfsdk:"ipv6_configuration_json"`
}

type networkBase struct {
	id         string
	name       string
	management string
	enabled    bool
	vlanID     int64
	ipv4       json.RawMessage
	ipv6       json.RawMessage
}

func NewNetworkResource() resource.Resource {
	return &networkResource{}
}

func (r *networkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

func (r *networkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"name": schema.StringAttribute{
				Required: true,
			},
			"management": schema.StringAttribute{
				Required: true,
			},
			"enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"vlan_id": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"ipv4_configuration_json": schema.StringAttribute{
				Optional: true,
			},
			"ipv6_configuration_json": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (r *networkResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *networkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan networkResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(plan.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	payload, diags := buildNetworkPayload(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var raw json.RawMessage
	path := fmt.Sprintf("/v1/sites/%s/networks", siteID)
	if err := r.client.Post(ctx, path, payload, &raw); err != nil {
		resp.Diagnostics.AddError("Unable to create network", err.Error())
		return
	}

	decoded, err := networks.DecodeCreateNetworkResponse(raw)
	if err != nil {
		resp.Diagnostics.AddError("Unable to decode network", err.Error())
		return
	}

	base, baseDiags := networkBaseFromAny(decoded)
	resp.Diagnostics.Append(baseDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := networkResourceModel{
		ID:                    types.StringValue(base.id),
		SiteID:                types.StringValue(siteID),
		Name:                  types.StringValue(base.name),
		Management:            types.StringValue(base.management),
		Enabled:               types.BoolValue(base.enabled),
		VlanID:                types.Int64Value(base.vlanID),
		IPv4ConfigurationJSON: plan.IPv4ConfigurationJSON,
		IPv6ConfigurationJSON: plan.IPv6ConfigurationJSON,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *networkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state networkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(state.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	if state.ID.IsNull() || state.ID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("id"), "Missing network ID", "The network ID is required to read the network.")
		return
	}

	var raw json.RawMessage
	path := fmt.Sprintf("/v1/sites/%s/networks/%s", siteID, state.ID.ValueString())
	if err := r.client.Get(ctx, path, &raw); err != nil {
		if errors.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read network", err.Error())
		return
	}

	decoded, err := networks.DecodeGetNetworkDetailsResponse(raw)
	if err != nil {
		resp.Diagnostics.AddError("Unable to decode network", err.Error())
		return
	}

	base, baseDiags := networkBaseFromAny(decoded)
	resp.Diagnostics.Append(baseDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.SiteID = types.StringValue(siteID)
	state.Name = types.StringValue(base.name)
	state.Management = types.StringValue(base.management)
	state.Enabled = types.BoolValue(base.enabled)
	state.VlanID = types.Int64Value(base.vlanID)
	state.IPv4ConfigurationJSON = rawMessageToOptionalString(base.ipv4)
	state.IPv6ConfigurationJSON = rawMessageToOptionalString(base.ipv6)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *networkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan networkResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(plan.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	if plan.ID.IsNull() || plan.ID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("id"), "Missing network ID", "The network ID is required to update the network.")
		return
	}

	payload, diags := buildNetworkPayload(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var raw json.RawMessage
	path := fmt.Sprintf("/v1/sites/%s/networks/%s", siteID, plan.ID.ValueString())
	if err := r.client.Put(ctx, path, payload, &raw); err != nil {
		resp.Diagnostics.AddError("Unable to update network", err.Error())
		return
	}

	decoded, err := networks.DecodeUpdateNetworkResponse(raw)
	if err != nil {
		resp.Diagnostics.AddError("Unable to decode network", err.Error())
		return
	}

	base, baseDiags := networkBaseFromAny(decoded)
	resp.Diagnostics.Append(baseDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := networkResourceModel{
		ID:                    types.StringValue(base.id),
		SiteID:                types.StringValue(siteID),
		Name:                  types.StringValue(base.name),
		Management:            types.StringValue(base.management),
		Enabled:               types.BoolValue(base.enabled),
		VlanID:                types.Int64Value(base.vlanID),
		IPv4ConfigurationJSON: plan.IPv4ConfigurationJSON,
		IPv6ConfigurationJSON: plan.IPv6ConfigurationJSON,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *networkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state networkResourceModel
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

	path := fmt.Sprintf("/v1/sites/%s/networks/%s", siteID, state.ID.ValueString())
	if err := r.client.Delete(ctx, path, nil); err != nil && !errors.IsNotFoundError(err) {
		resp.Diagnostics.AddError("Unable to delete network", err.Error())
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *networkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	siteID, networkID, err := splitImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import identifier", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site_id"), siteID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), networkID)...)
}

func buildNetworkPayload(plan networkResourceModel) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	payload := map[string]interface{}{
		"management": plan.Management.ValueString(),
		"name":       plan.Name.ValueString(),
		"enabled":    plan.Enabled.ValueBool(),
		"vlanId":     plan.VlanID.ValueInt64(),
	}

	if !plan.IPv4ConfigurationJSON.IsNull() && !plan.IPv4ConfigurationJSON.IsUnknown() {
		if err := requireValidJSON(plan.IPv4ConfigurationJSON.ValueString()); err != nil {
			diags.AddAttributeError(path.Root("ipv4_configuration_json"), "Invalid ipv4_configuration_json", err.Error())
			return nil, diags
		}
		payload["ipv4Configuration"] = json.RawMessage(plan.IPv4ConfigurationJSON.ValueString())
	}

	if !plan.IPv6ConfigurationJSON.IsNull() && !plan.IPv6ConfigurationJSON.IsUnknown() {
		if err := requireValidJSON(plan.IPv6ConfigurationJSON.ValueString()); err != nil {
			diags.AddAttributeError(path.Root("ipv6_configuration_json"), "Invalid ipv6_configuration_json", err.Error())
			return nil, diags
		}
		payload["ipv6Configuration"] = json.RawMessage(plan.IPv6ConfigurationJSON.ValueString())
	}

	return payload, diags
}

func networkBaseFromAny(response any) (networkBase, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch result := response.(type) {
	case *networks.CreateNetworkResponseGateway:
		return networkBase{id: result.Id, name: result.Name, management: result.Management, enabled: result.Enabled, vlanID: result.VlanId}, diags
	case *networks.CreateNetworkResponseSwitch:
		return networkBase{id: result.Id, name: result.Name, management: result.Management, enabled: result.Enabled, vlanID: result.VlanId}, diags
	case *networks.CreateNetworkResponseUnmanaged:
		return networkBase{id: result.Id, name: result.Name, management: result.Management, enabled: result.Enabled, vlanID: result.VlanId}, diags
	case *networks.UpdateNetworkResponseGateway:
		return networkBase{id: result.Id, name: result.Name, management: result.Management, enabled: result.Enabled, vlanID: result.VlanId}, diags
	case *networks.UpdateNetworkResponseSwitch:
		return networkBase{id: result.Id, name: result.Name, management: result.Management, enabled: result.Enabled, vlanID: result.VlanId}, diags
	case *networks.UpdateNetworkResponseUnmanaged:
		return networkBase{id: result.Id, name: result.Name, management: result.Management, enabled: result.Enabled, vlanID: result.VlanId}, diags
	case *networks.GetNetworkDetailsResponseGateway:
		ipv4, ipv6 := marshalNetworkConfig(result.Ipv4Configuration, result.Ipv6Configuration, &diags)
		return networkBase{id: result.Id, name: result.Name, management: result.Management, enabled: result.Enabled, vlanID: result.VlanId, ipv4: ipv4, ipv6: ipv6}, diags
	case *networks.GetNetworkDetailsResponseSwitch:
		ipv4, _ := marshalNetworkConfig(result.Ipv4Configuration, nil, &diags)
		return networkBase{id: result.Id, name: result.Name, management: result.Management, enabled: result.Enabled, vlanID: result.VlanId, ipv4: ipv4}, diags
	case *networks.GetNetworkDetailsResponseUnmanaged:
		return networkBase{id: result.Id, name: result.Name, management: result.Management, enabled: result.Enabled, vlanID: result.VlanId}, diags
	default:
		diags.AddError("Unsupported network response", fmt.Sprintf("unexpected response type %T", response))
		return networkBase{}, diags
	}
}

func marshalNetworkConfig(ipv4 any, ipv6 any, diags *diag.Diagnostics) (json.RawMessage, json.RawMessage) {
	var out4 json.RawMessage
	var out6 json.RawMessage

	if ipv4 != nil {
		encoded, err := json.Marshal(ipv4)
		if err != nil {
			diags.AddError("Unable to encode IPv4 configuration", err.Error())
		} else if string(encoded) != "null" {
			out4 = encoded
		}
	}

	if ipv6 != nil {
		encoded, err := json.Marshal(ipv6)
		if err != nil {
			diags.AddError("Unable to encode IPv6 configuration", err.Error())
		} else if string(encoded) != "null" {
			out6 = encoded
		}
	}

	return out4, out6
}

func rawMessageToOptionalString(raw json.RawMessage) types.String {
	if len(raw) == 0 {
		return types.StringNull()
	}
	if string(raw) == "null" {
		return types.StringNull()
	}
	return types.StringValue(string(raw))
}
