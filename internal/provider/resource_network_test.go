package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestResolveIPv4HostPrefix(t *testing.T) {
	t.Run("network-address", func(t *testing.T) {
		model := &networkIPv4ConfigurationModel{
			CIDR: types.StringValue("192.168.1.0/24"),
		}
		var diags diag.Diagnostics
		host, prefix := resolveIPv4HostPrefix(model, &diags)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if host != "192.168.1.1" {
			t.Fatalf("expected host 192.168.1.1, got %s", host)
		}
		if prefix != 24 {
			t.Fatalf("expected prefix 24, got %d", prefix)
		}
	})

	t.Run("host-address", func(t *testing.T) {
		model := &networkIPv4ConfigurationModel{
			CIDR: types.StringValue("10.0.0.1/24"),
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

	t.Run("conflict", func(t *testing.T) {
		model := &networkIPv4ConfigurationModel{
			CIDR:          types.StringValue("10.0.0.1/24"),
			HostIPAddress: types.StringValue("10.0.0.1"),
		}
		var diags diag.Diagnostics
		resolveIPv4HostPrefix(model, &diags)
		if !diags.HasError() {
			t.Fatal("expected diagnostics for conflicting cidr and host_ip_address")
		}
	})
}
