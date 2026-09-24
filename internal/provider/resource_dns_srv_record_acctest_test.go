package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDNSSRVRecord_basic(t *testing.T) {
	resourceName := "unifi_dns_srv_record.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "unifi_dns_srv_record" "test" {
  enabled       = true
  domain        = "acctest-srv.example.com"
  service       = "_ldap"
  protocol      = "_tcp"
  server_domain = "srv-a.example.com"
  port          = 389
  priority      = 10
  weight        = 20
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "domain", "acctest-srv.example.com"),
					resource.TestCheckResourceAttr(resourceName, "service", "_ldap"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "_tcp"),
					resource.TestCheckResourceAttr(resourceName, "server_domain", "srv-a.example.com"),
					resource.TestCheckResourceAttr(resourceName, "port", "389"),
					resource.TestCheckResourceAttr(resourceName, "priority", "10"),
					resource.TestCheckResourceAttr(resourceName, "weight", "20"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "site_id"),
					resource.TestCheckResourceAttrSet(resourceName, "origin"),
				),
			},
			{
				Config: testAccProviderConfig() + `
resource "unifi_dns_srv_record" "test" {
  enabled       = false
  domain        = "acctest-srv.example.com"
  service       = "_ldap"
  protocol      = "_tcp"
  server_domain = "srv-b.example.com"
  port          = 636
  priority      = 5
  weight        = 30
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "server_domain", "srv-b.example.com"),
					resource.TestCheckResourceAttr(resourceName, "port", "636"),
					resource.TestCheckResourceAttr(resourceName, "priority", "5"),
					resource.TestCheckResourceAttr(resourceName, "weight", "30"),
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
