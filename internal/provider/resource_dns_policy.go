package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
	"github.com/ilmax/unifi-client-go/pkg/errors"
	networkapi "github.com/ilmax/unifi-client-go/pkg/network"
)

type dnsPolicyResource struct {
	client *unifi.Client
	siteID string
}

type dnsPolicyResourceModel struct {
	ID      types.String `tfsdk:"id"`
	SiteID  types.String `tfsdk:"site_id"`
	Type    types.String `tfsdk:"type"`
	Enabled types.Bool   `tfsdk:"enabled"`
	Domain  types.String `tfsdk:"domain"`
	Origin  types.String `tfsdk:"origin"`
}

func NewDNSPolicyResource() resource.Resource {
	return &dnsPolicyResource{}
}

func (r *dnsPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_policy"
}

func (r *dnsPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Required: true,
			},
			"enabled": schema.BoolAttribute{
				Required: true,
			},
			"domain": schema.StringAttribute{
				Computed: true,
			},
			"origin": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *dnsPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *dnsPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dnsPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(plan.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	payload := networkapi.CreateOrUpdateDNSPolicy{
		Type:    plan.Type.ValueString(),
		Enabled: plan.Enabled.ValueBool(),
	}

	apiPath := fmt.Sprintf("/v1/sites/%s/dns-policies", siteID)
	var result networkapi.DNSPolicy
	if err := r.client.Post(ctx, apiPath, payload, &result); err != nil {
		resp.Diagnostics.AddError("Unable to create DNS policy", err.Error())
		return
	}

	state := dnsPolicyStateFromAPI(siteID, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dnsPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dnsPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(state.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	if state.ID.IsNull() || state.ID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("id"), "Missing DNS policy ID", "The DNS policy ID is required to read the resource.")
		return
	}

	apiPath := fmt.Sprintf("/v1/sites/%s/dns-policies/%s", siteID, state.ID.ValueString())
	var result networkapi.DNSPolicy
	if err := r.client.Get(ctx, apiPath, &result); err != nil {
		if errors.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read DNS policy", err.Error())
		return
	}

	state = dnsPolicyStateFromAPI(siteID, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dnsPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dnsPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(plan.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	if plan.ID.IsNull() || plan.ID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("id"), "Missing DNS policy ID", "The DNS policy ID is required to update the resource.")
		return
	}

	payload := networkapi.CreateOrUpdateDNSPolicy{
		Type:    plan.Type.ValueString(),
		Enabled: plan.Enabled.ValueBool(),
	}

	apiPath := fmt.Sprintf("/v1/sites/%s/dns-policies/%s", siteID, plan.ID.ValueString())
	var result networkapi.DNSPolicy
	if err := r.client.Put(ctx, apiPath, payload, &result); err != nil {
		resp.Diagnostics.AddError("Unable to update DNS policy", err.Error())
		return
	}

	state := dnsPolicyStateFromAPI(siteID, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dnsPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dnsPolicyResourceModel
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

	apiPath := fmt.Sprintf("/v1/sites/%s/dns-policies/%s", siteID, state.ID.ValueString())
	if err := r.client.Delete(ctx, apiPath, nil); err != nil && !errors.IsNotFoundError(err) {
		resp.Diagnostics.AddError("Unable to delete DNS policy", err.Error())
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *dnsPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	siteID, dnsPolicyID, err := splitImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import identifier", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site_id"), siteID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), dnsPolicyID)...)
}

func dnsPolicyStateFromAPI(siteID string, policy networkapi.DNSPolicy) dnsPolicyResourceModel {
	domain := types.StringNull()
	if policy.Domain != nil {
		domain = stringValueOrNull(*policy.Domain)
	}

	return dnsPolicyResourceModel{
		ID:      types.StringValue(policy.Id.String()),
		SiteID:  types.StringValue(siteID),
		Type:    stringValueOrNull(policy.Type),
		Enabled: types.BoolValue(policy.Enabled),
		Domain:  domain,
		Origin:  stringValueOrNull(policy.Metadata.Origin),
	}
}
