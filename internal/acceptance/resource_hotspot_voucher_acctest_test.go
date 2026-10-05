package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccHotspotVoucher_basic covers create, import, and destroy — there's
// no update test because the API has no update endpoint for vouchers
// (every attribute is RequiresReplace by design; see resource_hotspot_voucher.go).
func TestAccHotspotVoucher_basic(t *testing.T) {
	resourceName := "unifi_hotspot_voucher.test"
	name := "tf-acc-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: testAccCheckResourceDestroyed("unifi_hotspot_voucher", func(siteID, id string) string {
			return fmt.Sprintf("/v1/sites/%s/hotspot/vouchers/%s", siteID, id)
		}),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_hotspot_voucher" "test" {
  name                    = %q
  time_limit_minutes      = 60
  authorized_guest_limit  = 1
  data_usage_limit_mbytes = 1024
  rx_rate_limit_kbps      = 5000
  tx_rate_limit_kbps      = 5000
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "time_limit_minutes", "60"),
					resource.TestCheckResourceAttr(resourceName, "authorized_guest_limit", "1"),
					resource.TestCheckResourceAttr(resourceName, "authorized_guest_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "expired", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "site_id"),
					resource.TestCheckResourceAttrSet(resourceName, "code"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckNoResourceAttr(resourceName, "activated_at"),
					resource.TestCheckNoResourceAttr(resourceName, "expires_at"),
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

// TestAccHotspotVoucher_minimal exercises the resource with only the two
// required attributes set, confirming the optional rate/usage/guest-limit
// fields correctly stay null rather than resolving to a bogus zero value.
func TestAccHotspotVoucher_minimal(t *testing.T) {
	resourceName := "unifi_hotspot_voucher.test"
	name := "tf-acc-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: testAccCheckResourceDestroyed("unifi_hotspot_voucher", func(siteID, id string) string {
			return fmt.Sprintf("/v1/sites/%s/hotspot/vouchers/%s", siteID, id)
		}),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_hotspot_voucher" "test" {
  name               = %q
  time_limit_minutes = 30
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "time_limit_minutes", "30"),
					resource.TestCheckNoResourceAttr(resourceName, "authorized_guest_limit"),
					resource.TestCheckNoResourceAttr(resourceName, "data_usage_limit_mbytes"),
					resource.TestCheckNoResourceAttr(resourceName, "rx_rate_limit_kbps"),
					resource.TestCheckNoResourceAttr(resourceName, "tx_rate_limit_kbps"),
				),
			},
		},
	})
}
