package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDNSAAAARecord_basic(t *testing.T) {
	resourceName := "unifi_dns_aaaa_record.test"
	domain := acctest.RandomWithPrefix("tf-acc-aaaa") + ".example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDNSPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_dns_aaaa_record" "test" {
  enabled      = true
  domain       = %q
  ipv6_address = "2001:db8::10"
  ttl_seconds  = 300
}
`, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "domain", domain),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address", "2001:db8::10"),
					resource.TestCheckResourceAttr(resourceName, "ttl_seconds", "300"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "site_id"),
					resource.TestCheckResourceAttrSet(resourceName, "origin"),
				),
			},
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_dns_aaaa_record" "test" {
  enabled      = false
  domain       = %q
  ipv6_address = "2001:db8::20"
  ttl_seconds  = 600
}
`, domain),
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
