package provider

import (
	"testing"

	"github.com/google/uuid"
	networkapi "github.com/ilmax/unifi-client-go/pkg/network"
)

func TestDNSPolicyStateFromAPI(t *testing.T) {
	domain := "example.com"
	policyID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	state := dnsPolicyStateFromAPI("site-id", networkapi.DNSPolicy{
		Id:      policyID,
		Type:    "BLOCK",
		Enabled: true,
		Domain:  &domain,
		Metadata: networkapi.UserDefinedEntityMetadata{
			Origin: "USER_DEFINED",
		},
	})

	if state.ID.ValueString() != policyID.String() {
		t.Fatalf("unexpected id: %s", state.ID.ValueString())
	}
	if state.SiteID.ValueString() != "site-id" {
		t.Fatalf("unexpected site_id: %s", state.SiteID.ValueString())
	}
	if state.Type.ValueString() != "BLOCK" {
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

	state := dnsPolicyStateFromAPI("site-id", networkapi.DNSPolicy{
		Id:      policyID,
		Type:    "ALLOW",
		Enabled: false,
		Metadata: networkapi.UserDefinedEntityMetadata{
			Origin: "SYSTEM_DEFINED",
		},
	})

	if !state.Domain.IsNull() {
		t.Fatalf("expected null domain, got %s", state.Domain.ValueString())
	}
	if state.Enabled.ValueBool() {
		t.Fatal("expected enabled=false")
	}
}
