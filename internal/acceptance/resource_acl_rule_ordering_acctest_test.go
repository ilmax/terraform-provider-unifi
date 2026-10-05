package acceptance

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// aclRuleOrderingImportStateIdFunc returns the bare site_id as the import
// ID, matching aclRuleOrderingResource.ImportState's `<site_id>` format
// (unlike the "<site_id>/<id>" format most other resources use).
func aclRuleOrderingImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found in state: %s", resourceName)
		}
		return rs.Primary.Attributes["site_id"], nil
	}
}

// TestAccACLRuleOrdering_basic pre-populates two unifi_firewall_rule
// resources, orders them, then swaps the order — a realistic "reprioritize
// my rules" workflow.
//
// No CheckDestroy here deliberately: aclRuleOrderingResource.Delete (see
// resource_acl_rule_ordering.go) only calls resp.State.RemoveResource — it
// makes no API call at all, since "ordering" isn't really a deletable
// object server-side (there's nothing to un-order back to). A CheckDestroy
// would have nothing real to verify.
func TestAccACLRuleOrdering_basic(t *testing.T) {
	resourceName := "unifi_acl_rule_ordering.test"
	nameA := firewallRuleTestName("tf-aclord-a")
	nameB := firewallRuleTestName("tf-aclord-b")

	rulesConfig := fmt.Sprintf(`
resource "unifi_firewall_rule" "a" {
  type    = "IPV4"
  name    = %q
  action  = "ALLOW"
  enabled = true
}

resource "unifi_firewall_rule" "b" {
  type    = "IPV4"
  name    = %q
  action  = "ALLOW"
  enabled = true
}
`, nameA, nameB)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + rulesConfig + `
resource "unifi_acl_rule_ordering" "test" {
  ordered_acl_rule_ids = [unifi_firewall_rule.a.id, unifi_firewall_rule.b.id]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "ordered_acl_rule_ids.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "ordered_acl_rule_ids.0", "unifi_firewall_rule.a", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "ordered_acl_rule_ids.1", "unifi_firewall_rule.b", "id"),
				),
			},
			{
				// Swap the order in place.
				Config: testAccProviderConfig() + rulesConfig + `
resource "unifi_acl_rule_ordering" "test" {
  ordered_acl_rule_ids = [unifi_firewall_rule.b.id, unifi_firewall_rule.a.id]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "ordered_acl_rule_ids.0", "unifi_firewall_rule.b", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "ordered_acl_rule_ids.1", "unifi_firewall_rule.a", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: aclRuleOrderingImportStateIdFunc(resourceName),
			},
		},
	})
}

// TestAccACLRuleOrdering_invalidUUID asserts a non-UUID list entry is
// rejected at plan time with a clear error, never reaching the API.
func TestAccACLRuleOrdering_invalidUUID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "unifi_acl_rule_ordering" "test" {
  ordered_acl_rule_ids = ["not-a-uuid"]
}
`,
				ExpectError: regexp.MustCompile(`(?i)not\s+a\s+valid\s+UUID`),
			},
		},
	})
}
