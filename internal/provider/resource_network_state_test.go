package provider

import (
	"context"
	"testing"

	"github.com/ilmax/unifi-client-go/pkg/networks"
)

func TestNetworkStateFromResponseGateway(t *testing.T) {
	raw := []byte(`{
  "management": "GATEWAY",
  "id": "b83f633d-5d2e-4a47-a59c-df6d5e7b69ca",
  "name": "Homelab",
  "enabled": true,
  "vlanId": 20,
  "metadata": {
    "origin": "USER_DEFINED"
  },
  "isolationEnabled": false,
  "cellularBackupEnabled": true,
  "zoneId": "ed426e98-252f-4ff2-982b-d475007f085f",
  "internetAccessEnabled": true,
  "multicastDnsEnable": true,
  "ipv4Configuration": {
    "autoScaleEnabled": false,
    "hostIpAddress": "192.168.20.1",
    "prefixLength": 24,
    "dhcpConfiguration": {
      "mode": "SERVER",
      "ipAddressRange": {},
      "dnsServerIpAddressesOverride": [
        "192.168.1.250"
      ],
      "leaseTimeSeconds": 86400,
      "domainName": "lan",
      "pingConflictDetectionEnabled": true
    },
    "natOutboundIpAddressConfiguration": [
      {
        "type": "AUTO",
        "wanInterfaceId": "90032e74-5fb6-44fd-8cda-5842749f5a0e",
        "ipAddressSelectionMode": "MAIN"
      }
    ]
  }
}`)

	decoded, err := networks.DecodeGetNetworkDetailsResponse(raw)
	if err != nil {
		t.Fatalf("decode response: %v", err)
	}

	state := networkResourceModel{}
	diags := networkStateFromResponse(&state, "site-id", decoded)
	if diags.HasError() {
		t.Fatalf("network state diagnostics: %v", diags)
	}
	state.MulticastDNSEnabled = readMulticastDNSEnable(raw)

	if state.IPv4Configuration == nil {
		t.Fatal("expected ipv4_configuration to be populated")
	}
	if state.IPv4Configuration.DHCPConfiguration == nil {
		t.Fatal("expected ipv4_configuration.dhcp_configuration to be populated")
	}
	if state.IPv4Configuration.DHCPConfiguration.Mode.IsNull() || state.IPv4Configuration.DHCPConfiguration.Mode.ValueString() != "SERVER" {
		t.Fatalf("expected DHCP mode SERVER, got %q", state.IPv4Configuration.DHCPConfiguration.Mode.ValueString())
	}
	if state.MulticastDNSEnabled.IsNull() || !state.MulticastDNSEnabled.ValueBool() {
		t.Fatal("expected multicast_dns_enabled to be true")
	}

	var dns []string
	if diag := state.IPv4Configuration.DHCPConfiguration.DNSServers.ElementsAs(context.Background(), &dns, false); diag.HasError() {
		t.Fatalf("decode dns servers: %v", diag)
	}
	if len(dns) != 1 || dns[0] != "192.168.1.250" {
		t.Fatalf("expected dns servers [192.168.1.250], got %v", dns)
	}
}
