package acceptance

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccWanDataSource_found reads the real WAN interface the adopted
// gateway reports (test-infra/unifi's unifi-emu fleet — see its README's
// "Adopted device fleet" section) — this used to be impossible on a bare
// console with no adopted gateway, same situation unifi_device was in
// before it had one.
func TestAccWanDataSource_found(t *testing.T) {
	testAccPreCheck(t) // must run before findWANID's own API call
	dataSourceName := "data.unifi_wan.test"
	wanID := findWANID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
data "unifi_wan" "test" {
  wan_id = %q
}
`, wanID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "name"),
					resource.TestCheckResourceAttr(dataSourceName, "wan_id", wanID),
				),
			},
		},
	})
}

// findWANID looks up a real WAN interface ID directly via the API, since
// unifi_wan requires an exact ID and there's no lookup-by-name — skips if
// none exists (a bare console with no adopted gateway; see
// test-infra/unifi/README.md's "Adopted device fleet" section).
func findWANID(t *testing.T) string {
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
	if err := client.Get(context.Background(), fmt.Sprintf("/v1/sites/%s/wans", siteID), &page); err != nil {
		t.Fatalf("list wans: %v", err)
	}
	if len(page.Data) == 0 {
		t.Skip("no WAN interface found — see test-infra/unifi/README.md's \"Adopted device fleet\" section")
	}
	return page.Data[0].Id
}

// unifi_wan needs a real configured WAN interface — TestAccWanDataSource_found
// covers that now that the fleet provides one; this covers the clean
// not-found path for a WAN ID that doesn't exist.
func TestAccWanDataSource_notFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "unifi_wan" "test" {
  wan_id = "00000000-0000-0000-0000-000000000001"
}
`,
				ExpectError: regexp.MustCompile(`(?i)wan\s+interface\s+not\s+found`),
			},
		},
	})
}
