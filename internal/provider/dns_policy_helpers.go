package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
	"github.com/ilmax/unifi-client-go/pkg/errors"
	networkapi "github.com/ilmax/terraform-provider-unifi/internal/sdkcompat/dns"
)

type dnsPolicyPlanModel interface {
	dnsPolicySiteID() types.String
}

type dnsPolicyStateModel interface {
	dnsPolicyPlanModel
	dnsPolicyID() types.String
}

type dnsPolicyResourceBase struct {
	client *unifi.Client
	siteID string
}

func (r *dnsPolicyResourceBase) configure(req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func dnsPolicyIDAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
}

func dnsPolicySiteIDAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		Optional: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
}

func dnsPolicyEnabledAttribute() schema.BoolAttribute {
	return schema.BoolAttribute{Required: true}
}

func dnsPolicyDomainAttribute(optional bool) schema.StringAttribute {
	return schema.StringAttribute{Optional: optional, Computed: !optional}
}

func dnsPolicyOriginAttribute() schema.StringAttribute {
	return schema.StringAttribute{Computed: true}
}

func createTypedDNSPolicyResource[T dnsPolicyPlanModel](
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
	client *unifi.Client,
	providerSiteID string,
	buildPayload func(T) (networkapi.CreateOrUpdateDNSPolicy, error),
	stateFromAPI func(string, networkapi.DNSPolicy) (T, error),
) {
	var plan T
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(plan.dnsPolicySiteID(), providerSiteID, &resp.Diagnostics)
	if !ok {
		return
	}

	payload, err := buildPayload(plan)
	if err != nil {
		resp.Diagnostics.AddError("Unable to encode DNS policy", err.Error())
		return
	}

	apiPath := fmt.Sprintf("/v1/sites/%s/dns-policies", siteID)
	var result networkapi.DNSPolicy
	if err := client.Post(ctx, apiPath, payload, &result); err != nil {
		resp.Diagnostics.AddError("Unable to create DNS policy", err.Error())
		return
	}

	state, err := stateFromAPI(siteID, result)
	if err != nil {
		resp.Diagnostics.AddError("Unable to decode DNS policy", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func readTypedDNSPolicyResource[T dnsPolicyStateModel](
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
	client *unifi.Client,
	providerSiteID string,
	stateFromAPI func(string, networkapi.DNSPolicy) (T, error),
) {
	var state T
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(state.dnsPolicySiteID(), providerSiteID, &resp.Diagnostics)
	if !ok {
		return
	}

	id := state.dnsPolicyID()
	if id.IsNull() || id.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("id"), "Missing DNS policy ID", "The DNS policy ID is required to read the resource.")
		return
	}

	apiPath := fmt.Sprintf("/v1/sites/%s/dns-policies/%s", siteID, id.ValueString())
	var result networkapi.DNSPolicy
	if err := client.Get(ctx, apiPath, &result); err != nil {
		if errors.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Unable to read DNS policy", err.Error())
		return
	}

	updatedState, err := stateFromAPI(siteID, result)
	if err != nil {
		resp.Diagnostics.AddError("Unable to decode DNS policy", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func updateTypedDNSPolicyResource[T dnsPolicyStateModel](
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
	client *unifi.Client,
	providerSiteID string,
	buildPayload func(T) (networkapi.CreateOrUpdateDNSPolicy, error),
	stateFromAPI func(string, networkapi.DNSPolicy) (T, error),
) {
	var plan T
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(plan.dnsPolicySiteID(), providerSiteID, &resp.Diagnostics)
	if !ok {
		return
	}

	id := plan.dnsPolicyID()
	if id.IsNull() || id.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("id"), "Missing DNS policy ID", "The DNS policy ID is required to update the resource.")
		return
	}

	payload, err := buildPayload(plan)
	if err != nil {
		resp.Diagnostics.AddError("Unable to encode DNS policy", err.Error())
		return
	}

	apiPath := fmt.Sprintf("/v1/sites/%s/dns-policies/%s", siteID, id.ValueString())
	var result networkapi.DNSPolicy
	if err := client.Put(ctx, apiPath, payload, &result); err != nil {
		resp.Diagnostics.AddError("Unable to update DNS policy", err.Error())
		return
	}

	state, err := stateFromAPI(siteID, result)
	if err != nil {
		resp.Diagnostics.AddError("Unable to decode DNS policy", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func deleteTypedDNSPolicyResource[T dnsPolicyStateModel](
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
	client *unifi.Client,
	providerSiteID string,
) {
	var state T
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(state.dnsPolicySiteID(), providerSiteID, &resp.Diagnostics)
	if !ok {
		return
	}

	id := state.dnsPolicyID()
	if id.IsNull() || id.IsUnknown() {
		return
	}

	apiPath := fmt.Sprintf("/v1/sites/%s/dns-policies/%s", siteID, id.ValueString())
	if err := client.Delete(ctx, apiPath, nil); err != nil && !errors.IsNotFoundError(err) {
		resp.Diagnostics.AddError("Unable to delete DNS policy", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

func importTypedDNSPolicyResource(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	siteID, dnsPolicyID, err := splitImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import identifier", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site_id"), siteID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), dnsPolicyID)...)
}

func dnsPolicyBaseFromAPI(policy networkapi.DNSPolicy) (networkapi.DNSPolicyBase, error) {
	raw, err := json.Marshal(policy)
	if err != nil {
		return networkapi.DNSPolicyBase{}, fmt.Errorf("encode dns policy response: %w", err)
	}

	var base networkapi.DNSPolicyBase
	if err := json.Unmarshal(raw, &base); err != nil {
		return networkapi.DNSPolicyBase{}, fmt.Errorf("decode dns policy response: %w", err)
	}

	return base, nil
}

func expectDNSPolicyType(policy networkapi.DNSPolicy, expected string) error {
	actual, err := policy.Discriminator()
	if err != nil {
		return fmt.Errorf("decode dns policy discriminator: %w", err)
	}
	if actual != expected {
		return fmt.Errorf("expected DNS policy type %q, got %q", expected, actual)
	}
	return nil
}

func terraformStringPointer(value types.String) *string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := value.ValueString()
	return &v
}

func terraformInt32Pointer(value types.Int64, attributePath path.Path) (*int32, diag.Diagnostics) {
	var diags diag.Diagnostics
	if value.IsNull() || value.IsUnknown() {
		return nil, diags
	}

	v := value.ValueInt64()
	if v < math.MinInt32 || v > math.MaxInt32 {
		diags.AddAttributeError(attributePath, "Value out of range", "The value must fit in a signed 32-bit integer.")
		return nil, diags
	}

	out := int32(v)
	return &out, diags
}

func diagnosticsError(diags diag.Diagnostics) error {
	if len(diags) == 0 {
		return nil
	}
	return fmt.Errorf("%s", diags[0].Summary())
}

func stringPointerValueOrNull(value *string) types.String {
	if value == nil {
		return types.StringNull()
	}
	return stringValueOrNull(*value)
}

func int32PointerValueOrNull(value *int32) types.Int64 {
	if value == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*value))
}
