package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/ilmax/terraform-provider-unifi/internal/sdkcompat/firewallpolicies"
)

func reportFlowTestPlan(reportFlow types.Bool) firewallPolicyResourceModel {
	return firewallPolicyResourceModel{
		Name:                         types.StringValue("block-guests"),
		Action:                       types.StringValue("BLOCK"),
		SourceZoneID:                 types.StringValue("zone-src"),
		DestinationZoneID:            types.StringValue("zone-dst"),
		SourceTrafficFilterJSON:      types.StringNull(),
		DestinationTrafficFilterJSON: types.StringNull(),
		ConnectionStateFilter:        types.SetNull(types.StringType),
		ReportFlowEnabled:            reportFlow,
	}
}

func TestBuildFirewallPolicyPayloadReportFlowEnabled(t *testing.T) {
	tests := []struct {
		name       string
		value      types.Bool
		wantSent   bool
		wantResult bool
	}{
		// Unset (null on create) and unknown (computed, not yet planned) must
		// not be sent: the API then applies its own default (true).
		{"unset is omitted", types.BoolNull(), false, false},
		{"unknown is omitted", types.BoolUnknown(), false, false},
		{"true is sent", types.BoolValue(true), true, true},
		{"false is sent", types.BoolValue(false), true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, diags := buildFirewallPolicyPayload(reportFlowTestPlan(tt.value))
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}

			got, sent := payload["reportFlowEnabled"]
			if sent != tt.wantSent {
				t.Fatalf("reportFlowEnabled sent = %v, want %v (payload %v)", sent, tt.wantSent, payload)
			}

			if sent && got != tt.wantResult {
				t.Errorf("reportFlowEnabled = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func TestFirewallPolicyStateFromAPIReportFlowEnabled(t *testing.T) {
	for _, want := range []bool{true, false} {
		state := firewallPolicyStateFromAPI("site-1", firewallpolicies.Policy{
			Id:                "policy-1",
			Name:              "p",
			ReportFlowEnabled: want,
		})

		if state.ReportFlowEnabled.IsNull() || state.ReportFlowEnabled.IsUnknown() {
			t.Fatalf("report_flow_enabled must always be known after a read, got %v", state.ReportFlowEnabled)
		}

		if got := state.ReportFlowEnabled.ValueBool(); got != want {
			t.Errorf("report_flow_enabled = %v, want %v", got, want)
		}
	}
}

// The attribute is optional and computed with no schema default: the API reports
// false for predefined policies it never reports, so a fixed default would show
// a permanent diff on an imported predefined policy.
func TestFirewallPolicySchemaReportFlowEnabled(t *testing.T) {
	var resp resource.SchemaResponse
	NewFirewallPolicyResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	attr, ok := resp.Schema.Attributes["report_flow_enabled"].(schema.BoolAttribute)
	if !ok {
		t.Fatalf("report_flow_enabled is missing or not a bool attribute")
	}

	if !attr.Optional || !attr.Computed || attr.Required {
		t.Errorf("report_flow_enabled must be optional and computed, got optional=%v computed=%v required=%v",
			attr.Optional, attr.Computed, attr.Required)
	}

	if attr.Default != nil {
		t.Errorf("report_flow_enabled must not have a schema default")
	}

	if len(attr.PlanModifiers) == 0 {
		t.Errorf("report_flow_enabled must keep the state value across updates (UseStateForUnknown)")
	}
}
