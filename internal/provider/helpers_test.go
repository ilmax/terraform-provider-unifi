package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestResolveSiteIDPrefersLocalValue(t *testing.T) {
	var diags diag.Diagnostics

	siteID, ok := resolveSiteID(types.StringValue("resource-site"), "provider-site", &diags)
	if !ok {
		t.Fatal("expected resolveSiteID to succeed")
	}
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if siteID != "resource-site" {
		t.Fatalf("expected resource site id, got %q", siteID)
	}
}

func TestResolveSiteIDFallsBackToProvider(t *testing.T) {
	var diags diag.Diagnostics

	siteID, ok := resolveSiteID(types.StringNull(), "provider-site", &diags)
	if !ok {
		t.Fatal("expected resolveSiteID to succeed")
	}
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if siteID != "provider-site" {
		t.Fatalf("expected provider site id, got %q", siteID)
	}
}

func TestResolveSiteIDRequiresEitherLocalOrProviderValue(t *testing.T) {
	var diags diag.Diagnostics

	_, ok := resolveSiteID(types.StringNull(), "", &diags)
	if ok {
		t.Fatal("expected resolveSiteID to fail")
	}
	if !diags.HasError() {
		t.Fatal("expected diagnostics when site_id is missing")
	}
}
