package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/sdkcompat/switching"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
	"github.com/ilmax/unifi-client-go/pkg/errors"
)

// switchStackDataSource exposes
// /v1/sites/{siteId}/switching/switch-stacks/{switchStackId}, a GET-only
// endpoint in the real API (no create/update/delete), as a data source.
// Requires stackable switch hardware to have anything to read.
type switchStackDataSource struct {
	client *unifi.Client
	siteID string
}

type switchStackDataSourceModel struct {
	ID            types.String                         `tfsdk:"id"`
	SiteID        types.String                         `tfsdk:"site_id"`
	SwitchStackID types.String                         `tfsdk:"switch_stack_id"`
	Name          types.String                         `tfsdk:"name"`
	DeviceID      types.String                         `tfsdk:"device_id"`
	Origin        types.String                         `tfsdk:"origin"`
	Lags          []switchStackLagLocalDataSourceModel `tfsdk:"lags"`
	Units         []switchStackUnitDataSourceModel     `tfsdk:"units"`
}

type switchStackLagLocalDataSourceModel struct {
	ID      types.String                          `tfsdk:"id"`
	Origin  types.String                          `tfsdk:"origin"`
	Members []switchStackLagMemberDataSourceModel `tfsdk:"members"`
}

type switchStackLagMemberDataSourceModel struct {
	PortIdxs       types.List   `tfsdk:"port_idxs"`
	UnitID         types.Int64  `tfsdk:"unit_id"`
	UnitMacAddress types.String `tfsdk:"unit_mac_address"`
}

type switchStackUnitDataSourceModel struct {
	ID         types.Int64  `tfsdk:"id"`
	MacAddress types.String `tfsdk:"mac_address"`
	Order      types.Int64  `tfsdk:"order"`
	Role       types.String `tfsdk:"role"`
}

func NewSwitchStackDataSource() datasource.DataSource {
	return &switchStackDataSource{}
}

func (d *switchStackDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_switch_stack"
}

func (d *switchStackDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A stack of physically linked switches acting as one logical switch. Requires stackable " +
			"switch hardware to have anything to read. This endpoint is GET-only in the real API, so it's a data " +
			"source rather than a resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Switch stack identifier.",
				Computed:            true,
			},
			"site_id": schema.StringAttribute{
				MarkdownDescription: "Site identifier. Defaults to provider `site_id` when omitted. A local `site_id` overrides the provider setting.",
				Optional:            true,
			},
			"switch_stack_id": schema.StringAttribute{
				MarkdownDescription: "Switch stack identifier.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Switch stack name.",
				Computed:            true,
			},
			"device_id": schema.StringAttribute{
				MarkdownDescription: "Primary device ID for this stack, if reported.",
				Computed:            true,
			},
			"origin": schema.StringAttribute{
				MarkdownDescription: "Entity origin metadata.",
				Computed:            true,
			},
			"lags": schema.ListNestedAttribute{
				MarkdownDescription: "LAGs local to this stack.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "LAG identifier.",
							Computed:            true,
						},
						"origin": schema.StringAttribute{
							MarkdownDescription: "Entity origin metadata.",
							Computed:            true,
						},
						"members": schema.ListNestedAttribute{
							MarkdownDescription: "LAG member ports.",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"port_idxs": schema.ListAttribute{
										MarkdownDescription: "Port indexes making up this member.",
										Computed:            true,
										ElementType:         types.Int64Type,
									},
									"unit_id": schema.Int64Attribute{
										MarkdownDescription: "Stack unit ID this member port belongs to.",
										Computed:            true,
									},
									"unit_mac_address": schema.StringAttribute{
										MarkdownDescription: "MAC address of the stack unit this member port belongs to.",
										Computed:            true,
									},
								},
							},
						},
					},
				},
			},
			"units": schema.ListNestedAttribute{
				MarkdownDescription: "Physical switch units making up this stack.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "Stack unit ID.",
							Computed:            true,
						},
						"mac_address": schema.StringAttribute{
							MarkdownDescription: "Unit MAC address.",
							Computed:            true,
						},
						"order": schema.Int64Attribute{
							MarkdownDescription: "Unit's position in the stack, if reported.",
							Computed:            true,
						},
						"role": schema.StringAttribute{
							MarkdownDescription: "Unit role (e.g. master/member), if reported.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *switchStackDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *switchStackDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config switchStackDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(config.SiteID, d.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	switchStackID := config.SwitchStackID.ValueString()
	if switchStackID == "" {
		resp.Diagnostics.AddAttributeError(path.Root("switch_stack_id"), "Missing switch_stack_id", "switch_stack_id must be provided.")
		return
	}

	var result switching.SwitchStack
	apiPath := fmt.Sprintf("/v1/sites/%s/switching/switch-stacks/%s", siteID, switchStackID)
	if err := d.client.Get(ctx, apiPath, &result); err != nil {
		if errors.IsNotFoundError(err) {
			resp.Diagnostics.AddError("Switch stack not found", fmt.Sprintf("Switch stack %q was not found in site %q.", switchStackID, siteID))
			return
		}
		resp.Diagnostics.AddError("Unable to read switch stack", err.Error())
		return
	}

	state := switchStackDataSourceModel{
		ID:            types.StringValue(result.Id),
		SiteID:        types.StringValue(siteID),
		SwitchStackID: types.StringValue(result.Id),
		Name:          types.StringValue(result.Name),
		DeviceID:      stringValueOrNull(result.DeviceId),
		Origin:        stringValueOrNull(result.Metadata.Origin),
		Lags:          make([]switchStackLagLocalDataSourceModel, 0, len(result.Lags)),
		Units:         make([]switchStackUnitDataSourceModel, 0, len(result.Units)),
	}
	for _, lag := range result.Lags {
		members := make([]switchStackLagMemberDataSourceModel, 0, len(lag.Members))
		for _, member := range lag.Members {
			members = append(members, switchStackLagMemberDataSourceModel{
				PortIdxs:       int64sToList(member.PortIdxs),
				UnitID:         types.Int64Value(member.UnitId),
				UnitMacAddress: types.StringValue(member.UnitMacAddress),
			})
		}
		state.Lags = append(state.Lags, switchStackLagLocalDataSourceModel{
			ID:      types.StringValue(lag.Id),
			Origin:  stringValueOrNull(lag.Metadata.Origin),
			Members: members,
		})
	}
	for _, unit := range result.Units {
		u := switchStackUnitDataSourceModel{
			ID:         types.Int64Value(unit.Id),
			MacAddress: types.StringValue(unit.MacAddress),
			Order:      types.Int64Null(),
			Role:       stringValueOrNull(unit.Role),
		}
		if unit.Order != nil {
			u.Order = types.Int64Value(*unit.Order)
		}
		state.Units = append(state.Units, u)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
