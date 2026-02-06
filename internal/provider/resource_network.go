package provider

import (
	"context"
	"fmt"

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
	ID         types.String `tfsdk:"id"`
	SiteID     types.String `tfsdk:"site_id"`
	Name       types.String `tfsdk:"name"`
	Management types.String `tfsdk:"management"`
	Enabled    types.Bool   `tfsdk:"enabled"`
	VlanID     types.Int64  `tfsdk:"vlan_id"`
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

	request := &networks.CreateNetworkRequest{
		Management: plan.Management.ValueString(),
		Name:       plan.Name.ValueString(),
		Enabled:    plan.Enabled.ValueBool(),
		VlanId:     plan.VlanID.ValueInt64(),
	}

	var result networks.CreateNetworkResponse
	path := fmt.Sprintf("/v1/sites/%s/networks", siteID)
	if err := r.client.Post(ctx, path, request, &result); err != nil {
		resp.Diagnostics.AddError("Unable to create network", err.Error())
		return
	}

	state := networkResourceModel{
		ID:         types.StringValue(result.Id),
		SiteID:     types.StringValue(siteID),
		Name:       types.StringValue(result.Name),
		Management: types.StringValue(result.Management),
		Enabled:    types.BoolValue(result.Enabled),
		VlanID:     types.Int64Value(result.VlanId),
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

	networkID := state.ID.ValueString()
	var result networks.GetNetworkDetailsResponse
	path := fmt.Sprintf("/v1/sites/%s/networks/%s", siteID, networkID)
	if err := r.client.Get(ctx, path, &result); err != nil {
		if errors.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read network", err.Error())
		return
	}

	state.SiteID = types.StringValue(siteID)
	state.Name = types.StringValue(result.Name)
	state.Management = types.StringValue(result.Management)
	state.Enabled = types.BoolValue(result.Enabled)
	state.VlanID = types.Int64Value(result.VlanId)

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

	request := &networks.UpdateNetworkRequest{
		Management: plan.Management.ValueString(),
		Name:       plan.Name.ValueString(),
		Enabled:    plan.Enabled.ValueBool(),
		VlanId:     plan.VlanID.ValueInt64(),
	}

	var result networks.UpdateNetworkResponse
	path := fmt.Sprintf("/v1/sites/%s/networks/%s", siteID, plan.ID.ValueString())
	if err := r.client.Put(ctx, path, request, &result); err != nil {
		resp.Diagnostics.AddError("Unable to update network", err.Error())
		return
	}

	state := networkResourceModel{
		ID:         types.StringValue(result.Id),
		SiteID:     types.StringValue(siteID),
		Name:       types.StringValue(result.Name),
		Management: types.StringValue(result.Management),
		Enabled:    types.BoolValue(result.Enabled),
		VlanID:     types.Int64Value(result.VlanId),
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
