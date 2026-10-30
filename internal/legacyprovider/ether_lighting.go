package legacyprovider

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/legacy"
	"github.com/ilmax/terraform-provider-unifi/internal/siteid"
)

const etherLightingNotes = "\n\n~> **Temporary:** this uses the legacy controller API (the official Integrations API has no EtherLighting support yet). " +
	"The console only shows EtherLighting for switches that advertise it."

var colorPattern = regexp.MustCompile(`^[0-9A-Fa-f]{6}$`)

// ---- site settings ----

type etherLightingSettingsModel struct {
	ID               types.String `tfsdk:"id"`
	SiteID           types.String `tfsdk:"site_id"`
	SpeedOverrides   types.Map    `tfsdk:"speed_overrides"`
	NetworkOverrides types.Map    `tfsdk:"network_overrides"`
	SpeedDefaults    types.Map    `tfsdk:"speed_defaults"`
	NetworkDefaults  types.Map    `tfsdk:"network_defaults"`
}

func colorMap(entries []legacy.ColorEntry) types.Map {
	m := make(map[string]attr.Value, len(entries))
	for _, e := range entries {
		m[e.Key] = types.StringValue(e.RawColorHex)
	}
	return types.MapValueMust(types.StringType, m)
}

func colorEntries(m types.Map) []legacy.ColorEntry {
	out := []legacy.ColorEntry{}
	for k, v := range m.Elements() {
		out = append(out, legacy.ColorEntry{Key: k, RawColorHex: v.(types.String).ValueString()})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

func (m *etherLightingSettingsModel) set(siteID string, s legacy.EtherLightingSettings) {
	m.ID = types.StringValue(s.ID)
	m.SiteID = types.StringValue(siteID)
	m.SpeedOverrides = colorMap(s.SpeedOverrides)
	m.NetworkOverrides = colorMap(s.NetworkOverrides)
	m.SpeedDefaults = colorMap(s.SpeedDefaults)
	m.NetworkDefaults = colorMap(s.NetworkDefaults)
}

type etherLightingSettingsResource struct{ data *providerData }

func NewEtherLightingSettingsResource() resource.Resource { return &etherLightingSettingsResource{} }

func (r *etherLightingSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ether_lighting_settings"
}

func (r *etherLightingSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	colors := func(desc string) schema.MapAttribute {
		return schema.MapAttribute{MarkdownDescription: desc, ElementType: types.StringType, Optional: true, Computed: true}
	}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the site-wide EtherLighting colors. This is a per-site singleton: the controller owns the default colors and this resource owns the overrides. " +
			"Destroying it clears the overrides." + etherLightingNotes,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Identifier of the controller's `ether_lighting` setting document.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"site_id": schema.StringAttribute{
				MarkdownDescription: "Site identifier. Falls back to the provider `site_id`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown(), stringplanmodifier.RequiresReplace()},
			},
			"speed_overrides":   colors("Color per link speed (`FE`, `GbE`, `2.5GbE`, `5GbE`, `10GbE`, `25GbE`, `40GbE`, `100GbE`) as six hex digits, used when a switch's mode is `speed`. When unset, the controller's current overrides are kept."),
			"network_overrides": colors("Color per legacy network id as six hex digits, used when a switch's mode is `network`. When unset, the controller's current overrides are kept."),
			"speed_defaults":    schema.MapAttribute{MarkdownDescription: "Controller default color per link speed.", ElementType: types.StringType, Computed: true},
			"network_defaults":  schema.MapAttribute{MarkdownDescription: "Controller default color per network.", ElementType: types.StringType, Computed: true},
		},
	}
}

func (r *etherLightingSettingsResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var cfg etherLightingSettingsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	for name, m := range map[string]types.Map{"speed_overrides": cfg.SpeedOverrides, "network_overrides": cfg.NetworkOverrides} {
		if m.IsNull() || m.IsUnknown() {
			continue
		}
		for k, v := range m.Elements() {
			if s, ok := v.(types.String); ok && !s.IsUnknown() && !colorPattern.MatchString(s.ValueString()) {
				resp.Diagnostics.AddAttributeError(path.Root(name).AtMapKey(k), "Invalid color", "A color must be six hex digits such as 0544ff.")
			}
		}
	}
}

func (r *etherLightingSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*providerData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("got %T", req.ProviderData))
		return
	}
	r.data = data
}

func (r *etherLightingSettingsResource) apply(ctx context.Context, plan *etherLightingSettingsModel, diags *diag.Diagnostics) bool {
	site, ok := siteid.Resolve(plan.SiteID, r.data.siteID, diags)
	if diags.HasError() || !ok {
		return false
	}
	cur, err := r.data.client.GetEtherLightingSettings(ctx, site)
	if err != nil {
		diags.AddError("Unable to read EtherLighting settings", err.Error())
		return false
	}
	speed, network := cur.SpeedOverrides, cur.NetworkOverrides
	if !plan.SpeedOverrides.IsNull() && !plan.SpeedOverrides.IsUnknown() {
		speed = colorEntries(plan.SpeedOverrides)
	}
	if !plan.NetworkOverrides.IsNull() && !plan.NetworkOverrides.IsUnknown() {
		network = colorEntries(plan.NetworkOverrides)
	}
	if _, err := r.data.client.UpdateEtherLightingOverrides(ctx, site, cur.ID, speed, network); err != nil {
		diags.AddError("Unable to update EtherLighting settings", err.Error())
		return false
	}
	// Read back what the controller stored rather than trusting the PUT echo.
	updated, err := r.data.client.GetEtherLightingSettings(ctx, site)
	if err != nil {
		diags.AddError("Unable to read EtherLighting settings", err.Error())
		return false
	}
	plan.set(site, updated)
	return true
}

func (r *etherLightingSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan etherLightingSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if r.apply(ctx, &plan, &resp.Diagnostics) {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *etherLightingSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan etherLightingSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if r.apply(ctx, &plan, &resp.Diagnostics) {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *etherLightingSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state etherLightingSettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	site, ok := siteid.Resolve(state.SiteID, r.data.siteID, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !ok {
		return
	}
	s, err := r.data.client.GetEtherLightingSettings(ctx, site)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read EtherLighting settings", err.Error())
		return
	}
	state.set(site, s)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *etherLightingSettingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state etherLightingSettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	site, ok := siteid.Resolve(state.SiteID, r.data.siteID, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !ok {
		return
	}
	if _, err := r.data.client.UpdateEtherLightingOverrides(ctx, site, state.ID.ValueString(), nil, nil); err != nil {
		resp.Diagnostics.AddError("Unable to clear EtherLighting overrides", err.Error())
	}
}

// ImportState takes the site_id.
func (r *etherLightingSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site_id"), req.ID)...)
}

// ---- per-device ----

var (
	ledModes  = []string{"standard", "etherlighting"}
	elModes   = []string{"speed", "network"}
	behaviors = []string{"breath", "steady"}
)

type deviceEtherLightingModel struct {
	ID         types.String `tfsdk:"id"`
	SiteID     types.String `tfsdk:"site_id"`
	MAC        types.String `tfsdk:"mac"`
	LEDMode    types.String `tfsdk:"led_mode"`
	Mode       types.String `tfsdk:"mode"`
	Brightness types.Int64  `tfsdk:"brightness"`
	Behavior   types.String `tfsdk:"behavior"`
}

func (m deviceEtherLightingModel) toAPI() legacy.DeviceEtherLighting {
	return legacy.DeviceEtherLighting{
		LEDMode: m.LEDMode.ValueString(), Mode: m.Mode.ValueString(),
		Brightness: int(m.Brightness.ValueInt64()), Behavior: m.Behavior.ValueString(),
	}
}

func (m *deviceEtherLightingModel) set(siteID string, d legacy.Device) {
	m.ID = types.StringValue(d.ID)
	m.SiteID = types.StringValue(siteID)
	m.MAC = types.StringValue(strings.ToLower(d.MAC))
	if el := d.EtherLighting; el != nil {
		m.LEDMode = types.StringValue(el.LEDMode)
		m.Mode = types.StringValue(el.Mode)
		m.Brightness = types.Int64Value(int64(el.Brightness))
		m.Behavior = types.StringValue(el.Behavior)
	}
}

type deviceEtherLightingResource struct{ data *providerData }

func NewDeviceEtherLightingResource() resource.Resource { return &deviceEtherLightingResource{} }

func (r *deviceEtherLightingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_ether_lighting"
}

func (r *deviceEtherLightingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the EtherLighting configuration of one adopted switch. Destroying it switches the device back to `standard` LED mode and leaves the other values as they are." + etherLightingNotes,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Identifier of the device on the controller.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"site_id": schema.StringAttribute{
				MarkdownDescription: "Site identifier. Falls back to the provider `site_id`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown(), stringplanmodifier.RequiresReplace()},
			},
			"mac": schema.StringAttribute{
				MarkdownDescription: "MAC address of the adopted switch, lowercase with colons.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"led_mode": schema.StringAttribute{
				MarkdownDescription: "`standard` (regular port LEDs) or `etherlighting`.",
				Required:            true,
			},
			"mode": schema.StringAttribute{
				MarkdownDescription: "What the color follows: `speed` (link speed) or `network` (the port's network). Defaults to `speed`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("speed"),
			},
			"brightness": schema.Int64Attribute{
				MarkdownDescription: "Brightness from 1 to 100. Defaults to 100.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(100),
			},
			"behavior": schema.StringAttribute{
				MarkdownDescription: "`breath` or `steady`. Defaults to `steady`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("steady"),
			},
		},
	}
}

func oneOf(v types.String, name string, allowed []string, diags *diag.Diagnostics) {
	if v.IsNull() || v.IsUnknown() {
		return
	}
	for _, a := range allowed {
		if v.ValueString() == a {
			return
		}
	}
	diags.AddAttributeError(path.Root(name), "Invalid "+name, fmt.Sprintf("%s must be one of: %s.", name, strings.Join(allowed, ", ")))
}

func (r *deviceEtherLightingResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var cfg deviceEtherLightingModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	checkMAC(cfg.MAC, &resp.Diagnostics)
	oneOf(cfg.LEDMode, "led_mode", ledModes, &resp.Diagnostics)
	oneOf(cfg.Mode, "mode", elModes, &resp.Diagnostics)
	oneOf(cfg.Behavior, "behavior", behaviors, &resp.Diagnostics)
	if !cfg.Brightness.IsNull() && !cfg.Brightness.IsUnknown() && (cfg.Brightness.ValueInt64() < 1 || cfg.Brightness.ValueInt64() > 100) {
		resp.Diagnostics.AddAttributeError(path.Root("brightness"), "Invalid brightness", "brightness must be between 1 and 100.")
	}
}

func (r *deviceEtherLightingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*providerData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("got %T", req.ProviderData))
		return
	}
	r.data = data
}

func (r *deviceEtherLightingResource) apply(ctx context.Context, plan *deviceEtherLightingModel, diags *diag.Diagnostics) bool {
	site, ok := siteid.Resolve(plan.SiteID, r.data.siteID, diags)
	if diags.HasError() || !ok {
		return false
	}
	dev, err := r.data.client.FindDevice(ctx, site, plan.MAC.ValueString())
	if err != nil {
		diags.AddError("Unable to look up the device", err.Error())
		return false
	}
	if dev == nil {
		diags.AddAttributeError(path.Root("mac"), "Device not found", fmt.Sprintf("no adopted device with MAC %s on this site", plan.MAC.ValueString()))
		return false
	}
	if _, err := r.data.client.SetDeviceEtherLighting(ctx, site, dev.ID, plan.toAPI()); err != nil {
		diags.AddError("Unable to update EtherLighting", err.Error())
		return false
	}
	updated, err := r.data.client.FindDevice(ctx, site, plan.MAC.ValueString())
	if err != nil || updated == nil {
		diags.AddError("Unable to read the device back", fmt.Sprint(err))
		return false
	}
	plan.set(site, *updated)
	return true
}

func (r *deviceEtherLightingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan deviceEtherLightingModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if r.apply(ctx, &plan, &resp.Diagnostics) {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *deviceEtherLightingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan deviceEtherLightingModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if r.apply(ctx, &plan, &resp.Diagnostics) {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *deviceEtherLightingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state deviceEtherLightingModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	site, ok := siteid.Resolve(state.SiteID, r.data.siteID, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !ok {
		return
	}
	dev, err := r.data.client.FindDevice(ctx, site, state.MAC.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read the device", err.Error())
		return
	}
	if dev == nil || dev.EtherLighting == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	state.set(site, *dev)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *deviceEtherLightingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state deviceEtherLightingModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	site, ok := siteid.Resolve(state.SiteID, r.data.siteID, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !ok {
		return
	}
	el := state.toAPI()
	el.LEDMode = "standard"
	if _, err := r.data.client.SetDeviceEtherLighting(ctx, site, state.ID.ValueString(), el); err != nil {
		resp.Diagnostics.AddError("Unable to reset EtherLighting", err.Error())
	}
}

// ImportState takes "<site_id>/<mac>".
func (r *deviceEtherLightingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	site, mac, ok := strings.Cut(req.ID, "/")
	if !ok || site == "" || mac == "" {
		resp.Diagnostics.AddError("Invalid import ID", "expected <site_id>/<mac>")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site_id"), site)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("mac"), strings.ToLower(mac))...)
}
