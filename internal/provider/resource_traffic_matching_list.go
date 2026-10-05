package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/sdkcompat/trafficmatching"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
	"github.com/ilmax/unifi-client-go/pkg/errors"
)

type trafficMatchingListResource struct {
	client *unifi.Client
	siteID string
}

type trafficMatchingListResourceModel struct {
	ID        types.String `tfsdk:"id"`
	SiteID    types.String `tfsdk:"site_id"`
	Name      types.String `tfsdk:"name"`
	Type      types.String `tfsdk:"type"`
	ItemsJSON types.String `tfsdk:"items_json"`
}

func NewTrafficMatchingListResource() resource.Resource {
	return &trafficMatchingListResource{}
}

func (r *trafficMatchingListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_traffic_matching_list"
}

func (r *trafficMatchingListResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A reusable named list of IPv4 addresses, IPv6 addresses, or ports, referenced from " +
			"traffic filters (e.g. unifi_firewall_policy's source_traffic_filter_json/destination_traffic_filter_json) " +
			"instead of repeating the same match criteria inline. Each entry in the list is itself a small " +
			"discriminated union (a single address, an address range, or a subnet for IPv4; a single address or " +
			"subnet for IPv6; a single port or a port range for PORTS), so items_json takes them as a raw JSON " +
			"array string — the same opaque-JSON escape hatch used elsewhere in this provider for deeply nested " +
			"discriminated unions — rather than modeling every item variant as native attributes.",
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
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "IPV4_ADDRESSES, IPV6_ADDRESSES, or PORTS. Changing this requires replacement, since it changes the shape every item in items_json must match.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"items_json": schema.StringAttribute{
				Required: true,
				Description: "Raw JSON array of items matching the API's shape for the chosen type, e.g. " +
					`[{"type":"IP_ADDRESS","value":"192.168.1.5"},{"type":"SUBNET","value":"10.0.0.0/24"}] for ` +
					"IPV4_ADDRESSES, or [{\"type\":\"PORT_NUMBER\",\"value\":443}] for PORTS.",
			},
		},
	}
}

func (r *trafficMatchingListResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *trafficMatchingListResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan trafficMatchingListResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(plan.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	payload, diags := buildTrafficMatchingListPayload(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result trafficmatching.List
	apiPath := fmt.Sprintf("/v1/sites/%s/traffic-matching-lists", siteID)
	if err := r.client.Post(ctx, apiPath, payload, &result); err != nil {
		resp.Diagnostics.AddError("Unable to create traffic matching list", err.Error())
		return
	}

	state := trafficMatchingListStateFromAPI(siteID, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *trafficMatchingListResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state trafficMatchingListResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(state.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	if state.ID.IsNull() || state.ID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("id"), "Missing traffic matching list ID", "The traffic matching list ID is required to read the resource.")
		return
	}

	var result trafficmatching.List
	apiPath := fmt.Sprintf("/v1/sites/%s/traffic-matching-lists/%s", siteID, state.ID.ValueString())
	if err := r.client.Get(ctx, apiPath, &result); err != nil {
		if errors.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read traffic matching list", err.Error())
		return
	}

	updated := trafficMatchingListStateFromAPI(siteID, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &updated)...)
}

func (r *trafficMatchingListResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan trafficMatchingListResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(plan.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	if plan.ID.IsNull() || plan.ID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("id"), "Missing traffic matching list ID", "The traffic matching list ID is required to update the resource.")
		return
	}

	payload, diags := buildTrafficMatchingListPayload(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result trafficmatching.List
	apiPath := fmt.Sprintf("/v1/sites/%s/traffic-matching-lists/%s", siteID, plan.ID.ValueString())
	if err := r.client.Put(ctx, apiPath, payload, &result); err != nil {
		resp.Diagnostics.AddError("Unable to update traffic matching list", err.Error())
		return
	}

	state := trafficMatchingListStateFromAPI(siteID, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *trafficMatchingListResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state trafficMatchingListResourceModel
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

	apiPath := fmt.Sprintf("/v1/sites/%s/traffic-matching-lists/%s", siteID, state.ID.ValueString())
	if err := r.client.Delete(ctx, apiPath, nil); err != nil && !errors.IsNotFoundError(err) {
		resp.Diagnostics.AddError("Unable to delete traffic matching list", err.Error())
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *trafficMatchingListResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	siteID, listID, err := splitImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import identifier", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site_id"), siteID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), listID)...)
}

func buildTrafficMatchingListPayload(plan trafficMatchingListResourceModel) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics

	items, itemsDiags := decodeJSONRaw(plan.ItemsJSON, path.Root("items_json"))
	diags.Append(itemsDiags...)
	if diags.HasError() {
		return nil, diags
	}

	listType := strings.ToUpper(plan.Type.ValueString())
	switch listType {
	case "IPV4_ADDRESSES", "IPV6_ADDRESSES", "PORTS":
	default:
		diags.AddAttributeError(path.Root("type"), "Invalid type", "Supported values are IPV4_ADDRESSES, IPV6_ADDRESSES, or PORTS.")
		return nil, diags
	}

	payload := map[string]any{
		"name": plan.Name.ValueString(),
		"type": listType,
	}
	if items != nil {
		payload["items"] = items
	}

	return payload, diags
}

func trafficMatchingListStateFromAPI(siteID string, l trafficmatching.List) trafficMatchingListResourceModel {
	return trafficMatchingListResourceModel{
		ID:        types.StringValue(l.Id),
		SiteID:    types.StringValue(siteID),
		Name:      types.StringValue(l.Name),
		Type:      types.StringValue(l.Type),
		ItemsJSON: normalizeJSONString(l.Items),
	}
}
