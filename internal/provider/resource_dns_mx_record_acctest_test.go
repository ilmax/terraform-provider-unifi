package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDNSMXRecord_basic(t *testing.T) {
	resourceName := "unifi_dns_mx_record.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "unifi_dns_mx_record" "test" {
  enabled             = true
  domain              = "acctest-mx.example.com"
  mail_server_domain  = "mail-a.example.com"
  priority            = 10
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "domain", "acctest-mx.example.com"),
					resource.TestCheckResourceAttr(resourceName, "mail_server_domain", "mail-a.example.com"),
					resource.TestCheckResourceAttr(resourceName, "priority", "10"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "site_id"),
					resource.TestCheckResourceAttrSet(resourceName, "origin"),
				),
			},
			{
				Config: testAccProviderConfig() + `
resource "unifi_dns_mx_record" "test" {
  enabled            = false
  domain             = "acctest-mx.example.com"
  mail_server_domain = "mail-b.example.com"
  priority           = 20
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "mail_server_domain", "mail-b.example.com"),
					resource.TestCheckResourceAttr(resourceName, "priority", "20"),
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
