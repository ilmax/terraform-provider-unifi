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
			},
			"zone_id": schema.StringAttribute{
				Optional: true,
			},
			"device_id": schema.StringAttribute{
				Optional: true,
			},
			"isolation_enabled": schema.BoolAttribute{
				Optional: true,
			},
			"cellular_backup_enabled": schema.BoolAttribute{
				Optional: true,
			},
			"internet_access_enabled": schema.BoolAttribute{
				Optional: true,
			},
			"mdns_forwarding_enabled": schema.BoolAttribute{
				Optional: true,
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
					},
					"cidr": schema.StringAttribute{
						Optional: true,
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
							},
							"ip_address_range": schema.SingleNestedAttribute{
								Optional: true,
								Attributes: map[string]schema.Attribute{
									"start": schema.StringAttribute{
										Optional: true,
									},
									"stop": schema.StringAttribute{
										Optional: true,
									},
								},
							},
							"gateway_ip_address_override": schema.StringAttribute{
								Optional: true,
							},
							"dns_servers": schema.ListAttribute{
								Optional:    true,
								ElementType: types.StringType,
							},
							"lease_time_seconds": schema.Int64Attribute{
								Optional: true,
							},
							"domain_name": schema.StringAttribute{
								Optional: true,
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
					},
					"prefix_delegation_wan_interface_id": schema.StringAttribute{
						Optional: true,
					},
					"dns_servers": schema.ListAttribute{
						Optional:    true,
						ElementType: types.StringType,
					},
					"additional_host_ip_subnets": schema.ListAttribute{
						Optional:    true,
						ElementType: types.StringType,
					},
					"client_address_assignment": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"slaac_enabled": schema.BoolAttribute{
								Optional: true,
							},
							"dhcp_configuration": schema.SingleNestedAttribute{
								Optional: true,
								Attributes: map[string]schema.Attribute{
									"ip_address_suffix_range": schema.SingleNestedAttribute{
										Optional: true,
										Attributes: map[string]schema.Attribute{
											"start": schema.StringAttribute{
												Optional: true,
											},
											"stop": schema.StringAttribute{
												Optional: true,
											},
										},
									},
									"lease_time_seconds": schema.Int64Attribute{
										Optional: true,
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
	if err := r.client.Post(ctx, path, payload, nil); err != nil {
		resp.Diagnostics.AddError("Unable to create network", err.Error())
		return
	}

	state := plan
	state.SiteID = types.StringValue(siteID)
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

	path := fmt.Sprintf("/v1/sites/%s/networks/%s", siteID, state.ID.ValueString())
	var raw json.RawMessage
	if err := r.client.Get(ctx, path, &raw); err != nil {
		if errors.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read network", err.Error())
		return
	}

	decoded, err := networks.DecodeGetNetworkDetailsResponse(raw)
	if err != nil {
		resp.Diagnostics.AddError("Unable to decode network", err.Error())
		return
	}

	resp.Diagnostics.Append(networkStateFromResponse(&state, siteID, decoded)...)
	if resp.Diagnostics.HasError() {
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
	state.SiteID = types.StringValue(siteID)
	prevIPv4 := state.IPv4Configuration

	switch result := response.(type) {
	case *networks.GetNetworkDetailsResponseGateway:
		state.ID = types.StringValue(result.Id)
		state.Name = types.StringValue(result.Name)
		state.Management = types.StringValue(result.Management)
		state.Enabled = types.BoolValue(result.Enabled)
		state.VlanID = types.Int64Value(result.VlanId)
		state.ZoneID = types.StringValue(result.ZoneId)
		state.IsolationEnabled = types.BoolValue(result.IsolationEnabled)
		state.CellularBackupEnabled = types.BoolValue(result.CellularBackupEnabled)
		state.InternetAccessEnabled = types.BoolValue(result.InternetAccessEnabled)
		state.MDNSForwardingEnabled = types.BoolValue(result.MdnsForwardingEnabled)
		state.DHCPGuarding = readDHCPGuarding(result.DhcpGuarding)
		state.IPv4Configuration = readIPv4Config(result.Ipv4Configuration, prevIPv4)
		state.IPv6Configuration = readIPv6Config(result.Ipv6Configuration)
	case *networks.GetNetworkDetailsResponseSwitch:
		state.ID = types.StringValue(result.Id)
		state.Name = types.StringValue(result.Name)
		state.Management = types.StringValue(result.Management)
		state.Enabled = types.BoolValue(result.Enabled)
		state.VlanID = types.Int64Value(result.VlanId)
		state.DeviceID = types.StringValue(result.DeviceId)
		state.IsolationEnabled = types.BoolValue(result.IsolationEnabled)
		state.CellularBackupEnabled = types.BoolValue(result.CellularBackupEnabled)
		state.DHCPGuarding = readDHCPGuarding(result.DhcpGuarding)
		state.IPv4Configuration = readIPv4Config(result.Ipv4Configuration, prevIPv4)
		state.IPv6Configuration = nil
	case *networks.GetNetworkDetailsResponseUnmanaged:
		state.ID = types.StringValue(result.Id)
		state.Name = types.StringValue(result.Name)
		state.Management = types.StringValue(result.Management)
		state.Enabled = types.BoolValue(result.Enabled)
		state.VlanID = types.Int64Value(result.VlanId)
		state.DHCPGuarding = readDHCPGuarding(result.DhcpGuarding)
		state.IPv4Configuration = nil
		state.IPv6Configuration = nil
	default:
		diags.AddError("Unsupported network response", fmt.Sprintf("unexpected response type %T", response))
	}

	return diags
}

func readDHCPGuarding(guarding *networks.GetNetworkDetailsDhcpGuarding) *networkDHCPGuardingModel {
	if guarding == nil {
		return nil
	}

	list := rawMessagesToList(guarding.TrustedDhcpServerIpAddresses)
	return &networkDHCPGuardingModel{TrustedDHCPServerIPAddresses: list}
}

func readIPv4Config(cfg *networks.GetNetworkDetailsIpv4Configuration, prev *networkIPv4ConfigurationModel) *networkIPv4ConfigurationModel {
	if cfg == nil {
		return nil
	}

	cidr := types.StringNull()
	if prev != nil && !prev.CIDR.IsNull() && !prev.CIDR.IsUnknown() {
		cidr = prev.CIDR
	} else if cfg.HostIpAddress != "" && cfg.PrefixLength != 0 {
		cidr = types.StringValue(fmt.Sprintf("%s/%d", cfg.HostIpAddress, cfg.PrefixLength))
	}

	model := &networkIPv4ConfigurationModel{
		AutoScaleEnabled:        types.BoolValue(cfg.AutoScaleEnabled),
		CIDR:                    cidr,
		HostIPAddress:           types.StringValue(cfg.HostIpAddress),
		PrefixLength:            types.Int64Value(cfg.PrefixLength),
		AdditionalHostIPSubnets: rawMessagesToList(cfg.AdditionalHostIpSubnets),
	}
	model.DHCPConfiguration = readIPv4DHCPConfig(cfg.DhcpConfiguration)
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

func readIPv4DHCPConfig(cfg *networks.GetNetworkDetailsIpv4ConfigurationDhcpConfiguration) *networkIPv4DHCPConfigurationModel {
	if cfg == nil {
		return nil
	}

	model := &networkIPv4DHCPConfigurationModel{
		Mode:                     types.StringValue(cfg.Mode),
		GatewayIPAddressOverride: types.StringValue(cfg.GatewayIpAddressOverride),
		DNSServers:               rawMessagesToList(cfg.DnsServerIpAddressesOverride),
		LeaseTimeSeconds:         types.Int64Value(cfg.LeaseTimeSeconds),
		DomainName:               types.StringValue(cfg.DomainName),
	}
	if cfg.IpAddressRange != nil {
		model.IPAddressRange = &networkIPAddressRangeModel{
			Start: types.StringValue(cfg.IpAddressRange.Start),
			Stop:  types.StringValue(cfg.IpAddressRange.Stop),
		}
	}
	return model
}

func readIPv6Config(cfg *networks.GetNetworkDetailsIpv6Configuration) *networkIPv6ConfigurationModel {
	if cfg == nil {
		return nil
	}

	model := &networkIPv6ConfigurationModel{
		InterfaceType:                  types.StringValue(cfg.InterfaceType),
		PrefixDelegationWANInterfaceID: types.StringValue(cfg.PrefixDelegationWanInterfaceId),
		DNSServers:                     rawMessagesToList(cfg.DnsServerIpAddressesOverride),
		AdditionalHostIPSubnets:        rawMessagesToList(cfg.AdditionalHostIpSubnets),
	}
	model.ClientAddressAssignment = readIPv6ClientAssignment(cfg.ClientAddressAssignment)
	model.RouterAdvertisement = readIPv6RouterAdvertisement(cfg.RouterAdvertisement)
	return model
}

func readIPv6ClientAssignment(cfg *networks.GetNetworkDetailsIpv6ConfigurationClientAddressAssignment) *networkIPv6ClientAddressAssignmentModel {
	if cfg == nil {
		return nil
	}

	model := &networkIPv6ClientAddressAssignmentModel{
		SLAACEnabled: types.BoolValue(cfg.SlaacEnabled),
	}
	model.DHCPConfiguration = readIPv6DHCPConfig(cfg.DhcpConfiguration)
	return model
}

func readIPv6DHCPConfig(cfg *networks.GetNetworkDetailsIpv6ConfigurationClientAddressAssignmentDhcpConfiguration) *networkIPv6DHCPConfigurationModel {
	if cfg == nil {
		return nil
	}

	model := &networkIPv6DHCPConfigurationModel{
		LeaseTimeSeconds: types.Int64Value(cfg.LeaseTimeSeconds),
	}
	if cfg.IpAddressSuffixRange != nil {
		model.IPAddressSuffixRange = &networkIPAddressRangeModel{
			Start: types.StringValue(cfg.IpAddressSuffixRange.Start),
			Stop:  types.StringValue(cfg.IpAddressSuffixRange.Stop),
		}
	}
	return model
}

func readIPv6RouterAdvertisement(cfg *networks.GetNetworkDetailsIpv6ConfigurationRouterAdvertisement) *networkIPv6RouterAdvertisementModel {
	if cfg == nil {
		return nil
	}

	return &networkIPv6RouterAdvertisementModel{Priority: types.StringValue(cfg.Priority)}
}
