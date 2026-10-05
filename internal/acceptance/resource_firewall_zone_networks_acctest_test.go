package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccFirewallZoneNetworks_basic assigns a real network to a real custom
// zone and then unassigns it — only possible once Zone Based Firewall is
// enabled (see README.md's "Zone Based Firewall" section). Two things
// verified directly against the API before writing this:
//
//   - Only a GATEWAY-managed network can belong to a zone at all — an
//     UNMANAGED one is rejected with the confusing
//     api.firewall.zone.network-does-not-exist ("Configured network does
//     not exist"), even though the network genuinely exists.
//   - Unassigning (an empty network_ids) doesn't leave the network
//     zoneless — the real API falls back to the site's Internal zone,
//     since a GATEWAY network's zone_id is mandatory once Zone Based
//     Firewall is enabled (see TestAccNetwork_gatewayBasic).
func TestAccFirewallZoneNetworks_basic(t *testing.T) {
	testAccPreCheck(t) // must run before hasZoneBasedFirewall's own API call
	if !hasZoneBasedFirewall(t) {
		t.Skip("Zone Based Firewall is not enabled on this console — see test-infra/unifi/README.md, `make enable-zbf`")
	}

	zoneName := "tf-acc-zonenet-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	netName := networkTestName("tf-zonenetnet")
	resourceName := "unifi_firewall_zone_networks.test"

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

resource "unifi_network" "helper" {
  name       = %q
  management = "GATEWAY"
  enabled    = true
  vlan_id    = 960
  zone_id    = unifi_firewall_zone.test.id

  isolation_enabled        = false
  cellular_backup_enabled  = false
  internet_access_enabled  = true

  ipv4_configuration = {
    host_ip_address = "10.251.1.1"
    prefix_length   = 24
  }
}

resource "unifi_firewall_zone_networks" "test" {
  zone_id     = unifi_firewall_zone.test.id
  network_ids = [unifi_network.helper.id]
}
`, zoneName, netName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "zone_id", "unifi_firewall_zone.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "network_ids.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "network_ids.0", "unifi_network.helper", "id"),
					resource.TestCheckResourceAttrPair("unifi_network.helper", "zone_id", "unifi_firewall_zone.test", "id"),
				),
			},
		},
	})
}
