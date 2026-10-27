package acceptance

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccFirewallZonesDataSource_found lists the system-default zones —
// only possible once Zone Based Firewall is enabled (see README.md's
// "Zone Based Firewall" section and firewall_zone_gated_acctest_test.go for
// the clean-error path this guards on a console where it isn't). A fresh
// UniFi Network 11 console has six (Internal, External, Gateway, Vpn,
// Hotspot, Dmz); the older consoles this suite ran against reported seven.
// Asserting "at least six" keeps the check meaningful on both without
// assuming nothing else has added custom zones to the site.
func TestAccFirewallZonesDataSource_found(t *testing.T) {
	testAccPreCheck(t) // must run before hasZoneBasedFirewall's own API call
	if !hasZoneBasedFirewall(t) {
		t.Skip("Zone Based Firewall is not enabled on this console — see test-infra/unifi/README.md, `make enable-zbf`")
	}

	dataSourceName := "data.unifi_firewall_zones.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "unifi_firewall_zones" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(
						dataSourceName, "zones.#", regexp.MustCompile(`^([6-9]|[1-9][0-9]+)$`),
					),
				),
			},
		},
	})
}
