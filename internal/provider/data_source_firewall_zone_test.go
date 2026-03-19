package provider

import (
	"testing"

	"github.com/ilmax/terraform-provider-unifi/internal/sdkcompat/zones"
)

func TestMatchFirewallZonesByName(t *testing.T) {
	items := []zones.ListFirewallZonesData{
		{Id: "zone-1", Name: "Trusted"},
		{Id: "zone-2", Name: "IoT"},
		{Id: "zone-3", Name: "trusted"},
	}

	matches := matchFirewallZonesByName(items, " trusted ")
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
	if matches[0].Id != "zone-1" || matches[1].Id != "zone-3" {
		t.Fatalf("unexpected matches: %#v", matches)
	}
}
