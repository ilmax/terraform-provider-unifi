package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/sdkcompat/vouchers"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
	"github.com/ilmax/unifi-client-go/pkg/errors"
)

type hotspotVoucherResource struct {
	client *unifi.Client
	siteID string
}

type hotspotVoucherResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	SiteID               types.String `tfsdk:"site_id"`
	Name                 types.String `tfsdk:"name"`
	TimeLimitMinutes     types.Int64  `tfsdk:"time_limit_minutes"`
	AuthorizedGuestLimit types.Int64  `tfsdk:"authorized_guest_limit"`
	DataUsageLimitMBytes types.Int64  `tfsdk:"data_usage_limit_mbytes"`
	RxRateLimitKbps      types.Int64  `tfsdk:"rx_rate_limit_kbps"`
	TxRateLimitKbps      types.Int64  `tfsdk:"tx_rate_limit_kbps"`
	Code                 types.String `tfsdk:"code"`
	AuthorizedGuestCount types.Int64  `tfsdk:"authorized_guest_count"`
	CreatedAt            types.String `tfsdk:"created_at"`
	ActivatedAt          types.String `tfsdk:"activated_at"`
	ExpiresAt            types.String `tfsdk:"expires_at"`
	Expired              types.Bool   `tfsdk:"expired"`
}

func NewHotspotVoucherResource() resource.Resource {
	return &hotspotVoucherResource{}
}

func (r *hotspotVoucherResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hotspot_voucher"
}

func (r *hotspotVoucherResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A UniFi Hotspot guest WiFi voucher. The API has no update endpoint for vouchers " +
			"(they're immutable once created — matching how a printed/issued voucher code works in practice), " +
			"so every attribute forces replacement.",
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
				Required:    true,
				Description: "Voucher note. May be duplicated across multiple vouchers — it's a label, not an identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"time_limit_minutes": schema.Int64Attribute{
				Required:    true,
				Description: "How long, in minutes, the voucher grants access starting from when the first guest authorizes with it.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"authorized_guest_limit": schema.Int64Attribute{
				Optional:    true,
				Description: "Limit on how many different guests can authorize with this same voucher.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"data_usage_limit_mbytes": schema.Int64Attribute{
				Optional:    true,
				Description: "Data usage limit in megabytes.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"rx_rate_limit_kbps": schema.Int64Attribute{
				Optional:    true,
				Description: "Download rate limit in kilobits per second.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"tx_rate_limit_kbps": schema.Int64Attribute{
				Optional:    true,
				Description: "Upload rate limit in kilobits per second.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"code": schema.StringAttribute{
				Computed:    true,
				Description: "Secret code guests use to activate the voucher on the Hotspot portal.",
			},
			"authorized_guest_count": schema.Int64Attribute{
				Computed:    true,
				Description: "How many guests have used this voucher to authorize network access so far.",
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"activated_at": schema.StringAttribute{
				Computed:    true,
				Description: "When the first guest authorized with this voucher. Null until then.",
			},
			"expires_at": schema.StringAttribute{
				Computed:    true,
				Description: "When the voucher stops granting access. Null until activated.",
			},
			"expired": schema.BoolAttribute{
				Computed: true,
			},
		},
	}
}

func (r *hotspotVoucherResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *hotspotVoucherResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan hotspotVoucherResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(plan.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	payload := vouchers.CreateVoucherRequest{
		Name:                 plan.Name.ValueString(),
		Count:                1,
		TimeLimitMinutes:     plan.TimeLimitMinutes.ValueInt64(),
		AuthorizedGuestLimit: plan.AuthorizedGuestLimit.ValueInt64(),
		DataUsageLimitMBytes: plan.DataUsageLimitMBytes.ValueInt64(),
		RxRateLimitKbps:      plan.RxRateLimitKbps.ValueInt64(),
		TxRateLimitKbps:      plan.TxRateLimitKbps.ValueInt64(),
	}

	var result vouchers.CreateVoucherResponse
	apiPath := fmt.Sprintf("/v1/sites/%s/hotspot/vouchers", siteID)
	if err := r.client.Post(ctx, apiPath, payload, &result); err != nil {
		resp.Diagnostics.AddError("Unable to create hotspot voucher", err.Error())
		return
	}
	if len(result.Vouchers) != 1 {
		resp.Diagnostics.AddError("Unable to create hotspot voucher", fmt.Sprintf("expected exactly 1 voucher in response, got %d", len(result.Vouchers)))
		return
	}

	state := hotspotVoucherStateFromAPI(siteID, result.Vouchers[0])
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *hotspotVoucherResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state hotspotVoucherResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(state.SiteID, r.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	if state.ID.IsNull() || state.ID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("id"), "Missing voucher ID", "The voucher ID is required to read the resource.")
		return
	}

	var result vouchers.VoucherDetails
	apiPath := fmt.Sprintf("/v1/sites/%s/hotspot/vouchers/%s", siteID, state.ID.ValueString())
	if err := r.client.Get(ctx, apiPath, &result); err != nil {
		if errors.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read hotspot voucher", err.Error())
		return
	}

	updated := hotspotVoucherStateFromAPI(siteID, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &updated)...)
}

func (r *hotspotVoucherResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Hotspot voucher updates are not supported",
		"The API has no update endpoint for vouchers — every attribute forces replacement, so this should be unreachable.",
	)
}

func (r *hotspotVoucherResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state hotspotVoucherResourceModel
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

	apiPath := fmt.Sprintf("/v1/sites/%s/hotspot/vouchers/%s", siteID, state.ID.ValueString())
	if err := r.client.Delete(ctx, apiPath, nil); err != nil && !errors.IsNotFoundError(err) {
		resp.Diagnostics.AddError("Unable to delete hotspot voucher", err.Error())
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *hotspotVoucherResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	siteID, voucherID, err := splitImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import identifier", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site_id"), siteID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), voucherID)...)
}

func hotspotVoucherStateFromAPI(siteID string, v vouchers.VoucherDetails) hotspotVoucherResourceModel {
	state := hotspotVoucherResourceModel{
		ID:                   types.StringValue(v.Id),
		SiteID:               types.StringValue(siteID),
		Name:                 types.StringValue(v.Name),
		TimeLimitMinutes:     types.Int64Value(v.TimeLimitMinutes),
		Code:                 types.StringValue(v.Code),
		AuthorizedGuestCount: types.Int64Value(v.AuthorizedGuestCount),
		CreatedAt:            types.StringValue(v.CreatedAt.UTC().Format(time.RFC3339)),
		Expired:              types.BoolValue(v.Expired),
		ActivatedAt:          types.StringNull(),
		ExpiresAt:            types.StringNull(),
	}

	if v.AuthorizedGuestLimit > 0 {
		state.AuthorizedGuestLimit = types.Int64Value(v.AuthorizedGuestLimit)
	} else {
		state.AuthorizedGuestLimit = types.Int64Null()
	}
	if v.DataUsageLimitMBytes > 0 {
		state.DataUsageLimitMBytes = types.Int64Value(v.DataUsageLimitMBytes)
	} else {
		state.DataUsageLimitMBytes = types.Int64Null()
	}
	if v.RxRateLimitKbps > 0 {
		state.RxRateLimitKbps = types.Int64Value(v.RxRateLimitKbps)
	} else {
		state.RxRateLimitKbps = types.Int64Null()
	}
	if v.TxRateLimitKbps > 0 {
		state.TxRateLimitKbps = types.Int64Value(v.TxRateLimitKbps)
	} else {
		state.TxRateLimitKbps = types.Int64Null()
	}
	if v.ActivatedAt != nil {
		state.ActivatedAt = types.StringValue(v.ActivatedAt.UTC().Format(time.RFC3339))
	}
	if v.ExpiresAt != nil {
		state.ExpiresAt = types.StringValue(v.ExpiresAt.UTC().Format(time.RFC3339))
	}

	return state
}
