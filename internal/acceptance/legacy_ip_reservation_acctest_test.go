package acceptance

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/ilmax/terraform-provider-unifi/internal/legacy"
	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
)

func testAccLegacyClient(t *testing.T) *legacy.Client {
	t.Helper()
	c, err := legacy.NewClient(unifi.Config{
		APIKey:        os.Getenv("UNIFI_API_KEY"),
		BaseURL:       os.Getenv("UNIFI_API_URL"),
		AllowInsecure: os.Getenv("UNIFI_ALLOW_INSECURE") == "true",
	})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// testAccLegacyLAN returns the id and the first three octets of the first
// corporate network, so reservations land inside a real subnet.
func testAccLegacyLAN(t *testing.T, c *legacy.Client) (id, prefix string) {
	t.Helper()
	var nets []struct {
		ID       string `json:"_id"`
		Purpose  string `json:"purpose"`
		IPSubnet string `json:"ip_subnet"`
	}
	if err := c.List(context.Background(), os.Getenv("UNIFI_SITE_ID"), "networkconf", &nets); err != nil {
		t.Fatal(err)
	}
	for _, n := range nets {
		if n.Purpose == "corporate" && n.IPSubnet != "" {
			octets := strings.Split(strings.Split(n.IPSubnet, "/")[0], ".")
			return n.ID, strings.Join(octets[:3], ".")
		}
	}
	t.Skip("no corporate network with an ip_subnet on this console")
	return "", ""
}

// TestAccLegacyIPReservation_basic reserves an IP for a MAC the console has
// never seen, updates it, reads it back, imports it, and checks destroy only
// releases the fixed IP.
func TestAccLegacyIPReservation_basic(t *testing.T) {
	const (
		res = "unifilegacy_ip_reservation.test"
		mac = "02:00:00:aa:bb:01"
	)
	c := testAccLegacyClient(t)
	networkID, prefix := testAccLegacyLAN(t, c)
	site := os.Getenv("UNIFI_SITE_ID")

	config := func(ip, extra string) string {
		return testAccLegacyProviderConfig() + fmt.Sprintf(`
resource "unifilegacy_ip_reservation" "test" {
  mac        = %q
  ip_address = %q
  network_id = %q
  %s
}

data "unifilegacy_ip_reservation" "test" {
  mac        = unifilegacy_ip_reservation.test.mac
  depends_on = [unifilegacy_ip_reservation.test]
}
`, mac, ip, networkID, extra)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccLegacyProviderFactories,
		CheckDestroy: func(*terraform.State) error {
			r, err := c.GetReservation(context.Background(), site, mac)
			if err != nil {
				return err
			}
			if r != nil {
				return fmt.Errorf("reservation for %s still active after destroy", mac)
			}
			// Leave the console as found: the destroy keeps the client record.
			return c.Command(context.Background(), site, "stamgr", map[string]any{"cmd": "forget-sta", "macs": []string{mac}})
		},
		Steps: []resource.TestStep{
			{
				Config: config(prefix+".201", `name = "acctest-reservation"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res, "ip_address", prefix+".201"),
					resource.TestCheckResourceAttr(res, "name", "acctest-reservation"),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttr("data.unifilegacy_ip_reservation.test", "ip_address", prefix+".201"),
					resource.TestCheckResourceAttrPair(res, "id", "data.unifilegacy_ip_reservation.test", "id"),
				),
			},
			{
				Config: config(prefix+".202", `name = "acctest-reservation"
  note = "updated"
  local_dns_record = "acctest-res.lan"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res, "ip_address", prefix+".202"),
					resource.TestCheckResourceAttr(res, "note", "updated"),
					resource.TestCheckResourceAttr(res, "local_dns_record", "acctest-res.lan"),
				),
			},
			{
				ResourceName:      res,
				ImportState:       true,
				ImportStateId:     site + "/" + mac,
				ImportStateVerify: true,
			},
		},
	})
}
