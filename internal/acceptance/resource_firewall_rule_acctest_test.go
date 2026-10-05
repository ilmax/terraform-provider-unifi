package acceptance

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// firewallRuleTestName generates a short unique name (ACL rule names have no
// documented length cap, but staying short avoids ever hitting one).
func firewallRuleTestName(prefix string) string {
	return prefix + "-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
}

// TestAccFirewallRule_ipv4Basic exercises an IPV4 rule with both source and
// destination filters plus a protocol filter, then updates the action and
// filters in place, then imports.
func TestAccFirewallRule_ipv4Basic(t *testing.T) {
	resourceName := "unifi_firewall_rule.test"
	name := firewallRuleTestName("tf-acl-ipv4")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: testAccCheckResourceDestroyed("unifi_firewall_rule", func(siteID, id string) string {
			return fmt.Sprintf("/v1/sites/%s/acl-rules/%s", siteID, id)
		}),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_firewall_rule" "test" {
  type    = "IPV4"
  name    = %q
  action  = "ALLOW"
  enabled = true

  source_filter = {
    type                    = "IP_ADDRESSES_OR_SUBNETS"
    ip_addresses_or_subnets = ["10.0.0.0/24"]
  }
  destination_filter = {
    type        = "PORTS"
    port_filter = [443]
  }
  protocol_filter = ["TCP"]
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "type", "IPV4"),
					resource.TestCheckResourceAttr(resourceName, "action", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "protocol_filter.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "protocol_filter.*", "TCP"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "site_id"),
				),
			},
			{
				// Update in place: flip action, widen protocol filter.
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_firewall_rule" "test" {
  type    = "IPV4"
  name    = %q
  action  = "BLOCK"
  enabled = true

  source_filter = {
    type                    = "IP_ADDRESSES_OR_SUBNETS"
    ip_addresses_or_subnets = ["10.0.0.0/24"]
  }
  destination_filter = {
    type        = "PORTS"
    port_filter = [443]
  }
  protocol_filter = ["TCP", "UDP"]
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "action", "BLOCK"),
					resource.TestCheckResourceAttr(resourceName, "protocol_filter.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "protocol_filter.*", "TCP"),
					resource.TestCheckTypeSetElemAttr(resourceName, "protocol_filter.*", "UDP"),
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

// TestAccFirewallRule_macWithNetworkFilter pre-populates a unifi_network,
// then creates a MAC-type ACL rule referencing it via network_id_filter —
// this is the field that was silently dropped after create/update before
// CreateACLRuleResponseMac/UpdateACLRuleResponseMac gained the field, so a
// perpetual diff here would mean that regression came back.
func TestAccFirewallRule_macWithNetworkFilter(t *testing.T) {
	resourceName := "unifi_firewall_rule.test"
	ruleName := firewallRuleTestName("tf-acl-mac")
	netName := networkTestName("tf-aclnet")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: testAccCheckResourceDestroyed("unifi_firewall_rule", func(siteID, id string) string {
			return fmt.Sprintf("/v1/sites/%s/acl-rules/%s", siteID, id)
		}),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_network" "helper" {
  name       = %q
  management = "UNMANAGED"
  enabled    = true
  vlan_id    = 950
}

resource "unifi_firewall_rule" "test" {
  type    = "MAC"
  name    = %q
  action  = "BLOCK"
  enabled = true

  network_id_filter = unifi_network.helper.id

  source_filter = {
    type          = "MAC_ADDRESSES"
    mac_addresses = ["aa:bb:cc:dd:ee:ff"]
  }

  depends_on = [unifi_network.helper]
}
`, netName, ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "type", "MAC"),
					resource.TestCheckResourceAttrPair(resourceName, "network_id_filter", "unifi_network.helper", "id"),
				),
			},
		},
	})
}

// TestAccFirewallRule_invalidType asserts the API's real "unknown type"
// error surfaces cleanly for an unsupported ACL rule type.
func TestAccFirewallRule_invalidType(t *testing.T) {
	name := firewallRuleTestName("tf-acl-badtype")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_firewall_rule" "test" {
  type    = "BOGUS"
  name    = %q
  action  = "ALLOW"
  enabled = true
}
`, name),
				ExpectError: regexp.MustCompile(`(?i)invalid\s+.*type.*value\s+'BOGUS'`),
			},
		},
	})
}

// TestAccFirewallRule_invalidAction asserts an unsupported action value is
// rejected cleanly instead of silently coercing to a default.
func TestAccFirewallRule_invalidAction(t *testing.T) {
	name := firewallRuleTestName("tf-acl-badaction")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_firewall_rule" "test" {
  type    = "IPV4"
  name    = %q
  action  = "BOGUS"
  enabled = true
}
`, name),
				ExpectError: regexp.MustCompile(`(?i)type\s+mismatch|action`),
			},
		},
	})
}

// TestAccFirewallRule_invalidFilterType asserts an unsupported
// source_filter.type value is rejected at plan time with a clear error,
// never reaching the API.
func TestAccFirewallRule_invalidFilterType(t *testing.T) {
	name := firewallRuleTestName("tf-acl-badfiltertype")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_firewall_rule" "test" {
  type    = "IPV4"
  name    = %q
  action  = "ALLOW"
  enabled = true

  source_filter = {
    type = "BOGUS"
  }
}
`, name),
				ExpectError: regexp.MustCompile(`(?i)invalid\s+filter\s+type`),
			},
		},
	})
}

// TestAccFirewallRule_filterMissingRequiredField asserts that omitting the
// field a given source_filter.type requires (here, port_filter for type =
// PORTS) is rejected at plan time instead of silently sending an incomplete
// filter to the API.
func TestAccFirewallRule_filterMissingRequiredField(t *testing.T) {
	name := firewallRuleTestName("tf-acl-missingfield")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_firewall_rule" "test" {
  type    = "IPV4"
  name    = %q
  action  = "ALLOW"
  enabled = true

  destination_filter = {
    type = "PORTS"
  }
}
`, name),
				ExpectError: regexp.MustCompile(`(?i)missing\s+port_filter`),
			},
		},
	})
}
