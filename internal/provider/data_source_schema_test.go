package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

type dataSourceSchemaCheck struct {
	name            string
	dataSource      datasource.DataSource
	requiredStrings []string
}

func TestDataSourceSchemas(t *testing.T) {
	checks := []dataSourceSchemaCheck{
		{
			name:            "device",
			dataSource:      NewDeviceDataSource(),
			requiredStrings: []string{"device_id"},
		},
		{
			name:            "wan",
			dataSource:      NewWanDataSource(),
			requiredStrings: []string{"wan_id"},
		},
	}

	for _, check := range checks {
		resp := &datasource.SchemaResponse{}
		check.dataSource.Schema(context.Background(), datasource.SchemaRequest{}, resp)
		if resp.Diagnostics.HasError() {
			t.Fatalf("%s schema diagnostics: %v", check.name, resp.Diagnostics)
		}

		for _, attrName := range check.requiredStrings {
			attr, ok := resp.Schema.Attributes[attrName].(schema.StringAttribute)
			if !ok {
				t.Fatalf("%s: %s attribute missing or not string", check.name, attrName)
			}
			if !attr.Required {
				t.Fatalf("%s: %s should be required", check.name, attrName)
			}
		}
	}
}
