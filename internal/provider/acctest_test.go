package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testAccProtoV6ProviderFactories wires the real provider into
// terraform-plugin-testing's acceptance test runner.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"unifi": providerserver.NewProtocol6WithError(New("acctest")()),
}

// testAccPreCheck gates every acceptance test on TF_ACC (handled by
// resource.Test itself) plus the credentials this provider actually needs.
// Unlike a hard failure, this skips: acceptance tests are opt-in and must
// not break `go test ./...` in environments (including CI) that have no
// UniFi console to talk to. See AGENTS.md's "UniFi Network API Reference"
// section for why these credentials can't be provisioned automatically.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	for _, name := range []string{"UNIFI_API_KEY", "UNIFI_SITE_ID"} {
		if os.Getenv(name) == "" {
			t.Skipf("%s must be set to run acceptance tests", name)
		}
	}
}

// testAccProviderConfig renders the `provider "unifi" {}` block from
// environment variables. It deliberately sets site_id at the provider
// level (not per-resource) so every acceptance test also exercises the
// provider -> resource site_id fallback documented in docs/index.md.
func testAccProviderConfig() string {
	allowInsecure := os.Getenv("UNIFI_ALLOW_INSECURE") == "true"
	return fmt.Sprintf(`
provider "unifi" {
  api_key        = %q
  api_url        = %q
  site_id        = %q
  allow_insecure = %t
}
`, os.Getenv("UNIFI_API_KEY"), os.Getenv("UNIFI_API_URL"), os.Getenv("UNIFI_SITE_ID"), allowInsecure)
}

// testAccSiteScopedImportStateIdFunc builds the "<site_id>/<id>" import
// identifier documented for every DNS resource. Falls back to the
// provider-level UNIFI_SITE_ID when the resource itself omits site_id
// (the common case exercised by these tests), matching resolveSiteID's
// own fallback behavior in internal/provider/helpers.go.
func testAccSiteScopedImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found in state: %s", resourceName)
		}
		siteID := rs.Primary.Attributes["site_id"]
		if siteID == "" {
			siteID = os.Getenv("UNIFI_SITE_ID")
		}
		if siteID == "" {
			return "", fmt.Errorf("no site_id available for import (resource %s)", resourceName)
		}
		return fmt.Sprintf("%s/%s", siteID, rs.Primary.ID), nil
	}
}
