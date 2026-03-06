package provider

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildFirewallZoneCreateRequestUsesEmptyAssignments(t *testing.T) {
	plan := firewallZoneResourceModel{
		Name: types.StringValue("trusted"),
	}

	req := buildFirewallZoneCreateRequest(plan)
	if req.Name != "trusted" {
		t.Fatalf("expected name trusted, got %s", req.Name)
	}
	if len(req.NetworkIds) != 0 {
		t.Fatalf("expected empty network ids, got %d", len(req.NetworkIds))
	}
}

func TestBuildFirewallZoneUpdateRequestPreservesAssignments(t *testing.T) {
	plan := firewallZoneResourceModel{
		Name: types.StringValue("trusted"),
	}
	current := []json.RawMessage{
		json.RawMessage(`"network-1"`),
		json.RawMessage(`"network-2"`),
	}

	req := buildFirewallZoneUpdateRequest(plan, current)
	if req.Name != "trusted" {
		t.Fatalf("expected name trusted, got %s", req.Name)
	}
	if len(req.NetworkIds) != 2 {
		t.Fatalf("expected two network ids, got %d", len(req.NetworkIds))
	}
}
