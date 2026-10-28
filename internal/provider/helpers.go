package provider

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/siteid"
)

func resolveSiteID(siteAttr types.String, providerSiteID string, diags *diag.Diagnostics) (string, bool) {
	return siteid.Resolve(siteAttr, providerSiteID, diags)
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

func splitImportID3(id string) (string, string, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("expected import identifier to be in the format <site_id>/<part_one>/<part_two>")
	}
	if parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("site_id, part_one, and part_two must be non-empty")
	}
	return parts[0], parts[1], parts[2], nil
}

func requireValidJSON(raw string) error {
	if !json.Valid([]byte(raw)) {
		return fmt.Errorf("invalid JSON value")
	}
	return nil
}
