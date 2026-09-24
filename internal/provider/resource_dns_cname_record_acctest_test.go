package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDNSCNAMERecord_basic(t *testing.T) {
	resourceName := "unifi_dns_cname_record.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "unifi_dns_cname_record" "test" {
  enabled       = true
  domain        = "acctest-cname.example.com"
  target_domain = "target-a.example.com"
  ttl_seconds   = 300
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "domain", "acctest-cname.example.com"),
					resource.TestCheckResourceAttr(resourceName, "target_domain", "target-a.example.com"),
					resource.TestCheckResourceAttr(resourceName, "ttl_seconds", "300"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "site_id"),
					resource.TestCheckResourceAttrSet(resourceName, "origin"),
				),
			},
			{
				Config: testAccProviderConfig() + `
resource "unifi_dns_cname_record" "test" {
  enabled       = false
  domain        = "acctest-cname.example.com"
  target_domain = "target-b.example.com"
  ttl_seconds   = 600
}
`,
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
