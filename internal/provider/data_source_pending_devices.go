package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/sdkcompat/pendingdevices"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
)

// pendingDevicesDataSource exposes /v1/pending-devices, a GET-only,
// non-site-scoped endpoint listing devices that have sent an inform but
// haven't been adopted into any site yet.
type pendingDevicesDataSource struct {
	client *unifi.Client
}

type pendingDevicesDataSourceModel struct {
	ID      types.String                       `tfsdk:"id"`
	Devices []pendingDeviceDataSourceItemModel `tfsdk:"devices"`
}

type pendingDeviceDataSourceItemModel struct {
	MacAddress            types.String `tfsdk:"mac_address"`
	IPAddress             types.String `tfsdk:"ip_address"`
	Model                 types.String `tfsdk:"model"`
	State                 types.String `tfsdk:"state"`
	FirmwareVersion       types.String `tfsdk:"firmware_version"`
	FirmwareUpdatable     types.Bool   `tfsdk:"firmware_updatable"`
	Supported             types.Bool   `tfsdk:"supported"`
	AdoptionTargetSiteIDs types.List   `tfsdk:"adoption_target_site_ids"`
	Features              types.List   `tfsdk:"features"`
}

func NewPendingDevicesDataSource() datasource.DataSource {
	return &pendingDevicesDataSource{}
}

func (d *pendingDevicesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pending_devices"
}

func (d *pendingDevicesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "All devices that have sent an inform to the controller but haven't been adopted into any " +
			"site yet. This is global (not scoped to a site) and GET-only in the real API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"devices": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
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
						"adoption_target_site_ids": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
						},
						"features": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

func (d *pendingDevicesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*providerData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", "Expected provider data to be of type *providerData.")
		return
	}
	d.client = data.client
}

func (d *pendingDevicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var result pendingdevices.ListPendingDevicesResponse
	if err := d.client.Get(ctx, "/v1/pending-devices", &result); err != nil {
		resp.Diagnostics.AddError("Unable to list pending devices", err.Error())
		return
	}

	state := pendingDevicesDataSourceModel{
		ID:      types.StringValue("pending-devices"),
		Devices: make([]pendingDeviceDataSourceItemModel, 0, len(result.Data)),
	}
	for _, item := range result.Data {
		state.Devices = append(state.Devices, pendingDeviceDataSourceItemModel{
			MacAddress:            types.StringValue(item.MacAddress),
			IPAddress:             types.StringValue(item.IpAddress),
			Model:                 types.StringValue(item.Model),
			State:                 types.StringValue(item.State),
			FirmwareVersion:       stringValueOrNull(item.FirmwareVersion),
			FirmwareUpdatable:     types.BoolValue(item.FirmwareUpdatable),
			Supported:             types.BoolValue(item.Supported),
			AdoptionTargetSiteIDs: stringsToList(item.AdoptionTargetSiteIds),
			Features:              stringsToList(item.Features),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
