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
	"github.com/ilmax/unifi-client-go/pkg/wans"
)

type wanResource struct {
	client *unifi.Client
	siteID string
}

type wanResourceModel struct {
	ID     types.String `tfsdk:"id"`
	SiteID types.String `tfsdk:"site_id"`
	WanID  types.String `tfsdk:"wan_id"`
	Name   types.String `tfsdk:"name"`
}

func NewWanResource() resource.Resource {
	return &wanResource{}
}

func (r *wanResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wan"
}

func (r *wanResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"wan_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *wanResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *wanResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan wanResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(plan.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	wanID := plan.WanID.ValueString()
	if wanID == "" {
		resp.Diagnostics.AddAttributeError(path.Root("wan_id"), "Missing wan_id", "wan_id must be provided.")
		return
	}

	state, found, diags := r.readWan(ctx, siteID, wanID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("WAN interface not found", "The specified WAN interface was not found.")
		return
	}

	state.SiteID = types.StringValue(siteID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *wanResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state wanResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(state.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	wanID := state.WanID.ValueString()
	if wanID == "" {
		resp.Diagnostics.AddAttributeError(path.Root("wan_id"), "Missing wan_id", "wan_id must be provided.")
		return
	}

	newState, found, diags := r.readWan(ctx, siteID, wanID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	newState.SiteID = types.StringValue(siteID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *wanResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("WAN updates are not supported", "Update operations are not supported for unifi_wan resources.")
}

func (r *wanResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *wanResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	siteID, wanID, err := splitImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import identifier", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site_id"), siteID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("wan_id"), wanID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), wanID)...)
}

func (r *wanResource) readWan(ctx context.Context, siteID, wanID string) (wanResourceModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	var result wans.ListWANInterfacesResponse
	path := fmt.Sprintf("/v1/sites/%s/wans", siteID)
	if err := r.client.Get(ctx, path, &result); err != nil {
		if errors.IsNotFoundError(err) {
			return wanResourceModel{}, false, diags
		}
		diags.AddError("Unable to list WAN interfaces", err.Error())
		return wanResourceModel{}, false, diags
	}

	for _, item := range result.Data {
		if item.Id == wanID {
			state := wanResourceModel{
				ID:    types.StringValue(item.Id),
				WanID: types.StringValue(item.Id),
				Name:  types.StringValue(item.Name),
			}
			return state, true, diags
		}
	}

	return wanResourceModel{}, false, diags
}
