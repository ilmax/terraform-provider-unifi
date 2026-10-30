package legacyprovider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// TestTeleportSchemas pins the site_id contract (optional + computed on both,
// per AGENTS.md) and that the schemas are internally valid.
func TestTeleportSchemas(t *testing.T) {
	ctx := context.Background()

	var rr resource.SchemaResponse
	NewTeleportResource().Schema(ctx, resource.SchemaRequest{}, &rr)
	if d := rr.Schema.ValidateImplementation(ctx); d.HasError() {
		t.Fatalf("resource schema: %v", d)
	}
	site, ok := rr.Schema.Attributes["site_id"].(schema.StringAttribute)
	if !ok || !site.Optional || !site.Computed {
		t.Errorf("resource site_id must be Optional+Computed, got %#v", rr.Schema.Attributes["site_id"])
	}
	for _, name := range []string{"enabled", "subnet_cidr"} {
		if !rr.Schema.Attributes[name].IsOptional() || !rr.Schema.Attributes[name].IsComputed() {
			t.Errorf("resource %s must be Optional+Computed so unset keeps the controller value", name)
		}
	}

	var dr datasource.SchemaResponse
	NewTeleportDataSource().Schema(ctx, datasource.SchemaRequest{}, &dr)
	if d := dr.Schema.ValidateImplementation(ctx); d.HasError() {
		t.Fatalf("data source schema: %v", d)
	}
	dsite, ok := dr.Schema.Attributes["site_id"].(dsschema.StringAttribute)
	if !ok || !dsite.Optional || !dsite.Computed {
		t.Errorf("data source site_id must be Optional+Computed, got %#v", dr.Schema.Attributes["site_id"])
	}
}

func TestIPReservationSchema(t *testing.T) {
	ctx := context.Background()
	var rr resource.SchemaResponse
	NewIPReservationResource().Schema(ctx, resource.SchemaRequest{}, &rr)
	if d := rr.Schema.ValidateImplementation(ctx); d.HasError() {
		t.Fatalf("resource schema: %v", d)
	}
	site, ok := rr.Schema.Attributes["site_id"].(schema.StringAttribute)
	if !ok || !site.Optional || !site.Computed {
		t.Errorf("site_id must be Optional+Computed, got %#v", rr.Schema.Attributes["site_id"])
	}
	for _, name := range []string{"mac", "ip_address", "network_id"} {
		if !rr.Schema.Attributes[name].IsRequired() {
			t.Errorf("%s must be required", name)
		}
	}

	var dr datasource.SchemaResponse
	NewIPReservationDataSource().Schema(ctx, datasource.SchemaRequest{}, &dr)
	if d := dr.Schema.ValidateImplementation(ctx); d.HasError() {
		t.Fatalf("data source schema: %v", d)
	}
}

func TestMACPattern(t *testing.T) {
	for mac, want := range map[string]bool{
		"02:00:00:aa:bb:01": true,
		"02:00:00:AA:BB:01": false, // controller stores lowercase; uppercase would diff forever
		"02-00-00-aa-bb-01": false,
		"02:00:00:aa:bb":    false,
	} {
		if got := macPattern.MatchString(mac); got != want {
			t.Errorf("macPattern(%q) = %v, want %v", mac, got, want)
		}
	}
}
