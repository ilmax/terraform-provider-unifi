package acceptance

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func wifiTestName(prefix string) string {
	return prefix + "-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
}

// TestAccWifi_standardOpen exercises a STANDARD/OPEN broadcast — the
// richest schema path — including advertise_device_name, then recreates it
// via a RequiresReplace-triggering change (every settable attribute on this
// resource forces replacement; Update() is intentionally unreachable).
func TestAccWifi_standardOpen(t *testing.T) {
	resourceName := "unifi_wifi.test"
	name := wifiTestName("tf-wifi-std")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: testAccCheckResourceDestroyed("unifi_wifi", func(siteID, id string) string {
			return fmt.Sprintf("/v1/sites/%s/wifi/broadcasts/%s", siteID, id)
		}),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_wifi" "test" {
  type          = "STANDARD"
  name          = %q
  enabled       = true
  security_type = "OPEN"
  network_type  = "NATIVE"

  broadcasting_frequencies_ghz = ["2.4"]
  advertise_device_name        = true
  hide_name                    = false
  client_isolation_enabled     = false
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "type", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, "security_type", "OPEN"),
					resource.TestCheckResourceAttr(resourceName, "advertise_device_name", "true"),
					resource.TestCheckResourceAttr(resourceName, "broadcasting_frequencies_ghz.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "site_id"),
				),
			},
			{
				// enabled has RequiresReplace: this must recreate cleanly,
				// not fail through the "WiFi updates are not supported"
				// Update() error path.
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_wifi" "test" {
  type          = "STANDARD"
  name          = %q
  enabled       = false
  security_type = "OPEN"
  network_type  = "NATIVE"

  broadcasting_frequencies_ghz = ["2.4"]
}
`, name),
				Check: resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
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

// TestAccWifi_iotOptimizedOpen exercises the IOT_OPTIMIZED type, whose
// required-boolean set is smaller than STANDARD's — this is what regression
// -locks the requiredBoolDefaults fix for the 4 booleans required on both
// broadcast types.
func TestAccWifi_iotOptimizedOpen(t *testing.T) {
	resourceName := "unifi_wifi.test"
	name := wifiTestName("tf-wifi-iot")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: testAccCheckResourceDestroyed("unifi_wifi", func(siteID, id string) string {
			return fmt.Sprintf("/v1/sites/%s/wifi/broadcasts/%s", siteID, id)
		}),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_wifi" "test" {
  type          = "IOT_OPTIMIZED"
  name          = %q
  enabled       = true
  security_type = "OPEN"
  network_type  = "NATIVE"
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "type", "IOT_OPTIMIZED"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}

// TestAccWifi_standardRequiresFrequencies asserts the real API error for a
// STANDARD broadcast missing broadcasting_frequencies_ghz surfaces cleanly
// — this field has no sensible universal default (unlike the booleans
// above), so it's a genuine required-input case, not a provider bug.
func TestAccWifi_standardRequiresFrequencies(t *testing.T) {
	name := wifiTestName("tf-wifi-nofreq")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_wifi" "test" {
  type          = "STANDARD"
  name          = %q
  enabled       = true
  security_type = "OPEN"
}
`, name),
				ExpectError: regexp.MustCompile(`(?i)broadcastingFrequenciesGHz\s+must\s+not\s+be\s+empty`),
			},
		},
	})
}

// TestAccWifi_advertiseDeviceNameOnlyForStandard asserts the plan-time
// guard rejects advertise_device_name on a non-STANDARD broadcast instead
// of silently ignoring it or sending an inapplicable field to the API.
func TestAccWifi_advertiseDeviceNameOnlyForStandard(t *testing.T) {
	name := wifiTestName("tf-wifi-badadn")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_wifi" "test" {
  type                   = "IOT_OPTIMIZED"
  name                   = %q
  enabled                = true
  security_type          = "OPEN"
  advertise_device_name  = true
}
`, name),
				ExpectError: regexp.MustCompile(`(?i)only\s+supported\s+for\s+STANDARD`),
			},
		},
	})
}
