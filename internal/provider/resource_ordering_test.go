package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestListToUUIDs(t *testing.T) {
	list, diags := types.ListValueFrom(context.Background(), types.StringType, []string{
		"11111111-1111-1111-1111-111111111111",
		"22222222-2222-2222-2222-222222222222",
	})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	values, parseDiags := listToUUIDs(list, path.Root("ordered_acl_rule_ids"))
	if parseDiags.HasError() {
		t.Fatalf("unexpected parse diagnostics: %v", parseDiags)
	}
	if len(values) != 2 {
		t.Fatalf("expected two UUID values, got %d", len(values))
	}
}

func TestListToUUIDsInvalid(t *testing.T) {
	list, diags := types.ListValueFrom(context.Background(), types.StringType, []string{"not-a-uuid"})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	_, parseDiags := listToUUIDs(list, path.Root("ordered_acl_rule_ids"))
	if !parseDiags.HasError() {
		t.Fatal("expected diagnostics for invalid UUID")
	}
}

func TestSplitImportID3(t *testing.T) {
	siteID, partOne, partTwo, err := splitImportID3("site/source/destination")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if siteID != "site" || partOne != "source" || partTwo != "destination" {
		t.Fatalf("unexpected import values: %s %s %s", siteID, partOne, partTwo)
	}
}
