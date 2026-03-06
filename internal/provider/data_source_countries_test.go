package provider

import (
	"testing"

	networkapi "github.com/ilmax/unifi-client-go/pkg/network"
)

func TestCountriesFromAPISortsByName(t *testing.T) {
	items := []networkapi.CountryDefinition{
		{Code: "US", Name: "United States"},
		{Code: "AL", Name: "Albania"},
		{Code: "IT", Name: "Italy"},
	}

	countries := countriesFromAPI(items)
	if len(countries) != 3 {
		t.Fatalf("expected 3 countries, got %d", len(countries))
	}
	if countries[0].Name.ValueString() != "Albania" {
		t.Fatalf("expected first country Albania, got %s", countries[0].Name.ValueString())
	}
	if countries[2].Code.ValueString() != "US" {
		t.Fatalf("expected last code US, got %s", countries[2].Code.ValueString())
	}
}
