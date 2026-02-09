package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
)

type actionSchemaCheck struct {
	name            string
	action          action.Action
	requiredStrings []string
}

func TestActionSchemas(t *testing.T) {
	checks := []actionSchemaCheck{
		{
			name:            "execute_port_action",
			action:          NewExecutePortAction(),
			requiredStrings: []string{"device_id", "port_idx", "action"},
		},
		{
			name:            "execute_adopted_device_action",
			action:          NewExecuteAdoptedDeviceAction(),
			requiredStrings: []string{"device_id", "action"},
		},
	}

	for _, check := range checks {
		resp := &action.SchemaResponse{}
		check.action.Schema(context.Background(), action.SchemaRequest{}, resp)
		if resp.Diagnostics.HasError() {
			t.Fatalf("%s schema diagnostics: %v", check.name, resp.Diagnostics)
		}

		for _, attrName := range check.requiredStrings {
			switch attr := resp.Schema.Attributes[attrName].(type) {
			case schema.StringAttribute:
				if !attr.Required {
					t.Fatalf("%s: %s should be required", check.name, attrName)
				}
			case schema.Int64Attribute:
				if !attr.Required {
					t.Fatalf("%s: %s should be required", check.name, attrName)
				}
			default:
				t.Fatalf("%s: %s attribute missing or unexpected type", check.name, attrName)
			}
		}
	}
}
