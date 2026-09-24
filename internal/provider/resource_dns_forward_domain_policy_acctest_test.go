package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDNSForwardDomainPolicy_basic(t *testing.T) {
	resourceName := "unifi_dns_forward_domain_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "unifi_dns_forward_domain_policy" "test" {
  enabled    = true
  domain     = "acctest-forward.example.com"
  ip_address = "192.0.2.53"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "domain", "acctest-forward.example.com"),
					resource.TestCheckResourceAttr(resourceName, "ip_address", "192.0.2.53"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "site_id"),
					resource.TestCheckResourceAttrSet(resourceName, "origin"),
				),
			},
			{
				Config: testAccProviderConfig() + `
resource "unifi_dns_forward_domain_policy" "test" {
  enabled    = false
  domain     = "acctest-forward.example.com"
  ip_address = "192.0.2.54"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "ip_address", "192.0.2.54"),
					resource.TestCheckResourceAttrSet(resourceName, "origin"),
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
