package acceptance

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// unifi_device_tag has no create/update path anywhere in the API (GET-only,
// hence a data source rather than a resource — see
// internal/sdkcompat/devicetags), and this console has none configured, so
// only the not-found path is testable here.
func TestAccDeviceTagDataSource_notFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "unifi_device_tag" "test" {
  device_tag_id = "00000000-0000-0000-0000-000000000001"
}
`,
				ExpectError: regexp.MustCompile(`(?i)device\s+tag\s+not\s+found`),
			},
		},
	})
}
