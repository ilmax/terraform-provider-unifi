package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
	"github.com/ilmax/unifi-client-go/pkg/devices"
)

type executeAdoptedDeviceAction struct {
	client *unifi.Client
	siteID string
}

type executeAdoptedDeviceActionModel struct {
	SiteID         types.String `tfsdk:"site_id"`
	DeviceID       types.String `tfsdk:"device_id"`
	Action         types.String `tfsdk:"action"`
	TimeoutSeconds types.Int64  `tfsdk:"timeout_seconds"`
}

func NewExecuteAdoptedDeviceAction() action.Action {
	return &executeAdoptedDeviceAction{}
}

func (a *executeAdoptedDeviceAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_execute_adopted_device_action"
}

func (a *executeAdoptedDeviceAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Executes an action on an adopted device.",
		Attributes: map[string]schema.Attribute{
			"site_id": schema.StringAttribute{
				Optional:    true,
				Description: "Site identifier (defaults to provider site_id).",
			},
			"device_id": schema.StringAttribute{
				Required:    true,
				Description: "Adopted device identifier.",
			},
			"action": schema.StringAttribute{
				Required:    true,
				Description: "Device action to execute (supported: RESTART).",
			},
			"timeout_seconds": schema.Int64Attribute{
				Optional:    true,
				Description: "Timeout in seconds for the action request (default: 1800).",
			},
		},
	}
}

func (a *executeAdoptedDeviceAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*providerData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", "Expected provider data to be of type *providerData.")
		return
	}
	a.client = data.client
	a.siteID = data.siteID
}

func (a *executeAdoptedDeviceAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config executeAdoptedDeviceActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(config.SiteID, a.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	deviceID := config.DeviceID.ValueString()
	if deviceID == "" {
		resp.Diagnostics.AddAttributeError(path.Root("device_id"), "Missing device_id", "device_id must be provided.")
		return
	}

	actionValue := config.Action.ValueString()
	if actionValue == "" {
		resp.Diagnostics.AddAttributeError(path.Root("action"), "Missing action", "action must be provided.")
		return
	}
	if actionValue != string(devices.ExecuteAdoptedDeviceActionActionRestart) {
		resp.Diagnostics.AddAttributeError(path.Root("action"), "Unsupported action", "Supported values: RESTART.")
		return
	}

	timeoutSeconds := int64(1800)
	if !config.TimeoutSeconds.IsNull() && !config.TimeoutSeconds.IsUnknown() {
		timeoutSeconds = config.TimeoutSeconds.ValueInt64()
	}
	if timeoutSeconds <= 0 {
		resp.Diagnostics.AddAttributeError(path.Root("timeout_seconds"), "Invalid timeout_seconds", "timeout_seconds must be greater than 0.")
		return
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	if resp.SendProgress != nil {
		resp.SendProgress(action.InvokeProgressEvent{Message: "Executing device action..."})
	}

	path := fmt.Sprintf("/v1/sites/%s/devices/%s/actions", siteID, deviceID)
	payload := &devices.ExecuteAdoptedDeviceActionRequest{Action: actionValue}
	if err := a.client.Post(ctx, path, payload, nil); err != nil {
		resp.Diagnostics.AddError("Unable to execute adopted device action", err.Error())
		return
	}

	if resp.SendProgress != nil {
		resp.SendProgress(action.InvokeProgressEvent{Message: "Device action completed."})
	}
}
