package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/sdkcompat/devices"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
	"github.com/ilmax/unifi-client-go/pkg/errors"
)

type deviceDataSource struct {
	client *unifi.Client
	siteID string
}

type deviceDataSourceModel struct {
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

func NewDeviceDataSource() datasource.DataSource {
	return &deviceDataSource{}
}

func (d *deviceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device"
}

func (d *deviceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"site_id": schema.StringAttribute{
				Optional: true,
			},
			"device_id": schema.StringAttribute{
				Required: true,
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

func (d *deviceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*providerData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", "Expected provider data to be of type *providerData.")
		return
	}
	d.client = data.client
	d.siteID = data.siteID
}

func (d *deviceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config deviceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(config.SiteID, d.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	deviceID := config.DeviceID.ValueString()
	if deviceID == "" {
		resp.Diagnostics.AddAttributeError(path.Root("device_id"), "Missing device_id", "device_id must be provided.")
		return
	}

	state, found, err := d.readDevice(ctx, siteID, deviceID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read device", err.Error())
		return
	}
	if !found {
		resp.Diagnostics.AddError("Device not found", fmt.Sprintf("Device %q was not found in site %q.", deviceID, siteID))
		return
	}

	state.SiteID = types.StringValue(siteID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *deviceDataSource) readDevice(ctx context.Context, siteID, deviceID string) (deviceDataSourceModel, bool, error) {
	var result devices.GetAdoptedDeviceDetailsResponse
	path := fmt.Sprintf("/v1/sites/%s/devices/%s", siteID, deviceID)
	if err := d.client.Get(ctx, path, &result); err != nil {
		if errors.IsNotFoundError(err) {
			return deviceDataSourceModel{}, false, nil
		}
		return deviceDataSourceModel{}, false, err
	}

	state := deviceDataSourceModel{
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
	return state, true, nil
}
