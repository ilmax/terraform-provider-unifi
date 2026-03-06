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
	optionalStrings []string
}

func TestDataSourceSchemas(t *testing.T) {
	checks := []dataSourceSchemaCheck{
		{
			name:            "client",
			dataSource:      NewClientDataSource(),
			optionalStrings: []string{"client_id", "mac_address"},
		},
		{
			name:            "device",
			dataSource:      NewDeviceDataSource(),
			requiredStrings: []string{"device_id"},
		},
		{
			name:            "firewall_zone",
			dataSource:      NewFirewallZoneDataSource(),
			requiredStrings: []string{"zone_id"},
		},
		{
			name:            "firewall_zones",
			dataSource:      NewFirewallZonesDataSource(),
			optionalStrings: []string{"name"},
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

		for _, attrName := range check.optionalStrings {
			attr, ok := resp.Schema.Attributes[attrName].(schema.StringAttribute)
			if !ok {
				t.Fatalf("%s: %s attribute missing or not string", check.name, attrName)
			}
			if !attr.Optional {
				t.Fatalf("%s: %s should be optional", check.name, attrName)
			}
			if attr.Required {
				t.Fatalf("%s: %s should not be required", check.name, attrName)
			}
		}
	}
}

func TestFirewallZonesDataSourceSchema(t *testing.T) {
	resp := &datasource.SchemaResponse{}
	NewFirewallZonesDataSource().Schema(context.Background(), datasource.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("firewall_zones schema diagnostics: %v", resp.Diagnostics)
	}

	attr, ok := resp.Schema.Attributes["zones"].(schema.ListNestedAttribute)
	if !ok {
		t.Fatal("zones attribute missing or not list nested")
	}

	if _, ok := attr.NestedObject.Attributes["network_ids"].(schema.ListAttribute); !ok {
		t.Fatal("zones.network_ids attribute missing or not list")
	}
}
