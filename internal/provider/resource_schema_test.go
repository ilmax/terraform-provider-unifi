package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type resourceSchemaCheck struct {
	name            string
	resource        resource.Resource
	requiredStrings []string
	requiredLists   []string
}

func TestResourceSchemas(t *testing.T) {
	checks := []resourceSchemaCheck{
		{
			name:          "acl_rule_ordering",
			resource:      NewACLRuleOrderingResource(),
			requiredLists: []string{"ordered_acl_rule_ids"},
		},
		{
			name:            "dns_policy",
			resource:        NewDNSPolicyResource(),
			requiredStrings: []string{"type"},
		},
		{
			name:            "firewall_policy_ordering",
			resource:        NewFirewallPolicyOrderingResource(),
			requiredStrings: []string{"source_firewall_zone_id", "destination_firewall_zone_id"},
			requiredLists:   []string{"before_system_defined", "after_system_defined"},
		},
		{
			name:            "network",
			resource:        NewNetworkResource(),
			requiredStrings: []string{"name", "management"},
		},
		{
			name:            "wifi",
			resource:        NewWifiResource(),
			requiredStrings: []string{"name", "type", "security_type"},
		},
		{
			name:            "firewall_rule",
			resource:        NewFirewallRuleResource(),
			requiredStrings: []string{"type", "name", "action"},
		},
		{
			name:            "firewall_zone",
			resource:        NewFirewallZoneResource(),
			requiredStrings: []string{"name"},
		},
	}

	for _, check := range checks {
		resp := &resource.SchemaResponse{}
		check.resource.Schema(context.Background(), resource.SchemaRequest{}, resp)
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

		for _, attrName := range check.requiredLists {
			attr, ok := resp.Schema.Attributes[attrName].(schema.ListAttribute)
			if !ok {
				t.Fatalf("%s: %s attribute missing or not list", check.name, attrName)
			}
			if !attr.Required {
				t.Fatalf("%s: %s should be required", check.name, attrName)
			}
		}
	}
}

func TestNetworkSchemaNested(t *testing.T) {
	resp := &resource.SchemaResponse{}
	NewNetworkResource().Schema(context.Background(), resource.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("network schema diagnostics: %v", resp.Diagnostics)
	}

	ipv4Attr, ok := resp.Schema.Attributes["ipv4_configuration"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("ipv4_configuration attribute missing or not single nested")
	}

	dhcpAttr, ok := ipv4Attr.Attributes["dhcp_configuration"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("ipv4_configuration.dhcp_configuration attribute missing or not single nested")
	}

	if _, ok := dhcpAttr.Attributes["ip_address_range"].(schema.SingleNestedAttribute); !ok {
		t.Fatal("ipv4_configuration.dhcp_configuration.ip_address_range attribute missing or not single nested")
	}
	if _, ok := dhcpAttr.Attributes["dns_servers"].(schema.ListAttribute); !ok {
		t.Fatal("ipv4_configuration.dhcp_configuration.dns_servers attribute missing or not list")
	}

	if _, ok := resp.Schema.Attributes["ipv6_configuration"].(schema.SingleNestedAttribute); !ok {
		t.Fatal("ipv6_configuration attribute missing or not single nested")
	}
	ipv6Attr := resp.Schema.Attributes["ipv6_configuration"].(schema.SingleNestedAttribute)
	if _, ok := ipv6Attr.Attributes["dns_servers"].(schema.ListAttribute); !ok {
		t.Fatal("ipv6_configuration.dns_servers attribute missing or not list")
	}

	if _, ok := resp.Schema.Attributes["dhcp_guarding"].(schema.SingleNestedAttribute); !ok {
		t.Fatal("dhcp_guarding attribute missing or not single nested")
	}
}

func TestWifiSchemaAttributes(t *testing.T) {
	resp := &resource.SchemaResponse{}
	NewWifiResource().Schema(context.Background(), resource.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("wifi schema diagnostics: %v", resp.Diagnostics)
	}

	if _, ok := resp.Schema.Attributes["multicast_to_unicast_conversion_enabled"].(schema.BoolAttribute); !ok {
		t.Fatal("multicast_to_unicast_conversion_enabled attribute missing or not bool")
	}

	if _, ok := resp.Schema.Attributes["broadcasting_frequencies_ghz"].(schema.ListAttribute); !ok {
		t.Fatal("broadcasting_frequencies_ghz attribute missing or not list")
	}

	if _, ok := resp.Schema.Attributes["basic_data_rate_kbps_by_frequency_ghz"].(schema.SingleNestedAttribute); !ok {
		t.Fatal("basic_data_rate_kbps_by_frequency_ghz attribute missing or not single nested")
	}

	if _, ok := resp.Schema.Attributes["client_filtering_policy"].(schema.SingleNestedAttribute); !ok {
		t.Fatal("client_filtering_policy attribute missing or not single nested")
	}

	blackoutAttr, ok := resp.Schema.Attributes["blackout_schedule_configuration"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("blackout_schedule_configuration attribute missing or not single nested")
	}

	if _, ok := blackoutAttr.Attributes["days"].(schema.ListNestedAttribute); !ok {
		t.Fatal("blackout_schedule_configuration.days attribute missing or not list nested")
	}
}

func TestFirewallZoneSchema(t *testing.T) {
	resp := &resource.SchemaResponse{}
	NewFirewallZoneResource().Schema(context.Background(), resource.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("firewall zone schema diagnostics: %v", resp.Diagnostics)
	}

	attr, ok := resp.Schema.Attributes["network_ids"].(schema.ListAttribute)
	if !ok {
		t.Fatal("network_ids attribute missing or not list")
	}
	if !attr.Required {
		t.Fatal("network_ids should be required")
	}
}

func TestDNSPolicySchema(t *testing.T) {
	resp := &resource.SchemaResponse{}
	NewDNSPolicyResource().Schema(context.Background(), resource.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("dns policy schema diagnostics: %v", resp.Diagnostics)
	}

	enabledAttr, ok := resp.Schema.Attributes["enabled"].(schema.BoolAttribute)
	if !ok {
		t.Fatal("enabled attribute missing or not bool")
	}
	if !enabledAttr.Required {
		t.Fatal("enabled should be required")
	}

	domainAttr, ok := resp.Schema.Attributes["domain"].(schema.StringAttribute)
	if !ok {
		t.Fatal("domain attribute missing or not string")
	}
	if !domainAttr.Computed {
		t.Fatal("domain should be computed")
	}
}
