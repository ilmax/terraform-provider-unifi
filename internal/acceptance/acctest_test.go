package acceptance

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/ilmax/terraform-provider-unifi/internal/provider"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
	"github.com/ilmax/unifi-client-go/pkg/errors"
)

// testAccProtoV6ProviderFactories wires the real provider into
// terraform-plugin-testing's acceptance test runner.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"unifi": providerserver.NewProtocol6WithError(provider.New("acctest")()),
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

// testAccNewClient builds a raw unifi.Client from the same environment
// variables the provider itself uses, for CheckDestroy funcs (and anything
// else) that needs to talk to the API directly instead of through
// Terraform's state.
func testAccNewClient() (*unifi.Client, error) {
	return unifi.NewClient(unifi.Config{
		APIKey:        os.Getenv("UNIFI_API_KEY"),
		BaseURL:       os.Getenv("UNIFI_API_URL"),
		AllowInsecure: os.Getenv("UNIFI_ALLOW_INSECURE") == "true",
	})
}

// hasZoneBasedFirewall reports whether Zone Based Firewall is available on
// this site — used to skip tests that assert the "not configured" error
// path once test-infra/unifi's enable_zbf.sh (see README.md's "Zone Based
// Firewall" section; run via `make enable-zbf`) has force-enabled it, since
// that error can no longer occur. It also backs findFirewallZoneIDByName,
// since both need the same zone list.
func hasZoneBasedFirewall(t *testing.T) bool {
	t.Helper()
	return len(listFirewallZones(t)) > 0
}

// findFirewallZoneIDByName looks up a firewall zone's ID by its exact name
// directly via the API, since some tests need a real zone ID and there's no
// guarantee its UUID is stable across consoles. Matches the 7 system-default
// zone names enable_zbf.sh bootstraps (Internal/External/Gateway/VPN/
// Hotspot/DMZ/Management) — skips the test if Zone Based Firewall isn't
// enabled at all, or if that specific zone is missing for some other reason.
func findFirewallZoneIDByName(t *testing.T, name string) string {
	t.Helper()
	zones := listFirewallZones(t)
	if len(zones) == 0 {
		t.Skip("Zone Based Firewall is not enabled on this console — see test-infra/unifi/README.md, `make enable-zbf`")
	}
	for _, z := range zones {
		if z.Name == name {
			return z.Id
		}
	}
	t.Skipf("no firewall zone named %q found", name)
	return ""
}

func listFirewallZones(t *testing.T) []struct {
	Id   string `json:"id"`
	Name string `json:"name"`
} {
	t.Helper()
	client, err := testAccNewClient()
	if err != nil {
		t.Fatalf("build client: %v", err)
	}

	var page struct {
		Data []struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	siteID := os.Getenv("UNIFI_SITE_ID")
	if err := client.Get(context.Background(), fmt.Sprintf("/v1/sites/%s/firewall/zones", siteID), &page); err != nil {
		t.Fatalf("list firewall zones: %v", err)
	}
	return page.Data
}

// hasAdoptedGateway reports whether the site has at least one adopted
// device — used to skip tests that assert the "no gateway" error path once
// test-infra/unifi's unifi-emu fleet (see its README's "Adopted device
// fleet" section) has actually adopted one, since that error can no longer
// occur.
func hasAdoptedGateway(t *testing.T) bool {
	t.Helper()
	client, err := testAccNewClient()
	if err != nil {
		t.Fatalf("build client: %v", err)
	}

	var page struct {
		Data []struct {
			Id string `json:"id"`
		} `json:"data"`
	}
	siteID := os.Getenv("UNIFI_SITE_ID")
	if err := client.Get(context.Background(), fmt.Sprintf("/v1/sites/%s/devices", siteID), &page); err != nil {
		t.Fatalf("list devices: %v", err)
	}
	return len(page.Data) > 0
}

// testAccCheckResourceDestroyed builds a CheckDestroy func for a single
// resource type whose REST path isn't shared with any other type (unlike
// the DNS resources, which all live under the polymorphic dns/policies
// endpoint and share testAccCheckDNSPolicyDestroy in dns_acctest_test.go).
func testAccCheckResourceDestroyed(resourceType string, buildPath func(siteID, id string) string) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		client, err := testAccNewClient()
		if err != nil {
			return fmt.Errorf("build client for CheckDestroy: %w", err)
		}

		for _, rs := range s.RootModule().Resources {
			if rs.Type != resourceType {
				continue
			}

			siteID := rs.Primary.Attributes["site_id"]
			if siteID == "" {
				siteID = os.Getenv("UNIFI_SITE_ID")
			}

			apiPath := buildPath(siteID, rs.Primary.ID)
			if err := client.Get(context.Background(), apiPath, nil); err == nil {
				return fmt.Errorf("%s %s still exists", rs.Type, rs.Primary.ID)
			} else if !errors.IsNotFoundError(err) {
				return fmt.Errorf("unexpected error checking %s %s destroyed: %w", rs.Type, rs.Primary.ID, err)
			}
		}

		return nil
	}
}
