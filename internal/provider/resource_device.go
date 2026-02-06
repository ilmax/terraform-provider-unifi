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
	"github.com/ilmax/unifi-client-go/pkg/devices"
	"github.com/ilmax/unifi-client-go/pkg/errors"
)

type deviceResource struct {
	client *unifi.Client
	siteID string
}

type deviceResourceModel struct {
	ID                types.String `tfsdk:"id"`
	SiteID            types.String `tfsdk:"site_id"`
	DeviceID          types.String `tfsdk:"device_id"`
	Name              types.String `tfsdk:"name"`
	MacAddress        types.String `tfsdk:"mac_address"`
	IPAddress         types.String `tfsdk:"ip_address"`
	Model             types.String `tfsdk:"model"`
	State             types.String `tfsdk:"state"`
	FirmwareVersion   types.String `tfsdk:"firmware_version"`
	FirmwareUpdatable types.Bool   `tfsdk:"firmware_updatable"`
	Supported         types.Bool   `tfsdk:"supported"`
}

func NewDeviceResource() resource.Resource {
	return &deviceResource{}
}

func (r *deviceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device"
}

func (r *deviceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"device_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"mac_address": schema.StringAttribute{
				Computed: true,
			},
			"ip_address": schema.StringAttribute{
				Computed: true,
			},
			"model": schema.StringAttribute{
				Computed: true,
			},
			"state": schema.StringAttribute{
				Computed: true,
			},
			"firmware_version": schema.StringAttribute{
				Computed: true,
			},
			"firmware_updatable": schema.BoolAttribute{
				Computed: true,
			},
			"supported": schema.BoolAttribute{
				Computed: true,
			},
		},
	}
}

func (r *deviceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *deviceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan deviceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(plan.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	deviceID := plan.DeviceID.ValueString()
	if deviceID == "" {
		resp.Diagnostics.AddAttributeError(path.Root("device_id"), "Missing device_id", "device_id must be provided.")
		return
	}

	state, found, diags := r.readDevice(ctx, siteID, deviceID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("Device not found", "The specified device was not found.")
		return
	}

	state.SiteID = types.StringValue(siteID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *deviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state deviceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(state.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	deviceID := state.DeviceID.ValueString()
	if deviceID == "" {
		resp.Diagnostics.AddAttributeError(path.Root("device_id"), "Missing device_id", "device_id must be provided.")
		return
	}

	newState, found, diags := r.readDevice(ctx, siteID, deviceID)
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

func (r *deviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Device updates are not supported", "Update operations are not supported for unifi_device resources.")
}

func (r *deviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *deviceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	siteID, deviceID, err := splitImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import identifier", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site_id"), siteID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("device_id"), deviceID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), deviceID)...)
}

func (r *deviceResource) readDevice(ctx context.Context, siteID, deviceID string) (deviceResourceModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	var result devices.GetAdoptedDeviceDetailsResponse
	path := fmt.Sprintf("/v1/sites/%s/devices/%s", siteID, deviceID)
	if err := r.client.Get(ctx, path, &result); err != nil {
		if errors.IsNotFoundError(err) {
			return deviceResourceModel{}, false, diags
		}
		diags.AddError("Unable to read device", err.Error())
		return deviceResourceModel{}, false, diags
	}

	state := deviceResourceModel{
		ID:                types.StringValue(result.Id),
		DeviceID:          types.StringValue(result.Id),
		Name:              types.StringValue(result.Name),
		MacAddress:        types.StringValue(result.MacAddress),
		IPAddress:         types.StringValue(result.IpAddress),
		Model:             types.StringValue(result.Model),
		State:             types.StringValue(result.State),
		FirmwareVersion:   types.StringValue(result.FirmwareVersion),
		FirmwareUpdatable: types.BoolValue(result.FirmwareUpdatable),
		Supported:         types.BoolValue(result.Supported),
	}
	return state, true, diags
}
