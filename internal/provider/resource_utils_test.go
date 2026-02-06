package provider

import (
	"context"
	"encoding/json"
	"testing"
)

func TestRawMessagesToListSkipsNulls(t *testing.T) {
	items := []json.RawMessage{
		json.RawMessage(`"1.1.1.1"`),
		json.RawMessage(`null`),
		json.RawMessage(`""`),
	}

	list := rawMessagesToList(items)
	if list.IsNull() {
		t.Fatal("expected list to be non-null")
	}

	var values []string
	diags := list.ElementsAs(context.Background(), &values, false)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if len(values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(values))
	}
	if values[0] != "1.1.1.1" {
		t.Fatalf("expected first value to be 1.1.1.1, got %s", values[0])
	}
	if values[1] != "" {
		t.Fatalf("expected second value to be empty string, got %s", values[1])
	}
}
