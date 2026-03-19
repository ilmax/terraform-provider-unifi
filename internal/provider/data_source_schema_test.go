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
			name:       "acl_rules",
			dataSource: NewACLRulesDataSource(),
		},
		{
			name:       "countries",
			dataSource: NewCountriesDataSource(),
		},
		{
			name:            "dns_policy",
			dataSource:      NewDNSPolicyDataSource(),
			requiredStrings: []string{"policy_id"},
		},
		{
			name:            "site",
			dataSource:      NewSiteDataSource(),
			requiredStrings: []string{"name"},
		},
		{
			name:            "dns_policies",
			dataSource:      NewDNSPoliciesDataSource(),
			optionalStrings: []string{"type", "domain"},
		},
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
			requiredStrings: []string{"name"},
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

func TestSiteScopedDataSourcesUseOptionalSiteID(t *testing.T) {
	checks := []struct {
		name       string
		dataSource datasource.DataSource
	}{
		{name: "acl_rules", dataSource: NewACLRulesDataSource()},
		{name: "client", dataSource: NewClientDataSource()},
		{name: "device", dataSource: NewDeviceDataSource()},
		{name: "dns_policy", dataSource: NewDNSPolicyDataSource()},
		{name: "dns_policies", dataSource: NewDNSPoliciesDataSource()},
		{name: "firewall_zone", dataSource: NewFirewallZoneDataSource()},
		{name: "firewall_zones", dataSource: NewFirewallZonesDataSource()},
		{name: "wan", dataSource: NewWanDataSource()},
	}

	for _, check := range checks {
		resp := &datasource.SchemaResponse{}
		check.dataSource.Schema(context.Background(), datasource.SchemaRequest{}, resp)
		if resp.Diagnostics.HasError() {
			t.Fatalf("%s schema diagnostics: %v", check.name, resp.Diagnostics)
		}

		attr, ok := resp.Schema.Attributes["site_id"].(schema.StringAttribute)
		if !ok {
			t.Fatalf("%s: site_id attribute missing or not string", check.name)
		}
		if !attr.Optional {
			t.Fatalf("%s: site_id should be optional", check.name)
		}
		if attr.Required {
			t.Fatalf("%s: site_id should not be required", check.name)
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

func TestFirewallZoneDataSourceSchema(t *testing.T) {
	resp := &datasource.SchemaResponse{}
	NewFirewallZoneDataSource().Schema(context.Background(), datasource.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("firewall_zone schema diagnostics: %v", resp.Diagnostics)
	}

	nameAttr, ok := resp.Schema.Attributes["name"].(schema.StringAttribute)
	if !ok {
		t.Fatal("name attribute missing or not string")
	}
	if !nameAttr.Required {
		t.Fatal("name should be required")
	}

	zoneIDAttr, ok := resp.Schema.Attributes["zone_id"].(schema.StringAttribute)
	if !ok {
		t.Fatal("zone_id attribute missing or not string")
	}
	if !zoneIDAttr.Computed {
		t.Fatal("zone_id should be computed")
	}
}

func TestDNSPoliciesDataSourceSchema(t *testing.T) {
	resp := &datasource.SchemaResponse{}
	NewDNSPoliciesDataSource().Schema(context.Background(), datasource.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("dns_policies schema diagnostics: %v", resp.Diagnostics)
	}

	attr, ok := resp.Schema.Attributes["policies"].(schema.ListNestedAttribute)
	if !ok {
		t.Fatal("policies attribute missing or not list nested")
	}

	if _, ok := attr.NestedObject.Attributes["domain"].(schema.StringAttribute); !ok {
		t.Fatal("policies.domain attribute missing or not string")
	}
}

func TestCountriesDataSourceSchema(t *testing.T) {
	resp := &datasource.SchemaResponse{}
	NewCountriesDataSource().Schema(context.Background(), datasource.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("countries schema diagnostics: %v", resp.Diagnostics)
	}

	attr, ok := resp.Schema.Attributes["countries"].(schema.MapNestedAttribute)
	if !ok {
		t.Fatal("countries attribute missing or not map nested")
	}

	if _, ok := attr.NestedObject.Attributes["code"].(schema.StringAttribute); !ok {
		t.Fatal("countries.code attribute missing or not string")
	}
}

func TestACLRulesDataSourceSchema(t *testing.T) {
	resp := &datasource.SchemaResponse{}
	NewACLRulesDataSource().Schema(context.Background(), datasource.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("acl_rules schema diagnostics: %v", resp.Diagnostics)
	}

	attr, ok := resp.Schema.Attributes["acl_rules"].(schema.ListNestedAttribute)
	if !ok {
		t.Fatal("acl_rules attribute missing or not list nested")
	}
	if _, ok := attr.NestedObject.Attributes["id"].(schema.StringAttribute); !ok {
		t.Fatal("acl_rules.id attribute missing or not string")
	}
	if _, ok := attr.NestedObject.Attributes["index"].(schema.Int64Attribute); !ok {
		t.Fatal("acl_rules.index attribute missing or not int64")
	}
}
