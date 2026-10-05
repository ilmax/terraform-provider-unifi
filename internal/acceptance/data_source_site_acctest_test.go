package acceptance

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccSiteDataSource_basic reads the console's pre-existing "Default"
// site by name — no resource needs to pre-populate anything, since every
// UniFi console always has at least this one site.
func TestAccSiteDataSource_basic(t *testing.T) {
	dataSourceName := "data.unifi_site.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "unifi_site" "test" {
  name = "Default"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", "Default"),
					resource.TestCheckResourceAttr(dataSourceName, "internal_reference", "default"),
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "site_id"),
				),
			},
		},
	})
}

// TestAccSiteDataSource_caseInsensitive locks in matchSitesByName's
// strings.EqualFold matching: a differently-cased lookup must still find
// the site.
func TestAccSiteDataSource_caseInsensitive(t *testing.T) {
	dataSourceName := "data.unifi_site.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "unifi_site" "test" {
  name = "dEfAuLt"
}
`,
				Check: resource.TestCheckResourceAttr(dataSourceName, "internal_reference", "default"),
			},
		},
	})
}

// TestAccSiteDataSource_notFound asserts a clean, specific error instead of
// an empty/zero-value result when no site matches.
func TestAccSiteDataSource_notFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "unifi_site" "test" {
  name = "does-not-exist-site"
}
`,
				ExpectError: regexp.MustCompile(`(?i)site\s+not\s+found`),
			},
		},
	})
}
