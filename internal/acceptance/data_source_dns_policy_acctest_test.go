package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDNSPolicyDataSource_basic(t *testing.T) {
	dataSourceName := "data.unifi_dns_policy.test"
	domain := acctest.RandomWithPrefix("tf-acc-ds-policy") + ".example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_dns_txt_record" "lookup" {
  enabled = true
  domain  = %q
  text    = "acctest-lookup"
}

data "unifi_dns_policy" "test" {
  policy_id = unifi_dns_txt_record.lookup.id
}
`, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "type", "TXT_RECORD"),
					resource.TestCheckResourceAttr(dataSourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "domain", domain),
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "origin"),
				),
			},
		},
	})
}

func TestAccDNSPoliciesDataSource_basic(t *testing.T) {
	dataSourceName := "data.unifi_dns_policies.test"
	domain := acctest.RandomWithPrefix("tf-acc-ds-policies") + ".example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_dns_txt_record" "lookup" {
  enabled = true
  domain  = %q
  text    = "acctest-lookup"
}

data "unifi_dns_policies" "test" {
  domain = unifi_dns_txt_record.lookup.domain

  depends_on = [unifi_dns_txt_record.lookup]
}
`, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "domain", domain),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "policies.*", map[string]string{
						"type":   "TXT_RECORD",
						"domain": domain,
					}),
				),
			},
		},
	})
}
