package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestResolveClientLookup(t *testing.T) {
	cases := []struct {
		name       string
		clientID   types.String
		macAddress types.String
		expectOK   bool
		expectErr  bool
		expectID   string
		expectMAC  string
	}{
		{
			name:       "missing identifiers",
			clientID:   types.StringNull(),
			macAddress: types.StringNull(),
			expectOK:   false,
			expectErr:  true,
		},
		{
			name:       "both identifiers set",
			clientID:   types.StringValue("client-1"),
			macAddress: types.StringValue("aa:bb:cc:dd:ee:ff"),
			expectOK:   false,
			expectErr:  true,
		},
		{
			name:       "invalid mac",
			clientID:   types.StringNull(),
			macAddress: types.StringValue("not-a-mac"),
			expectOK:   false,
			expectErr:  true,
		},
		{
			name:       "mac normalization",
			clientID:   types.StringNull(),
			macAddress: types.StringValue("AA-BB-CC-DD-EE-FF"),
			expectOK:   true,
			expectErr:  false,
			expectMAC:  "aa:bb:cc:dd:ee:ff",
		},
		{
			name:       "client id lookup",
			clientID:   types.StringValue("client-123"),
			macAddress: types.StringNull(),
			expectOK:   true,
			expectErr:  false,
			expectID:   "client-123",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			model := clientDataSourceModel{
				ClientID:   tc.clientID,
				MacAddress: tc.macAddress,
			}
			var diags diag.Diagnostics
			clientID, macAddress, ok := resolveClientLookup(model, &diags)
			if ok != tc.expectOK {
				t.Fatalf("expected ok=%v, got %v", tc.expectOK, ok)
			}
			if diags.HasError() != tc.expectErr {
				t.Fatalf("expected error=%v, got %v (diags=%v)", tc.expectErr, diags.HasError(), diags)
			}
			if tc.expectID != "" && clientID != tc.expectID {
				t.Fatalf("expected client id %q, got %q", tc.expectID, clientID)
			}
			if tc.expectMAC != "" && macAddress != tc.expectMAC {
				t.Fatalf("expected mac %q, got %q", tc.expectMAC, macAddress)
			}
		})
	}
}
