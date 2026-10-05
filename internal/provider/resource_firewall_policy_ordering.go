package provider

import (
	"context"
	"fmt"
	"net/url"

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

type firewallPolicyOrderingResource struct {
	client *unifi.Client
	siteID string
}

type firewallPolicyOrderingResourceModel struct {
	ID                        types.String `tfsdk:"id"`
	SiteID                    types.String `tfsdk:"site_id"`
	SourceFirewallZoneID      types.String `tfsdk:"source_firewall_zone_id"`
	DestinationFirewallZoneID types.String `tfsdk:"destination_firewall_zone_id"`
	BeforeSystemDefined       types.List   `tfsdk:"before_system_defined"`
	AfterSystemDefined        types.List   `tfsdk:"after_system_defined"`
}

func NewFirewallPolicyOrderingResource() resource.Resource {
	return &firewallPolicyOrderingResource{}
}

func (r *firewallPolicyOrderingResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_policy_ordering"
}

func (r *firewallPolicyOrderingResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"source_firewall_zone_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"destination_firewall_zone_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"before_system_defined": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
			},
			"after_system_defined": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *firewallPolicyOrderingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *firewallPolicyOrderingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan firewallPolicyOrderingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(plan.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	state, diags := r.applyOrdering(ctx, siteID, plan.SourceFirewallZoneID, plan.DestinationFirewallZoneID, plan.BeforeSystemDefined, plan.AfterSystemDefined)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *firewallPolicyOrderingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state firewallPolicyOrderingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(state.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	_, sourceZoneID, parseDiags := parseUUIDString(state.SourceFirewallZoneID, path.Root("source_firewall_zone_id"))
	resp.Diagnostics.Append(parseDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, destinationZoneID, parseDiags := parseUUIDString(state.DestinationFirewallZoneID, path.Root("destination_firewall_zone_id"))
	resp.Diagnostics.Append(parseDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiPath := fmt.Sprintf(
		"/v1/sites/%s/firewall/policies/ordering?sourceFirewallZoneId=%s&destinationFirewallZoneId=%s",
		siteID,
		url.QueryEscape(sourceZoneID.String()),
		url.QueryEscape(destinationZoneID.String()),
	)

	var result networkapi.IntegrationFirewallPolicyOrderingDto
	if err := r.client.Get(ctx, apiPath, &result); err != nil {
		resp.Diagnostics.AddError("Unable to read firewall policy ordering", err.Error())
		return
	}

	state.SiteID = types.StringValue(siteID)
	state.ID = types.StringValue(firewallPolicyOrderingID(siteID, sourceZoneID.String(), destinationZoneID.String()))
	state.BeforeSystemDefined = uuidListToStringList(result.OrderedFirewallPolicyIds.BeforeSystemDefined)
	state.AfterSystemDefined = uuidListToStringList(result.OrderedFirewallPolicyIds.AfterSystemDefined)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *firewallPolicyOrderingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan firewallPolicyOrderingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(plan.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	state, diags := r.applyOrdering(ctx, siteID, plan.SourceFirewallZoneID, plan.DestinationFirewallZoneID, plan.BeforeSystemDefined, plan.AfterSystemDefined)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *firewallPolicyOrderingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *firewallPolicyOrderingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	siteID, sourceZoneID, destinationZoneID, err := splitImportID3(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import identifier", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site_id"), siteID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("source_firewall_zone_id"), sourceZoneID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("destination_firewall_zone_id"), destinationZoneID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), firewallPolicyOrderingID(siteID, sourceZoneID, destinationZoneID))...)
}

func (r *firewallPolicyOrderingResource) applyOrdering(ctx context.Context, siteID string, sourceAttr, destinationAttr types.String, before, after types.List) (firewallPolicyOrderingResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	sourceString, sourceUUID, parseDiags := parseUUIDString(sourceAttr, path.Root("source_firewall_zone_id"))
	diags.Append(parseDiags...)
	destinationString, destinationUUID, parseDiags := parseUUIDString(destinationAttr, path.Root("destination_firewall_zone_id"))
	diags.Append(parseDiags...)
	if diags.HasError() {
		return firewallPolicyOrderingResourceModel{}, diags
	}

	beforeIDs, parseDiags := listToUUIDs(before, path.Root("before_system_defined"))
	diags.Append(parseDiags...)
	afterIDs, parseDiags := listToUUIDs(after, path.Root("after_system_defined"))
	diags.Append(parseDiags...)
	if diags.HasError() {
		return firewallPolicyOrderingResourceModel{}, diags
	}

	payload := networkapi.IntegrationFirewallPolicyOrderingDto{
		OrderedFirewallPolicyIds: networkapi.OrderedFirewallPolicyIDs{
			BeforeSystemDefined: beforeIDs,
			AfterSystemDefined:  afterIDs,
		},
	}

	apiPath := fmt.Sprintf(
		"/v1/sites/%s/firewall/policies/ordering?sourceFirewallZoneId=%s&destinationFirewallZoneId=%s",
		siteID,
		url.QueryEscape(sourceUUID.String()),
		url.QueryEscape(destinationUUID.String()),
	)

	var result networkapi.IntegrationFirewallPolicyOrderingDto
	if err := r.client.Put(ctx, apiPath, payload, &result); err != nil {
		diags.AddError("Unable to update firewall policy ordering", err.Error())
		return firewallPolicyOrderingResourceModel{}, diags
	}

	return firewallPolicyOrderingResourceModel{
		ID:                        types.StringValue(firewallPolicyOrderingID(siteID, sourceString, destinationString)),
		SiteID:                    types.StringValue(siteID),
		SourceFirewallZoneID:      types.StringValue(sourceString),
		DestinationFirewallZoneID: types.StringValue(destinationString),
		BeforeSystemDefined:       uuidListToStringList(result.OrderedFirewallPolicyIds.BeforeSystemDefined),
		AfterSystemDefined:        uuidListToStringList(result.OrderedFirewallPolicyIds.AfterSystemDefined),
	}, diags
}

func parseUUIDString(value types.String, attrPath path.Path) (string, openapi_types.UUID, diag.Diagnostics) {
	var diags diag.Diagnostics
	if value.IsNull() || value.IsUnknown() || value.ValueString() == "" {
		diags.AddAttributeError(attrPath, "Missing UUID", "A UUID value is required.")
		return "", uuid.Nil, diags
	}

	parsed, err := uuid.Parse(value.ValueString())
	if err != nil {
		diags.AddAttributeError(attrPath, "Invalid UUID", fmt.Sprintf("Value %q is not a valid UUID: %s", value.ValueString(), err))
		return "", uuid.Nil, diags
	}

	return parsed.String(), parsed, diags
}

func firewallPolicyOrderingID(siteID, sourceZoneID, destinationZoneID string) string {
	return fmt.Sprintf("%s/%s/%s", siteID, sourceZoneID, destinationZoneID)
}
