package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/unifi-client-go/pkg/broadcasts"
	"github.com/ilmax/unifi-client-go/pkg/errors"
	"github.com/massimilianodonini/terraform-provider-unifi/internal/unifi"
)

type wifiResource struct {
	client *unifi.Client
	siteID string
}

type wifiResourceModel struct {
	ID           types.String `tfsdk:"id"`
	SiteID       types.String `tfsdk:"site_id"`
	Name         types.String `tfsdk:"name"`
	Type         types.String `tfsdk:"type"`
	Enabled      types.Bool   `tfsdk:"enabled"`
	SecurityType types.String `tfsdk:"security_type"`
	NetworkType  types.String `tfsdk:"network_type"`
}

func NewWifiResource() resource.Resource {
	return &wifiResource{}
}

func (r *wifiResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wifi"
}

func (r *wifiResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"security_type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"network_type": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *wifiResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *wifiResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan wifiResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(plan.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	request := &broadcasts.CreateWifiBroadcastRequest{
		Type:    plan.Type.ValueString(),
		Name:    plan.Name.ValueString(),
		Enabled: plan.Enabled.ValueBool(),
		SecurityConfiguration: &broadcasts.CreateWifiBroadcastSecurityConfiguration{
			Type: plan.SecurityType.ValueString(),
		},
	}

	if !plan.NetworkType.IsNull() && !plan.NetworkType.IsUnknown() {
		request.Network = &broadcasts.ClientAccess{
			Type: plan.NetworkType.ValueString(),
		}
	}

	var result broadcasts.CreateWifiBroadcastResponse
	path := fmt.Sprintf("/v1/sites/%s/wifi/broadcasts", siteID)
	if err := r.client.Post(ctx, path, request, &result); err != nil {
		resp.Diagnostics.AddError("Unable to create WiFi broadcast", err.Error())
		return
	}

	state := wifiResourceModel{
		ID:           types.StringValue(result.Id),
		SiteID:       types.StringValue(siteID),
		Name:         types.StringValue(result.Name),
		Type:         types.StringValue(result.Type),
		Enabled:      types.BoolValue(result.Enabled),
		SecurityType: types.StringValue(result.SecurityConfiguration.Type),
	}
	if result.Network != nil {
		state.NetworkType = types.StringValue(result.Network.Type)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *wifiResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state wifiResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(state.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	if state.ID.IsNull() || state.ID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("id"), "Missing WiFi broadcast ID", "The WiFi broadcast ID is required to read the resource.")
		return
	}

	var result broadcasts.GetWifiBroadcastDetailsResponse
	path := fmt.Sprintf("/v1/sites/%s/wifi/broadcasts/%s", siteID, state.ID.ValueString())
	if err := r.client.Get(ctx, path, &result); err != nil {
		if errors.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read WiFi broadcast", err.Error())
		return
	}

	state.SiteID = types.StringValue(siteID)
	state.Name = types.StringValue(result.Name)
	state.Type = types.StringValue(result.Type)
	state.Enabled = types.BoolValue(result.Enabled)
	if result.SecurityConfiguration != nil {
		state.SecurityType = types.StringValue(result.SecurityConfiguration.Type)
	}
	if result.Network != nil {
		state.NetworkType = types.StringValue(result.Network.Type)
	} else {
		state.NetworkType = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *wifiResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("WiFi updates are not supported", "Update operations are not supported for unifi_wifi resources. Configure changes will force replacement.")
}

func (r *wifiResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state wifiResourceModel
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

	path := fmt.Sprintf("/v1/sites/%s/wifi/broadcasts/%s", siteID, state.ID.ValueString())
	if err := r.client.Delete(ctx, path, nil); err != nil && !errors.IsNotFoundError(err) {
		resp.Diagnostics.AddError("Unable to delete WiFi broadcast", err.Error())
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *wifiResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	siteID, broadcastID, err := splitImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import identifier", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site_id"), siteID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), broadcastID)...)
}
