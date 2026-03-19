package provider

import (
	"encoding/json"
	"fmt"

	networkapi "github.com/ilmax/unifi-client-go/pkg/network"
)

func buildDNSPolicyPayload(policyType string, enabled bool) (networkapi.CreateOrUpdateDNSPolicy, error) {
	raw, err := json.Marshal(networkapi.CreateOrUpdateDNSPolicyBase{
		Type:    policyType,
		Enabled: enabled,
	})
	if err != nil {
		return networkapi.CreateOrUpdateDNSPolicy{}, fmt.Errorf("encode dns policy payload: %w", err)
	}

	var payload networkapi.CreateOrUpdateDNSPolicy
	if err := json.Unmarshal(raw, &payload); err != nil {
		return networkapi.CreateOrUpdateDNSPolicy{}, fmt.Errorf("decode dns policy payload union: %w", err)
	}

	return payload, nil
}

func dnsPolicyBaseFromAPI(policy networkapi.DNSPolicy) (networkapi.DNSPolicyBase, error) {
	raw, err := json.Marshal(policy)
	if err != nil {
		return networkapi.DNSPolicyBase{}, fmt.Errorf("encode dns policy response: %w", err)
	}

	var base networkapi.DNSPolicyBase
	if err := json.Unmarshal(raw, &base); err != nil {
		return networkapi.DNSPolicyBase{}, fmt.Errorf("decode dns policy response: %w", err)
	}

	return base, nil
}
