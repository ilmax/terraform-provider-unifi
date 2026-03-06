package provider

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/sdkcompat/acl_rules"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
)

const (
	aclRulesPageSize = int64(200)
	aclRulesMaxPages = 1000
)

type aclRulesDataSource struct {
	client *unifi.Client
	siteID string
}

type aclRulesDataSourceModel struct {
	ID       types.String                  `tfsdk:"id"`
	SiteID   types.String                  `tfsdk:"site_id"`
	ACLRules []aclRulesDataSourceRuleModel `tfsdk:"acl_rules"`
}

type aclRulesDataSourceRuleModel struct {
	ID          types.String `tfsdk:"id"`
	Type        types.String `tfsdk:"type"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Action      types.String `tfsdk:"action"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Index       types.Int64  `tfsdk:"index"`
	Origin      types.String `tfsdk:"origin"`
}

func NewACLRulesDataSource() datasource.DataSource {
	return &aclRulesDataSource{}
}

func (d *aclRulesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acl_rules"
}

func (d *aclRulesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"site_id": schema.StringAttribute{
				Optional: true,
			},
			"acl_rules": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"type": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"description": schema.StringAttribute{
							Computed: true,
						},
						"action": schema.StringAttribute{
							Computed: true,
						},
						"enabled": schema.BoolAttribute{
							Computed: true,
						},
						"index": schema.Int64Attribute{
							Computed: true,
						},
						"origin": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *aclRulesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *aclRulesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config aclRulesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, ok := resolveSiteID(config.SiteID, d.siteID, &resp.Diagnostics)
	if !ok {
		return
	}

	rules, err := d.listAllACLRules(ctx, siteID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list ACL rules", err.Error())
		return
	}

	state := aclRulesDataSourceModel{
		ID:       types.StringValue(siteID),
		SiteID:   types.StringValue(siteID),
		ACLRules: aclRulesFromAPI(rules),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *aclRulesDataSource) listAllACLRules(ctx context.Context, siteID string) ([]acl_rules.ListACLRulesData, error) {
	return collectACLRulesPages(aclRulesPageSize, aclRulesMaxPages, func(offset, limit int64) (acl_rules.ListACLRulesResponse, error) {
		apiPath := fmt.Sprintf("/v1/sites/%s/acl-rules?offset=%d&limit=%d", siteID, offset, limit)
		var page acl_rules.ListACLRulesResponse
		if err := d.client.Get(ctx, apiPath, &page); err != nil {
			return acl_rules.ListACLRulesResponse{}, err
		}
		return page, nil
	})
}

func collectACLRulesPages(pageSize int64, maxPages int, fetch func(offset, limit int64) (acl_rules.ListACLRulesResponse, error)) ([]acl_rules.ListACLRulesData, error) {
	offset := int64(0)
	allRules := make([]acl_rules.ListACLRulesData, 0)

	for page := 0; page < maxPages; page++ {
		result, err := fetch(offset, pageSize)
		if err != nil {
			return nil, err
		}

		pageCount := int64(len(result.Data))
		if pageCount == 0 {
			return allRules, nil
		}

		allRules = append(allRules, result.Data...)
		nextOffset := offset + pageCount
		if nextOffset < 0 {
			return nil, fmt.Errorf("acl rules pagination offset overflow: %d", nextOffset)
		}

		if result.TotalCount > 0 && nextOffset >= result.TotalCount {
			return allRules, nil
		}
		if pageCount < pageSize {
			return allRules, nil
		}
		offset = nextOffset
	}

	return nil, fmt.Errorf("acl rules pagination exceeded %d pages", maxPages)
}

func aclRulesFromAPI(items []acl_rules.ListACLRulesData) []aclRulesDataSourceRuleModel {
	rules := make([]aclRulesDataSourceRuleModel, 0, len(items))
	for _, item := range items {
		rule := aclRulesDataSourceRuleModel{
			ID:          stringValueOrNull(item.Id),
			Type:        stringValueOrNull(item.Type),
			Name:        stringValueOrNull(item.Name),
			Description: stringValueOrNull(item.Description),
			Action:      stringValueOrNull(item.Action),
			Enabled:     types.BoolValue(item.Enabled),
			Index:       types.Int64Value(item.Index),
			Origin:      types.StringNull(),
		}
		if item.Metadata != nil {
			rule.Origin = stringValueOrNull(item.Metadata.Origin)
		}
		rules = append(rules, rule)
	}

	sort.SliceStable(rules, func(i, j int) bool {
		if rules[i].Index.ValueInt64() == rules[j].Index.ValueInt64() {
			return rules[i].ID.ValueString() < rules[j].ID.ValueString()
		}
		return rules[i].Index.ValueInt64() < rules[j].Index.ValueInt64()
	})

	return rules
}
