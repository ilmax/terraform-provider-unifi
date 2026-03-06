package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/sdkcompat/zones"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
	"github.com/ilmax/unifi-client-go/pkg/errors"
)

type firewallZoneNetworksResource struct {
	client *unifi.Client
	siteID string
}

type firewallZoneNetworksResourceModel struct {
	ID         types.String `tfsdk:"id"`
	SiteID     types.String `tfsdk:"site_id"`
	ZoneID     types.String `tfsdk:"zone_id"`
	NetworkIDs types.List   `tfsdk:"network_ids"`
}

func NewFirewallZoneNetworksResource() resource.Resource {
	return &firewallZoneNetworksResource{}
}

func (r *firewallZoneNetworksResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_zone_networks"
}

func (r *firewallZoneNetworksResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"zone_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"network_ids": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *firewallZoneNetworksResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *firewallZoneNetworksResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan firewallZoneNetworksResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state, diags := r.applyZoneNetworks(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *firewallZoneNetworksResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state firewallZoneNetworksResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(state.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	if state.ZoneID.IsNull() || state.ZoneID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("zone_id"), "Missing zone ID", "The zone_id is required to read firewall zone networks.")
		return
	}

	apiPath := fmt.Sprintf("/v1/sites/%s/firewall/zones/%s", siteID, state.ZoneID.ValueString())
	var zone zones.GetFirewallZoneResponse
	if err := r.client.Get(ctx, apiPath, &zone); err != nil {
		if errors.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read firewall zone", err.Error())
		return
	}

	state = firewallZoneNetworksState(siteID, zone.Id, zone.NetworkIds)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *firewallZoneNetworksResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan firewallZoneNetworksResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state, diags := r.applyZoneNetworks(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *firewallZoneNetworksResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *firewallZoneNetworksResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	siteID, zoneID, err := splitImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import identifier", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site_id"), siteID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("zone_id"), zoneID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), firewallZoneNetworksID(siteID, zoneID))...)
}

func (r *firewallZoneNetworksResource) applyZoneNetworks(ctx context.Context, plan firewallZoneNetworksResourceModel) (firewallZoneNetworksResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	siteID, ok := resolveSiteID(plan.SiteID, r.siteID, &diags)
	if !ok {
		return firewallZoneNetworksResourceModel{}, diags
	}

	if plan.ZoneID.IsNull() || plan.ZoneID.IsUnknown() || plan.ZoneID.ValueString() == "" {
		diags.AddAttributeError(path.Root("zone_id"), "Missing zone ID", "zone_id must be provided.")
		return firewallZoneNetworksResourceModel{}, diags
	}
	zoneID := plan.ZoneID.ValueString()

	networkIDs, listDiags := listToRawMessages(plan.NetworkIDs, path.Root("network_ids"))
	diags.Append(listDiags...)
	if diags.HasError() {
		return firewallZoneNetworksResourceModel{}, diags
	}

	getPath := fmt.Sprintf("/v1/sites/%s/firewall/zones/%s", siteID, zoneID)
	var current zones.GetFirewallZoneResponse
	if err := r.client.Get(ctx, getPath, &current); err != nil {
		if errors.IsNotFoundError(err) {
			diags.AddError("Firewall zone not found", fmt.Sprintf("Firewall zone %q was not found in site %q.", zoneID, siteID))
			return firewallZoneNetworksResourceModel{}, diags
		}
		diags.AddError("Unable to read firewall zone", err.Error())
		return firewallZoneNetworksResourceModel{}, diags
	}

	updatePath := fmt.Sprintf("/v1/sites/%s/firewall/zones/%s", siteID, zoneID)
	payload := zones.UpdateFirewallZoneRequest{
		Name:       current.Name,
		NetworkIds: networkIDs,
	}
	var result zones.UpdateFirewallZoneResponse
	if err := r.client.Put(ctx, updatePath, payload, &result); err != nil {
		diags.AddError("Unable to update firewall zone networks", err.Error())
		return firewallZoneNetworksResourceModel{}, diags
	}

	return firewallZoneNetworksState(siteID, result.Id, result.NetworkIds), diags
}

func firewallZoneNetworksState(siteID, zoneID string, networkIDs []json.RawMessage) firewallZoneNetworksResourceModel {
	return firewallZoneNetworksResourceModel{
		ID:         types.StringValue(firewallZoneNetworksID(siteID, zoneID)),
		SiteID:     types.StringValue(siteID),
		ZoneID:     types.StringValue(zoneID),
		NetworkIDs: rawMessagesToStringList(networkIDs),
	}
}

func firewallZoneNetworksID(siteID, zoneID string) string {
	return fmt.Sprintf("%s/%s", siteID, zoneID)
}
