package acceptance

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccFirewallZoneDataSource_found reads a real system-default zone by
// name — only possible once Zone Based Firewall is enabled (see README.md's
// "Zone Based Firewall" section and firewall_zone_gated_acctest_test.go for
// the clean-error path this guards on a console where it isn't).
func TestAccFirewallZoneDataSource_found(t *testing.T) {
	testAccPreCheck(t) // must run before hasZoneBasedFirewall's own API call
	if !hasZoneBasedFirewall(t) {
		t.Skip("Zone Based Firewall is not enabled on this console — see test-infra/unifi/README.md, `make enable-zbf`")
	}

	dataSourceName := "data.unifi_firewall_zone.test"

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
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", "Internal"),
					resource.TestCheckResourceAttrSet(dataSourceName, "zone_id"),
					resource.TestCheckResourceAttr(dataSourceName, "origin", "SYSTEM_DEFINED"),
				),
			},
		},
	})
}
