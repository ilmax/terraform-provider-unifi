package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/action"
	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// newActionInvokeRequest builds an action.InvokeRequest directly against an
// action's schema, bypassing the Terraform CLI entirely. This exists
// because neither terraform-plugin-testing v1.16.0 (no ActionInvoke/action
// TestStep support at all) nor the Terraform CLI installed in this repo's
// dev environment (v1.5.7, predates the `action` block syntax) can drive a
// real acceptance test for Actions today. Once either gains that support,
// these validation cases belong in internal/acceptance instead.
func newActionInvokeRequest(t *testing.T, sch actionschema.Schema, attrs map[string]tftypes.Value) action.InvokeRequest {
	t.Helper()
	ctx := context.Background()

	tfType := sch.Type().TerraformType(ctx)
	objType, ok := tfType.(tftypes.Object)
	if !ok {
		t.Fatalf("expected schema type to be an object, got %T", tfType)
	}

	values := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, at := range objType.AttributeTypes {
		if v, ok := attrs[name]; ok {
			values[name] = v
		} else {
			values[name] = tftypes.NewValue(at, nil)
		}
	}

	return action.InvokeRequest{
		Config: tfsdk.Config{
			Raw:    tftypes.NewValue(objType, values),
			Schema: sch,
		},
	}
}

func actionDiagsContain(resp *action.InvokeResponse, substr string) bool {
	for _, d := range resp.Diagnostics {
		if strings.Contains(strings.ToLower(d.Summary()), strings.ToLower(substr)) ||
			strings.Contains(strings.ToLower(d.Detail()), strings.ToLower(substr)) {
			return true
		}
	}
	return false
}

func TestExecutePortActionInvokeValidation(t *testing.T) {
	sch := action.SchemaResponse{}
	a := &executePortAction{siteID: "test-site"}
	a.Schema(context.Background(), action.SchemaRequest{}, &sch)

	cases := []struct {
		name    string
		attrs   map[string]tftypes.Value
		wantErr string
	}{
		{
			name: "missing device_id",
			attrs: map[string]tftypes.Value{
				"device_id": tftypes.NewValue(tftypes.String, nil),
				"port_idx":  tftypes.NewValue(tftypes.Number, 1),
				"action":    tftypes.NewValue(tftypes.String, "POWER_CYCLE"),
			},
			wantErr: "Missing device_id",
		},
		{
			name: "invalid port_idx",
			attrs: map[string]tftypes.Value{
				"device_id": tftypes.NewValue(tftypes.String, "device-1"),
				"port_idx":  tftypes.NewValue(tftypes.Number, 0),
				"action":    tftypes.NewValue(tftypes.String, "POWER_CYCLE"),
			},
			wantErr: "Invalid port_idx",
		},
		{
			name: "missing action",
			attrs: map[string]tftypes.Value{
				"device_id": tftypes.NewValue(tftypes.String, "device-1"),
				"port_idx":  tftypes.NewValue(tftypes.Number, 1),
				"action":    tftypes.NewValue(tftypes.String, nil),
			},
			wantErr: "Missing action",
		},
		{
			name: "unsupported action",
			attrs: map[string]tftypes.Value{
				"device_id": tftypes.NewValue(tftypes.String, "device-1"),
				"port_idx":  tftypes.NewValue(tftypes.Number, 1),
				"action":    tftypes.NewValue(tftypes.String, "REBOOT"),
			},
			wantErr: "Unsupported action",
		},
		{
			name: "invalid timeout_seconds",
			attrs: map[string]tftypes.Value{
				"device_id":       tftypes.NewValue(tftypes.String, "device-1"),
				"port_idx":        tftypes.NewValue(tftypes.Number, 1),
				"action":          tftypes.NewValue(tftypes.String, "POWER_CYCLE"),
				"timeout_seconds": tftypes.NewValue(tftypes.Number, 0),
			},
			wantErr: "Invalid timeout_seconds",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := newActionInvokeRequest(t, sch.Schema, tc.attrs)
			resp := &action.InvokeResponse{}
			a.Invoke(context.Background(), req, resp)
			if !resp.Diagnostics.HasError() {
				t.Fatalf("expected an error, got none")
			}
			if !actionDiagsContain(resp, tc.wantErr) {
				t.Fatalf("expected diagnostics to mention %q, got: %v", tc.wantErr, resp.Diagnostics)
			}
		})
	}

	t.Run("missing site_id", func(t *testing.T) {
		noSite := &executePortAction{}
		req := newActionInvokeRequest(t, sch.Schema, map[string]tftypes.Value{
			"device_id": tftypes.NewValue(tftypes.String, "device-1"),
			"port_idx":  tftypes.NewValue(tftypes.Number, 1),
			"action":    tftypes.NewValue(tftypes.String, "POWER_CYCLE"),
		})
		resp := &action.InvokeResponse{}
		noSite.Invoke(context.Background(), req, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatalf("expected an error, got none")
		}
		if !actionDiagsContain(resp, "site_id") {
			t.Fatalf("expected diagnostics to mention site_id, got: %v", resp.Diagnostics)
		}
	})
}

func TestExecuteAdoptedDeviceActionInvokeValidation(t *testing.T) {
	sch := action.SchemaResponse{}
	a := &executeAdoptedDeviceAction{siteID: "test-site"}
	a.Schema(context.Background(), action.SchemaRequest{}, &sch)

	cases := []struct {
		name    string
		attrs   map[string]tftypes.Value
		wantErr string
	}{
		{
			name: "missing device_id",
			attrs: map[string]tftypes.Value{
				"device_id": tftypes.NewValue(tftypes.String, nil),
				"action":    tftypes.NewValue(tftypes.String, "RESTART"),
			},
			wantErr: "Missing device_id",
		},
		{
			name: "missing action",
			attrs: map[string]tftypes.Value{
				"device_id": tftypes.NewValue(tftypes.String, "device-1"),
				"action":    tftypes.NewValue(tftypes.String, nil),
			},
			wantErr: "Missing action",
		},
		{
			name: "unsupported action",
			attrs: map[string]tftypes.Value{
				"device_id": tftypes.NewValue(tftypes.String, "device-1"),
				"action":    tftypes.NewValue(tftypes.String, "POWER_CYCLE"),
			},
			wantErr: "Unsupported action",
		},
		{
			name: "invalid timeout_seconds",
			attrs: map[string]tftypes.Value{
				"device_id":       tftypes.NewValue(tftypes.String, "device-1"),
				"action":          tftypes.NewValue(tftypes.String, "RESTART"),
				"timeout_seconds": tftypes.NewValue(tftypes.Number, -5),
			},
			wantErr: "Invalid timeout_seconds",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := newActionInvokeRequest(t, sch.Schema, tc.attrs)
			resp := &action.InvokeResponse{}
			a.Invoke(context.Background(), req, resp)
			if !resp.Diagnostics.HasError() {
				t.Fatalf("expected an error, got none")
			}
			if !actionDiagsContain(resp, tc.wantErr) {
				t.Fatalf("expected diagnostics to mention %q, got: %v", tc.wantErr, resp.Diagnostics)
			}
		})
	}

	t.Run("missing site_id", func(t *testing.T) {
		noSite := &executeAdoptedDeviceAction{}
		req := newActionInvokeRequest(t, sch.Schema, map[string]tftypes.Value{
			"device_id": tftypes.NewValue(tftypes.String, "device-1"),
			"action":    tftypes.NewValue(tftypes.String, "RESTART"),
		})
		resp := &action.InvokeResponse{}
		noSite.Invoke(context.Background(), req, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatalf("expected an error, got none")
		}
		if !actionDiagsContain(resp, "site_id") {
			t.Fatalf("expected diagnostics to mention site_id, got: %v", resp.Diagnostics)
		}
	})
}
