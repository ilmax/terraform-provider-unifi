package provider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
	networkapi "github.com/ilmax/unifi-client-go/pkg/network"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type aclRuleOrderingResource struct {
	client *unifi.Client
	siteID string
}

type aclRuleOrderingResourceModel struct {
	ID                types.String `tfsdk:"id"`
	SiteID            types.String `tfsdk:"site_id"`
	OrderedACLRuleIDs types.List   `tfsdk:"ordered_acl_rule_ids"`
}

func NewACLRuleOrderingResource() resource.Resource {
	return &aclRuleOrderingResource{}
}

func (r *aclRuleOrderingResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acl_rule_ordering"
}

func (r *aclRuleOrderingResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ordered_acl_rule_ids": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *aclRuleOrderingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *aclRuleOrderingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan aclRuleOrderingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(plan.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	state, diags := r.applyOrdering(ctx, siteID, plan.OrderedACLRuleIDs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *aclRuleOrderingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state aclRuleOrderingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(state.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	apiPath := fmt.Sprintf("/v1/sites/%s/acl-rules/ordering", siteID)
	var result networkapi.ACLRuleOrdering
	if err := r.client.Get(ctx, apiPath, &result); err != nil {
		resp.Diagnostics.AddError("Unable to read ACL rule ordering", err.Error())
		return
	}

	state.SiteID = types.StringValue(siteID)
	state.ID = types.StringValue(siteID)
	state.OrderedACLRuleIDs = uuidListToStringList(result.OrderedAclRuleIds)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *aclRuleOrderingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan aclRuleOrderingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(plan.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	state, diags := r.applyOrdering(ctx, siteID, plan.OrderedACLRuleIDs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *aclRuleOrderingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *aclRuleOrderingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	siteID := req.ID
	if siteID == "" {
		resp.Diagnostics.AddError("Invalid import identifier", "Expected import identifier format: <site_id>")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site_id"), siteID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), siteID)...)
}

func (r *aclRuleOrderingResource) applyOrdering(ctx context.Context, siteID string, ordered types.List) (aclRuleOrderingResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	orderedIDs, parseDiags := listToUUIDs(ordered, path.Root("ordered_acl_rule_ids"))
	diags.Append(parseDiags...)
	if diags.HasError() {
		return aclRuleOrderingResourceModel{}, diags
	}

	payload := networkapi.ACLRuleOrdering{
		OrderedAclRuleIds: orderedIDs,
	}

	apiPath := fmt.Sprintf("/v1/sites/%s/acl-rules/ordering", siteID)
	var result networkapi.ACLRuleOrdering
	if err := r.client.Put(ctx, apiPath, payload, &result); err != nil {
		diags.AddError("Unable to update ACL rule ordering", err.Error())
		return aclRuleOrderingResourceModel{}, diags
	}

	return aclRuleOrderingResourceModel{
		ID:                types.StringValue(siteID),
		SiteID:            types.StringValue(siteID),
		OrderedACLRuleIDs: uuidListToStringList(result.OrderedAclRuleIds),
	}, diags
}

func listToUUIDs(list types.List, attrPath path.Path) ([]openapi_types.UUID, diag.Diagnostics) {
	var diags diag.Diagnostics

	var values []string
	diags.Append(list.ElementsAs(context.Background(), &values, false)...)
	if diags.HasError() {
		return nil, diags
	}

	uuids := make([]openapi_types.UUID, 0, len(values))
	for idx, value := range values {
		parsed, err := uuid.Parse(value)
		if err != nil {
			diags.AddAttributeError(
				attrPath.AtListIndex(idx),
				"Invalid UUID",
				fmt.Sprintf("Value %q is not a valid UUID: %s", value, err),
			)
			return nil, diags
		}
		uuids = append(uuids, parsed)
	}
	return uuids, diags
}

func uuidListToStringList(values []openapi_types.UUID) types.List {
	items := make([]string, 0, len(values))
	for _, value := range values {
		items = append(items, value.String())
	}
	list, _ := types.ListValueFrom(context.Background(), types.StringType, items)
	return list
}
