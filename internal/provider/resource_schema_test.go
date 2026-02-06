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
}

func TestResourceSchemas(t *testing.T) {
	checks := []resourceSchemaCheck{
		{
			name:            "device",
			resource:        NewDeviceResource(),
			requiredStrings: []string{"device_id"},
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
			name:            "firewall",
			resource:        NewFirewallResource(),
			requiredStrings: []string{"type", "name", "action"},
		},
		{
			name:            "wan",
			resource:        NewWanResource(),
			requiredStrings: []string{"wan_id"},
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
	}
}
