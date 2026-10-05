package acceptance

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// unifi_client needs a real connected client to return data for, which
// this container-based console has none of — see test-infra/unifi/README.md.
// These tests instead cover the realistic failure modes: no match, a
// malformed identifier, and the two "you gave me the wrong number of
// identifiers" validation paths, all of which are genuine user mistakes
// worth a clean error for.

func TestAccClientDataSource_notFoundByID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "unifi_client" "test" {
  client_id = "00000000-0000-0000-0000-000000000001"
}
`,
				ExpectError: regexp.MustCompile(`(?i)client\s+not\s+found`),
			},
		},
	})
}

func TestAccClientDataSource_notFoundByMAC(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "unifi_client" "test" {
  mac_address = "aa:bb:cc:dd:ee:ff"
}
`,
				ExpectError: regexp.MustCompile(`(?i)client\s+not\s+found`),
			},
		},
	})
}

func TestAccClientDataSource_ambiguousIdentifier(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "unifi_client" "test" {
  client_id   = "00000000-0000-0000-0000-000000000001"
  mac_address = "aa:bb:cc:dd:ee:ff"
}
`,
				ExpectError: regexp.MustCompile(`(?i)ambiguous\s+client\s+identifier`),
			},
		},
	})
}

func TestAccClientDataSource_missingIdentifier(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "unifi_client" "test" {}
`,
				ExpectError: regexp.MustCompile(`(?i)missing\s+client\s+identifier`),
			},
		},
	})
}

func TestAccClientDataSource_invalidMAC(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "unifi_client" "test" {
  mac_address = "not-a-mac-address"
}
`,
				ExpectError: regexp.MustCompile(`(?i)invalid\s+mac_address`),
			},
		},
	})
}
