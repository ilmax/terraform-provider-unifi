package provider

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	networkapi "github.com/ilmax/unifi-client-go/pkg/network"
)

func TestDNSPolicyStateFromAPI(t *testing.T) {
	domain := "example.com"
	policyID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	state, err := dnsPolicyStateFromAPI("site-id", mustDNSPolicyFromBase(t, networkapi.DNSPolicyBase{
		Id:      policyID,
		Type:    "FORWARD_DOMAIN",
		Enabled: true,
		Domain:  &domain,
		Metadata: networkapi.UserDefinedEntityMetadata{
			Origin: "USER_DEFINED",
		},
	}))
	if err != nil {
		t.Fatalf("dnsPolicyStateFromAPI returned error: %v", err)
	}

	if state.ID.ValueString() != policyID.String() {
		t.Fatalf("unexpected id: %s", state.ID.ValueString())
	}
	if state.SiteID.ValueString() != "site-id" {
		t.Fatalf("unexpected site_id: %s", state.SiteID.ValueString())
	}
	if state.Type.ValueString() != "FORWARD_DOMAIN" {
		t.Fatalf("unexpected type: %s", state.Type.ValueString())
	}
	if !state.Enabled.ValueBool() {
		t.Fatal("expected enabled=true")
	}
	if state.Domain.IsNull() || state.Domain.ValueString() != "example.com" {
		t.Fatalf("unexpected domain: %#v", state.Domain)
	}
	if state.Origin.ValueString() != "USER_DEFINED" {
		t.Fatalf("unexpected origin: %s", state.Origin.ValueString())
	}
}

func TestDNSPolicyStateFromAPINullDomain(t *testing.T) {
	policyID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	state, err := dnsPolicyStateFromAPI("site-id", mustDNSPolicyFromBase(t, networkapi.DNSPolicyBase{
		Id:      policyID,
		Type:    "A_RECORD",
		Enabled: false,
		Metadata: networkapi.UserDefinedEntityMetadata{
			Origin: "SYSTEM_DEFINED",
		},
	}))
	if err != nil {
		t.Fatalf("dnsPolicyStateFromAPI returned error: %v", err)
	}

	if !state.Domain.IsNull() {
		t.Fatalf("expected null domain, got %s", state.Domain.ValueString())
	}
	if state.Enabled.ValueBool() {
		t.Fatal("expected enabled=false")
	}
}

func mustDNSPolicyFromBase(t *testing.T, base networkapi.DNSPolicyBase) networkapi.DNSPolicy {
	t.Helper()

	raw, err := json.Marshal(base)
	if err != nil {
		t.Fatalf("marshal dns policy base: %v", err)
	}

	var policy networkapi.DNSPolicy
	if err := json.Unmarshal(raw, &policy); err != nil {
		t.Fatalf("unmarshal dns policy base into union: %v", err)
	}

	return policy
}
