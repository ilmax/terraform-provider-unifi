package provider

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
	networkapi "github.com/ilmax/unifi-client-go/pkg/network"
)

func TestBuildDNSARecordPayload(t *testing.T) {
	payload, err := buildDNSARecordPayload(dnsARecordResourceModel{
		Enabled:     boolValue(true),
		Domain:      stringValue("example.com"),
		IPv4Address: stringValue("192.0.2.10"),
		TTLSeconds:  int64Value(300),
	})
	if err != nil {
		t.Fatalf("buildDNSARecordPayload returned error: %v", err)
	}

	record, err := payload.AsIntegrationDnsARecordCreateUpdateDto()
	if err != nil {
		t.Fatalf("AsIntegrationDnsARecordCreateUpdateDto returned error: %v", err)
	}

	if record.Type != "A_RECORD" {
		t.Fatalf("unexpected type: %s", record.Type)
	}
	if record.Domain == nil || *record.Domain != "example.com" {
		t.Fatalf("unexpected domain: %#v", record.Domain)
	}
	if record.Ipv4Address == nil || *record.Ipv4Address != "192.0.2.10" {
		t.Fatalf("unexpected ipv4 address: %#v", record.Ipv4Address)
	}
	if record.TtlSeconds == nil || *record.TtlSeconds != 300 {
		t.Fatalf("unexpected ttl: %#v", record.TtlSeconds)
	}
}

func TestDNSARecordStateFromAPI(t *testing.T) {
	policy := networkapi.DNSPolicy{}
	if err := policy.FromIntegrationDnsARecordDto(networkapi.IntegrationDnsARecordDto{
		Id:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Type:        "A_RECORD",
		Enabled:     true,
		Domain:      stringPointer("example.com"),
		Ipv4Address: stringPointer("192.0.2.10"),
		TtlSeconds:  int32Pointer(300),
		Metadata:    networkapi.UserDefinedEntityMetadata{Origin: "USER_DEFINED"},
	}); err != nil {
		t.Fatalf("FromIntegrationDnsARecordDto returned error: %v", err)
	}

	state, err := dnsARecordStateFromAPI("site-id", policy)
	if err != nil {
		t.Fatalf("dnsARecordStateFromAPI returned error: %v", err)
	}

	if state.IPv4Address.ValueString() != "192.0.2.10" {
		t.Fatalf("unexpected ipv4_address: %s", state.IPv4Address.ValueString())
	}
	if state.TTLSeconds.ValueInt64() != 300 {
		t.Fatalf("unexpected ttl_seconds: %d", state.TTLSeconds.ValueInt64())
	}
}

func TestBuildDNSAAAARecordPayload(t *testing.T) {
	payload, err := buildDNSAAAARecordPayload(dnsAAAARecordResourceModel{
		Enabled:     boolValue(true),
		Domain:      stringValue("example.com"),
		IPv6Address: stringValue("2001:db8::10"),
		TTLSeconds:  int64Value(600),
	})
	if err != nil {
		t.Fatalf("buildDNSAAAARecordPayload returned error: %v", err)
	}

	record, err := payload.AsIntegrationDnsAaaaRecordCreateUpdateDto()
	if err != nil {
		t.Fatalf("AsIntegrationDnsAaaaRecordCreateUpdateDto returned error: %v", err)
	}

	if record.Type != "AAAA_RECORD" {
		t.Fatalf("unexpected type: %s", record.Type)
	}
	if record.Ipv6Address == nil || *record.Ipv6Address != "2001:db8::10" {
		t.Fatalf("unexpected ipv6 address: %#v", record.Ipv6Address)
	}
}

func TestDNSAAAARecordStateFromAPI(t *testing.T) {
	policy := networkapi.DNSPolicy{}
	if err := policy.FromIntegrationDnsAaaaRecordDto(networkapi.IntegrationDnsAaaaRecordDto{
		Id:          uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		Type:        "AAAA_RECORD",
		Enabled:     true,
		Domain:      stringPointer("example.com"),
		Ipv6Address: stringPointer("2001:db8::10"),
		TtlSeconds:  int32Pointer(600),
		Metadata:    networkapi.UserDefinedEntityMetadata{Origin: "USER_DEFINED"},
	}); err != nil {
		t.Fatalf("FromIntegrationDnsAaaaRecordDto returned error: %v", err)
	}

	state, err := dnsAAAARecordStateFromAPI("site-id", policy)
	if err != nil {
		t.Fatalf("dnsAAAARecordStateFromAPI returned error: %v", err)
	}

	if state.IPv6Address.ValueString() != "2001:db8::10" {
		t.Fatalf("unexpected ipv6_address: %s", state.IPv6Address.ValueString())
	}
}

func TestBuildDNSCNAMERecordPayload(t *testing.T) {
	payload, err := buildDNSCNAMERecordPayload(dnsCNAMERecordResourceModel{
		Enabled:      boolValue(true),
		Domain:       stringValue("app.example.com"),
		TargetDomain: stringValue("target.example.com"),
		TTLSeconds:   int64Value(900),
	})
	if err != nil {
		t.Fatalf("buildDNSCNAMERecordPayload returned error: %v", err)
	}

	record, err := payload.AsIntegrationDnsCnameRecordCreateUpdateDto()
	if err != nil {
		t.Fatalf("AsIntegrationDnsCnameRecordCreateUpdateDto returned error: %v", err)
	}

	if record.Type != "CNAME_RECORD" {
		t.Fatalf("unexpected type: %s", record.Type)
	}
	if record.TargetDomain == nil || *record.TargetDomain != "target.example.com" {
		t.Fatalf("unexpected target domain: %#v", record.TargetDomain)
	}
}

func TestDNSCNAMERecordStateFromAPI(t *testing.T) {
	policy := networkapi.DNSPolicy{}
	if err := policy.FromIntegrationDnsCnameRecordDto(networkapi.IntegrationDnsCnameRecordDto{
		Id:           uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		Type:         "CNAME_RECORD",
		Enabled:      true,
		Domain:       stringPointer("app.example.com"),
		TargetDomain: stringPointer("target.example.com"),
		TtlSeconds:   int32Pointer(900),
		Metadata:     networkapi.UserDefinedEntityMetadata{Origin: "USER_DEFINED"},
	}); err != nil {
		t.Fatalf("FromIntegrationDnsCnameRecordDto returned error: %v", err)
	}

	state, err := dnsCNAMERecordStateFromAPI("site-id", policy)
	if err != nil {
		t.Fatalf("dnsCNAMERecordStateFromAPI returned error: %v", err)
	}

	if state.TargetDomain.ValueString() != "target.example.com" {
		t.Fatalf("unexpected target_domain: %s", state.TargetDomain.ValueString())
	}
}

func stringValue(value string) types.String {
	return types.StringValue(value)
}

func boolValue(value bool) types.Bool {
	return types.BoolValue(value)
}

func int64Value(value int64) types.Int64 {
	return types.Int64Value(value)
}

func stringPointer(value string) *string {
	return &value
}

func int32Pointer(value int32) *int32 {
	return &value
}
