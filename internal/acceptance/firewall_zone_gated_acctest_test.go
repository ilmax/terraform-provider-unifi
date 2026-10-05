package acceptance

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// This file covers the clean-error path for unifi_firewall_zone,
// unifi_firewall_zone_networks, unifi_firewall_policy,
// unifi_firewall_policy_ordering, and the unifi_firewall_zone(s) data
// sources: what a console reports when Zone Based Firewall genuinely isn't
// configured. On this test console it always is, once `make enable-zbf` has
// run (see README.md's "Zone Based Firewall" section for why that's a
// database write rather than an API call, and resource_firewall_*_acctest_test.go
// / data_source_firewall_zone*_acctest_test.go for the real CRUD happy path
// these tests can't otherwise exercise here) — so every test below
// self-skips once it detects that, and only actually runs against a bare
// console that hasn't had Zone Based Firewall force-enabled yet.
//
// Fake-but-well-formed UUIDs are used for zone/network IDs below because
// the API rejects these calls before it would ever validate that the IDs
// refer to real objects.

const fakeUUID = "00000000-0000-0000-0000-000000000001"

func TestAccFirewallZone_requiresZoneBasedFirewall(t *testing.T) {
	testAccPreCheck(t) // must run before hasZoneBasedFirewall's own API call
	if hasZoneBasedFirewall(t) {
		t.Skip("Zone Based Firewall is enabled on this console; see resource_firewall_zone_acctest_test.go for the happy path this error path guards")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "unifi_firewall_zone" "test" {
  name = "tf-acc-zone"
}
`,
				ExpectError: regexp.MustCompile(`(?i)zone\s+based\s+firewall\s+is\s+not\s+configured`),
			},
		},
	})
}

func TestAccFirewallZoneNetworks_requiresZoneBasedFirewall(t *testing.T) {
	testAccPreCheck(t)
	if hasZoneBasedFirewall(t) {
		t.Skip("Zone Based Firewall is enabled on this console; see resource_firewall_zone_networks_acctest_test.go for the happy path this error path guards")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "unifi_firewall_zone_networks" "test" {
  zone_id     = "` + fakeUUID + `"
  network_ids = []
}
`,
				ExpectError: regexp.MustCompile(`(?i)zone\s+based\s+firewall\s+is\s+not\s+configured`),
			},
		},
	})
}

func TestAccFirewallPolicy_requiresZoneBasedFirewall(t *testing.T) {
	testAccPreCheck(t)
	if hasZoneBasedFirewall(t) {
		t.Skip("Zone Based Firewall is enabled on this console; see resource_firewall_policy_acctest_test.go for the happy path this error path guards")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "unifi_firewall_policy" "test" {
  name                = "tf-acc-policy"
  action              = "BLOCK"
  source_zone_id      = "` + fakeUUID + `"
  destination_zone_id = "00000000-0000-0000-0000-000000000002"
}
`,
				ExpectError: regexp.MustCompile(`(?i)zone\s+based\s+firewall\s+is\s+not\s+configured`),
			},
		},
	})
}

func TestAccFirewallPolicyOrdering_requiresZoneBasedFirewall(t *testing.T) {
	testAccPreCheck(t)
	if hasZoneBasedFirewall(t) {
		t.Skip("Zone Based Firewall is enabled on this console; see resource_firewall_policy_ordering_acctest_test.go for the happy path this error path guards")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "unifi_firewall_policy_ordering" "test" {
  source_firewall_zone_id      = "` + fakeUUID + `"
  destination_firewall_zone_id = "00000000-0000-0000-0000-000000000002"
  before_system_defined        = []
  after_system_defined         = []
}
`,
				ExpectError: regexp.MustCompile(`(?i)zone\s+based\s+firewall\s+is\s+not\s+configured`),
			},
		},
	})
}

func TestAccFirewallZoneDataSource_requiresZoneBasedFirewall(t *testing.T) {
	testAccPreCheck(t)
	if hasZoneBasedFirewall(t) {
		t.Skip("Zone Based Firewall is enabled on this console; see data_source_firewall_zone_acctest_test.go for the happy path this error path guards")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "unifi_firewall_zone" "test" {
  name = "Internal"
}
`,
				ExpectError: regexp.MustCompile(`(?i)zone\s+based\s+firewall\s+is\s+not\s+configured`),
			},
		},
	})
}

func TestAccFirewallZonesDataSource_requiresZoneBasedFirewall(t *testing.T) {
	testAccPreCheck(t)
	if hasZoneBasedFirewall(t) {
		t.Skip("Zone Based Firewall is enabled on this console; see data_source_firewall_zones_acctest_test.go for the happy path this error path guards")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "unifi_firewall_zones" "test" {}
`,
				ExpectError: regexp.MustCompile(`(?i)zone\s+based\s+firewall\s+is\s+not\s+configured`),
			},
		},
	})
}
