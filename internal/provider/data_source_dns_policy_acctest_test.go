package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDNSPolicyDataSource_basic(t *testing.T) {
	dataSourceName := "data.unifi_dns_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "unifi_dns_txt_record" "lookup" {
  enabled = true
  domain  = "acctest-ds-policy.example.com"
  text    = "acctest-lookup"
}

data "unifi_dns_policy" "test" {
  policy_id = unifi_dns_txt_record.lookup.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "type", "TXT_RECORD"),
					resource.TestCheckResourceAttr(dataSourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "domain", "acctest-ds-policy.example.com"),
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "origin"),
				),
			},
		},
	})
}

func TestAccDNSPoliciesDataSource_basic(t *testing.T) {
	dataSourceName := "data.unifi_dns_policies.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "unifi_dns_txt_record" "lookup" {
  enabled = true
  domain  = "acctest-ds-policies.example.com"
  text    = "acctest-lookup"
}

data "unifi_dns_policies" "test" {
  domain = unifi_dns_txt_record.lookup.domain

  depends_on = [unifi_dns_txt_record.lookup]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "domain", "acctest-ds-policies.example.com"),
					resource.TestCheckResourceAttr(dataSourceName, "policies.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "policies.0.type", "TXT_RECORD"),
					resource.TestCheckResourceAttr(dataSourceName, "policies.0.domain", "acctest-ds-policies.example.com"),
				),
			},
		},
	})
}
