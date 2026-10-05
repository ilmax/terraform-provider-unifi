package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/sdkcompat/broadcasts"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
	"github.com/ilmax/unifi-client-go/pkg/errors"
)

type wifiResource struct {
	client *unifi.Client
	siteID string
}

type wifiResourceModel struct {
	ID                                  types.String                            `tfsdk:"id"`
	SiteID                              types.String                            `tfsdk:"site_id"`
	Name                                types.String                            `tfsdk:"name"`
	Type                                types.String                            `tfsdk:"type"`
	Enabled                             types.Bool                              `tfsdk:"enabled"`
	SecurityType                        types.String                            `tfsdk:"security_type"`
	NetworkType                         types.String                            `tfsdk:"network_type"`
	MulticastToUnicastConversionEnabled types.Bool                              `tfsdk:"multicast_to_unicast_conversion_enabled"`
	ClientIsolationEnabled              types.Bool                              `tfsdk:"client_isolation_enabled"`
	HideName                            types.Bool                              `tfsdk:"hide_name"`
	UapsdEnabled                        types.Bool                              `tfsdk:"uapsd_enabled"`
	BroadcastingFrequenciesGHz          types.List                              `tfsdk:"broadcasting_frequencies_ghz"`
	MloEnabled                          types.Bool                              `tfsdk:"mlo_enabled"`
	BandSteeringEnabled                 types.Bool                              `tfsdk:"band_steering_enabled"`
	ArpProxyEnabled                     types.Bool                              `tfsdk:"arp_proxy_enabled"`
	BssTransitionEnabled                types.Bool                              `tfsdk:"bss_transition_enabled"`
	BasicDataRateKbpsByFrequencyGHz     *wifiBasicDataRateModel                 `tfsdk:"basic_data_rate_kbps_by_frequency_ghz"`
	ClientFilteringPolicy               *wifiClientFilteringPolicyModel         `tfsdk:"client_filtering_policy"`
	BlackoutScheduleConfiguration       *wifiBlackoutScheduleConfigurationModel `tfsdk:"blackout_schedule_configuration"`
	AdvertiseDeviceName                 types.Bool                              `tfsdk:"advertise_device_name"`
}

type wifiBasicDataRateModel struct {
	Rate24Kbps types.Int64 `tfsdk:"rate_2_4_kbps"`
	Rate5Kbps  types.Int64 `tfsdk:"rate_5_kbps"`
}

type wifiClientFilteringPolicyModel struct {
	Action           types.String `tfsdk:"action"`
	MacAddressFilter types.List   `tfsdk:"mac_address_filter"`
}

type wifiBlackoutScheduleConfigurationModel struct {
	Days []wifiBlackoutDayModel `tfsdk:"days"`
}

type wifiBlackoutDayModel struct {
	Day  types.String `tfsdk:"day"`
	Type types.String `tfsdk:"type"`
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
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
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
			"multicast_to_unicast_conversion_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					boolplanmodifier.RequiresReplace(),
				},
			},
			"client_isolation_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					boolplanmodifier.RequiresReplace(),
				},
			},
			"hide_name": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					boolplanmodifier.RequiresReplace(),
				},
			},
			"uapsd_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					boolplanmodifier.RequiresReplace(),
				},
			},
			"broadcasting_frequencies_ghz": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"mlo_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					boolplanmodifier.RequiresReplace(),
				},
			},
			"band_steering_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					boolplanmodifier.RequiresReplace(),
				},
			},
			"arp_proxy_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					boolplanmodifier.RequiresReplace(),
				},
			},
			"bss_transition_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					boolplanmodifier.RequiresReplace(),
				},
			},
			"advertise_device_name": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the device name is advertised in beacon frames. Required by the API for STANDARD broadcasts; omitting it sends false. Only supported for STANDARD broadcasts.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					boolplanmodifier.RequiresReplace(),
				},
			},
			"basic_data_rate_kbps_by_frequency_ghz": schema.SingleNestedAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"rate_2_4_kbps": schema.Int64Attribute{
						Optional: true,
					},
					"rate_5_kbps": schema.Int64Attribute{
						Optional: true,
					},
				},
			},
			"client_filtering_policy": schema.SingleNestedAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"action": schema.StringAttribute{
						Required: true,
					},
					"mac_address_filter": schema.ListAttribute{
						Optional:    true,
						ElementType: types.StringType,
					},
				},
			},
			"blackout_schedule_configuration": schema.SingleNestedAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"days": schema.ListNestedAttribute{
						Optional: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"day": schema.StringAttribute{
									Required: true,
								},
								"type": schema.StringAttribute{
									Required: true,
								},
							},
						},
					},
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

	payload, diags := buildWifiPayload(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var raw json.RawMessage
	path := fmt.Sprintf("/v1/sites/%s/wifi/broadcasts", siteID)
	if err := r.client.Post(ctx, path, payload, &raw); err != nil {
		resp.Diagnostics.AddError("Unable to create WiFi broadcast", err.Error())
		return
	}

	decoded, err := broadcasts.DecodeCreateWifiBroadcastResponse(raw)
	if err != nil {
		resp.Diagnostics.AddError("Unable to decode WiFi broadcast", err.Error())
		return
	}

	state, diags := wifiStateFromResponse(siteID, decoded)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
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

	var raw json.RawMessage
	path := fmt.Sprintf("/v1/sites/%s/wifi/broadcasts/%s", siteID, state.ID.ValueString())
	if err := r.client.Get(ctx, path, &raw); err != nil {
		if errors.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read WiFi broadcast", err.Error())
		return
	}

	decoded, err := broadcasts.DecodeGetWifiBroadcastDetailsResponse(raw)
	if err != nil {
		resp.Diagnostics.AddError("Unable to decode WiFi broadcast", err.Error())
		return
	}

	state, diags := wifiStateFromResponse(siteID, decoded)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
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

func buildWifiPayload(plan wifiResourceModel) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	payload := map[string]any{
		"type":    plan.Type.ValueString(),
		"name":    plan.Name.ValueString(),
		"enabled": plan.Enabled.ValueBool(),
		"securityConfiguration": map[string]any{
			"type": plan.SecurityType.ValueString(),
		},
	}

	if !plan.NetworkType.IsNull() && !plan.NetworkType.IsUnknown() {
		payload["network"] = map[string]any{
			"type": plan.NetworkType.ValueString(),
		}
	}

	setBoolIfKnown(payload, "multicastToUnicastConversionEnabled", plan.MulticastToUnicastConversionEnabled)
	setBoolIfKnown(payload, "clientIsolationEnabled", plan.ClientIsolationEnabled)
	setBoolIfKnown(payload, "hideName", plan.HideName)
	setBoolIfKnown(payload, "uapsdEnabled", plan.UapsdEnabled)

	if !plan.BroadcastingFrequenciesGHz.IsNull() && !plan.BroadcastingFrequenciesGHz.IsUnknown() {
		if strings.ToUpper(plan.Type.ValueString()) != "STANDARD" {
			diags.AddAttributeError(path.Root("broadcasting_frequencies_ghz"), "Unsupported field", "broadcasting_frequencies_ghz is only supported for STANDARD broadcasts.")
			return nil, diags
		}
		frequencies, listDiags := listToStrings(plan.BroadcastingFrequenciesGHz, path.Root("broadcasting_frequencies_ghz"))
		diags.Append(listDiags...)
		if diags.HasError() {
			return nil, diags
		}
		// The API types broadcastingFrequenciesGHz as an array of numbers
		// (2.4, 5, 6), not strings — sending Go strings here marshals to
		// JSON strings and the API rejects it as a type mismatch. Parse and
		// re-encode as raw JSON numbers instead.
		frequencyNumbers := make([]json.RawMessage, 0, len(frequencies))
		for _, frequency := range frequencies {
			if _, err := strconv.ParseFloat(frequency, 64); err != nil {
				diags.AddAttributeError(path.Root("broadcasting_frequencies_ghz"), "Invalid frequency", fmt.Sprintf("%q is not a valid number.", frequency))
				return nil, diags
			}
			frequencyNumbers = append(frequencyNumbers, json.RawMessage(frequency))
		}
		payload["broadcastingFrequenciesGHz"] = frequencyNumbers
	}

	isStandard := strings.ToUpper(plan.Type.ValueString()) == "STANDARD"

	if !isStandard {
		if hasAnyTrue(plan.MloEnabled, plan.BandSteeringEnabled, plan.ArpProxyEnabled, plan.BssTransitionEnabled, plan.AdvertiseDeviceName) {
			diags.AddAttributeError(path.Root("type"), "Unsupported fields", "mlo_enabled, band_steering_enabled, arp_proxy_enabled, bss_transition_enabled, and advertise_device_name are only supported for STANDARD broadcasts.")
			return nil, diags
		}
	}

	setBoolIfKnown(payload, "mloEnabled", plan.MloEnabled)
	setBoolIfKnown(payload, "bandSteeringEnabled", plan.BandSteeringEnabled)
	setBoolIfKnown(payload, "arpProxyEnabled", plan.ArpProxyEnabled)
	setBoolIfKnown(payload, "bssTransitionEnabled", plan.BssTransitionEnabled)
	setBoolIfKnown(payload, "advertiseDeviceName", plan.AdvertiseDeviceName)

	// The API requires these as explicit booleans on create — never merely
	// absent — for both broadcast types (plus three more for STANDARD).
	// setBoolIfKnown above only sets a key when the user gave an explicit
	// value, so fill in a safe `false` default for anything still missing.
	requiredBoolDefaults := []string{"multicastToUnicastConversionEnabled", "clientIsolationEnabled", "hideName", "uapsdEnabled"}
	if isStandard {
		requiredBoolDefaults = append(requiredBoolDefaults, "arpProxyEnabled", "bssTransitionEnabled", "advertiseDeviceName")
	}
	for _, key := range requiredBoolDefaults {
		if _, ok := payload[key]; !ok {
			payload[key] = false
		}
	}

	if plan.BasicDataRateKbpsByFrequencyGHz != nil {
		rates := map[string]any{}
		if !plan.BasicDataRateKbpsByFrequencyGHz.Rate24Kbps.IsNull() && !plan.BasicDataRateKbpsByFrequencyGHz.Rate24Kbps.IsUnknown() {
			rates["2.4"] = plan.BasicDataRateKbpsByFrequencyGHz.Rate24Kbps.ValueInt64()
		}
		if !plan.BasicDataRateKbpsByFrequencyGHz.Rate5Kbps.IsNull() && !plan.BasicDataRateKbpsByFrequencyGHz.Rate5Kbps.IsUnknown() {
			rates["5"] = plan.BasicDataRateKbpsByFrequencyGHz.Rate5Kbps.ValueInt64()
		}
		if len(rates) == 0 {
			diags.AddAttributeError(
				path.Root("basic_data_rate_kbps_by_frequency_ghz"),
				"Missing basic data rates",
				"At least one of rate_2_4_kbps or rate_5_kbps must be set when basic_data_rate_kbps_by_frequency_ghz is configured.",
			)
			return nil, diags
		}
		payload["basicDataRateKbpsByFrequencyGHz"] = rates
	}

	if plan.ClientFilteringPolicy != nil {
		clientFiltering := map[string]any{
			"action": plan.ClientFilteringPolicy.Action.ValueString(),
		}
		if !plan.ClientFilteringPolicy.MacAddressFilter.IsNull() && !plan.ClientFilteringPolicy.MacAddressFilter.IsUnknown() {
			macs, listDiags := listToStrings(plan.ClientFilteringPolicy.MacAddressFilter, path.Root("client_filtering_policy").AtName("mac_address_filter"))
			diags.Append(listDiags...)
			if diags.HasError() {
				return nil, diags
			}
			clientFiltering["macAddressFilter"] = macs
		}
		payload["clientFilteringPolicy"] = clientFiltering
	}

	if plan.BlackoutScheduleConfiguration != nil {
		days := make([]map[string]any, 0, len(plan.BlackoutScheduleConfiguration.Days))
		for idx, day := range plan.BlackoutScheduleConfiguration.Days {
			if day.Day.IsNull() || day.Day.IsUnknown() || strings.TrimSpace(day.Day.ValueString()) == "" {
				diags.AddAttributeError(
					path.Root("blackout_schedule_configuration").AtName("days").AtListIndex(idx).AtName("day"),
					"Missing blackout day",
					"day must be set for blackout_schedule_configuration.days entries.",
				)
				return nil, diags
			}
			if day.Type.IsNull() || day.Type.IsUnknown() || strings.TrimSpace(day.Type.ValueString()) == "" {
				diags.AddAttributeError(
					path.Root("blackout_schedule_configuration").AtName("days").AtListIndex(idx).AtName("type"),
					"Missing blackout type",
					"type must be set for blackout_schedule_configuration.days entries.",
				)
				return nil, diags
			}
			days = append(days, map[string]any{
				"day":  day.Day.ValueString(),
				"type": day.Type.ValueString(),
			})
		}
		payload["blackoutScheduleConfiguration"] = map[string]any{
			"days": days,
		}
	}

	return payload, diags
}

func setBoolIfKnown(payload map[string]any, key string, value types.Bool) {
	if value.IsNull() || value.IsUnknown() {
		return
	}
	payload[key] = value.ValueBool()
}

func hasAnyTrue(values ...types.Bool) bool {
	for _, value := range values {
		if value.IsNull() || value.IsUnknown() {
			continue
		}
		if value.ValueBool() {
			return true
		}
	}
	return false
}

func listToStrings(list types.List, attrPath path.Path) ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if list.IsNull() || list.IsUnknown() {
		return nil, diags
	}

	var values []string
	if diag := list.ElementsAs(context.Background(), &values, false); diag.HasError() {
		diags.Append(diag...)
		return nil, diags
	}

	return values, diags
}

func wifiStateFromResponse(siteID string, response any) (wifiResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch result := response.(type) {
	case *broadcasts.CreateWifiBroadcastResponseStandard:
		return wifiStateFromStandard(
			siteID,
			result.Id,
			result.Name,
			result.Type,
			result.Enabled,
			result.SecurityConfiguration.Type,
			result.Network,
			result.MulticastToUnicastConversionEnabled,
			result.ClientIsolationEnabled,
			result.HideName,
			result.UapsdEnabled,
			result.BroadcastingFrequenciesGHz,
			result.MloEnabled,
			result.BandSteeringEnabled,
			result.ArpProxyEnabled,
			result.BssTransitionEnabled,
			result.AdvertiseDeviceName,
			readCreateWifiBasicDataRate(result.BasicDataRateKbpsByFrequencyGHz),
			readCreateWifiClientFilteringPolicy(result.ClientFilteringPolicy),
			readCreateWifiBlackoutSchedule(result.BlackoutScheduleConfiguration),
		), diags
	case *broadcasts.CreateWifiBroadcastResponseIotOptimized:
		return wifiStateFromIot(
			siteID,
			result.Id,
			result.Name,
			result.Type,
			result.Enabled,
			result.SecurityConfiguration.Type,
			result.Network,
			result.MulticastToUnicastConversionEnabled,
			result.ClientIsolationEnabled,
			result.HideName,
			result.UapsdEnabled,
			readCreateWifiBasicDataRate(result.BasicDataRateKbpsByFrequencyGHz),
			readCreateWifiClientFilteringPolicy(result.ClientFilteringPolicy),
			readCreateWifiBlackoutSchedule(result.BlackoutScheduleConfiguration),
		), diags
	case *broadcasts.GetWifiBroadcastDetailsResponseStandard:
		return wifiStateFromStandard(
			siteID,
			result.Id,
			result.Name,
			result.Type,
			result.Enabled,
			result.SecurityConfiguration.Type,
			result.Network,
			result.MulticastToUnicastConversionEnabled,
			result.ClientIsolationEnabled,
			result.HideName,
			result.UapsdEnabled,
			result.BroadcastingFrequenciesGHz,
			result.MloEnabled,
			result.BandSteeringEnabled,
			result.ArpProxyEnabled,
			result.BssTransitionEnabled,
			result.AdvertiseDeviceName,
			readGetWifiBasicDataRate(result.BasicDataRateKbpsByFrequencyGHz),
			readGetWifiClientFilteringPolicy(result.ClientFilteringPolicy),
			readGetWifiBlackoutSchedule(result.BlackoutScheduleConfiguration),
		), diags
	case *broadcasts.GetWifiBroadcastDetailsResponseIotOptimized:
		return wifiStateFromIot(
			siteID,
			result.Id,
			result.Name,
			result.Type,
			result.Enabled,
			result.SecurityConfiguration.Type,
			result.Network,
			result.MulticastToUnicastConversionEnabled,
			result.ClientIsolationEnabled,
			result.HideName,
			result.UapsdEnabled,
			readGetWifiBasicDataRate(result.BasicDataRateKbpsByFrequencyGHz),
			readGetWifiClientFilteringPolicy(result.ClientFilteringPolicy),
			readGetWifiBlackoutSchedule(result.BlackoutScheduleConfiguration),
		), diags
	default:
		diags.AddError("Unsupported WiFi broadcast response", fmt.Sprintf("unexpected response type %T", response))
		return wifiResourceModel{}, diags
	}
}

func wifiStateCommon(siteID, id, name, broadcastType string, enabled bool, securityType string, network *broadcasts.ClientAccess) wifiResourceModel {
	state := wifiResourceModel{
		ID:      types.StringValue(id),
		SiteID:  types.StringValue(siteID),
		Name:    types.StringValue(name),
		Type:    types.StringValue(broadcastType),
		Enabled: types.BoolValue(enabled),
	}

	if securityType != "" {
		state.SecurityType = types.StringValue(securityType)
	}

	if network != nil {
		state.NetworkType = types.StringValue(network.Type)
	} else {
		state.NetworkType = types.StringNull()
	}

	state.MulticastToUnicastConversionEnabled = types.BoolNull()
	state.ClientIsolationEnabled = types.BoolNull()
	state.HideName = types.BoolNull()
	state.UapsdEnabled = types.BoolNull()
	state.BroadcastingFrequenciesGHz = types.ListNull(types.StringType)
	state.MloEnabled = types.BoolNull()
	state.BandSteeringEnabled = types.BoolNull()
	state.ArpProxyEnabled = types.BoolNull()
	state.BssTransitionEnabled = types.BoolNull()
	state.AdvertiseDeviceName = types.BoolNull()
	state.BasicDataRateKbpsByFrequencyGHz = nil
	state.ClientFilteringPolicy = nil
	state.BlackoutScheduleConfiguration = nil

	return state
}

func wifiStateFromStandard(siteID, id, name, broadcastType string, enabled bool, securityType string, network *broadcasts.ClientAccess, multicastToUnicast bool, clientIsolation bool, hideName bool, uapsdEnabled bool, frequencies []json.RawMessage, mloEnabled bool, bandSteering bool, arpProxy bool, bssTransition bool, advertiseDeviceName bool, basicDataRate *wifiBasicDataRateModel, clientFilteringPolicy *wifiClientFilteringPolicyModel, blackoutSchedule *wifiBlackoutScheduleConfigurationModel) wifiResourceModel {
	state := wifiStateCommon(siteID, id, name, broadcastType, enabled, securityType, network)
	state.MulticastToUnicastConversionEnabled = types.BoolValue(multicastToUnicast)
	state.ClientIsolationEnabled = types.BoolValue(clientIsolation)
	state.HideName = types.BoolValue(hideName)
	state.UapsdEnabled = types.BoolValue(uapsdEnabled)
	state.BroadcastingFrequenciesGHz = rawMessagesToStringList(frequencies)
	state.MloEnabled = types.BoolValue(mloEnabled)
	state.BandSteeringEnabled = types.BoolValue(bandSteering)
	state.ArpProxyEnabled = types.BoolValue(arpProxy)
	state.BssTransitionEnabled = types.BoolValue(bssTransition)
	state.AdvertiseDeviceName = types.BoolValue(advertiseDeviceName)
	state.BasicDataRateKbpsByFrequencyGHz = basicDataRate
	state.ClientFilteringPolicy = clientFilteringPolicy
	state.BlackoutScheduleConfiguration = blackoutSchedule
	return state
}

func wifiStateFromIot(siteID, id, name, broadcastType string, enabled bool, securityType string, network *broadcasts.ClientAccess, multicastToUnicast bool, clientIsolation bool, hideName bool, uapsdEnabled bool, basicDataRate *wifiBasicDataRateModel, clientFilteringPolicy *wifiClientFilteringPolicyModel, blackoutSchedule *wifiBlackoutScheduleConfigurationModel) wifiResourceModel {
	state := wifiStateCommon(siteID, id, name, broadcastType, enabled, securityType, network)
	state.MulticastToUnicastConversionEnabled = types.BoolValue(multicastToUnicast)
	state.ClientIsolationEnabled = types.BoolValue(clientIsolation)
	state.HideName = types.BoolValue(hideName)
	state.UapsdEnabled = types.BoolValue(uapsdEnabled)
	state.BasicDataRateKbpsByFrequencyGHz = basicDataRate
	state.ClientFilteringPolicy = clientFilteringPolicy
	state.BlackoutScheduleConfiguration = blackoutSchedule
	return state
}

func readCreateWifiBasicDataRate(value *broadcasts.CreateWifiBroadcastBasicDataRateKbpsByFrequencyGHz) *wifiBasicDataRateModel {
	if value == nil {
		return nil
	}
	return &wifiBasicDataRateModel{
		Rate24Kbps: types.Int64Value(value.N24),
		Rate5Kbps:  types.Int64Value(value.N5),
	}
}

func readGetWifiBasicDataRate(value *broadcasts.GetWifiBroadcastDetailsBasicDataRateKbpsByFrequencyGHz) *wifiBasicDataRateModel {
	if value == nil {
		return nil
	}
	return &wifiBasicDataRateModel{
		Rate24Kbps: types.Int64Value(value.N24),
		Rate5Kbps:  types.Int64Value(value.N5),
	}
}

func readCreateWifiClientFilteringPolicy(value *broadcasts.CreateWifiBroadcastClientFilteringPolicy) *wifiClientFilteringPolicyModel {
	if value == nil {
		return nil
	}
	return &wifiClientFilteringPolicyModel{
		Action:           types.StringValue(value.Action),
		MacAddressFilter: rawMessagesToStringList(value.MacAddressFilter),
	}
}

func readGetWifiClientFilteringPolicy(value *broadcasts.GetWifiBroadcastDetailsClientFilteringPolicy) *wifiClientFilteringPolicyModel {
	if value == nil {
		return nil
	}
	return &wifiClientFilteringPolicyModel{
		Action:           types.StringValue(value.Action),
		MacAddressFilter: rawMessagesToStringList(value.MacAddressFilter),
	}
}

func readCreateWifiBlackoutSchedule(value *broadcasts.CreateWifiBroadcastBlackoutScheduleConfiguration) *wifiBlackoutScheduleConfigurationModel {
	if value == nil {
		return nil
	}
	days := make([]wifiBlackoutDayModel, 0, len(value.Days))
	for _, day := range value.Days {
		days = append(days, wifiBlackoutDayModel{
			Day:  stringValueOrNull(day.Day),
			Type: stringValueOrNull(day.Type),
		})
	}
	return &wifiBlackoutScheduleConfigurationModel{Days: days}
}

func readGetWifiBlackoutSchedule(value *broadcasts.GetWifiBroadcastDetailsBlackoutScheduleConfiguration) *wifiBlackoutScheduleConfigurationModel {
	if value == nil {
		return nil
	}
	days := make([]wifiBlackoutDayModel, 0, len(value.Days))
	for _, day := range value.Days {
		days = append(days, wifiBlackoutDayModel{
			Day:  stringValueOrNull(day.Day),
			Type: stringValueOrNull(day.Type),
		})
	}
	return &wifiBlackoutScheduleConfigurationModel{Days: days}
}

func rawMessagesToStringList(raw []json.RawMessage) types.List {
	if len(raw) == 0 {
		return types.ListNull(types.StringType)
	}
	values := make([]string, 0, len(raw))
	for _, item := range raw {
		var decoded string
		if err := json.Unmarshal(item, &decoded); err != nil {
			values = append(values, string(item))
			continue
		}
		values = append(values, decoded)
	}
	list, _ := types.ListValueFrom(context.Background(), types.StringType, values)
	return list
}
