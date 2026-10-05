package acceptance

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccTrafficMatchingList_ipv4 covers a real create/update/import/destroy
// cycle for the IPV4_ADDRESSES variant. Unlike unifi_firewall_policy, this
// endpoint (/traffic-matching-lists) is not nested under /firewall/ and is
// not gated behind Zone Based Firewall, so it's fully testable on this
// console.
func TestAccTrafficMatchingList_ipv4(t *testing.T) {
	resourceName := "unifi_traffic_matching_list.test"
	name := "tf-acc-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: testAccCheckResourceDestroyed("unifi_traffic_matching_list", func(siteID, id string) string {
			return fmt.Sprintf("/v1/sites/%s/traffic-matching-lists/%s", siteID, id)
		}),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_traffic_matching_list" "test" {
  name = %q
  type = "IPV4_ADDRESSES"
  items_json = jsonencode([
    { type = "IP_ADDRESS", value = "192.168.1.5" },
    { type = "SUBNET", value = "10.0.0.0/24" },
    { type = "IP_ADDRESS_RANGE", start = "10.0.0.10", stop = "10.0.0.20" },
  ])
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "type", "IPV4_ADDRESSES"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "site_id"),
				),
			},
			{
				// Update: rename and shrink to a single item, verifying the
				// real PUT-based update applies in place (no replacement).
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_traffic_matching_list" "test" {
  name = %q
  type = "IPV4_ADDRESSES"
  items_json = jsonencode([
    { type = "IP_ADDRESS", value = "192.168.1.99" },
  ])
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

// TestAccTrafficMatchingList_ports covers the PORTS variant, which the IPv4
// test above doesn't exercise (single port + a port range).
func TestAccTrafficMatchingList_ports(t *testing.T) {
	resourceName := "unifi_traffic_matching_list.test"
	name := "tf-acc-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: testAccCheckResourceDestroyed("unifi_traffic_matching_list", func(siteID, id string) string {
			return fmt.Sprintf("/v1/sites/%s/traffic-matching-lists/%s", siteID, id)
		}),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_traffic_matching_list" "test" {
  name = %q
  type = "PORTS"
  items_json = jsonencode([
    { type = "PORT_NUMBER", value = 443 },
    { type = "PORT_NUMBER_RANGE", start = 8000, stop = 8100 },
  ])
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "type", "PORTS"),
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

// TestAccTrafficMatchingList_ipv6 covers the IPV6_ADDRESSES variant, which
// (unlike IPv4) has no IP_ADDRESS_RANGE item type.
func TestAccTrafficMatchingList_ipv6(t *testing.T) {
	resourceName := "unifi_traffic_matching_list.test"
	name := "tf-acc-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: testAccCheckResourceDestroyed("unifi_traffic_matching_list", func(siteID, id string) string {
			return fmt.Sprintf("/v1/sites/%s/traffic-matching-lists/%s", siteID, id)
		}),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_traffic_matching_list" "test" {
  name = %q
  type = "IPV6_ADDRESSES"
  items_json = jsonencode([
    { type = "IP_ADDRESS", value = "2001:db8::1" },
    { type = "SUBNET", value = "2001:db8::/32" },
  ])
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "type", "IPV6_ADDRESSES"),
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

// TestAccTrafficMatchingList_invalidType exercises the client-side
// validation in buildTrafficMatchingListPayload, firing before any API call.
func TestAccTrafficMatchingList_invalidType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "unifi_traffic_matching_list" "test" {
  name       = "tf-acc-badtype"
  type       = "BOGUS"
  items_json = jsonencode([{ type = "IP_ADDRESS", value = "10.0.0.1" }])
}
`,
				ExpectError: regexp.MustCompile(`(?i)invalid\s+type`),
			},
		},
	})
}
