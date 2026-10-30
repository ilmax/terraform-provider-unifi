// Package siteid holds the repo-wide site_id precedence rule shared by the
// main and the legacy provider (see AGENTS.md "Site ID Consistency Rules").
package siteid

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Resolve returns the resource-level site_id when set, otherwise the
// provider-level one, otherwise adds a diagnostic and reports false.
func Resolve(siteAttr types.String, providerSiteID string, diags *diag.Diagnostics) (string, bool) {
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
