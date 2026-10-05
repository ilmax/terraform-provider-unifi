package acceptance

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// unifi_switch_lag, unifi_switch_mc_lag_domain, and unifi_switch_stack all
// need real switch hardware (MC-LAG/stacking-capable, for the latter two)
// adopted into the site to return anything — this console has none, and
// unlike the gateway, this repo has no switch emulator to adopt one with.
// Only the clean not-found path is testable here; see
// internal/sdkcompat/switching for the full response shapes these data
// sources expose once real hardware is available.

func TestAccSwitchLagDataSource_notFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "unifi_switch_lag" "test" {
  lag_id = "00000000-0000-0000-0000-000000000001"
}
`,
				ExpectError: regexp.MustCompile(`(?i)switch\s+LAG\s+not\s+found`),
			},
		},
	})
}

func TestAccSwitchMcLagDomainDataSource_notFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "unifi_switch_mc_lag_domain" "test" {
  mc_lag_domain_id = "00000000-0000-0000-0000-000000000001"
}
`,
				ExpectError: regexp.MustCompile(`(?i)mc-lag\s+domain\s+not\s+found`),
			},
		},
	})
}

func TestAccSwitchStackDataSource_notFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "unifi_switch_stack" "test" {
  switch_stack_id = "00000000-0000-0000-0000-000000000001"
}
`,
				ExpectError: regexp.MustCompile(`(?i)switch\s+stack\s+not\s+found`),
			},
		},
	})
}
