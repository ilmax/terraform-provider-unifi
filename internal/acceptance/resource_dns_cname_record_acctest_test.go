package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDNSCNAMERecord_basic(t *testing.T) {
	resourceName := "unifi_dns_cname_record.test"
	domain := acctest.RandomWithPrefix("tf-acc-cname") + ".example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDNSPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_dns_cname_record" "test" {
  enabled       = true
  domain        = %q
  target_domain = "target-a.example.com"
  ttl_seconds   = 300
}
`, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "domain", domain),
					resource.TestCheckResourceAttr(resourceName, "target_domain", "target-a.example.com"),
					resource.TestCheckResourceAttr(resourceName, "ttl_seconds", "300"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "site_id"),
					resource.TestCheckResourceAttrSet(resourceName, "origin"),
				),
			},
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_dns_cname_record" "test" {
  enabled       = false
  domain        = %q
  target_domain = "target-b.example.com"
  ttl_seconds   = 600
}
`, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "target_domain", "target-b.example.com"),
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
