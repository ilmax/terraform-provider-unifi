package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccACLRulesDataSource_basic pre-populates two unifi_firewall_rule
// resources, then reads them back via the unifi_acl_rules data source —
// Terraform's dependency graph orders the resource creates before the data
// source read since the data source has no explicit reference to the
// resources, so this uses depends_on to force that ordering.
func TestAccACLRulesDataSource_basic(t *testing.T) {
	dataSourceName := "data.unifi_acl_rules.test"
	nameA := firewallRuleTestName("tf-aclds-a")
	nameB := firewallRuleTestName("tf-aclds-b")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_firewall_rule" "a" {
  type    = "IPV4"
  name    = %q
  action  = "ALLOW"
  enabled = true
}

resource "unifi_firewall_rule" "b" {
  type    = "IPV4"
  name    = %q
  action  = "BLOCK"
  enabled = false
}

data "unifi_acl_rules" "test" {
  depends_on = [unifi_firewall_rule.a, unifi_firewall_rule.b]
}
`, nameA, nameB),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "acl_rules.*", map[string]string{
						"name":    nameA,
						"action":  "ALLOW",
						"enabled": "true",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "acl_rules.*", map[string]string{
						"name":    nameB,
						"action":  "BLOCK",
						"enabled": "false",
					}),
				),
			},
		},
	})
}
