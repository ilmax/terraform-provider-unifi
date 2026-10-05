package acceptance

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccFirewallZonesDataSource_found lists the real 7 system-default
// zones — only possible once Zone Based Firewall is enabled (see README.md's
// "Zone Based Firewall" section and firewall_zone_gated_acctest_test.go for
// the clean-error path this guards on a console where it isn't). The exact
// count assumes nothing else has added custom zones to this site; if that
// assumption ever breaks, switch this to a >=7 check instead.
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
					resource.TestCheckResourceAttr(dataSourceName, "zones.#", "7"),
				),
			},
		},
	})
}
