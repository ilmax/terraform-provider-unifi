package provider

import (
	"testing"

	"github.com/google/uuid"
	networkapi "github.com/ilmax/unifi-client-go/pkg/network"
)

func TestDNSPolicyDataSourceStateFromAPI(t *testing.T) {
	domain := "ads.example"
	policyID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	state, err := dnsPolicyDataSourceStateFromAPI("site-id", mustDNSPolicyFromBase(t, networkapi.DNSPolicyBase{
		Id:      policyID,
		Type:    "FORWARD_DOMAIN",
		Enabled: true,
		Domain:  &domain,
		Metadata: networkapi.UserDefinedEntityMetadata{
			Origin: "USER_DEFINED",
		},
	}))
	if err != nil {
		t.Fatalf("dnsPolicyDataSourceStateFromAPI returned error: %v", err)
	}

	if state.PolicyID.ValueString() != policyID.String() {
		t.Fatalf("unexpected policy_id: %s", state.PolicyID.ValueString())
	}
	if state.Type.ValueString() != "FORWARD_DOMAIN" {
		t.Fatalf("unexpected type: %s", state.Type.ValueString())
	}
	if state.Domain.IsNull() || state.Domain.ValueString() != "ads.example" {
		t.Fatalf("unexpected domain: %#v", state.Domain)
	}
	if state.Origin.ValueString() != "USER_DEFINED" {
		t.Fatalf("unexpected origin: %s", state.Origin.ValueString())
	}
}

func TestDNSPolicyMatchesFilter(t *testing.T) {
	domain := "ads.example"
	policy := networkapi.DNSPolicyBase{
		Type:   "FORWARD_DOMAIN",
		Domain: &domain,
	}

	if !dnsPolicyMatchesFilter(policy, "forward_domain", "ADS.EXAMPLE") {
		t.Fatal("expected case-insensitive type+domain match")
	}

	if dnsPolicyMatchesFilter(policy, "allow", "ads.example") {
		t.Fatal("expected type mismatch to fail")
	}

	if dnsPolicyMatchesFilter(policy, "block", "other.example") {
		t.Fatal("expected domain mismatch to fail")
	}
}
