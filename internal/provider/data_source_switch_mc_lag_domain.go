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

// switchMcLagDomainDataSource exposes
// /v1/sites/{siteId}/switching/mc-lag-domains/{mcLagDomainId}, a GET-only
// endpoint in the real API (no create/update/delete), as a data source.
// Requires switch hardware supporting MC-LAG to return anything on a real
// console.
type switchMcLagDomainDataSource struct {
	client *unifi.Client
	siteID string
}

type switchMcLagDomainDataSourceModel struct {
	ID            types.String                `tfsdk:"id"`
	SiteID        types.String                `tfsdk:"site_id"`
	McLagDomainID types.String                `tfsdk:"mc_lag_domain_id"`
	Name          types.String                `tfsdk:"name"`
	Origin        types.String                `tfsdk:"origin"`
	Lags          []mcLagLocalDataSourceModel `tfsdk:"lags"`
	Peers         []mcLagPeerDataSourceModel  `tfsdk:"peers"`
}

type mcLagLocalDataSourceModel struct {
	ID      types.String               `tfsdk:"id"`
	Origin  types.String               `tfsdk:"origin"`
	Members []lagMemberDataSourceModel `tfsdk:"members"`
}

type lagMemberDataSourceModel struct {
	DeviceID types.String `tfsdk:"device_id"`
	PortIdxs types.List   `tfsdk:"port_idxs"`
}

type mcLagPeerDataSourceModel struct {
	DeviceID     types.String `tfsdk:"device_id"`
	LinkPortIdxs types.List   `tfsdk:"link_port_idxs"`
	Role         types.String `tfsdk:"role"`
}

func NewSwitchMcLagDomainDataSource() datasource.DataSource {
	return &switchMcLagDomainDataSource{}
}

func (d *switchMcLagDomainDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_switch_mc_lag_domain"
}

func lagMemberNestedObject() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"device_id": schema.StringAttribute{
				Computed: true,
			},
			"port_idxs": schema.ListAttribute{
				Computed:    true,
				ElementType: types.Int64Type,
			},
		},
	}
}

func (d *switchMcLagDomainDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "An MC-LAG (multi-chassis link aggregation) domain spanning two switches. Requires MC-LAG " +
			"capable switch hardware to have anything to read. This endpoint is GET-only in the real API, so it's " +
			"a data source rather than a resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"site_id": schema.StringAttribute{
				Optional: true,
			},
			"mc_lag_domain_id": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"origin": schema.StringAttribute{
				Computed: true,
			},
			"lags": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"origin": schema.StringAttribute{
							Computed: true,
						},
						"members": schema.ListNestedAttribute{
							Computed:     true,
							NestedObject: lagMemberNestedObject(),
						},
					},
				},
			},
			"peers": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"device_id": schema.StringAttribute{
							Computed: true,
						},
						"link_port_idxs": schema.ListAttribute{
							Computed:    true,
							ElementType: types.Int64Type,
						},
						"role": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *switchMcLagDomainDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *switchMcLagDomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config switchMcLagDomainDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(config.SiteID, d.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	mcLagDomainID := config.McLagDomainID.ValueString()
	if mcLagDomainID == "" {
		resp.Diagnostics.AddAttributeError(path.Root("mc_lag_domain_id"), "Missing mc_lag_domain_id", "mc_lag_domain_id must be provided.")
		return
	}

	var result switching.McLagDomain
	apiPath := fmt.Sprintf("/v1/sites/%s/switching/mc-lag-domains/%s", siteID, mcLagDomainID)
	if err := d.client.Get(ctx, apiPath, &result); err != nil {
		if errors.IsNotFoundError(err) {
			resp.Diagnostics.AddError("MC-LAG domain not found", fmt.Sprintf("MC-LAG domain %q was not found in site %q.", mcLagDomainID, siteID))
			return
		}
		resp.Diagnostics.AddError("Unable to read MC-LAG domain", err.Error())
		return
	}

	state := switchMcLagDomainDataSourceModel{
		ID:            types.StringValue(result.Id),
		SiteID:        types.StringValue(siteID),
		McLagDomainID: types.StringValue(result.Id),
		Name:          types.StringValue(result.Name),
		Origin:        stringValueOrNull(result.Metadata.Origin),
		Lags:          make([]mcLagLocalDataSourceModel, 0, len(result.Lags)),
		Peers:         make([]mcLagPeerDataSourceModel, 0, len(result.Peers)),
	}
	for _, lag := range result.Lags {
		members := make([]lagMemberDataSourceModel, 0, len(lag.Members))
		for _, member := range lag.Members {
			members = append(members, lagMemberDataSourceModel{
				DeviceID: types.StringValue(member.DeviceId),
				PortIdxs: int64sToList(member.PortIdxs),
			})
		}
		state.Lags = append(state.Lags, mcLagLocalDataSourceModel{
			ID:      types.StringValue(lag.Id),
			Origin:  stringValueOrNull(lag.Metadata.Origin),
			Members: members,
		})
	}
	for _, peer := range result.Peers {
		state.Peers = append(state.Peers, mcLagPeerDataSourceModel{
			DeviceID:     types.StringValue(peer.DeviceId),
			LinkPortIdxs: int64sToList(peer.LinkPortIdxs),
			Role:         types.StringValue(peer.Role),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func int64sToList(values []int64) types.List {
	if len(values) == 0 {
		return types.ListNull(types.Int64Type)
	}
	list, _ := types.ListValueFrom(context.Background(), types.Int64Type, values)
	return list
}
