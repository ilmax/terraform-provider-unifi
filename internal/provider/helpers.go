package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func resolveSiteID(siteAttr types.String, providerSiteID string, diags *diag.Diagnostics) (string, bool) {
	if !siteAttr.IsNull() && !siteAttr.IsUnknown() {
		value := siteAttr.ValueString()
		if value == "" {
			diags.AddAttributeError(
				path.Root("site_id"),
				"Empty site_id",
				"The site_id value cannot be empty when set.",
			)
			return "", false
		}
		return value, true
	}

	if providerSiteID != "" {
		return providerSiteID, true
	}

	diags.AddAttributeError(
		path.Root("site_id"),
		"Missing site_id",
		"The site_id must be set on the resource or provider configuration.",
	)
	return "", false
}

func splitImportID(id string) (string, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("expected import identifier to be in the format <site_id>/<resource_id>")
	}
	if parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("site_id and resource_id must be non-empty")
	}
	return parts[0], parts[1], nil
}

func requireValidJSON(raw string) error {
	if !json.Valid([]byte(raw)) {
		return fmt.Errorf("invalid JSON value")
	}
	return nil
}

type defaultInt64WhenNullPlanModifier struct {
	defaultValue int64
}

func (m defaultInt64WhenNullPlanModifier) Description(ctx context.Context) string {
	return fmt.Sprintf("Defaults to %d when not set.", m.defaultValue)
}

func (m defaultInt64WhenNullPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m defaultInt64WhenNullPlanModifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if req.ConfigValue.IsNull() {
		resp.PlanValue = types.Int64Value(m.defaultValue)
	}
}

func defaultInt64WhenNull(defaultValue int64) planmodifier.Int64 {
	return defaultInt64WhenNullPlanModifier{defaultValue: defaultValue}
}
