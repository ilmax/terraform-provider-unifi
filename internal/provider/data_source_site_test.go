package provider

import (
	"testing"

	"github.com/ilmax/unifi-client-go/pkg/sitemanager"
)

func TestBuildSiteListPath(t *testing.T) {
	path := buildSiteListPath(200, "")
	if path != "/v1/sites?pageSize=200" {
		t.Fatalf("unexpected path: %s", path)
	}

	path = buildSiteListPath(200, "next-token")
	if path != "/v1/sites?nextToken=next-token&pageSize=200" {
		t.Fatalf("unexpected paginated path: %s", path)
	}
}

func TestMatchSitesByName(t *testing.T) {
	sites := []sitemanager.Site{
		{SiteID: "1", Meta: sitemanager.SiteMeta{Name: "Home"}},
		{SiteID: "2", Meta: sitemanager.SiteMeta{Name: "Lab"}},
		{SiteID: "3", Meta: sitemanager.SiteMeta{Name: "home"}},
	}

	matches := matchSitesByName(sites, " HOME ")
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
	if matches[0].SiteID != "1" || matches[1].SiteID != "3" {
		t.Fatalf("unexpected matches: %#v", matches)
	}
}

func TestSiteStateFromAPI(t *testing.T) {
	state := siteStateFromAPI(sitemanager.Site{
		SiteID: "site-1",
		HostID: "host-1",
		Meta: sitemanager.SiteMeta{
			Name:     "Home",
			Desc:     "Primary site",
			Timezone: "Europe/Amsterdam",
		},
	})

	if state.ID.ValueString() != "site-1" {
		t.Fatalf("unexpected id: %s", state.ID.ValueString())
	}
	if state.Name.ValueString() != "Home" {
		t.Fatalf("unexpected name: %s", state.Name.ValueString())
	}
	if state.Description.ValueString() != "Primary site" {
		t.Fatalf("unexpected description: %s", state.Description.ValueString())
	}
	if state.Timezone.ValueString() != "Europe/Amsterdam" {
		t.Fatalf("unexpected timezone: %s", state.Timezone.ValueString())
	}
}
