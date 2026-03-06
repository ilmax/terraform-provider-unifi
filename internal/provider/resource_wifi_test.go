package provider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ilmax/terraform-provider-unifi/internal/sdkcompat/broadcasts"
)

func TestBuildWifiPayloadWithNewFields(t *testing.T) {
	macList, diags := types.ListValueFrom(context.Background(), types.StringType, []string{"aa:bb:cc:dd:ee:ff"})
	if diags.HasError() {
		t.Fatalf("unexpected list diagnostics: %v", diags)
	}

	plan := wifiResourceModel{
		Name:         types.StringValue("wifi"),
		Type:         types.StringValue("STANDARD"),
		Enabled:      types.BoolValue(true),
		SecurityType: types.StringValue("WPA2_PERSONAL"),
		BasicDataRateKbpsByFrequencyGHz: &wifiBasicDataRateModel{
			Rate24Kbps: types.Int64Value(6000),
			Rate5Kbps:  types.Int64Value(12000),
		},
		ClientFilteringPolicy: &wifiClientFilteringPolicyModel{
			Action:           types.StringValue("ALLOW"),
			MacAddressFilter: macList,
		},
		BlackoutScheduleConfiguration: &wifiBlackoutScheduleConfigurationModel{
			Days: []wifiBlackoutDayModel{
				{
					Day:  types.StringValue("MONDAY"),
					Type: types.StringValue("ALL_DAY"),
				},
			},
		},
	}

	payload, payloadDiags := buildWifiPayload(plan)
	if payloadDiags.HasError() {
		t.Fatalf("unexpected payload diagnostics: %v", payloadDiags)
	}

	rates, ok := payload["basicDataRateKbpsByFrequencyGHz"].(map[string]any)
	if !ok {
		t.Fatal("basicDataRateKbpsByFrequencyGHz missing")
	}
	if rates["2.4"] != int64(6000) {
		t.Fatalf("expected 2.4 rate 6000, got %#v", rates["2.4"])
	}
	if rates["5"] != int64(12000) {
		t.Fatalf("expected 5 rate 12000, got %#v", rates["5"])
	}

	clientFiltering, ok := payload["clientFilteringPolicy"].(map[string]any)
	if !ok {
		t.Fatal("clientFilteringPolicy missing")
	}
	if clientFiltering["action"] != "ALLOW" {
		t.Fatalf("expected action ALLOW, got %#v", clientFiltering["action"])
	}

	blackoutCfg, ok := payload["blackoutScheduleConfiguration"].(map[string]any)
	if !ok {
		t.Fatal("blackoutScheduleConfiguration missing")
	}
	days, ok := blackoutCfg["days"].([]map[string]any)
	if !ok || len(days) != 1 {
		t.Fatalf("expected one blackout day, got %#v", blackoutCfg["days"])
	}
}

func TestBuildWifiPayloadRejectsEmptyBasicDataRates(t *testing.T) {
	plan := wifiResourceModel{
		Name:         types.StringValue("wifi"),
		Type:         types.StringValue("STANDARD"),
		Enabled:      types.BoolValue(true),
		SecurityType: types.StringValue("WPA2_PERSONAL"),
		BasicDataRateKbpsByFrequencyGHz: &wifiBasicDataRateModel{
			Rate24Kbps: types.Int64Null(),
			Rate5Kbps:  types.Int64Null(),
		},
	}

	_, diags := buildWifiPayload(plan)
	if !diags.HasError() {
		t.Fatal("expected diagnostics for empty basic_data_rate_kbps_by_frequency_ghz")
	}
}

func TestWifiStateFromResponseWithNewFields(t *testing.T) {
	response := &broadcasts.GetWifiBroadcastDetailsResponseStandard{
		Type:    "STANDARD",
		Id:      "wifi-id",
		Name:    "wifi",
		Enabled: true,
		SecurityConfiguration: &broadcasts.GetWifiBroadcastDetailsSecurityConfiguration{
			Type: "WPA2_PERSONAL",
		},
		BasicDataRateKbpsByFrequencyGHz: &broadcasts.GetWifiBroadcastDetailsBasicDataRateKbpsByFrequencyGHz{
			N24: 6000,
			N5:  12000,
		},
		ClientFilteringPolicy: &broadcasts.GetWifiBroadcastDetailsClientFilteringPolicy{
			Action:           "ALLOW",
			MacAddressFilter: []json.RawMessage{json.RawMessage(`"aa:bb:cc:dd:ee:ff"`)},
		},
		BlackoutScheduleConfiguration: &broadcasts.GetWifiBroadcastDetailsBlackoutScheduleConfiguration{
			Days: []broadcasts.GetWifiBroadcastDetailsBlackoutScheduleConfigurationDays{
				{Day: "MONDAY", Type: "ALL_DAY"},
			},
		},
	}

	state, diags := wifiStateFromResponse("site-id", response)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if state.BasicDataRateKbpsByFrequencyGHz == nil {
		t.Fatal("expected basic_data_rate_kbps_by_frequency_ghz in state")
	}
	if state.BasicDataRateKbpsByFrequencyGHz.Rate24Kbps.ValueInt64() != 6000 {
		t.Fatalf("expected 2.4 rate 6000, got %d", state.BasicDataRateKbpsByFrequencyGHz.Rate24Kbps.ValueInt64())
	}
	if state.ClientFilteringPolicy == nil {
		t.Fatal("expected client_filtering_policy in state")
	}
	if state.BlackoutScheduleConfiguration == nil || len(state.BlackoutScheduleConfiguration.Days) != 1 {
		t.Fatal("expected blackout_schedule_configuration.days in state")
	}
}
