package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
)

func TestProviderSchema(t *testing.T) {
	p := New("test")()
	resp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}

	apiKeyAttr, ok := resp.Schema.Attributes["api_key"].(schema.StringAttribute)
	if !ok {
		t.Fatal("api_key attribute is missing or not a string attribute")
	}
	if !apiKeyAttr.Required {
		t.Fatal("api_key should be required")
	}

	apiURLAttr, ok := resp.Schema.Attributes["api_url"].(schema.StringAttribute)
	if !ok {
		t.Fatal("api_url attribute is missing or not a string attribute")
	}
	if !apiURLAttr.Optional {
		t.Fatal("api_url should be optional")
	}

	allowInsecureAttr, ok := resp.Schema.Attributes["allow_insecure"].(schema.BoolAttribute)
	if !ok {
		t.Fatal("allow_insecure attribute is missing or not a bool attribute")
	}
	if !allowInsecureAttr.Optional {
		t.Fatal("allow_insecure should be optional")
	}
}
