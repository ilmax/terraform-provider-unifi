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
	"github.com/ilmax/unifi-client-go/pkg/ports"
)

type executePortAction struct {
	client *unifi.Client
	siteID string
}

type executePortActionModel struct {
	SiteID         types.String `tfsdk:"site_id"`
	DeviceID       types.String `tfsdk:"device_id"`
	PortIdx        types.Int64  `tfsdk:"port_idx"`
	Action         types.String `tfsdk:"action"`
	TimeoutSeconds types.Int64  `tfsdk:"timeout_seconds"`
}

func NewExecutePortAction() action.Action {
	return &executePortAction{}
}

func (a *executePortAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_execute_port_action"
}

func (a *executePortAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Executes an action on a switch port.",
		Attributes: map[string]schema.Attribute{
			"site_id": schema.StringAttribute{
				Optional:    true,
				Description: "Site identifier (defaults to provider site_id).",
			},
			"device_id": schema.StringAttribute{
				Required:    true,
				Description: "Device identifier that owns the port.",
			},
			"port_idx": schema.Int64Attribute{
				Required:    true,
				Description: "Port index on the device.",
			},
			"action": schema.StringAttribute{
				Required:    true,
				Description: "Port action to execute (supported: POWER_CYCLE).",
			},
			"timeout_seconds": schema.Int64Attribute{
				Optional:    true,
				Description: "Timeout in seconds for the action request (default: 1800).",
			},
		},
	}
}

func (a *executePortAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
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

func (a *executePortAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config executePortActionModel
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

	if config.PortIdx.IsNull() || config.PortIdx.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("port_idx"), "Missing port_idx", "port_idx must be provided.")
		return
	}
	portIdx := config.PortIdx.ValueInt64()
	if portIdx <= 0 {
		resp.Diagnostics.AddAttributeError(path.Root("port_idx"), "Invalid port_idx", "port_idx must be greater than 0.")
		return
	}

	actionValue := config.Action.ValueString()
	if actionValue == "" {
		resp.Diagnostics.AddAttributeError(path.Root("action"), "Missing action", "action must be provided.")
		return
	}
	if actionValue != string(ports.ExecutePortActionActionPowerCycle) {
		resp.Diagnostics.AddAttributeError(path.Root("action"), "Unsupported action", "Supported values: POWER_CYCLE.")
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
		resp.SendProgress(action.InvokeProgressEvent{Message: "Executing port action..."})
	}

	path := fmt.Sprintf("/v1/sites/%s/devices/%s/interfaces/ports/%d/actions", siteID, deviceID, portIdx)
	payload := &ports.ExecutePortActionRequest{Action: actionValue}
	if err := a.client.Post(ctx, path, payload, nil); err != nil {
		resp.Diagnostics.AddError("Unable to execute port action", err.Error())
		return
	}

	if resp.SendProgress != nil {
		resp.SendProgress(action.InvokeProgressEvent{Message: "Port action completed."})
	}
}
