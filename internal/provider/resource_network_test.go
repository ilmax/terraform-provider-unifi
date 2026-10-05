package provider

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/sdkcompat/networks"
)

func TestResolveIPv4HostPrefix(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		model := &networkIPv4ConfigurationModel{
			HostIPAddress: types.StringValue("10.0.0.1"),
			PrefixLength:  types.Int64Value(24),
		}
		var diags diag.Diagnostics
		host, prefix := resolveIPv4HostPrefix(model, &diags)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if host != "10.0.0.1" {
			t.Fatalf("expected host 10.0.0.1, got %s", host)
		}
		if prefix != 24 {
			t.Fatalf("expected prefix 24, got %d", prefix)
		}
	})

	t.Run("missing-host", func(t *testing.T) {
		model := &networkIPv4ConfigurationModel{
			PrefixLength: types.Int64Value(24),
		}
		var diags diag.Diagnostics
		resolveIPv4HostPrefix(model, &diags)
		if !diags.HasError() {
			t.Fatal("expected diagnostics for missing host_ip_address")
		}
	})

	t.Run("missing-prefix", func(t *testing.T) {
		model := &networkIPv4ConfigurationModel{
			HostIPAddress: types.StringValue("10.0.0.1"),
		}
		var diags diag.Diagnostics
		resolveIPv4HostPrefix(model, &diags)
		if !diags.HasError() {
			t.Fatal("expected diagnostics for missing prefix_length")
		}
	})
}

func TestBoolOrDefault(t *testing.T) {
	if got := boolOrDefault(types.BoolNull(), true); got != true {
		t.Errorf("null value: expected default true, got %v", got)
	}
	if got := boolOrDefault(types.BoolUnknown(), true); got != true {
		t.Errorf("unknown value: expected default true, got %v", got)
	}
	if got := boolOrDefault(types.BoolValue(false), true); got != false {
		t.Errorf("explicit false: expected false to win over default, got %v", got)
	}
	if got := boolOrDefault(types.BoolValue(true), false); got != true {
		t.Errorf("explicit true: expected true to win over default, got %v", got)
	}
}

// TestBuildCreateNetworkRequestGateway locks in two real API requirements
// discovered against a live console: isolationEnabled, cellularBackupEnabled
// and internetAccessEnabled must be sent as explicit booleans (never
// omitted, even when false) and pingConflictDetectionEnabled is required
// inside dhcpConfiguration whenever mode is SERVER.
func TestBuildCreateNetworkRequestGateway(t *testing.T) {
	plan := networkResourceModel{
		Name:       types.StringValue("gw-test"),
		Management: types.StringValue("GATEWAY"),
		Enabled:    types.BoolValue(true),
		VlanID:     types.Int64Value(200),
		ZoneID:     types.StringValue("zone-1"),
		// Left null/unset deliberately to exercise boolOrDefault's fallback.
		IsolationEnabled:      types.BoolNull(),
		CellularBackupEnabled: types.BoolNull(),
		InternetAccessEnabled: types.BoolNull(),
		IPv4Configuration: &networkIPv4ConfigurationModel{
			HostIPAddress:           types.StringValue("10.0.0.1"),
			PrefixLength:            types.Int64Value(24),
			AdditionalHostIPSubnets: types.ListNull(types.StringType),
			DHCPConfiguration: &networkIPv4DHCPConfigurationModel{
				Mode: types.StringValue("SERVER"),
				IPAddressRange: &networkIPAddressRangeModel{
					Start: types.StringValue("10.0.0.100"),
					Stop:  types.StringValue("10.0.0.200"),
				},
				DNSServers:                   types.ListNull(types.StringType),
				LeaseTimeSeconds:             types.Int64Value(3600),
				PingConflictDetectionEnabled: types.BoolValue(true),
			},
		},
	}

	payload, diags := buildCreateNetworkRequest(plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	req, ok := payload.(*networks.CreateNetworkRequestGateway)
	if !ok {
		t.Fatalf("expected *networks.CreateNetworkRequestGateway, got %T", payload)
	}

	if req.IsolationEnabled {
		t.Errorf("expected IsolationEnabled to default to false when unset")
	}
	if req.CellularBackupEnabled {
		t.Errorf("expected CellularBackupEnabled to default to false when unset")
	}
	if !req.InternetAccessEnabled {
		t.Errorf("expected InternetAccessEnabled to default to true when unset")
	}
	if req.Ipv4Configuration == nil || req.Ipv4Configuration.DhcpConfiguration == nil {
		t.Fatalf("expected ipv4_configuration.dhcp_configuration to be set")
	}
	if !req.Ipv4Configuration.DhcpConfiguration.PingConflictDetectionEnabled {
		t.Errorf("expected PingConflictDetectionEnabled to round-trip as true")
	}

	// Regression lock for the omitempty-on-bool bug: false/default values
	// must still appear in the outgoing JSON, not be silently dropped.
	raw, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, field := range []string{"isolationEnabled", "cellularBackupEnabled", "internetAccessEnabled"} {
		if _, ok := decoded[field]; !ok {
			t.Errorf("expected JSON to include %q even though its value is false, but it was omitted", field)
		}
	}
}

func TestBuildCreateNetworkRequestSwitch(t *testing.T) {
	plan := networkResourceModel{
		Name:                  types.StringValue("sw-test"),
		Management:            types.StringValue("SWITCH"),
		Enabled:               types.BoolValue(true),
		VlanID:                types.Int64Value(300),
		DeviceID:              types.StringValue("device-1"),
		IsolationEnabled:      types.BoolNull(),
		CellularBackupEnabled: types.BoolNull(),
	}

	payload, diags := buildCreateNetworkRequest(plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	req, ok := payload.(*networks.CreateNetworkRequestSwitch)
	if !ok {
		t.Fatalf("expected *networks.CreateNetworkRequestSwitch, got %T", payload)
	}
	if req.DeviceId != "device-1" {
		t.Errorf("expected DeviceId to round-trip, got %q", req.DeviceId)
	}

	raw, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, field := range []string{"isolationEnabled", "cellularBackupEnabled"} {
		if _, ok := decoded[field]; !ok {
			t.Errorf("expected JSON to include %q even though its value is false, but it was omitted", field)
		}
	}
}

// TestNetworkStateFromResponseNullsUnsupportedFields regression-locks the
// fix for a real "Provider produced inconsistent result after apply" bug:
// each management type must explicitly null out the Computed attributes
// that don't apply to it, or Terraform sees them stuck at Unknown.
func TestNetworkStateFromResponseNullsUnsupportedFields(t *testing.T) {
	t.Run("gateway", func(t *testing.T) {
		state := &networkResourceModel{}
		diags := networkStateFromResponse(state, "site-1", &networks.GetNetworkDetailsResponseGateway{
			Management: "GATEWAY",
			Id:         "net-1",
			Name:       "gw",
			VlanId:     10,
		})
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if !state.DeviceID.IsNull() {
			t.Errorf("expected DeviceID null for GATEWAY, got %v", state.DeviceID)
		}
	})

	t.Run("switch", func(t *testing.T) {
		state := &networkResourceModel{}
		diags := networkStateFromResponse(state, "site-1", &networks.GetNetworkDetailsResponseSwitch{
			Management: "SWITCH",
			Id:         "net-2",
			Name:       "sw",
			VlanId:     20,
		})
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if !state.ZoneID.IsNull() {
			t.Errorf("expected ZoneID null for SWITCH, got %v", state.ZoneID)
		}
		if !state.InternetAccessEnabled.IsNull() {
			t.Errorf("expected InternetAccessEnabled null for SWITCH, got %v", state.InternetAccessEnabled)
		}
	})

	t.Run("unmanaged", func(t *testing.T) {
		state := &networkResourceModel{}
		diags := networkStateFromResponse(state, "site-1", &networks.GetNetworkDetailsResponseUnmanaged{
			Management: "UNMANAGED",
			Id:         "net-3",
			Name:       "un",
			VlanId:     30,
		})
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		for name, v := range map[string]types.String{"ZoneID": state.ZoneID, "DeviceID": state.DeviceID} {
			if !v.IsNull() {
				t.Errorf("expected %s null for UNMANAGED, got %v", name, v)
			}
		}
		for name, v := range map[string]types.Bool{
			"IsolationEnabled":      state.IsolationEnabled,
			"CellularBackupEnabled": state.CellularBackupEnabled,
			"InternetAccessEnabled": state.InternetAccessEnabled,
		} {
			if !v.IsNull() {
				t.Errorf("expected %s null for UNMANAGED, got %v", name, v)
			}
		}
	})
}

// TestReadIPv4DHCPConfigIncludesPingConflictDetection regression-locks the
// previously-missing ping_conflict_detection_enabled field on the read path.
func TestReadIPv4DHCPConfigIncludesPingConflictDetection(t *testing.T) {
	model := readIPv4DHCPConfig(&networks.GetNetworkDetailsIpv4ConfigurationDhcpConfiguration{
		Mode:                         "SERVER",
		LeaseTimeSeconds:             3600,
		PingConflictDetectionEnabled: true,
	})
	if model == nil {
		t.Fatal("expected non-nil model")
	}
	if !model.PingConflictDetectionEnabled.ValueBool() {
		t.Errorf("expected PingConflictDetectionEnabled true, got %v", model.PingConflictDetectionEnabled)
	}
}
