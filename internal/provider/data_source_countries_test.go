package provider

import (
	"errors"
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

func TestCollectCountriesPagesReadsAllPages(t *testing.T) {
	var offsets []int32

	countries, err := collectCountriesPages(2, 10, func(offset, limit int32) (networkapi.CountryDefinitionPage, error) {
		offsets = append(offsets, offset)
		switch offset {
		case 0:
			return networkapi.CountryDefinitionPage{
				TotalCount: 4,
				Data: []networkapi.CountryDefinition{
					{Code: "AL", Name: "Albania"},
					{Code: "AD", Name: "Andorra"},
				},
			}, nil
		case 2:
			return networkapi.CountryDefinitionPage{
				TotalCount: 4,
				Data: []networkapi.CountryDefinition{
					{Code: "US", Name: "United States"},
					{Code: "IT", Name: "Italy"},
				},
			}, nil
		default:
			return networkapi.CountryDefinitionPage{}, nil
		}
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(countries) != 4 {
		t.Fatalf("expected 4 countries, got %d", len(countries))
	}
	if len(offsets) != 2 || offsets[0] != 0 || offsets[1] != 2 {
		t.Fatalf("unexpected offsets: %#v", offsets)
	}
}

func TestCollectCountriesPagesStopsOnShortPageWithoutTotalCount(t *testing.T) {
	countries, err := collectCountriesPages(2, 10, func(offset, limit int32) (networkapi.CountryDefinitionPage, error) {
		switch offset {
		case 0:
			return networkapi.CountryDefinitionPage{
				Data: []networkapi.CountryDefinition{
					{Code: "AL", Name: "Albania"},
					{Code: "AD", Name: "Andorra"},
				},
			}, nil
		case 2:
			return networkapi.CountryDefinitionPage{
				Data: []networkapi.CountryDefinition{
					{Code: "US", Name: "United States"},
				},
			}, nil
		default:
			return networkapi.CountryDefinitionPage{}, nil
		}
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(countries) != 3 {
		t.Fatalf("expected 3 countries, got %d", len(countries))
	}
}

func TestCollectCountriesPagesPropagatesError(t *testing.T) {
	expectedErr := errors.New("boom")
	_, err := collectCountriesPages(2, 10, func(offset, limit int32) (networkapi.CountryDefinitionPage, error) {
		return networkapi.CountryDefinitionPage{}, expectedErr
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestCollectCountriesPagesReturnsErrorWhenMaxPagesExceeded(t *testing.T) {
	_, err := collectCountriesPages(1, 2, func(offset, limit int32) (networkapi.CountryDefinitionPage, error) {
		return networkapi.CountryDefinitionPage{
			Data: []networkapi.CountryDefinition{
				{Code: "AL", Name: "Albania"},
			},
		}, nil
	})
	if err == nil {
		t.Fatal("expected max pages exceeded error")
	}
}
