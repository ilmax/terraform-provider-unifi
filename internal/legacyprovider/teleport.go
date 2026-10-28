package legacyprovider

import (
	"context"
	"fmt"
	"net"

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

type teleportModel struct {
	ID         types.String `tfsdk:"id"`
	SiteID     types.String `tfsdk:"site_id"`
	Enabled    types.Bool   `tfsdk:"enabled"`
	SubnetCIDR types.String `tfsdk:"subnet_cidr"`
}

func (m *teleportModel) set(siteID string, t legacy.Teleport) {
	m.ID = types.StringValue(t.ID)
	m.SiteID = types.StringValue(siteID)
	m.Enabled = types.BoolValue(t.Enabled)
	m.SubnetCIDR = types.StringValue(t.SubnetCIDR)
}

const teleportNotes = "\n\n~> **Temporary:** this uses the legacy controller API and moves to the `unifi` provider once the official Integrations API exposes Teleport."

// ---- resource ----

type teleportResource struct{ data *providerData }

func NewTeleportResource() resource.Resource { return &teleportResource{} }

func (r *teleportResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_teleport"
}

func (r *teleportResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the site's Teleport VPN setting. Teleport is a per-site singleton: creating this resource takes it over, " +
			"and destroying it only removes it from Terraform state (the setting is left as it is)." + teleportNotes,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Identifier of the controller's `teleport` setting document.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"site_id": schema.StringAttribute{
				MarkdownDescription: "Site identifier. Falls back to the provider `site_id`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown(), stringplanmodifier.RequiresReplace()},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether Teleport is enabled. When unset, the controller's current value is kept.",
				Optional:            true,
				Computed:            true,
			},
			"subnet_cidr": schema.StringAttribute{
				MarkdownDescription: "Subnet Teleport clients are assigned addresses from, in CIDR notation. Unset means the controller's default; an empty string is rejected (the controller cannot parse it). Setting it together with `enabled = false` is rejected. When unset, the controller's current value is kept.",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

// ValidateConfig rejects an invalid or empty subnet (an empty one is stored by
// the controller and breaks its network listing), and a subnet sent together
// with enabled = false: the controller echoes it in the PUT response but does
// not keep it, which would otherwise surface as a permanent diff.
func (r *teleportResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var cfg teleportModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if !cfg.SubnetCIDR.IsNull() && !cfg.SubnetCIDR.IsUnknown() {
		if _, _, err := net.ParseCIDR(cfg.SubnetCIDR.ValueString()); err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("subnet_cidr"), "Invalid subnet_cidr",
				"subnet_cidr must be a CIDR such as 192.168.99.0/24; an empty value would make the controller fail to list networks.")
			return
		}
	}
	if cfg.Enabled.IsNull() || cfg.Enabled.IsUnknown() || cfg.Enabled.ValueBool() {
		return
	}
	if !cfg.SubnetCIDR.IsNull() && !cfg.SubnetCIDR.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("subnet_cidr"), "subnet_cidr requires Teleport to be enabled",
			"The controller does not keep subnet_cidr when it is sent together with enabled = false. Set enabled = true or omit subnet_cidr.")
	}
}

func (r *teleportResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// apply merges the plan over the controller's current setting and writes it.
func (r *teleportResource) apply(ctx context.Context, plan *teleportModel, diags *diag.Diagnostics, siteID string) bool {
	cur, err := r.data.client.GetTeleport(ctx, siteID)
	if err != nil {
		diags.AddError("Unable to read Teleport setting", err.Error())
		return false
	}
	want := cur
	if !plan.Enabled.IsNull() && !plan.Enabled.IsUnknown() {
		want.Enabled = plan.Enabled.ValueBool()
	}
	if !plan.SubnetCIDR.IsNull() && !plan.SubnetCIDR.IsUnknown() {
		want.SubnetCIDR = plan.SubnetCIDR.ValueString()
	}
	if !want.Enabled {
		want.SubnetCIDR = "" // sent with enabled=false the controller drops it
	}
	if _, err := r.data.client.UpdateTeleport(ctx, siteID, cur.ID, want); err != nil {
		diags.AddError("Unable to update Teleport setting", err.Error())
		return false
	}
	// The PUT response echoes values the controller then discards; read back
	// what it actually stored.
	updated, err := r.data.client.GetTeleport(ctx, siteID)
	if err != nil {
		diags.AddError("Unable to read Teleport setting", err.Error())
		return false
	}
	plan.set(siteID, updated)
	return true
}

func (r *teleportResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan teleportModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	siteID, ok := siteid.Resolve(plan.SiteID, r.data.siteID, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !ok {
		return
	}
	if r.apply(ctx, &plan, &resp.Diagnostics, siteID) {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *teleportResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state teleportModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	siteID, ok := siteid.Resolve(state.SiteID, r.data.siteID, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !ok {
		return
	}
	t, err := r.data.client.GetTeleport(ctx, siteID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Teleport setting", err.Error())
		return
	}
	state.set(siteID, t)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *teleportResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan teleportModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	siteID, ok := siteid.Resolve(plan.SiteID, r.data.siteID, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !ok {
		return
	}
	if r.apply(ctx, &plan, &resp.Diagnostics, siteID) {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

// Delete only drops the resource from state: the setting is a per-site
// singleton that always exists on the controller.
func (r *teleportResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
}

// ImportState takes the site_id.
func (r *teleportResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site_id"), req.ID)...)
}

// ---- data source ----

type teleportDataSource struct{ data *providerData }

func NewTeleportDataSource() datasource.DataSource { return &teleportDataSource{} }

func (d *teleportDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_teleport"
}

func (d *teleportDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads the site's Teleport VPN setting." + teleportNotes,
		Attributes: map[string]dsschema.Attribute{
			"id":          dsschema.StringAttribute{MarkdownDescription: "Identifier of the controller's `teleport` setting document.", Computed: true},
			"site_id":     dsschema.StringAttribute{MarkdownDescription: "Site identifier. Falls back to the provider `site_id`.", Optional: true, Computed: true},
			"enabled":     dsschema.BoolAttribute{MarkdownDescription: "Whether Teleport is enabled.", Computed: true},
			"subnet_cidr": dsschema.StringAttribute{MarkdownDescription: "Subnet Teleport clients are assigned addresses from, in CIDR notation.", Computed: true},
		},
	}
}

func (d *teleportDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *teleportDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg teleportModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	siteID, ok := siteid.Resolve(cfg.SiteID, d.data.siteID, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !ok {
		return
	}
	t, err := d.data.client.GetTeleport(ctx, siteID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Teleport setting", err.Error())
		return
	}
	cfg.set(siteID, t)
	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}
