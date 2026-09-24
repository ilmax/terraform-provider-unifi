package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDNSAAAARecord_basic(t *testing.T) {
	resourceName := "unifi_dns_aaaa_record.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "unifi_dns_aaaa_record" "test" {
  enabled      = true
  domain       = "acctest-aaaa.example.com"
  ipv6_address = "2001:db8::10"
  ttl_seconds  = 300
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "domain", "acctest-aaaa.example.com"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address", "2001:db8::10"),
					resource.TestCheckResourceAttr(resourceName, "ttl_seconds", "300"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "site_id"),
					resource.TestCheckResourceAttrSet(resourceName, "origin"),
				),
			},
			{
				Config: testAccProviderConfig() + `
resource "unifi_dns_aaaa_record" "test" {
  enabled      = false
  domain       = "acctest-aaaa.example.com"
  ipv6_address = "2001:db8::20"
  ttl_seconds  = 600
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address", "2001:db8::20"),
					resource.TestCheckResourceAttr(resourceName, "ttl_seconds", "600"),
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
