package provider

import (
	"testing"

	"github.com/google/uuid"
	networkapi "github.com/ilmax/unifi-client-go/pkg/network"
)

func TestBuildSiteListPath(t *testing.T) {
	path := buildSiteListPath(200, 0)
	if path != "/v1/sites?limit=200&offset=0" {
		t.Fatalf("unexpected path: %s", path)
	}

	path = buildSiteListPath(200, 200)
	if path != "/v1/sites?limit=200&offset=200" {
		t.Fatalf("unexpected paginated path: %s", path)
	}
}

func TestMatchSitesByName(t *testing.T) {
	sites := []networkapi.SiteOverview{
		{Id: uuid.MustParse("11111111-1111-1111-1111-111111111111"), Name: "Home"},
		{Id: uuid.MustParse("22222222-2222-2222-2222-222222222222"), Name: "Lab"},
		{Id: uuid.MustParse("33333333-3333-3333-3333-333333333333"), Name: "home"},
	}

	matches := matchSitesByName(sites, " HOME ")
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
	if matches[0].Id.String() != "11111111-1111-1111-1111-111111111111" || matches[1].Id.String() != "33333333-3333-3333-3333-333333333333" {
		t.Fatalf("unexpected matches: %#v", matches)
	}
}

func TestSiteStateFromAPI(t *testing.T) {
	state := siteStateFromAPI(networkapi.SiteOverview{
		Id:                uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Name:              "Home",
		InternalReference: "default",
	})

	if state.ID.ValueString() != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("unexpected id: %s", state.ID.ValueString())
	}
	if state.Name.ValueString() != "Home" {
		t.Fatalf("unexpected name: %s", state.Name.ValueString())
	}
	if state.InternalReference.ValueString() != "default" {
		t.Fatalf("unexpected internal reference: %s", state.InternalReference.ValueString())
	}
}
