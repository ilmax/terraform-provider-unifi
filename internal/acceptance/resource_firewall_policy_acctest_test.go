package acceptance

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// These two validation paths are checked in buildFirewallPolicyPayload
// before any API call, so — unlike the rest of unifi_firewall_policy — they
// don't need Zone Based Firewall to be configured and can run on any
// console, including this one (see firewall_zone_gated_acctest_test.go for
// the API-backed happy path this resource can't otherwise exercise here).

func TestAccFirewallPolicy_invalidAction(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "unifi_firewall_policy" "test" {
  name                = "tf-acc-policy-badaction"
  action              = "BOGUS"
  source_zone_id      = "` + fakeUUID + `"
  destination_zone_id = "00000000-0000-0000-0000-000000000002"
}
`,
				ExpectError: regexp.MustCompile(`(?i)invalid\s+action`),
			},
		},
	})
}

func TestAccFirewallPolicy_allowReturnTrafficOnlyForAllow(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "unifi_firewall_policy" "test" {
  name                  = "tf-acc-policy-badreturn"
  action                = "BLOCK"
  allow_return_traffic  = true
  source_zone_id        = "` + fakeUUID + `"
  destination_zone_id   = "00000000-0000-0000-0000-000000000002"
}
`,
				ExpectError: regexp.MustCompile(`(?i)only\s+supported\s+for\s+action\s+=\s+ALLOW`),
			},
		},
	})
}

// TestAccFirewallPolicy_allowBasic exercises a real create/update/import/
// destroy cycle for an ALLOW policy between two real zones — only possible
// once Zone Based Firewall is enabled (see README.md's "Zone Based
// Firewall" section). Deliberately never changes `action` across steps:
// allow_return_traffic is Computed with UseStateForUnknown, so switching
// action away from ALLOW without also clearing it in the same config change
// hits a real (separately tracked) plan-modifier interaction bug — see
// TestAccFirewallPolicy_blockUpdate for the non-ALLOW update path instead.
func TestAccFirewallPolicy_allowBasic(t *testing.T) {
	testAccPreCheck(t) // must run before findFirewallZoneIDByName's own API call
	sourceZoneID := findFirewallZoneIDByName(t, "Internal")
	destinationZoneID := findFirewallZoneIDByName(t, "External")
	resourceName := "unifi_firewall_policy.test"
	name := "tf-acc-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: testAccCheckResourceDestroyed("unifi_firewall_policy", func(siteID, id string) string {
			return fmt.Sprintf("/v1/sites/%s/firewall/policies/%s", siteID, id)
		}),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_firewall_policy" "test" {
  name                = %q
  action              = "ALLOW"
  source_zone_id      = %q
  destination_zone_id = %q
}
`, name, sourceZoneID, destinationZoneID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "action", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "ip_protocol_scope", "IPV4_AND_IPV6"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "index"),
					resource.TestCheckResourceAttrSet(resourceName, "origin"),
				),
			},
			{
				// Update in place: description, logging, and ip_protocol_scope
				// change; action stays ALLOW throughout.
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_firewall_policy" "test" {
  name                = %q
  description         = "updated by acceptance test"
  action              = "ALLOW"
  logging_enabled     = true
  ip_protocol_scope   = "IPV4"
  source_zone_id      = %q
  destination_zone_id = %q
}
`, name, sourceZoneID, destinationZoneID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "updated by acceptance test"),
					resource.TestCheckResourceAttr(resourceName, "logging_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "ip_protocol_scope", "IPV4"),
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

// TestAccFirewallPolicy_blockUpdate covers the same real update-in-place
// cycle as TestAccFirewallPolicy_allowBasic, but for a BLOCK policy —
// exercising connection_state_filter, which only makes sense outside
// action = ALLOW.
func TestAccFirewallPolicy_blockUpdate(t *testing.T) {
	testAccPreCheck(t)
	sourceZoneID := findFirewallZoneIDByName(t, "Internal")
	destinationZoneID := findFirewallZoneIDByName(t, "External")
	resourceName := "unifi_firewall_policy.test"
	name := "tf-acc-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: testAccCheckResourceDestroyed("unifi_firewall_policy", func(siteID, id string) string {
			return fmt.Sprintf("/v1/sites/%s/firewall/policies/%s", siteID, id)
		}),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_firewall_policy" "test" {
  name                     = %q
  action                   = "BLOCK"
  connection_state_filter  = ["NEW", "ESTABLISHED"]
  source_zone_id           = %q
  destination_zone_id      = %q
}
`, name, sourceZoneID, destinationZoneID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "action", "BLOCK"),
					resource.TestCheckResourceAttr(resourceName, "connection_state_filter.#", "2"),
					resource.TestCheckNoResourceAttr(resourceName, "allow_return_traffic"),
				),
			},
			{
				// Update in place: widen the connection state filter.
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_firewall_policy" "test" {
  name                     = %q
  action                   = "BLOCK"
  connection_state_filter  = ["NEW", "ESTABLISHED", "RELATED"]
  source_zone_id           = %q
  destination_zone_id      = %q
}
`, name, sourceZoneID, destinationZoneID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "connection_state_filter.#", "3"),
				),
			},
		},
	})
}
