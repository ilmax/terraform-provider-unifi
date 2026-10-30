package legacyprovider

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/legacy"
	"github.com/ilmax/terraform-provider-unifi/internal/siteid"
)

const ipReservationNotes = "\n\n~> **Temporary:** this uses the legacy controller API (the official Integrations API has no fixed-IP support yet)."

var macPattern = regexp.MustCompile(`^([0-9a-f]{2}:){5}[0-9a-f]{2}$`)

// checkMAC requires the lowercase colon form the controller stores, so the
// config never differs from the state only by case.
func checkMAC(mac types.String, diags *diag.Diagnostics) {
	if mac.IsNull() || mac.IsUnknown() || macPattern.MatchString(mac.ValueString()) {
		return
	}
	diags.AddAttributeError(path.Root("mac"), "Invalid mac", "mac must be a lowercase MAC address such as 02:00:00:aa:bb:01.")
}

type ipReservationModel struct {
	ID             types.String `tfsdk:"id"`
	SiteID         types.String `tfsdk:"site_id"`
	MAC            types.String `tfsdk:"mac"`
	IPAddress      types.String `tfsdk:"ip_address"`
	NetworkID      types.String `tfsdk:"network_id"`
	Name           types.String `tfsdk:"name"`
	Note           types.String `tfsdk:"note"`
	LocalDNSRecord types.String `tfsdk:"local_dns_record"`
}

func (m *ipReservationModel) set(siteID string, r legacy.Reservation) {
	m.ID = types.StringValue(r.ID)
	m.SiteID = types.StringValue(siteID)
	m.MAC = types.StringValue(r.MAC)
	m.IPAddress = types.StringValue(r.FixedIP)
	m.NetworkID = types.StringValue(r.NetworkID)
	m.Name = types.StringValue(r.Name)
	m.Note = types.StringValue(r.Note)
	if r.LocalDNSRecordEnabled {
		m.LocalDNSRecord = types.StringValue(r.LocalDNSRecord)
	} else {
		m.LocalDNSRecord = types.StringNull()
	}
}

func (m ipReservationModel) toAPI() legacy.Reservation {
	r := legacy.Reservation{
		MAC:       m.MAC.ValueString(),
		FixedIP:   m.IPAddress.ValueString(),
		NetworkID: m.NetworkID.ValueString(),
		Name:      m.Name.ValueString(),
		Note:      m.Note.ValueString(),
	}
	if !m.LocalDNSRecord.IsNull() && !m.LocalDNSRecord.IsUnknown() {
		r.LocalDNSRecordEnabled = true
		r.LocalDNSRecord = m.LocalDNSRecord.ValueString()
	}
	return r
}

// ---- resource ----

type ipReservationResource struct{ data *providerData }

func NewIPReservationResource() resource.Resource { return &ipReservationResource{} }

func (r *ipReservationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_reservation"
}

func (r *ipReservationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reserves a fixed IP address for a client MAC (a DHCP reservation). The MAC does not need to have connected before. " +
			"If the console already knows the client, its record is updated. Destroying the resource releases the fixed IP but keeps the client record and its history." + ipReservationNotes,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Identifier of the client record on the controller.",
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
				MarkdownDescription: "Client MAC address, lowercase with colons.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"ip_address": schema.StringAttribute{
				MarkdownDescription: "Fixed IPv4 address. It must be inside the network's subnet and not reserved by another client.",
				Required:            true,
			},
			"network_id": schema.StringAttribute{
				MarkdownDescription: "Identifier of the legacy network (LAN) the address belongs to.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Client name. When unset, the controller's current value is kept.",
				Optional:            true,
				Computed:            true,
			},
			"note": schema.StringAttribute{
				MarkdownDescription: "Free-form note. When unset, the controller's current value is kept.",
				Optional:            true,
				Computed:            true,
			},
			"local_dns_record": schema.StringAttribute{
				MarkdownDescription: "Local DNS name the controller serves for this client. Unset disables the record.",
				Optional:            true,
			},
		},
	}
}

func (r *ipReservationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ipReservationResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var cfg ipReservationModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	checkMAC(cfg.MAC, &resp.Diagnostics)
}

func (r *ipReservationResource) save(ctx context.Context, plan *ipReservationModel, diags *diag.Diagnostics) bool {
	site, ok := siteid.Resolve(plan.SiteID, r.data.siteID, diags)
	if diags.HasError() || !ok {
		return false
	}
	saved, err := r.data.client.SaveReservation(ctx, site, plan.toAPI())
	if err != nil {
		diags.AddError("Unable to save IP reservation", err.Error())
		return false
	}
	plan.set(site, saved)
	return true
}

func (r *ipReservationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ipReservationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if r.save(ctx, &plan, &resp.Diagnostics) {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *ipReservationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ipReservationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	site, ok := siteid.Resolve(state.SiteID, r.data.siteID, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !ok {
		return
	}
	got, err := r.data.client.GetReservation(ctx, site, state.MAC.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read IP reservation", err.Error())
		return
	}
	if got == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	state.set(site, *got)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ipReservationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ipReservationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if r.save(ctx, &plan, &resp.Diagnostics) {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *ipReservationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ipReservationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	site, ok := siteid.Resolve(state.SiteID, r.data.siteID, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !ok {
		return
	}
	if err := r.data.client.ReleaseReservation(ctx, site, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to release IP reservation", err.Error())
	}
}

// ImportState takes "<site_id>/<mac>".
func (r *ipReservationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	site, mac, ok := strings.Cut(req.ID, "/")
	if !ok || site == "" || mac == "" {
		resp.Diagnostics.AddError("Invalid import ID", "expected <site_id>/<mac>")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site_id"), site)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("mac"), strings.ToLower(mac))...)
}

// ---- data source ----

type ipReservationDataSource struct{ data *providerData }

func NewIPReservationDataSource() datasource.DataSource { return &ipReservationDataSource{} }

func (d *ipReservationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_reservation"
}

func (d *ipReservationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads the fixed-IP reservation of a client MAC." + ipReservationNotes,
		Attributes: map[string]dsschema.Attribute{
			"id":               dsschema.StringAttribute{MarkdownDescription: "Identifier of the client record on the controller.", Computed: true},
			"site_id":          dsschema.StringAttribute{MarkdownDescription: "Site identifier. Falls back to the provider `site_id`.", Optional: true, Computed: true},
			"mac":              dsschema.StringAttribute{MarkdownDescription: "Client MAC address, lowercase with colons.", Required: true},
			"ip_address":       dsschema.StringAttribute{MarkdownDescription: "Fixed IPv4 address.", Computed: true},
			"network_id":       dsschema.StringAttribute{MarkdownDescription: "Identifier of the legacy network the address belongs to.", Computed: true},
			"name":             dsschema.StringAttribute{MarkdownDescription: "Client name.", Computed: true},
			"note":             dsschema.StringAttribute{MarkdownDescription: "Free-form note.", Computed: true},
			"local_dns_record": dsschema.StringAttribute{MarkdownDescription: "Local DNS name served for this client; null when disabled.", Computed: true},
		},
	}
}

func (d *ipReservationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*providerData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("got %T", req.ProviderData))
		return
	}
	d.data = data
}

func (d *ipReservationDataSource) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var cfg ipReservationModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	checkMAC(cfg.MAC, &resp.Diagnostics)
}

func (d *ipReservationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg ipReservationModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	site, ok := siteid.Resolve(cfg.SiteID, d.data.siteID, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !ok {
		return
	}
	got, err := d.data.client.GetReservation(ctx, site, cfg.MAC.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read IP reservation", err.Error())
		return
	}
	if got == nil {
		resp.Diagnostics.AddError("IP reservation not found", fmt.Sprintf("no fixed-IP reservation for %s", cfg.MAC.ValueString()))
		return
	}
	cfg.set(site, *got)
	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}
