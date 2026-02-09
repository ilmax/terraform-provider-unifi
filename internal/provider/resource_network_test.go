package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestResolveIPv4HostPrefix(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		model := &networkIPv4ConfigurationModel{
			HostIPAddress: types.StringValue("10.0.0.1"),
			PrefixLength:  types.Int64Value(24),
		}
		var diags diag.Diagnostics
		host, prefix := resolveIPv4HostPrefix(model, &diags)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if host != "10.0.0.1" {
			t.Fatalf("expected host 10.0.0.1, got %s", host)
		}
		if prefix != 24 {
			t.Fatalf("expected prefix 24, got %d", prefix)
		}
	})

	t.Run("missing-host", func(t *testing.T) {
		model := &networkIPv4ConfigurationModel{
			PrefixLength: types.Int64Value(24),
		}
		var diags diag.Diagnostics
		resolveIPv4HostPrefix(model, &diags)
		if !diags.HasError() {
			t.Fatal("expected diagnostics for missing host_ip_address")
		}
	})

	t.Run("missing-prefix", func(t *testing.T) {
		model := &networkIPv4ConfigurationModel{
			HostIPAddress: types.StringValue("10.0.0.1"),
		}
		var diags diag.Diagnostics
		_, prefix := resolveIPv4HostPrefix(model, &diags)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if prefix != 24 {
			t.Fatalf("expected default prefix 24, got %d", prefix)
		}
	})
}
