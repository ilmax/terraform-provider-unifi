package acceptance

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/ilmax/unifi-client-go/pkg/errors"
)

// testAccCheckDNSPolicyDestroy verifies every unifi_dns_* resource left in
// state after a test run no longer exists on the server — catching a
// Delete call that returns success without actually deleting anything. All
// seven DNS record types share this one check because they all live under
// the polymorphic dns/policies endpoint (see internal/sdkcompat/dns).
func testAccCheckDNSPolicyDestroy(s *terraform.State) error {
	client, err := testAccNewClient()
	if err != nil {
		return fmt.Errorf("build client for CheckDestroy: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if !strings.HasPrefix(rs.Type, "unifi_dns_") {
			continue
		}

		siteID := rs.Primary.Attributes["site_id"]
		if siteID == "" {
			siteID = os.Getenv("UNIFI_SITE_ID")
		}

		apiPath := fmt.Sprintf("/v1/sites/%s/dns/policies/%s", siteID, rs.Primary.ID)
		if err := client.Get(context.Background(), apiPath, nil); err == nil {
			return fmt.Errorf("%s %s still exists", rs.Type, rs.Primary.ID)
		} else if !errors.IsNotFoundError(err) {
			return fmt.Errorf("unexpected error checking %s %s destroyed: %w", rs.Type, rs.Primary.ID, err)
		}
	}

	return nil
}
