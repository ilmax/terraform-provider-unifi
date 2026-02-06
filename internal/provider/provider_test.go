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

	baseURLAttr, ok := resp.Schema.Attributes["base_url"].(schema.StringAttribute)
	if !ok {
		t.Fatal("base_url attribute is missing or not a string attribute")
	}
	if !baseURLAttr.Optional {
		t.Fatal("base_url should be optional")
	}
}
