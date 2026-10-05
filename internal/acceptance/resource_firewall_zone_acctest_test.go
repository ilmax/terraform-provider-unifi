package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccFirewallZone_basic exercises a real create/update/import/destroy
// cycle for a custom firewall zone — only possible once Zone Based Firewall
// is enabled (see README.md's "Zone Based Firewall" section and
// firewall_zone_gated_acctest_test.go for the clean-error path this guards
// on a console where it isn't).
func TestAccFirewallZone_basic(t *testing.T) {
	testAccPreCheck(t) // must run before hasZoneBasedFirewall's own API call
	if !hasZoneBasedFirewall(t) {
		t.Skip("Zone Based Firewall is not enabled on this console — see test-infra/unifi/README.md, `make enable-zbf`")
	}

	resourceName := "unifi_firewall_zone.test"
	name := "tf-acc-zone-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: testAccCheckResourceDestroyed("unifi_firewall_zone", func(siteID, id string) string {
			return fmt.Sprintf("/v1/sites/%s/firewall/zones/%s", siteID, id)
		}),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_firewall_zone" "test" {
  name = %q
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "site_id"),
				),
			},
			{
				// Update in place: rename.
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_firewall_zone" "test" {
  name = %q
}
`, name+"-renamed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name+"-renamed"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccSiteScopedImportStateIdFunc(resourceName),
			},
		},
	})
}
