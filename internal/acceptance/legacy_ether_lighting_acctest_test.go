package acceptance

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccLegacyEtherLightingSettings_basic sets site-wide color overrides,
// updates them, imports the setting and checks destroy clears the overrides.
func TestAccLegacyEtherLightingSettings_basic(t *testing.T) {
	const res = "unifilegacy_ether_lighting_settings.test"
	c := testAccLegacyClient(t)
	networkID, _ := testAccLegacyLAN(t, c)
	site := os.Getenv("UNIFI_SITE_ID")

	config := func(speed string) string {
		return testAccLegacyProviderConfig() + fmt.Sprintf(`
resource "unifilegacy_ether_lighting_settings" "test" {
  speed_overrides   = { %s }
  network_overrides = { %q = "00FF00" }
}
`, speed, networkID)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccLegacyProviderFactories,
		CheckDestroy: func(*terraform.State) error {
			s, err := c.GetEtherLightingSettings(context.Background(), site)
			if err != nil {
				return err
			}
			if len(s.SpeedOverrides)+len(s.NetworkOverrides) != 0 {
				return fmt.Errorf("overrides still set after destroy: %+v", s)
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: config(`GbE = "FF0000"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res, "speed_overrides.GbE", "FF0000"),
					resource.TestCheckResourceAttr(res, "network_overrides."+networkID, "00FF00"),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttrSet(res, "speed_defaults.GbE"),
				),
			},
			{
				Config: config(`GbE = "FF0000", "2.5GbE" = "0000FF"`),
				Check:  resource.TestCheckResourceAttr(res, "speed_overrides.2.5GbE", "0000FF"),
			},
			{
				ResourceName:      res,
				ImportState:       true,
				ImportStateId:     site,
				ImportStateVerify: true,
			},
			{
				Config: testAccLegacyProviderConfig() + `
resource "unifilegacy_ether_lighting_settings" "test" {
  speed_overrides = { GbE = "red" }
}
`,
				ExpectError: regexp.MustCompile(`Invalid color`),
			},
			{
				Config: config(`GbE = "FF0000"`),
			},
		},
	})
}

// TestAccLegacyDeviceEtherLighting_basic configures the emulated switch. The
// controller stores the block even though the emulator does not advertise the
// capability, so this runs on the shared fleet.
func TestAccLegacyDeviceEtherLighting_basic(t *testing.T) {
	const res = "unifilegacy_device_ether_lighting.test"
	c := testAccLegacyClient(t)
	site := os.Getenv("UNIFI_SITE_ID")

	var devices []struct {
		MAC  string `json:"mac"`
		Type string `json:"type"`
	}
	if err := c.Stat(context.Background(), site, "device", &devices); err != nil {
		t.Fatal(err)
	}
	mac := ""
	for _, d := range devices {
		if d.Type == "usw" {
			mac = d.MAC
			break
		}
	}
	if mac == "" {
		t.Skip("no adopted switch on this console")
	}

	config := func(body string) string {
		return testAccLegacyProviderConfig() + fmt.Sprintf(`
resource "unifilegacy_device_ether_lighting" "test" {
  mac = %q
  %s
}
`, mac, body)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccLegacyProviderFactories,
		CheckDestroy: func(*terraform.State) error {
			d, err := c.FindDevice(context.Background(), site, mac)
			if err != nil {
				return err
			}
			if d == nil || d.EtherLighting == nil || d.EtherLighting.LEDMode != "standard" {
				return fmt.Errorf("device not back in standard LED mode after destroy: %+v", d)
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: config(`led_mode = "etherlighting"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res, "led_mode", "etherlighting"),
					resource.TestCheckResourceAttr(res, "mode", "speed"),
					resource.TestCheckResourceAttr(res, "brightness", "100"),
					resource.TestCheckResourceAttr(res, "behavior", "steady"),
					resource.TestCheckResourceAttrSet(res, "id"),
				),
			},
			{
				Config: config(`led_mode = "etherlighting"
  mode       = "network"
  brightness = 40
  behavior   = "breath"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res, "mode", "network"),
					resource.TestCheckResourceAttr(res, "brightness", "40"),
					resource.TestCheckResourceAttr(res, "behavior", "breath"),
				),
			},
			{
				ResourceName:      res,
				ImportState:       true,
				ImportStateId:     site + "/" + mac,
				ImportStateVerify: true,
			},
			{
				Config:      config(`led_mode = "rainbow"`),
				ExpectError: regexp.MustCompile(`Invalid led_mode`),
			},
			{
				Config: config(`led_mode = "etherlighting"`),
			},
		},
	})
}
