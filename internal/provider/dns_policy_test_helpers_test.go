package provider

import (
	"encoding/json"
	"testing"

	networkapi "github.com/ilmax/terraform-provider-unifi/internal/sdkcompat/dns"
)

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
