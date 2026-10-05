package acceptance

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccCountriesDataSource_basic reads the static, always-present country
// reference list — no resource needs to pre-populate anything.
func TestAccCountriesDataSource_basic(t *testing.T) {
	dataSourceName := "data.unifi_countries.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "unifi_countries" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceName, "countries.United States.code", "US"),
					resource.TestCheckResourceAttr(dataSourceName, "countries.Germany.code", "DE"),
				),
			},
		},
	})
}
