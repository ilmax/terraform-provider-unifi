package acceptance

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccDPIApplicationDataSource_found reads a real DPI application. This
// reference data is global (not scoped to a site or gated on any adopted
// hardware) and always populated on any console.
func TestAccDPIApplicationDataSource_found(t *testing.T) {
	testAccPreCheck(t) // must run before findDPIApplicationID's own API call
	dataSourceName := "data.unifi_dpi_application.test"
	appID := findDPIApplicationID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
data "unifi_dpi_application" "test" {
  application_id = %d
}
`, appID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "name"),
					resource.TestCheckResourceAttr(dataSourceName, "application_id", fmt.Sprintf("%d", appID)),
				),
			},
		},
	})
}

func findDPIApplicationID(t *testing.T) int64 {
	t.Helper()
	client, err := testAccNewClient()
	if err != nil {
		t.Fatalf("build client: %v", err)
	}

	var page struct {
		Data []struct {
			Id int64 `json:"id"`
		} `json:"data"`
	}
	if err := client.Get(context.Background(), "/v1/dpi/applications", &page); err != nil {
		t.Fatalf("list dpi applications: %v", err)
	}
	if len(page.Data) == 0 {
		t.Skip("no DPI application found")
	}
	return page.Data[0].Id
}

func TestAccDPIApplicationDataSource_notFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "unifi_dpi_application" "test" {
  application_id = -1
}
`,
				ExpectError: regexp.MustCompile(`(?i)dpi\s+application\s+not\s+found`),
			},
		},
	})
}

// TestAccDPICategoryDataSource_found reads a real DPI category, same
// rationale as TestAccDPIApplicationDataSource_found above.
func TestAccDPICategoryDataSource_found(t *testing.T) {
	testAccPreCheck(t) // must run before findDPICategoryID's own API call
	dataSourceName := "data.unifi_dpi_category.test"
	categoryID := findDPICategoryID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
data "unifi_dpi_category" "test" {
  category_id = %d
}
`, categoryID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "name"),
					resource.TestCheckResourceAttr(dataSourceName, "category_id", fmt.Sprintf("%d", categoryID)),
				),
			},
		},
	})
}

func findDPICategoryID(t *testing.T) int64 {
	t.Helper()
	client, err := testAccNewClient()
	if err != nil {
		t.Fatalf("build client: %v", err)
	}

	var page struct {
		Data []struct {
			Id int64 `json:"id"`
		} `json:"data"`
	}
	if err := client.Get(context.Background(), "/v1/dpi/categories", &page); err != nil {
		t.Fatalf("list dpi categories: %v", err)
	}
	if len(page.Data) == 0 {
		t.Skip("no DPI category found")
	}
	return page.Data[0].Id
}

func TestAccDPICategoryDataSource_notFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "unifi_dpi_category" "test" {
  category_id = -1
}
`,
				ExpectError: regexp.MustCompile(`(?i)dpi\s+category\s+not\s+found`),
			},
		},
	})
}
