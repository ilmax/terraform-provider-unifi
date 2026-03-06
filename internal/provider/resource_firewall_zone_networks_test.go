package provider

import (
	"context"
	"encoding/json"
	"testing"
)

func TestFirewallZoneNetworksState(t *testing.T) {
	state := firewallZoneNetworksState("site-1", "zone-1", []json.RawMessage{
		json.RawMessage(`"net-1"`),
		json.RawMessage(`"net-2"`),
	})

	if state.ID.ValueString() != "site-1/zone-1" {
		t.Fatalf("unexpected id: %s", state.ID.ValueString())
	}
	if state.SiteID.ValueString() != "site-1" {
		t.Fatalf("unexpected site_id: %s", state.SiteID.ValueString())
	}
	if state.ZoneID.ValueString() != "zone-1" {
		t.Fatalf("unexpected zone_id: %s", state.ZoneID.ValueString())
	}

	var ids []string
	diags := state.NetworkIDs.ElementsAs(context.Background(), &ids, false)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(ids) != 2 || ids[0] != "net-1" || ids[1] != "net-2" {
		t.Fatalf("unexpected network_ids: %#v", ids)
	}
}
