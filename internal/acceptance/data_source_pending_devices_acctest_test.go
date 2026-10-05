package acceptance

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccPendingDevicesDataSource_empty confirms unifi_pending_devices reads
// cleanly when nothing is pending adoption (this data source has no
// not-found path — it always returns 200 with a, possibly empty, list).
// The unifi-emu fleet (test-infra/unifi/README.md's "Adopted device fleet"
// section) passes through this exact state on its way to being adopted, but
// by the time TestAccNetwork_gatewayBasic and friends can run, it's already
// adopted — so this test's "0" assertion is expected in this environment's
// steady state, not a placeholder.
func TestAccPendingDevicesDataSource_empty(t *testing.T) {
	dataSourceName := "data.unifi_pending_devices.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "unifi_pending_devices" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceName, "devices.#", "0"),
				),
			},
		},
	})
}
