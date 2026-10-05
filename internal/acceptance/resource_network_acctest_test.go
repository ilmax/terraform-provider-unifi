package acceptance

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// networkCheckDestroy is the CheckDestroy shared by every unifi_network
// acceptance test in this file.
func networkCheckDestroy(s *terraform.State) error {
	return testAccCheckResourceDestroyed("unifi_network", func(siteID, id string) string {
		return fmt.Sprintf("/v1/sites/%s/networks/%s", siteID, id)
	})(s)
}

// networkTestName generates a short unique name, since unifi_network.name
// is capped at 32 characters by the API (RandomWithPrefix's numeric suffix
// alone can run past that once a "-renamed"-style suffix is appended).
func networkTestName(prefix string) string {
	return prefix + "-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
}

func TestAccNetwork_unmanagedBasic(t *testing.T) {
	resourceName := "unifi_network.test"
	name := networkTestName("tf-net")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             networkCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_network" "test" {
  name       = %q
  management = "UNMANAGED"
  enabled    = true
  vlan_id    = 100
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "management", "UNMANAGED"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "vlan_id", "100"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "site_id"),
				),
			},
			{
				// Update in place: name and enabled change, vlan_id stays the
				// same (changing it on an UNMANAGED network isn't guaranteed
				// idempotent against the real switch/port model).
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_network" "test" {
  name       = %q
  management = "UNMANAGED"
  enabled    = false
  vlan_id    = 100
}
`, name+"-renamed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name+"-renamed"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
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

// TestAccNetwork_gatewayBasic exercises the GATEWAY management type end to
// end, including the static IPv4 + DHCP server + ping-conflict-detection
// path that was completely untestable before test-infra/unifi's unifi-emu
// fleet (see the README's "Adopted device fleet" section) gave this console
// an adopted gateway — every one of these previously failed with "No
// gateway found on the site for gateway-managed network". zone_id is
// required here (not just accepted) once Zone Based Firewall is enabled
// (see README.md's "Zone Based Firewall" section) — the real API rejects a
// GATEWAY network with none.
func TestAccNetwork_gatewayBasic(t *testing.T) {
	testAccPreCheck(t) // must run before findFirewallZoneIDByName's own API call
	resourceName := "unifi_network.test"
	name := networkTestName("tf-netgw")
	zoneID := findFirewallZoneIDByName(t, "Internal")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             networkCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_network" "test" {
  name       = %q
  management = "GATEWAY"
  enabled    = true
  vlan_id    = 200
  zone_id    = %q

  isolation_enabled        = false
  cellular_backup_enabled  = false
  internet_access_enabled  = true

  ipv4_configuration = {
    host_ip_address = "10.200.0.1"
    prefix_length   = 24

    dhcp_configuration = {
      mode = "SERVER"
      ip_address_range = {
        start = "10.200.0.100"
        stop  = "10.200.0.200"
      }
      lease_time_seconds              = 86400
      ping_conflict_detection_enabled = true
    }
  }
}
`, name, zoneID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "management", "GATEWAY"),
					resource.TestCheckResourceAttr(resourceName, "internet_access_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "zone_id", zoneID),
					resource.TestCheckResourceAttr(resourceName, "ipv4_configuration.host_ip_address", "10.200.0.1"),
					resource.TestCheckResourceAttr(resourceName, "ipv4_configuration.dhcp_configuration.ping_conflict_detection_enabled", "true"),
				),
			},
			{
				// Update in place: widen the DHCP range, flip ping-conflict
				// detection off.
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_network" "test" {
  name       = %q
  management = "GATEWAY"
  enabled    = true
  vlan_id    = 200
  zone_id    = %q

  isolation_enabled        = false
  cellular_backup_enabled  = false
  internet_access_enabled  = true

  ipv4_configuration = {
    host_ip_address = "10.200.0.1"
    prefix_length   = 24

    dhcp_configuration = {
      mode = "SERVER"
      ip_address_range = {
        start = "10.200.0.50"
        stop  = "10.200.0.250"
      }
      lease_time_seconds              = 43200
      ping_conflict_detection_enabled = false
    }
  }
}
`, name, zoneID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "ipv4_configuration.dhcp_configuration.ip_address_range.start", "10.200.0.50"),
					resource.TestCheckResourceAttr(resourceName, "ipv4_configuration.dhcp_configuration.ping_conflict_detection_enabled", "false"),
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

// TestAccNetwork_gatewayRequiresAdoptedGateway used to be the only possible
// GATEWAY-type test, asserting the clean error this console gave before it
// had an adopted device. Kept as a regression check that removing the
// gateway emulator (or running against a fresh console that hasn't adopted
// one yet) still fails cleanly rather than with a confusing wrapped error —
// skips itself once a gateway is actually present, since then this
// specific error can no longer occur.
func TestAccNetwork_gatewayRequiresAdoptedGateway(t *testing.T) {
	testAccPreCheck(t) // must run before hasAdoptedGateway's own API call
	if hasAdoptedGateway(t) {
		t.Skip("a gateway is adopted on this console; see TestAccNetwork_gatewayBasic for the happy path this error path guards")
	}

	name := networkTestName("tf-netgw2")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_network" "test" {
  name       = %q
  management = "GATEWAY"
  enabled    = true
  vlan_id    = 201

  ipv4_configuration = {
    host_ip_address = "10.201.0.1"
    prefix_length   = 24
  }
}
`, name),
				ExpectError: regexp.MustCompile(`(?i)no\s+gateway\s+found`),
			},
		},
	})
}

// TestAccNetwork_invalidManagement asserts an unsupported management value
// is rejected with a clean plan-time error instead of reaching the API (or
// panicking) — buildCreateNetworkRequest's default case in
// resource_network.go.
func TestAccNetwork_invalidManagement(t *testing.T) {
	name := networkTestName("tf-netbad")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_network" "test" {
  name       = %q
  management = "BOGUS"
  enabled    = true
}
`, name),
				ExpectError: regexp.MustCompile(`(?i)invalid management type`),
			},
		},
	})
}

// TestAccNetwork_dhcpGuarding exercises dhcp_guarding.trusted_dhcp_server_ip_addresses,
// including updating the list in place (add one entry, drop another).
func TestAccNetwork_dhcpGuarding(t *testing.T) {
	resourceName := "unifi_network.test"
	name := networkTestName("tf-netdg")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             networkCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_network" "test" {
  name       = %q
  management = "UNMANAGED"
  enabled    = true
  vlan_id    = 401

  dhcp_guarding = {
    trusted_dhcp_server_ip_addresses = ["10.0.0.53", "10.0.0.54"]
  }
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "dhcp_guarding.trusted_dhcp_server_ip_addresses.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "dhcp_guarding.trusted_dhcp_server_ip_addresses.*", "10.0.0.53"),
					resource.TestCheckTypeSetElemAttr(resourceName, "dhcp_guarding.trusted_dhcp_server_ip_addresses.*", "10.0.0.54"),
				),
			},
			{
				// Update in place: drop .53, keep .54, add .55.
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_network" "test" {
  name       = %q
  management = "UNMANAGED"
  enabled    = true
  vlan_id    = 401

  dhcp_guarding = {
    trusted_dhcp_server_ip_addresses = ["10.0.0.54", "10.0.0.55"]
  }
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "dhcp_guarding.trusted_dhcp_server_ip_addresses.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "dhcp_guarding.trusted_dhcp_server_ip_addresses.*", "10.0.0.54"),
					resource.TestCheckTypeSetElemAttr(resourceName, "dhcp_guarding.trusted_dhcp_server_ip_addresses.*", "10.0.0.55"),
				),
			},
		},
	})
}

// TestAccNetwork_vlanBoundaries pins down the real, undocumented VLAN ID
// range for non-default networks discovered empirically against the live
// console: [2, 4009] (not the 802.1Q [1, 4094] range one might assume).
func TestAccNetwork_vlanBoundaries(t *testing.T) {
	for _, vlan := range []int{2, 4009} {
		vlan := vlan
		t.Run(fmt.Sprintf("vlan_%d", vlan), func(t *testing.T) {
			resourceName := "unifi_network.test"
			name := networkTestName("tf-netvb")

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				CheckDestroy:             networkCheckDestroy,
				Steps: []resource.TestStep{
					{
						Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_network" "test" {
  name       = %q
  management = "UNMANAGED"
  enabled    = true
  vlan_id    = %d
}
`, name, vlan),
						Check: resource.TestCheckResourceAttr(resourceName, "vlan_id", fmt.Sprintf("%d", vlan)),
					},
				},
			})
		})
	}
}

// TestAccNetwork_vlanOutOfRange asserts values outside [2, 4009] are
// rejected with the API's real error message surfacing cleanly, for both
// the "too low" (1 is reserved for the default network) and "too high" ends.
func TestAccNetwork_vlanOutOfRange(t *testing.T) {
	cases := []struct {
		name    string
		vlan    int
		wantErr string
	}{
		{"tooLow", 1, `(?i)must\s+be\s+greater\s+than\s+1|vlanId\s+must\s+be\s+greater\s+than\s+or\s+equal\s+to`},
		{"tooHigh", 4010, `(?i)vlanId\s+must\s+be\s+less\s+than\s+or\s+equal\s+to`},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			name := networkTestName("tf-netvr")

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_network" "test" {
  name       = %q
  management = "UNMANAGED"
  enabled    = true
  vlan_id    = %d
}
`, name, tc.vlan),
						ExpectError: regexp.MustCompile(tc.wantErr),
					},
				},
			})
		})
	}
}

// TestAccNetwork_duplicateVlan asserts that reusing a VLAN ID already
// assigned to another network on the same site fails cleanly instead of
// silently succeeding or corrupting either network — a realistic user
// mistake (typo'd or copy-pasted VLAN ID).
func TestAccNetwork_duplicateVlan(t *testing.T) {
	firstName := networkTestName("tf-netdv1")
	secondName := networkTestName("tf-netdv2")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             networkCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_network" "first" {
  name       = %q
  management = "UNMANAGED"
  enabled    = true
  vlan_id    = 601
}

resource "unifi_network" "second" {
  name       = %q
  management = "UNMANAGED"
  enabled    = true
  vlan_id    = 601

  depends_on = [unifi_network.first]
}
`, firstName, secondName),
				ExpectError: regexp.MustCompile(`(?i)vlan\s+id\s+601\s+is\s+already\s+in\s+use`),
			},
		},
	})
}

// TestAccNetwork_nameLength pins down the real, undocumented name length
// range discovered empirically against the live console: [2, 32].
func TestAccNetwork_nameLength(t *testing.T) {
	t.Run("maxValid", func(t *testing.T) {
		resourceName := "unifi_network.test"
		// Exactly 32 characters, the documented (empirically) maximum.
		name := "n" + acctest.RandStringFromCharSet(31, acctest.CharSetAlphaNum)

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			CheckDestroy:             networkCheckDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_network" "test" {
  name       = %q
  management = "UNMANAGED"
  enabled    = true
  vlan_id    = 701
}
`, name),
					Check: resource.TestCheckResourceAttr(resourceName, "name", name),
				},
			},
		})
	})

	t.Run("tooLong", func(t *testing.T) {
		name := "n" + acctest.RandStringFromCharSet(32, acctest.CharSetAlphaNum) // 33 chars

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_network" "test" {
  name       = %q
  management = "UNMANAGED"
  enabled    = true
  vlan_id    = 702
}
`, name),
					ExpectError: regexp.MustCompile(`(?i)name\s+length\s+must\s+be\s+between\s+2\s+and\s+32`),
				},
			},
		})
	})

	t.Run("tooShort", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccProviderConfig() + `
resource "unifi_network" "test" {
  name       = "n"
  management = "UNMANAGED"
  enabled    = true
  vlan_id    = 703
}
`,
					ExpectError: regexp.MustCompile(`(?i)name\s+length\s+must\s+be\s+between\s+2\s+and\s+32`),
				},
			},
		})
	})
}

// TestAccNetwork_driftDetection deletes the network out-of-band (directly
// via the API, bypassing Terraform) between two plans and asserts Terraform
// notices — i.e. Read correctly removes it from state instead of erroring,
// and the next plan is non-empty (proposing to recreate it) rather than
// reporting "no changes" against a resource that no longer exists.
func TestAccNetwork_driftDetection(t *testing.T) {
	name := networkTestName("tf-netdr")
	config := testAccProviderConfig() + fmt.Sprintf(`
resource "unifi_network" "test" {
  name       = %q
  management = "UNMANAGED"
  enabled    = true
  vlan_id    = 801
}
`, name)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             networkCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				PreConfig: func() { deleteNetworkOutOfBand(t, 801) },
				Config:    config,
				// The prior apply's state still names an ID that no longer
				// exists; Read must drop it and the plan must show a create
				// (not "no changes", and not an error).
				ExpectNonEmptyPlan: true,
				PlanOnly:           true,
			},
		},
	})
}

// deleteNetworkOutOfBand deletes the network with the given VLAN ID
// directly via the API, bypassing Terraform, to simulate drift. It looks
// the network up by VLAN ID rather than taking an ID directly because
// PreConfig runs before the test framework re-exposes prior-step state.
func deleteNetworkOutOfBand(t *testing.T, vlanID int64) {
	t.Helper()
	client, err := testAccNewClient()
	if err != nil {
		t.Fatalf("build client: %v", err)
	}

	siteID := os.Getenv("UNIFI_SITE_ID")
	var page struct {
		Data []struct {
			Id     string `json:"id"`
			VlanId int64  `json:"vlanId"`
		} `json:"data"`
	}
	if err := client.Get(context.Background(), fmt.Sprintf("/v1/sites/%s/networks", siteID), &page); err != nil {
		t.Fatalf("list networks: %v", err)
	}
	for _, n := range page.Data {
		if n.VlanId == vlanID {
			if err := client.Delete(context.Background(), fmt.Sprintf("/v1/sites/%s/networks/%s", siteID, n.Id), nil); err != nil {
				t.Fatalf("delete network out of band: %v", err)
			}
			return
		}
	}
	t.Fatalf("no network with vlan_id %d found to delete out of band", vlanID)
}
