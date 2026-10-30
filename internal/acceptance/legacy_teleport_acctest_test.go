package acceptance

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/ilmax/terraform-provider-unifi/internal/legacyprovider"
)

var testAccLegacyProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"unifilegacy": providerserver.NewProtocol6WithError(legacyprovider.New("acctest")()),
}

func testAccLegacyProviderConfig() string {
	return fmt.Sprintf(`
provider "unifilegacy" {
  api_key        = %q
  api_url        = %q
  site_id        = %q
  allow_insecure = %t
}
`, os.Getenv("UNIFI_API_KEY"), os.Getenv("UNIFI_API_URL"), os.Getenv("UNIFI_SITE_ID"), os.Getenv("UNIFI_ALLOW_INSECURE") == "true")
}

// TestAccLegacyTeleport_basic changes the per-site Teleport setting, reads it
// back through the data source, imports it and finally restores the console's
// defaults (destroy leaves the singleton setting untouched).
func TestAccLegacyTeleport_basic(t *testing.T) {
	const res = "unifilegacy_teleport.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccLegacyProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLegacyProviderConfig() + `
resource "unifilegacy_teleport" "test" {
  enabled     = true
  subnet_cidr = "192.168.99.0/24"
}

data "unifilegacy_teleport" "test" {
  depends_on = [unifilegacy_teleport.test]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res, "enabled", "true"),
					resource.TestCheckResourceAttr(res, "subnet_cidr", "192.168.99.0/24"),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttr("data.unifilegacy_teleport.test", "enabled", "true"),
					resource.TestCheckResourceAttr("data.unifilegacy_teleport.test", "subnet_cidr", "192.168.99.0/24"),
					resource.TestCheckResourceAttrPair(res, "id", "data.unifilegacy_teleport.test", "id"),
				),
			},
			{
				ResourceName:      res,
				ImportState:       true,
				ImportStateId:     os.Getenv("UNIFI_SITE_ID"),
				ImportStateVerify: true,
			},
			{
				// Disabling alone keeps the controller's stored subnet.
				Config: testAccLegacyProviderConfig() + `
resource "unifilegacy_teleport" "test" {
  enabled = false
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res, "enabled", "false"),
					resource.TestCheckResourceAttr(res, "subnet_cidr", "192.168.99.0/24"),
				),
			},
			{
				Config: testAccLegacyProviderConfig() + `
resource "unifilegacy_teleport" "test" {
  enabled     = false
  subnet_cidr = "192.168.99.0/24"
}
`,
				ExpectError: regexp.MustCompile(`subnet_cidr requires Teleport to be enabled`),
			},
			{
				Config: testAccLegacyProviderConfig() + `
resource "unifilegacy_teleport" "test" {
  enabled     = true
  subnet_cidr = ""
}
`,
				ExpectError: regexp.MustCompile(`Invalid subnet_cidr`),
			},
			{
				// Re-enabling clears the stored subnet, which also leaves the console
				// as a fresh one starts. Never send an empty subnet to get there: the
				// controller stores it and then fails to list networks (HTTP 500).
				Config: testAccLegacyProviderConfig() + `
resource "unifilegacy_teleport" "test" {
  enabled = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res, "enabled", "true"),
					resource.TestCheckResourceAttr(res, "subnet_cidr", ""),
				),
			},
		},
	})
}
