package provider

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
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
	ID                    types.String                   `tfsdk:"id"`
	SiteID                types.String                   `tfsdk:"site_id"`
	Name                  types.String                   `tfsdk:"name"`
	Management            types.String                   `tfsdk:"management"`
	Enabled               types.Bool                     `tfsdk:"enabled"`
	VlanID                types.Int64                    `tfsdk:"vlan_id"`
	ZoneID                types.String                   `tfsdk:"zone_id"`
	DeviceID              types.String                   `tfsdk:"device_id"`
	IsolationEnabled      types.Bool                     `tfsdk:"isolation_enabled"`
	CellularBackupEnabled types.Bool                     `tfsdk:"cellular_backup_enabled"`
	InternetAccessEnabled types.Bool                     `tfsdk:"internet_access_enabled"`
	MDNSForwardingEnabled types.Bool                     `tfsdk:"mdns_forwarding_enabled"`
	DHCPGuarding          *networkDHCPGuardingModel      `tfsdk:"dhcp_guarding"`
	IPv4Configuration     *networkIPv4ConfigurationModel `tfsdk:"ipv4_configuration"`
	IPv6Configuration     *networkIPv6ConfigurationModel `tfsdk:"ipv6_configuration"`
}

type networkDHCPGuardingModel struct {
	TrustedDHCPServerIPAddresses types.List `tfsdk:"trusted_dhcp_server_ip_addresses"`
}

type networkIPv4ConfigurationModel struct {
	AutoScaleEnabled        types.Bool                         `tfsdk:"auto_scale_enabled"`
	CIDR                    types.String                       `tfsdk:"cidr"`
	HostIPAddress           types.String                       `tfsdk:"host_ip_address"`
	PrefixLength            types.Int64                        `tfsdk:"prefix_length"`
	AdditionalHostIPSubnets types.List                         `tfsdk:"additional_host_ip_subnets"`
	DHCPConfiguration       *networkIPv4DHCPConfigurationModel `tfsdk:"dhcp_configuration"`
}

type networkIPv4DHCPConfigurationModel struct {
	Mode                     types.String                `tfsdk:"mode"`
	IPAddressRange           *networkIPAddressRangeModel `tfsdk:"ip_address_range"`
	GatewayIPAddressOverride types.String                `tfsdk:"gateway_ip_address_override"`
	DNSServers               types.List                  `tfsdk:"dns_servers"`
	LeaseTimeSeconds         types.Int64                 `tfsdk:"lease_time_seconds"`
	DomainName               types.String                `tfsdk:"domain_name"`
}

type networkIPv6ConfigurationModel struct {
	InterfaceType                  types.String                             `tfsdk:"interface_type"`
	PrefixDelegationWANInterfaceID types.String                             `tfsdk:"prefix_delegation_wan_interface_id"`
	DNSServers                     types.List                               `tfsdk:"dns_servers"`
	AdditionalHostIPSubnets        types.List                               `tfsdk:"additional_host_ip_subnets"`
	ClientAddressAssignment        *networkIPv6ClientAddressAssignmentModel `tfsdk:"client_address_assignment"`
	RouterAdvertisement            *networkIPv6RouterAdvertisementModel     `tfsdk:"router_advertisement"`
}

type networkIPv6ClientAddressAssignmentModel struct {
	SLAACEnabled      types.Bool                         `tfsdk:"slaac_enabled"`
	DHCPConfiguration *networkIPv6DHCPConfigurationModel `tfsdk:"dhcp_configuration"`
}

type networkIPv6DHCPConfigurationModel struct {
	IPAddressSuffixRange *networkIPAddressRangeModel `tfsdk:"ip_address_suffix_range"`
	LeaseTimeSeconds     types.Int64                 `tfsdk:"lease_time_seconds"`
}

type networkIPv6RouterAdvertisementModel struct {
	Priority types.String `tfsdk:"priority"`
}

type networkIPAddressRangeModel struct {
	Start types.String `tfsdk:"start"`
	Stop  types.String `tfsdk:"stop"`
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
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"zone_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"device_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"isolation_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"cellular_backup_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"internet_access_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"mdns_forwarding_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"dhcp_guarding": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"trusted_dhcp_server_ip_addresses": schema.ListAttribute{
						Optional:    true,
						ElementType: types.StringType,
					},
				},
			},
			"ipv4_configuration": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"auto_scale_enabled": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"cidr": schema.StringAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"host_ip_address": schema.StringAttribute{
						Optional: true,
					},
					"prefix_length": schema.Int64Attribute{
						Optional: true,
					},
					"additional_host_ip_subnets": schema.ListAttribute{
						Optional:    true,
						ElementType: types.StringType,
					},
					"dhcp_configuration": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"mode": schema.StringAttribute{
								Optional: true,
								Computed: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"ip_address_range": schema.SingleNestedAttribute{
								Optional: true,
								Computed: true,
								PlanModifiers: []planmodifier.Object{
									objectplanmodifier.UseStateForUnknown(),
								},
								Attributes: map[string]schema.Attribute{
									"start": schema.StringAttribute{
										Optional: true,
										Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"stop": schema.StringAttribute{
										Optional: true,
										Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
								},
							},
							"gateway_ip_address_override": schema.StringAttribute{
								Optional: true,
								Computed: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"dns_servers": schema.ListAttribute{
								Optional:    true,
								Computed:    true,
								ElementType: types.StringType,
								PlanModifiers: []planmodifier.List{
									listplanmodifier.UseStateForUnknown(),
								},
							},
							"lease_time_seconds": schema.Int64Attribute{
								Optional: true,
								Computed: true,
								PlanModifiers: []planmodifier.Int64{
									int64planmodifier.UseStateForUnknown(),
								},
							},
							"domain_name": schema.StringAttribute{
								Optional: true,
								Computed: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
						},
					},
				},
			},
			"ipv6_configuration": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"interface_type": schema.StringAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"prefix_delegation_wan_interface_id": schema.StringAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"dns_servers": schema.ListAttribute{
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
						PlanModifiers: []planmodifier.List{
							listplanmodifier.UseStateForUnknown(),
						},
					},
					"additional_host_ip_subnets": schema.ListAttribute{
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
						PlanModifiers: []planmodifier.List{
							listplanmodifier.UseStateForUnknown(),
						},
					},
					"client_address_assignment": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"slaac_enabled": schema.BoolAttribute{
								Optional: true,
								Computed: true,
								PlanModifiers: []planmodifier.Bool{
									boolplanmodifier.UseStateForUnknown(),
								},
							},
							"dhcp_configuration": schema.SingleNestedAttribute{
								Optional: true,
								Attributes: map[string]schema.Attribute{
									"ip_address_suffix_range": schema.SingleNestedAttribute{
										Optional: true,
										Computed: true,
										PlanModifiers: []planmodifier.Object{
											objectplanmodifier.UseStateForUnknown(),
										},
										Attributes: map[string]schema.Attribute{
											"start": schema.StringAttribute{
												Optional: true,
												Computed: true,
												PlanModifiers: []planmodifier.String{
													stringplanmodifier.UseStateForUnknown(),
												},
											},
											"stop": schema.StringAttribute{
												Optional: true,
												Computed: true,
												PlanModifiers: []planmodifier.String{
													stringplanmodifier.UseStateForUnknown(),
												},
											},
										},
									},
									"lease_time_seconds": schema.Int64Attribute{
										Optional: true,
										Computed: true,
										PlanModifiers: []planmodifier.Int64{
											int64planmodifier.UseStateForUnknown(),
										},
									},
								},
							},
						},
					},
					"router_advertisement": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"priority": schema.StringAttribute{
								Optional: true,
								Computed: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
						},
					},
				},
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

	payload, diags := buildCreateNetworkRequest(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("/v1/sites/%s/networks", siteID)
	var raw json.RawMessage
	if err := r.client.Post(ctx, path, payload, &raw); err != nil {
		resp.Diagnostics.AddError("Unable to create network", err.Error())
		return
	}

	networkID, idDiags := networkIDFromCreateResponse(raw)
	resp.Diagnostics.Append(idDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := plan
	state.SiteID = types.StringValue(siteID)
	state.ID = types.StringValue(networkID)

	if !r.readNetwork(ctx, siteID, networkID, &state, &resp.Diagnostics) || resp.Diagnostics.HasError() {
		return
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

	found := r.readNetwork(ctx, siteID, state.ID.ValueString(), &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

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

	payload, diags := buildUpdateNetworkRequest(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("/v1/sites/%s/networks/%s", siteID, plan.ID.ValueString())
	if err := r.client.Put(ctx, path, payload, nil); err != nil {
		resp.Diagnostics.AddError("Unable to update network", err.Error())
		return
	}

	state := plan
	state.SiteID = types.StringValue(siteID)

	if !r.readNetwork(ctx, siteID, plan.ID.ValueString(), &state, &resp.Diagnostics) || resp.Diagnostics.HasError() {
		return
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

	state := networkResourceModel{
		SiteID: types.StringValue(siteID),
		ID:     types.StringValue(networkID),
	}

	found := r.readNetwork(ctx, siteID, networkID, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("Network not found", "The network could not be found during import.")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *networkResource) readNetwork(ctx context.Context, siteID, networkID string, state *networkResourceModel, diags *diag.Diagnostics) bool {
	if state == nil {
		diags.AddError("Unable to read network", "state is nil")
		return false
	}

	path := fmt.Sprintf("/v1/sites/%s/networks/%s", siteID, networkID)
	var raw json.RawMessage
	if err := r.client.Get(ctx, path, &raw); err != nil {
		if errors.IsNotFoundError(err) {
			return false
		}
		diags.AddError("Unable to read network", err.Error())
		return false
	}

	decoded, err := networks.DecodeGetNetworkDetailsResponse(raw)
	if err != nil {
		diags.AddError("Unable to decode network", err.Error())
		return false
	}

	diags.Append(networkStateFromResponse(state, siteID, decoded)...)
	return !diags.HasError()
}

func networkIDFromCreateResponse(raw json.RawMessage) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(raw) == 0 {
		diags.AddError("Unable to determine network ID", "Create network response was empty.")
		return "", diags
	}

	decoded, err := networks.DecodeCreateNetworkResponse(raw)
	if err != nil {
		diags.AddError("Unable to decode create network response", err.Error())
		return "", diags
	}

	switch result := decoded.(type) {
	case *networks.CreateNetworkResponseGateway:
		return result.Id, diags
	case *networks.CreateNetworkResponseSwitch:
		return result.Id, diags
	case *networks.CreateNetworkResponseUnmanaged:
		return result.Id, diags
	default:
		diags.AddError("Unable to determine network ID", fmt.Sprintf("unexpected create response type %T", decoded))
		return "", diags
	}
}

func buildCreateNetworkRequest(plan networkResourceModel) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	management := strings.ToUpper(plan.Management.ValueString())

	switch management {
	case "GATEWAY":
		request := &networks.CreateNetworkRequestGateway{
			Management:            management,
			Name:                  plan.Name.ValueString(),
			Enabled:               plan.Enabled.ValueBool(),
			VlanId:                plan.VlanID.ValueInt64(),
			IsolationEnabled:      plan.IsolationEnabled.ValueBool(),
			CellularBackupEnabled: plan.CellularBackupEnabled.ValueBool(),
			ZoneId:                plan.ZoneID.ValueString(),
			InternetAccessEnabled: plan.InternetAccessEnabled.ValueBool(),
			MdnsForwardingEnabled: plan.MDNSForwardingEnabled.ValueBool(),
		}
		request.DhcpGuarding = buildCreateDHCPGuarding(plan.DHCPGuarding, &diags)
		request.Ipv4Configuration = buildCreateIPv4Config(plan.IPv4Configuration, &diags)
		request.Ipv6Configuration = buildCreateIPv6Config(plan.IPv6Configuration, &diags)
		return request, diags
	case "SWITCH":
		request := &networks.CreateNetworkRequestSwitch{
			Management:            management,
			Name:                  plan.Name.ValueString(),
			Enabled:               plan.Enabled.ValueBool(),
			VlanId:                plan.VlanID.ValueInt64(),
			IsolationEnabled:      plan.IsolationEnabled.ValueBool(),
			CellularBackupEnabled: plan.CellularBackupEnabled.ValueBool(),
			DeviceId:              plan.DeviceID.ValueString(),
		}
		request.DhcpGuarding = buildCreateDHCPGuarding(plan.DHCPGuarding, &diags)
		request.Ipv4Configuration = buildCreateIPv4Config(plan.IPv4Configuration, &diags)
		return request, diags
	case "UNMANAGED":
		request := &networks.CreateNetworkRequestUnmanaged{
			Management: management,
			Name:       plan.Name.ValueString(),
			Enabled:    plan.Enabled.ValueBool(),
			VlanId:     plan.VlanID.ValueInt64(),
		}
		request.DhcpGuarding = buildCreateDHCPGuarding(plan.DHCPGuarding, &diags)
		return request, diags
	default:
		diags.AddAttributeError(path.Root("management"), "Invalid management type", "Supported values are GATEWAY, SWITCH, or UNMANAGED.")
		return nil, diags
	}
}

func buildUpdateNetworkRequest(plan networkResourceModel) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	management := strings.ToUpper(plan.Management.ValueString())

	switch management {
	case "GATEWAY":
		request := &networks.UpdateNetworkRequestGateway{
			Management:            management,
			Name:                  plan.Name.ValueString(),
			Enabled:               plan.Enabled.ValueBool(),
			VlanId:                plan.VlanID.ValueInt64(),
			IsolationEnabled:      plan.IsolationEnabled.ValueBool(),
			CellularBackupEnabled: plan.CellularBackupEnabled.ValueBool(),
			ZoneId:                plan.ZoneID.ValueString(),
			InternetAccessEnabled: plan.InternetAccessEnabled.ValueBool(),
			MdnsForwardingEnabled: plan.MDNSForwardingEnabled.ValueBool(),
		}
		request.DhcpGuarding = buildUpdateDHCPGuarding(plan.DHCPGuarding, &diags)
		request.Ipv4Configuration = buildUpdateIPv4Config(plan.IPv4Configuration, &diags)
		request.Ipv6Configuration = buildUpdateIPv6Config(plan.IPv6Configuration, &diags)
		return request, diags
	case "SWITCH":
		request := &networks.UpdateNetworkRequestSwitch{
			Management:            management,
			Name:                  plan.Name.ValueString(),
			Enabled:               plan.Enabled.ValueBool(),
			VlanId:                plan.VlanID.ValueInt64(),
			IsolationEnabled:      plan.IsolationEnabled.ValueBool(),
			CellularBackupEnabled: plan.CellularBackupEnabled.ValueBool(),
			DeviceId:              plan.DeviceID.ValueString(),
		}
		request.DhcpGuarding = buildUpdateDHCPGuarding(plan.DHCPGuarding, &diags)
		request.Ipv4Configuration = buildUpdateIPv4Config(plan.IPv4Configuration, &diags)
		return request, diags
	case "UNMANAGED":
		request := &networks.UpdateNetworkRequestUnmanaged{
			Management: management,
			Name:       plan.Name.ValueString(),
			Enabled:    plan.Enabled.ValueBool(),
			VlanId:     plan.VlanID.ValueInt64(),
		}
		request.DhcpGuarding = buildUpdateDHCPGuarding(plan.DHCPGuarding, &diags)
		return request, diags
	default:
		diags.AddAttributeError(path.Root("management"), "Invalid management type", "Supported values are GATEWAY, SWITCH, or UNMANAGED.")
		return nil, diags
	}
}

func buildCreateDHCPGuarding(model *networkDHCPGuardingModel, diags *diag.Diagnostics) *networks.CreateNetworkDhcpGuarding {
	if model == nil {
		return nil
	}

	trusted, listDiags := listToRawMessages(model.TrustedDHCPServerIPAddresses, path.Root("dhcp_guarding").AtName("trusted_dhcp_server_ip_addresses"))
	diags.Append(listDiags...)
	if diags.HasError() {
		return nil
	}

	return &networks.CreateNetworkDhcpGuarding{TrustedDhcpServerIpAddresses: trusted}
}

func buildUpdateDHCPGuarding(model *networkDHCPGuardingModel, diags *diag.Diagnostics) *networks.UpdateNetworkDhcpGuarding {
	if model == nil {
		return nil
	}

	trusted, listDiags := listToRawMessages(model.TrustedDHCPServerIPAddresses, path.Root("dhcp_guarding").AtName("trusted_dhcp_server_ip_addresses"))
	diags.Append(listDiags...)
	if diags.HasError() {
		return nil
	}

	return &networks.UpdateNetworkDhcpGuarding{TrustedDhcpServerIpAddresses: trusted}
}

func buildCreateIPv4Config(model *networkIPv4ConfigurationModel, diags *diag.Diagnostics) *networks.CreateNetworkIpv4Configuration {
	if model == nil {
		return nil
	}

	additional, listDiags := listToRawMessages(model.AdditionalHostIPSubnets, path.Root("ipv4_configuration").AtName("additional_host_ip_subnets"))
	diags.Append(listDiags...)

	hostIP, prefixLength := resolveIPv4HostPrefix(model, diags)
	if diags.HasError() {
		return nil
	}

	cfg := &networks.CreateNetworkIpv4Configuration{
		AutoScaleEnabled:        model.AutoScaleEnabled.ValueBool(),
		HostIpAddress:           hostIP,
		PrefixLength:            prefixLength,
		AdditionalHostIpSubnets: additional,
	}
	cfg.DhcpConfiguration = buildCreateIPv4DHCPConfig(model.DHCPConfiguration, diags)
	return cfg
}

func buildUpdateIPv4Config(model *networkIPv4ConfigurationModel, diags *diag.Diagnostics) *networks.UpdateNetworkIpv4Configuration {
	if model == nil {
		return nil
	}

	additional, listDiags := listToRawMessages(model.AdditionalHostIPSubnets, path.Root("ipv4_configuration").AtName("additional_host_ip_subnets"))
	diags.Append(listDiags...)

	hostIP, prefixLength := resolveIPv4HostPrefix(model, diags)
	if diags.HasError() {
		return nil
	}

	cfg := &networks.UpdateNetworkIpv4Configuration{
		AutoScaleEnabled:        model.AutoScaleEnabled.ValueBool(),
		HostIpAddress:           hostIP,
		PrefixLength:            prefixLength,
		AdditionalHostIpSubnets: additional,
	}
	cfg.DhcpConfiguration = buildUpdateIPv4DHCPConfig(model.DHCPConfiguration, diags)
	return cfg
}

func buildCreateIPv4DHCPConfig(model *networkIPv4DHCPConfigurationModel, diags *diag.Diagnostics) *networks.CreateNetworkIpv4ConfigurationDhcpConfiguration {
	if model == nil {
		return nil
	}

	dns, dnsDiags := listToRawMessages(model.DNSServers, path.Root("ipv4_configuration").AtName("dhcp_configuration").AtName("dns_servers"))
	diags.Append(dnsDiags...)

	cfg := &networks.CreateNetworkIpv4ConfigurationDhcpConfiguration{
		Mode:                         model.Mode.ValueString(),
		GatewayIpAddressOverride:     model.GatewayIPAddressOverride.ValueString(),
		DnsServerIpAddressesOverride: dns,
		LeaseTimeSeconds:             model.LeaseTimeSeconds.ValueInt64(),
		DomainName:                   model.DomainName.ValueString(),
	}

	if model.IPAddressRange != nil {
		cfg.IpAddressRange = &networks.CreateNetworkIpv4ConfigurationDhcpConfigurationIpAddressRange{
			Start: model.IPAddressRange.Start.ValueString(),
			Stop:  model.IPAddressRange.Stop.ValueString(),
		}
	}

	return cfg
}

func buildUpdateIPv4DHCPConfig(model *networkIPv4DHCPConfigurationModel, diags *diag.Diagnostics) *networks.UpdateNetworkIpv4ConfigurationDhcpConfiguration {
	if model == nil {
		return nil
	}

	dns, dnsDiags := listToRawMessages(model.DNSServers, path.Root("ipv4_configuration").AtName("dhcp_configuration").AtName("dns_servers"))
	diags.Append(dnsDiags...)

	cfg := &networks.UpdateNetworkIpv4ConfigurationDhcpConfiguration{
		Mode:                         model.Mode.ValueString(),
		GatewayIpAddressOverride:     model.GatewayIPAddressOverride.ValueString(),
		DnsServerIpAddressesOverride: dns,
		LeaseTimeSeconds:             model.LeaseTimeSeconds.ValueInt64(),
		DomainName:                   model.DomainName.ValueString(),
	}

	if model.IPAddressRange != nil {
		cfg.IpAddressRange = &networks.UpdateNetworkIpv4ConfigurationDhcpConfigurationIpAddressRange{
			Start: model.IPAddressRange.Start.ValueString(),
			Stop:  model.IPAddressRange.Stop.ValueString(),
		}
	}

	return cfg
}

func buildCreateIPv6Config(model *networkIPv6ConfigurationModel, diags *diag.Diagnostics) *networks.CreateNetworkIpv6Configuration {
	if model == nil {
		return nil
	}

	dns, dnsDiags := listToRawMessages(model.DNSServers, path.Root("ipv6_configuration").AtName("dns_servers"))
	diags.Append(dnsDiags...)
	additional, listDiags := listToRawMessages(model.AdditionalHostIPSubnets, path.Root("ipv6_configuration").AtName("additional_host_ip_subnets"))
	diags.Append(listDiags...)

	cfg := &networks.CreateNetworkIpv6Configuration{
		InterfaceType:                  model.InterfaceType.ValueString(),
		PrefixDelegationWanInterfaceId: model.PrefixDelegationWANInterfaceID.ValueString(),
		DnsServerIpAddressesOverride:   dns,
		AdditionalHostIpSubnets:        additional,
	}
	cfg.ClientAddressAssignment = buildCreateIPv6ClientAssignment(model.ClientAddressAssignment, diags)
	cfg.RouterAdvertisement = buildCreateIPv6RouterAdvertisement(model.RouterAdvertisement)
	return cfg
}

func buildUpdateIPv6Config(model *networkIPv6ConfigurationModel, diags *diag.Diagnostics) *networks.UpdateNetworkIpv6Configuration {
	if model == nil {
		return nil
	}

	dns, dnsDiags := listToRawMessages(model.DNSServers, path.Root("ipv6_configuration").AtName("dns_servers"))
	diags.Append(dnsDiags...)
	additional, listDiags := listToRawMessages(model.AdditionalHostIPSubnets, path.Root("ipv6_configuration").AtName("additional_host_ip_subnets"))
	diags.Append(listDiags...)

	cfg := &networks.UpdateNetworkIpv6Configuration{
		InterfaceType:                  model.InterfaceType.ValueString(),
		PrefixDelegationWanInterfaceId: model.PrefixDelegationWANInterfaceID.ValueString(),
		DnsServerIpAddressesOverride:   dns,
		AdditionalHostIpSubnets:        additional,
	}
	cfg.ClientAddressAssignment = buildUpdateIPv6ClientAssignment(model.ClientAddressAssignment, diags)
	cfg.RouterAdvertisement = buildUpdateIPv6RouterAdvertisement(model.RouterAdvertisement)
	return cfg
}

func buildCreateIPv6ClientAssignment(model *networkIPv6ClientAddressAssignmentModel, diags *diag.Diagnostics) *networks.CreateNetworkIpv6ConfigurationClientAddressAssignment {
	if model == nil {
		return nil
	}

	cfg := &networks.CreateNetworkIpv6ConfigurationClientAddressAssignment{
		SlaacEnabled: model.SLAACEnabled.ValueBool(),
	}
	cfg.DhcpConfiguration = buildCreateIPv6DHCPConfig(model.DHCPConfiguration)
	return cfg
}

func buildUpdateIPv6ClientAssignment(model *networkIPv6ClientAddressAssignmentModel, diags *diag.Diagnostics) *networks.UpdateNetworkIpv6ConfigurationClientAddressAssignment {
	if model == nil {
		return nil
	}

	cfg := &networks.UpdateNetworkIpv6ConfigurationClientAddressAssignment{
		SlaacEnabled: model.SLAACEnabled.ValueBool(),
	}
	cfg.DhcpConfiguration = buildUpdateIPv6DHCPConfig(model.DHCPConfiguration)
	return cfg
}

func buildCreateIPv6DHCPConfig(model *networkIPv6DHCPConfigurationModel) *networks.CreateNetworkIpv6ConfigurationClientAddressAssignmentDhcpConfiguration {
	if model == nil {
		return nil
	}

	cfg := &networks.CreateNetworkIpv6ConfigurationClientAddressAssignmentDhcpConfiguration{
		LeaseTimeSeconds: model.LeaseTimeSeconds.ValueInt64(),
	}
	if model.IPAddressSuffixRange != nil {
		cfg.IpAddressSuffixRange = &networks.CreateNetworkIpv6ConfigurationClientAddressAssignmentDhcpConfigurationIpAddressSuffixRange{
			Start: model.IPAddressSuffixRange.Start.ValueString(),
			Stop:  model.IPAddressSuffixRange.Stop.ValueString(),
		}
	}
	return cfg
}

func buildUpdateIPv6DHCPConfig(model *networkIPv6DHCPConfigurationModel) *networks.UpdateNetworkIpv6ConfigurationClientAddressAssignmentDhcpConfiguration {
	if model == nil {
		return nil
	}

	cfg := &networks.UpdateNetworkIpv6ConfigurationClientAddressAssignmentDhcpConfiguration{
		LeaseTimeSeconds: model.LeaseTimeSeconds.ValueInt64(),
	}
	if model.IPAddressSuffixRange != nil {
		cfg.IpAddressSuffixRange = &networks.UpdateNetworkIpv6ConfigurationClientAddressAssignmentDhcpConfigurationIpAddressSuffixRange{
			Start: model.IPAddressSuffixRange.Start.ValueString(),
			Stop:  model.IPAddressSuffixRange.Stop.ValueString(),
		}
	}
	return cfg
}

func buildCreateIPv6RouterAdvertisement(model *networkIPv6RouterAdvertisementModel) *networks.CreateNetworkIpv6ConfigurationRouterAdvertisement {
	if model == nil {
		return nil
	}

	return &networks.CreateNetworkIpv6ConfigurationRouterAdvertisement{Priority: model.Priority.ValueString()}
}

func buildUpdateIPv6RouterAdvertisement(model *networkIPv6RouterAdvertisementModel) *networks.UpdateNetworkIpv6ConfigurationRouterAdvertisement {
	if model == nil {
		return nil
	}

	return &networks.UpdateNetworkIpv6ConfigurationRouterAdvertisement{Priority: model.Priority.ValueString()}
}

func listToRawMessages(list types.List, attrPath path.Path) ([]json.RawMessage, diag.Diagnostics) {
	var diags diag.Diagnostics
	if list.IsNull() || list.IsUnknown() {
		return nil, diags
	}

	var values []string
	if diag := list.ElementsAs(context.Background(), &values, false); diag.HasError() {
		diags.Append(diag...)
		return nil, diags
	}

	result := make([]json.RawMessage, 0, len(values))
	for _, value := range values {
		encoded, err := json.Marshal(value)
		if err != nil {
			diags.AddError("Invalid list value", fmt.Sprintf("%s: %s", attrPath.String(), err.Error()))
			return nil, diags
		}
		result = append(result, json.RawMessage(encoded))
	}

	return result, diags
}

func networkStateFromResponse(state *networkResourceModel, siteID string, response any) diag.Diagnostics {
	var diags diag.Diagnostics
	prevState := *state
	state.SiteID = types.StringValue(siteID)

	switch result := response.(type) {
	case *networks.GetNetworkDetailsResponseGateway:
		state.ID = types.StringValue(result.Id)
		state.Name = types.StringValue(result.Name)
		state.Management = types.StringValue(result.Management)
		state.Enabled = types.BoolValue(result.Enabled)
		state.VlanID = types.Int64Value(result.VlanId)
		state.ZoneID = optionalString(prevState.ZoneID, result.ZoneId)
		state.IsolationEnabled = optionalBool(prevState.IsolationEnabled, result.IsolationEnabled)
		state.CellularBackupEnabled = optionalBool(prevState.CellularBackupEnabled, result.CellularBackupEnabled)
		state.InternetAccessEnabled = optionalBool(prevState.InternetAccessEnabled, result.InternetAccessEnabled)
		state.MDNSForwardingEnabled = optionalBool(prevState.MDNSForwardingEnabled, result.MdnsForwardingEnabled)
		state.DHCPGuarding = readDHCPGuarding(result.DhcpGuarding, prevState.DHCPGuarding)
		state.IPv4Configuration = readIPv4Config(result.Ipv4Configuration, prevState.IPv4Configuration)
		state.IPv6Configuration = readIPv6Config(result.Ipv6Configuration, prevState.IPv6Configuration)
	case *networks.GetNetworkDetailsResponseSwitch:
		state.ID = types.StringValue(result.Id)
		state.Name = types.StringValue(result.Name)
		state.Management = types.StringValue(result.Management)
		state.Enabled = types.BoolValue(result.Enabled)
		state.VlanID = types.Int64Value(result.VlanId)
		state.DeviceID = optionalString(prevState.DeviceID, result.DeviceId)
		state.IsolationEnabled = optionalBool(prevState.IsolationEnabled, result.IsolationEnabled)
		state.CellularBackupEnabled = optionalBool(prevState.CellularBackupEnabled, result.CellularBackupEnabled)
		state.DHCPGuarding = readDHCPGuarding(result.DhcpGuarding, prevState.DHCPGuarding)
		state.IPv4Configuration = readIPv4Config(result.Ipv4Configuration, prevState.IPv4Configuration)
		state.IPv6Configuration = nil
	case *networks.GetNetworkDetailsResponseUnmanaged:
		state.ID = types.StringValue(result.Id)
		state.Name = types.StringValue(result.Name)
		state.Management = types.StringValue(result.Management)
		state.Enabled = types.BoolValue(result.Enabled)
		state.VlanID = types.Int64Value(result.VlanId)
		state.DHCPGuarding = readDHCPGuarding(result.DhcpGuarding, prevState.DHCPGuarding)
		state.IPv4Configuration = nil
		state.IPv6Configuration = nil
	default:
		diags.AddError("Unsupported network response", fmt.Sprintf("unexpected response type %T", response))
	}

	return diags
}

func readDHCPGuarding(guarding *networks.GetNetworkDetailsDhcpGuarding, prev *networkDHCPGuardingModel) *networkDHCPGuardingModel {
	if guarding == nil || prev == nil {
		return nil
	}

	list := rawMessagesToList(guarding.TrustedDhcpServerIpAddresses)
	return &networkDHCPGuardingModel{TrustedDHCPServerIPAddresses: list}
}

func readIPv4Config(cfg *networks.GetNetworkDetailsIpv4Configuration, prev *networkIPv4ConfigurationModel) *networkIPv4ConfigurationModel {
	if cfg == nil || prev == nil {
		return nil
	}

	cidr := types.StringNull()
	if prev != nil && !prev.CIDR.IsNull() && !prev.CIDR.IsUnknown() {
		cidr = prev.CIDR
	}

	model := &networkIPv4ConfigurationModel{
		AutoScaleEnabled:        optionalBool(prev.AutoScaleEnabled, cfg.AutoScaleEnabled),
		CIDR:                    cidr,
		HostIPAddress:           optionalString(prev.HostIPAddress, cfg.HostIpAddress),
		PrefixLength:            optionalInt64(prev.PrefixLength, cfg.PrefixLength),
		AdditionalHostIPSubnets: optionalList(prev.AdditionalHostIPSubnets, cfg.AdditionalHostIpSubnets),
	}
	model.DHCPConfiguration = readIPv4DHCPConfig(cfg.DhcpConfiguration, prev.DHCPConfiguration)
	return model
}

func resolveIPv4HostPrefix(model *networkIPv4ConfigurationModel, diags *diag.Diagnostics) (string, int64) {
	if model == nil {
		return "", 0
	}

	if model.CIDR.IsUnknown() {
		return model.HostIPAddress.ValueString(), model.PrefixLength.ValueInt64()
	}

	cidrValue := strings.TrimSpace(model.CIDR.ValueString())
	if cidrValue == "" {
		return model.HostIPAddress.ValueString(), model.PrefixLength.ValueInt64()
	}

	if !model.HostIPAddress.IsNull() && !model.HostIPAddress.IsUnknown() && model.HostIPAddress.ValueString() != "" {
		diags.AddAttributeError(
			path.Root("ipv4_configuration").AtName("cidr"),
			"Conflicting IPv4 configuration",
			"cidr cannot be set with host_ip_address. Use cidr only.",
		)
		return "", 0
	}
	if !model.PrefixLength.IsNull() && !model.PrefixLength.IsUnknown() && model.PrefixLength.ValueInt64() != 0 {
		diags.AddAttributeError(
			path.Root("ipv4_configuration").AtName("cidr"),
			"Conflicting IPv4 configuration",
			"cidr cannot be set with prefix_length. Use cidr only.",
		)
		return "", 0
	}

	ip, ipNet, err := net.ParseCIDR(cidrValue)
	if err != nil {
		diags.AddAttributeError(
			path.Root("ipv4_configuration").AtName("cidr"),
			"Invalid CIDR",
			err.Error(),
		)
		return "", 0
	}

	ones, _ := ipNet.Mask.Size()
	ipv4 := ip.To4()
	base := ipNet.IP.To4()
	if ipv4 == nil || base == nil {
		diags.AddAttributeError(
			path.Root("ipv4_configuration").AtName("cidr"),
			"Invalid CIDR",
			"cidr must be an IPv4 CIDR block.",
		)
		return "", 0
	}

	hostIP := ipv4
	if ipv4.Equal(base) && ones < 31 {
		calculated, err := ipv4Host(ipNet, 1)
		if err != nil {
			diags.AddAttributeError(
				path.Root("ipv4_configuration").AtName("cidr"),
				"Invalid CIDR",
				err.Error(),
			)
			return "", 0
		}
		hostIP = calculated
	}

	return hostIP.String(), int64(ones)
}

func ipv4Host(ipNet *net.IPNet, host uint32) (net.IP, error) {
	if ipNet == nil {
		return nil, fmt.Errorf("invalid CIDR")
	}

	base := ipNet.IP.To4()
	if base == nil {
		return nil, fmt.Errorf("cidr must be IPv4")
	}

	ones, bits := ipNet.Mask.Size()
	if bits != 32 {
		return nil, fmt.Errorf("cidr must be IPv4")
	}

	if host >= 1<<uint32(32-ones) {
		return nil, fmt.Errorf("host index %d out of range for /%d", host, ones)
	}

	baseInt := binary.BigEndian.Uint32(base)
	ipInt := baseInt + host
	result := make(net.IP, 4)
	binary.BigEndian.PutUint32(result, ipInt)
	return result, nil
}

func readIPv4DHCPConfig(cfg *networks.GetNetworkDetailsIpv4ConfigurationDhcpConfiguration, prev *networkIPv4DHCPConfigurationModel) *networkIPv4DHCPConfigurationModel {
	if cfg == nil || prev == nil {
		return nil
	}

	model := &networkIPv4DHCPConfigurationModel{
		Mode:                     optionalString(prev.Mode, cfg.Mode),
		GatewayIPAddressOverride: optionalString(prev.GatewayIPAddressOverride, cfg.GatewayIpAddressOverride),
		DNSServers:               optionalList(prev.DNSServers, cfg.DnsServerIpAddressesOverride),
		LeaseTimeSeconds:         optionalInt64(prev.LeaseTimeSeconds, cfg.LeaseTimeSeconds),
		DomainName:               optionalString(prev.DomainName, cfg.DomainName),
	}
	if prev.IPAddressRange != nil && cfg.IpAddressRange != nil {
		model.IPAddressRange = &networkIPAddressRangeModel{
			Start: types.StringValue(cfg.IpAddressRange.Start),
			Stop:  types.StringValue(cfg.IpAddressRange.Stop),
		}
	}
	return model
}

func readIPv6Config(cfg *networks.GetNetworkDetailsIpv6Configuration, prev *networkIPv6ConfigurationModel) *networkIPv6ConfigurationModel {
	if cfg == nil || prev == nil {
		return nil
	}

	model := &networkIPv6ConfigurationModel{
		InterfaceType:                  optionalString(prev.InterfaceType, cfg.InterfaceType),
		PrefixDelegationWANInterfaceID: optionalString(prev.PrefixDelegationWANInterfaceID, cfg.PrefixDelegationWanInterfaceId),
		DNSServers:                     optionalList(prev.DNSServers, cfg.DnsServerIpAddressesOverride),
		AdditionalHostIPSubnets:        optionalList(prev.AdditionalHostIPSubnets, cfg.AdditionalHostIpSubnets),
	}
	model.ClientAddressAssignment = readIPv6ClientAssignment(cfg.ClientAddressAssignment, prev.ClientAddressAssignment)
	model.RouterAdvertisement = readIPv6RouterAdvertisement(cfg.RouterAdvertisement, prev.RouterAdvertisement)
	return model
}

func readIPv6ClientAssignment(cfg *networks.GetNetworkDetailsIpv6ConfigurationClientAddressAssignment, prev *networkIPv6ClientAddressAssignmentModel) *networkIPv6ClientAddressAssignmentModel {
	if cfg == nil || prev == nil {
		return nil
	}

	model := &networkIPv6ClientAddressAssignmentModel{
		SLAACEnabled: optionalBool(prev.SLAACEnabled, cfg.SlaacEnabled),
	}
	model.DHCPConfiguration = readIPv6DHCPConfig(cfg.DhcpConfiguration, prev.DHCPConfiguration)
	return model
}

func readIPv6DHCPConfig(cfg *networks.GetNetworkDetailsIpv6ConfigurationClientAddressAssignmentDhcpConfiguration, prev *networkIPv6DHCPConfigurationModel) *networkIPv6DHCPConfigurationModel {
	if cfg == nil || prev == nil {
		return nil
	}

	model := &networkIPv6DHCPConfigurationModel{
		LeaseTimeSeconds: optionalInt64(prev.LeaseTimeSeconds, cfg.LeaseTimeSeconds),
	}
	if prev.IPAddressSuffixRange != nil && cfg.IpAddressSuffixRange != nil {
		model.IPAddressSuffixRange = &networkIPAddressRangeModel{
			Start: types.StringValue(cfg.IpAddressSuffixRange.Start),
			Stop:  types.StringValue(cfg.IpAddressSuffixRange.Stop),
		}
	}
	return model
}

func readIPv6RouterAdvertisement(cfg *networks.GetNetworkDetailsIpv6ConfigurationRouterAdvertisement, prev *networkIPv6RouterAdvertisementModel) *networkIPv6RouterAdvertisementModel {
	if cfg == nil || prev == nil {
		return nil
	}

	return &networkIPv6RouterAdvertisementModel{Priority: optionalString(prev.Priority, cfg.Priority)}
}

func optionalBool(prev types.Bool, value bool) types.Bool {
	if prev.IsNull() || prev.IsUnknown() {
		return types.BoolNull()
	}
	return types.BoolValue(value)
}

func optionalString(prev types.String, value string) types.String {
	if prev.IsNull() || prev.IsUnknown() {
		return types.StringNull()
	}
	return types.StringValue(value)
}

func optionalInt64(prev types.Int64, value int64) types.Int64 {
	if prev.IsNull() || prev.IsUnknown() {
		return types.Int64Null()
	}
	return types.Int64Value(value)
}

func optionalList(prev types.List, raw []json.RawMessage) types.List {
	if prev.IsNull() || prev.IsUnknown() {
		return types.ListNull(types.StringType)
	}
	return rawMessagesToList(raw)
}
