package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
	"github.com/ilmax/unifi-client-go/pkg/errors"
	"github.com/ilmax/unifi-client-go/pkg/zones"
)

type firewallZoneResource struct {
	client *unifi.Client
	siteID string
}

type firewallZoneResourceModel struct {
	ID         types.String `tfsdk:"id"`
	SiteID     types.String `tfsdk:"site_id"`
	Name       types.String `tfsdk:"name"`
	NetworkIDs types.List   `tfsdk:"network_ids"`
}

func NewFirewallZoneResource() resource.Resource {
	return &firewallZoneResource{}
}

func (r *firewallZoneResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_zone"
}

func (r *firewallZoneResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"network_ids": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *firewallZoneResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *firewallZoneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan firewallZoneResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(plan.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	reqBody, diags := buildFirewallZoneCreateRequest(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("/v1/sites/%s/firewall/zones", siteID)
	var result zones.CreateCustomFirewallZoneResponse
	if err := r.client.Post(ctx, path, reqBody, &result); err != nil {
		resp.Diagnostics.AddError("Unable to create firewall zone", err.Error())
		return
	}

	state := firewallZoneResourceModel{
		ID:         types.StringValue(result.Id),
		SiteID:     types.StringValue(siteID),
		Name:       types.StringValue(result.Name),
		NetworkIDs: rawMessagesToStringList(result.NetworkIds),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *firewallZoneResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state firewallZoneResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(state.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	if state.ID.IsNull() || state.ID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("id"), "Missing firewall zone ID", "The firewall zone ID is required to read the resource.")
		return
	}

	path := fmt.Sprintf("/v1/sites/%s/firewall/zones/%s", siteID, state.ID.ValueString())
	var result zones.GetFirewallZoneResponse
	if err := r.client.Get(ctx, path, &result); err != nil {
		if errors.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read firewall zone", err.Error())
		return
	}

	state.SiteID = types.StringValue(siteID)
	state.Name = types.StringValue(result.Name)
	state.NetworkIDs = rawMessagesToStringList(result.NetworkIds)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *firewallZoneResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan firewallZoneResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(plan.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	if plan.ID.IsNull() || plan.ID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("id"), "Missing firewall zone ID", "The firewall zone ID is required to update the resource.")
		return
	}

	reqBody, diags := buildFirewallZoneUpdateRequest(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("/v1/sites/%s/firewall/zones/%s", siteID, plan.ID.ValueString())
	var result zones.UpdateFirewallZoneResponse
	if err := r.client.Put(ctx, path, reqBody, &result); err != nil {
		resp.Diagnostics.AddError("Unable to update firewall zone", err.Error())
		return
	}

	state := firewallZoneResourceModel{
		ID:         types.StringValue(result.Id),
		SiteID:     types.StringValue(siteID),
		Name:       types.StringValue(result.Name),
		NetworkIDs: rawMessagesToStringList(result.NetworkIds),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *firewallZoneResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state firewallZoneResourceModel
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

	path := fmt.Sprintf("/v1/sites/%s/firewall/zones/%s", siteID, state.ID.ValueString())
	if err := r.client.Delete(ctx, path, nil); err != nil && !errors.IsNotFoundError(err) {
		resp.Diagnostics.AddError("Unable to delete firewall zone", err.Error())
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *firewallZoneResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	siteID, zoneID, err := splitImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import identifier", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site_id"), siteID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), zoneID)...)
}

func buildFirewallZoneCreateRequest(plan firewallZoneResourceModel) (*zones.CreateCustomFirewallZoneRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	ids, listDiags := listToRawMessages(plan.NetworkIDs, path.Root("network_ids"))
	diags.Append(listDiags...)
	if diags.HasError() {
		return nil, diags
	}

	return &zones.CreateCustomFirewallZoneRequest{
		Name:       plan.Name.ValueString(),
		NetworkIds: ids,
	}, diags
}

func buildFirewallZoneUpdateRequest(plan firewallZoneResourceModel) (*zones.UpdateFirewallZoneRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	ids, listDiags := listToRawMessages(plan.NetworkIDs, path.Root("network_ids"))
	diags.Append(listDiags...)
	if diags.HasError() {
		return nil, diags
	}

	return &zones.UpdateFirewallZoneRequest{
		Name:       plan.Name.ValueString(),
		NetworkIds: ids,
	}, diags
}
