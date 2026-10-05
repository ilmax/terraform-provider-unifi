package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccFirewallPolicyOrdering_basic exercises a real create/update cycle
// against the system-default Internal/External zone pair — only possible
// once Zone Based Firewall is enabled (see README.md's "Zone Based
// Firewall" section). Empty before/after lists are a legitimate ordering
// (no custom policies to place relative to the zone pair's system-defined
// ones), so this doesn't need any unifi_firewall_policy to already exist.
func TestAccFirewallPolicyOrdering_basic(t *testing.T) {
	testAccPreCheck(t) // must run before findFirewallZoneIDByName's own API call
	sourceZoneID := findFirewallZoneIDByName(t, "Internal")
	destinationZoneID := findFirewallZoneIDByName(t, "External")
	resourceName := "unifi_firewall_policy_ordering.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_firewall_policy_ordering" "test" {
  source_firewall_zone_id      = %q
  destination_firewall_zone_id = %q
  before_system_defined        = []
  after_system_defined         = []
}
`, sourceZoneID, destinationZoneID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "source_firewall_zone_id", sourceZoneID),
					resource.TestCheckResourceAttr(resourceName, "destination_firewall_zone_id", destinationZoneID),
					resource.TestCheckResourceAttr(resourceName, "before_system_defined.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "after_system_defined.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// id is already "<site_id>/<source_zone_id>/<destination_zone_id>"
				// (see firewallPolicyOrderingID in resource_firewall_policy_ordering.go),
				// exactly the import identifier splitImportID3 expects.
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("resource not found in state: %s", resourceName)
					}
					return rs.Primary.ID, nil
				},
			},
		},
	})
}
