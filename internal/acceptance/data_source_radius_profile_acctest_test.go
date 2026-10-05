package acceptance

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccRadiusProfileDataSource_found reads a real RADIUS profile. Unlike
// unifi_vpn_server/unifi_device_tag, this list is never empty: every site
// has at least a SYSTEM_DEFINED "Default" profile out of the box, with no
// gateway or manual configuration required.
func TestAccRadiusProfileDataSource_found(t *testing.T) {
	testAccPreCheck(t) // must run before findRadiusProfileID's own API call
	dataSourceName := "data.unifi_radius_profile.test"
	profileID := findRadiusProfileID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
data "unifi_radius_profile" "test" {
  radius_profile_id = %q
}
`, profileID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "name"),
					resource.TestCheckResourceAttr(dataSourceName, "radius_profile_id", profileID),
					resource.TestCheckResourceAttrSet(dataSourceName, "origin"),
				),
			},
		},
	})
}

// findRadiusProfileID looks up a real RADIUS profile ID directly via the
// API, since unifi_radius_profile requires an exact ID and there's no
// lookup-by-name.
func findRadiusProfileID(t *testing.T) string {
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
	if err := client.Get(context.Background(), fmt.Sprintf("/v1/sites/%s/radius/profiles", siteID), &page); err != nil {
		t.Fatalf("list radius profiles: %v", err)
	}
	if len(page.Data) == 0 {
		t.Skip("no RADIUS profile found")
	}
	return page.Data[0].Id
}

func TestAccRadiusProfileDataSource_notFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "unifi_radius_profile" "test" {
  radius_profile_id = "00000000-0000-0000-0000-000000000001"
}
`,
				ExpectError: regexp.MustCompile(`(?i)radius\s+profile\s+not\s+found`),
			},
		},
	})
}
