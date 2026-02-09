package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
	"github.com/ilmax/unifi-client-go/pkg/clients"
	"github.com/ilmax/unifi-client-go/pkg/errors"
)

type clientDataSource struct {
	client *unifi.Client
	siteID string
}

type clientDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	SiteID         types.String `tfsdk:"site_id"`
	ClientID       types.String `tfsdk:"client_id"`
	Name           types.String `tfsdk:"name"`
	Type           types.String `tfsdk:"type"`
	ConnectedAt    types.String `tfsdk:"connected_at"`
	IPAddress      types.String `tfsdk:"ip_address"`
	AccessType     types.String `tfsdk:"access_type"`
	MacAddress     types.String `tfsdk:"mac_address"`
	UplinkDeviceID types.String `tfsdk:"uplink_device_id"`
}

func NewClientDataSource() datasource.DataSource {
	return &clientDataSource{}
}

func (d *clientDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_client"
}

func (d *clientDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"site_id": schema.StringAttribute{
				Optional: true,
			},
			"client_id": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"type": schema.StringAttribute{
				Computed: true,
			},
			"connected_at": schema.StringAttribute{
				Computed: true,
			},
			"ip_address": schema.StringAttribute{
				Computed: true,
			},
			"access_type": schema.StringAttribute{
				Computed: true,
			},
			"mac_address": schema.StringAttribute{
				Computed: true,
			},
			"uplink_device_id": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *clientDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *clientDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config clientDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(config.SiteID, d.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	clientID := config.ClientID.ValueString()
	if clientID == "" {
		resp.Diagnostics.AddAttributeError(path.Root("client_id"), "Missing client_id", "client_id must be provided.")
		return
	}

	state, found, err := d.readClient(ctx, siteID, clientID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read client", err.Error())
		return
	}
	if !found {
		resp.Diagnostics.AddError("Client not found", fmt.Sprintf("Client %q was not found in site %q.", clientID, siteID))
		return
	}

	state.SiteID = types.StringValue(siteID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *clientDataSource) readClient(ctx context.Context, siteID, clientID string) (clientDataSourceModel, bool, error) {
	path := fmt.Sprintf("/v1/sites/%s/clients/%s", siteID, clientID)
	var raw json.RawMessage
	if err := d.client.Get(ctx, path, &raw); err != nil {
		if errors.IsNotFoundError(err) {
			return clientDataSourceModel{}, false, nil
		}
		return clientDataSourceModel{}, false, err
	}

	decoded, err := clients.DecodeGetConnectedClientDetailsResponse(raw)
	if err != nil {
		return clientDataSourceModel{}, false, err
	}

	state := clientDataSourceModel{
		ID:       types.StringValue(clientID),
		ClientID: types.StringValue(clientID),
	}

	switch result := decoded.(type) {
	case *clients.GetConnectedClientDetailsResponseTeleport:
		state.Name = types.StringValue(result.Name)
		state.Type = types.StringValue("TELEPORT")
		state.ConnectedAt = timeToString(result.ConnectedAt)
		state.IPAddress = stringValueOrNull(result.IpAddress)
		state.AccessType = clientAccessType(result.Access)
	case *clients.GetConnectedClientDetailsResponseVpn:
		state.Name = types.StringValue(result.Name)
		state.Type = types.StringValue(result.Type)
		state.ConnectedAt = timeToString(result.ConnectedAt)
		state.IPAddress = stringValueOrNull(result.IpAddress)
		state.AccessType = clientAccessType(result.Access)
	case *clients.GetConnectedClientDetailsResponseWired:
		state.Name = types.StringValue(result.Name)
		state.Type = types.StringValue(result.Type)
		state.ConnectedAt = timeToString(result.ConnectedAt)
		state.IPAddress = stringValueOrNull(result.IpAddress)
		state.AccessType = clientAccessType(result.Access)
		state.MacAddress = stringValueOrNull(result.MacAddress)
		state.UplinkDeviceID = stringValueOrNull(result.UplinkDeviceId)
	case *clients.GetConnectedClientDetailsResponseWireless:
		state.Name = types.StringValue(result.Name)
		state.Type = types.StringValue(result.Type)
		state.ConnectedAt = timeToString(result.ConnectedAt)
		state.IPAddress = stringValueOrNull(result.IpAddress)
		state.AccessType = clientAccessType(result.Access)
		state.MacAddress = stringValueOrNull(result.MacAddress)
		state.UplinkDeviceID = stringValueOrNull(result.UplinkDeviceId)
	default:
		return clientDataSourceModel{}, false, fmt.Errorf("unsupported client response type %T", decoded)
	}

	return state, true, nil
}

func timeToString(value *time.Time) types.String {
	if value == nil {
		return types.StringNull()
	}
	return types.StringValue(value.UTC().Format(time.RFC3339))
}

func clientAccessType(access *clients.ClientAccess) types.String {
	if access == nil {
		return types.StringNull()
	}
	return stringValueOrNull(access.Type)
}
