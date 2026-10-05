package acceptance

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// unifi_vpn_server and unifi_vpn_site_to_site_tunnel have no create/update
// path anywhere in the API (GET-only, hence data sources rather than
// resources — see internal/sdkcompat/vpn), and this console has none
// configured, so only the not-found path is testable here. A console with
// VPN configured through the UI would exercise the happy path the same way
// TestAccWanDataSource_found does once a gateway is adopted.

func TestAccVPNServerDataSource_notFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "unifi_vpn_server" "test" {
  vpn_server_id = "00000000-0000-0000-0000-000000000001"
}
`,
				ExpectError: regexp.MustCompile(`(?i)vpn\s+server\s+not\s+found`),
			},
		},
	})
}

func TestAccVPNSiteToSiteTunnelDataSource_notFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "unifi_vpn_site_to_site_tunnel" "test" {
  tunnel_id = "00000000-0000-0000-0000-000000000001"
}
`,
				ExpectError: regexp.MustCompile(`(?i)vpn\s+site-to-site\s+tunnel\s+not\s+found`),
			},
		},
	})
}
