package acceptance

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// The three models below must match test-infra/unifi/docker-compose.yml's
// unifi-emu service (SIM_MODELS=UXG,USM8P,U7PRO) — see its README.md's
// "Adopted device fleet" section. Each is reported by the real API under
// its display name, not the SIM_MODELS short code.
const (
	adoptedGatewayModel = "Gateway Lite"
	adoptedSwitchModel  = "USW Ultra"
	adoptedAPModel      = "U7 Pro"
)

// TestAccDeviceDataSource_found reads the real adopted gateway from
// test-infra/unifi's unifi-emu fleet (see its README) — one of three device
// types adopted there; see _foundSwitch/_foundAP below for the other two.
func TestAccDeviceDataSource_found(t *testing.T) {
	testAccPreCheck(t) // must run before findAdoptedDeviceIDByModel's own API call
	dataSourceName := "data.unifi_device.test"
	deviceID := findAdoptedDeviceIDByModel(t, adoptedGatewayModel)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
data "unifi_device" "test" {
  device_id = %q
}
`, deviceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "model", adoptedGatewayModel),
					resource.TestCheckResourceAttrSet(dataSourceName, "mac_address"),
					resource.TestCheckResourceAttrSet(dataSourceName, "state"),
				),
			},
		},
	})
}

// TestAccDeviceDataSource_foundSwitch reads the real adopted switch — a
// second, differently-shaped adopted device (features: ["switching"])
// alongside the gateway, unlike the single-USG3-only fixture this replaced.
func TestAccDeviceDataSource_foundSwitch(t *testing.T) {
	testAccPreCheck(t)
	dataSourceName := "data.unifi_device.test"
	deviceID := findAdoptedDeviceIDByModel(t, adoptedSwitchModel)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
data "unifi_device" "test" {
  device_id = %q
}
`, deviceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "model", adoptedSwitchModel),
					resource.TestCheckResourceAttrSet(dataSourceName, "mac_address"),
					resource.TestCheckResourceAttrSet(dataSourceName, "state"),
				),
			},
		},
	})
}

// TestAccDeviceDataSource_foundAP reads the real adopted access point.
func TestAccDeviceDataSource_foundAP(t *testing.T) {
	testAccPreCheck(t)
	dataSourceName := "data.unifi_device.test"
	deviceID := findAdoptedDeviceIDByModel(t, adoptedAPModel)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
data "unifi_device" "test" {
  device_id = %q
}
`, deviceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "model", adoptedAPModel),
					resource.TestCheckResourceAttrSet(dataSourceName, "mac_address"),
					resource.TestCheckResourceAttrSet(dataSourceName, "state"),
				),
			},
		},
	})
}

// findAdoptedDeviceIDByModel looks up an adopted device's ID by its exact
// model display name directly via the API, since unifi_device requires an
// exact ID and there's no lookup-by-name — skips the test if no device with
// that model is adopted (a bare console, or the unifi-emu fleet not
// started; see test-infra/unifi/README.md).
func findAdoptedDeviceIDByModel(t *testing.T, model string) string {
	t.Helper()
	client, err := testAccNewClient()
	if err != nil {
		t.Fatalf("build client: %v", err)
	}

	var page struct {
		Data []struct {
			Id    string `json:"id"`
			Model string `json:"model"`
		} `json:"data"`
	}
	siteID := os.Getenv("UNIFI_SITE_ID")
	if err := client.Get(context.Background(), fmt.Sprintf("/v1/sites/%s/devices", siteID), &page); err != nil {
		t.Fatalf("list devices: %v", err)
	}
	for _, d := range page.Data {
		if d.Model == model {
			return d.Id
		}
	}
	t.Skipf("no adopted %q device found — see test-infra/unifi/README.md's \"Adopted device fleet\" section", model)
	return ""
}

// unifi_device needs a real adopted device — TestAccDeviceDataSource_found
// and friends cover that now that the fleet provides three; this covers the
// clean not-found path for a device ID that doesn't exist.
func TestAccDeviceDataSource_notFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "unifi_device" "test" {
  device_id = "00000000-0000-0000-0000-000000000001"
}
`,
				ExpectError: regexp.MustCompile(`(?i)device\s+not\s+found`),
			},
		},
	})
}
