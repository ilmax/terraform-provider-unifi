package legacy

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ilmax/terraform-provider-unifi/internal/unifi"
)

func TestConsoleURL(t *testing.T) {
	for in, want := range map[string]string{
		"https://c":                                  "https://c",
		"https://c/":                                 "https://c",
		"https://c/proxy/network":                    "https://c",
		"https://c/proxy/network/integrations":       "https://c",
		"https://c/proxy/network/integration/":       "https://c",
		"https://c:11443/proxy/network/integrations": "https://c:11443",
	} {
		if got := ConsoleURL(in); got != want {
			t.Errorf("ConsoleURL(%q) = %q, want %q", in, got, want)
		}
	}
}

// fakeConsole serves the three legacy calls Teleport needs and records the PUT body.
func fakeConsole(t *testing.T, putBody *map[string]any) (*Client, string) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/proxy/network/api/self/sites", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "key" {
			http.Error(w, "no key", http.StatusUnauthorized)
			return
		}
		io.WriteString(w, `{"meta":{"rc":"ok"},"data":[{"external_id":"uuid-1","name":"default"}]}`)
	})
	mux.HandleFunc("/proxy/network/api/s/default/rest/setting", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"meta":{"rc":"ok"},"data":[{"_id":"a","key":"mgmt"},{"_id":"t1","key":"teleport","enabled":true,"subnet_cidr":"10.0.0.0/24"}]}`)
	})
	mux.HandleFunc("/proxy/network/api/s/default/rest/setting/teleport/t1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "method", http.StatusMethodNotAllowed)
			return
		}
		json.NewDecoder(r.Body).Decode(putBody)
		io.WriteString(w, `{"meta":{"rc":"ok"},"data":[{"_id":"t1","key":"teleport","enabled":false,"subnet_cidr":"10.9.0.0/24"}]}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	c, err := NewClient(unifi.Config{APIKey: "key", BaseURL: srv.URL + "/proxy/network/integrations"})
	if err != nil {
		t.Fatal(err)
	}
	return c, srv.URL
}

func TestTeleportRoundTrip(t *testing.T) {
	var put map[string]any
	c, _ := fakeConsole(t, &put)
	ctx := context.Background()

	got, err := c.GetTeleport(ctx, "uuid-1")
	if err != nil {
		t.Fatal(err)
	}
	if got != (Teleport{ID: "t1", Enabled: true, SubnetCIDR: "10.0.0.0/24"}) {
		t.Fatalf("GetTeleport = %+v", got)
	}

	updated, err := c.UpdateTeleport(ctx, "uuid-1", got.ID, Teleport{Enabled: false, SubnetCIDR: "10.9.0.0/24"})
	if err != nil {
		t.Fatal(err)
	}
	if put["enabled"] != false || put["subnet_cidr"] != "10.9.0.0/24" || len(put) != 2 {
		t.Errorf("PUT body = %v, want exactly enabled=false subnet_cidr=10.9.0.0/24", put)
	}
	if updated.Enabled || updated.SubnetCIDR != "10.9.0.0/24" {
		t.Errorf("UpdateTeleport = %+v", updated)
	}

	// An empty subnet must not be sent at all (the controller would store it
	// and later fail to list networks).
	put = nil
	if _, err := c.UpdateTeleport(ctx, "uuid-1", got.ID, Teleport{Enabled: true}); err != nil {
		t.Fatal(err)
	}
	if _, has := put["subnet_cidr"]; has || put["enabled"] != true {
		t.Errorf("PUT body = %v, want only enabled=true", put)
	}
}

func TestUnknownSiteAndBadKey(t *testing.T) {
	var put map[string]any
	c, url := fakeConsole(t, &put)
	if _, err := c.GetTeleport(context.Background(), "nope"); err == nil || !strings.Contains(err.Error(), "no site") {
		t.Errorf("unknown site error = %v", err)
	}

	bad, _ := NewClient(unifi.Config{APIKey: "wrong", BaseURL: url})
	if _, err := bad.GetTeleport(context.Background(), "uuid-1"); err == nil {
		t.Error("expected an error for a rejected API key")
	}
}

// fakeUsers is a rest/user collection that behaves like the controller: POST
// of a known MAC fails with MacUsed, PUT merges fields.
func fakeUsers(t *testing.T) (*Client, *[]map[string]any, *[]string) {
	t.Helper()
	users := []map[string]any{}
	var calls []string
	mux := http.NewServeMux()
	mux.HandleFunc("/proxy/network/api/self/sites", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"meta":{"rc":"ok"},"data":[{"external_id":"uuid-1","name":"default"}]}`)
	})
	reply := func(w http.ResponseWriter, docs ...map[string]any) {
		json.NewEncoder(w).Encode(map[string]any{"meta": map[string]any{"rc": "ok"}, "data": docs})
	}
	mux.HandleFunc("/proxy/network/api/s/default/rest/user", func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" user")
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(map[string]any{"meta": map[string]any{"rc": "ok"}, "data": users})
			return
		}
		var doc map[string]any
		json.NewDecoder(r.Body).Decode(&doc)
		for _, u := range users {
			if u["mac"] == doc["mac"] {
				io.WriteString(w, `{"meta":{"rc":"error","msg":"api.err.MacUsed"},"data":[]}`)
				return
			}
		}
		doc["_id"] = "id-" + doc["mac"].(string)
		users = append(users, doc)
		reply(w, doc)
	})
	mux.HandleFunc("/proxy/network/api/s/default/rest/user/", func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" user/id")
		var patch map[string]any
		json.NewDecoder(r.Body).Decode(&patch)
		for _, u := range users {
			if "/proxy/network/api/s/default/rest/user/"+u["_id"].(string) == r.URL.Path {
				for k, v := range patch {
					u[k] = v
				}
				reply(w, u)
				return
			}
		}
		io.WriteString(w, `{"meta":{"rc":"error","msg":"api.err.NotFound"},"data":[]}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c, err := NewClient(unifi.Config{APIKey: "key", BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	return c, &users, &calls
}

func TestReservationLifecycle(t *testing.T) {
	c, users, calls := fakeUsers(t)
	ctx := context.Background()

	got, err := c.SaveReservation(ctx, "uuid-1", Reservation{MAC: "02:00:00:AA:BB:01", FixedIP: "192.168.1.50", NetworkID: "n1", Name: "a"})
	if err != nil {
		t.Fatal(err)
	}
	if got.ID == "" || got.MAC != "02:00:00:aa:bb:01" || !got.UseFixedIP {
		t.Fatalf("created = %+v (MAC must be lowercased and use_fixedip forced on)", got)
	}

	// Saving the same MAC again must update the existing record, not POST again.
	*calls = nil
	if _, err := c.SaveReservation(ctx, "uuid-1", Reservation{MAC: "02:00:00:aa:bb:01", FixedIP: "192.168.1.51", NetworkID: "n1"}); err != nil {
		t.Fatal(err)
	}
	for _, call := range *calls {
		if call == "POST user" {
			t.Errorf("update path POSTed a new document: %v", *calls)
		}
	}
	if (*users)[0]["fixed_ip"] != "192.168.1.51" || len(*users) != 1 {
		t.Errorf("users = %v", *users)
	}

	r, err := c.GetReservation(ctx, "uuid-1", "02:00:00:aa:bb:01")
	if err != nil || r == nil || r.FixedIP != "192.168.1.51" {
		t.Fatalf("GetReservation = %+v, %v", r, err)
	}

	if err := c.ReleaseReservation(ctx, "uuid-1", r.ID); err != nil {
		t.Fatal(err)
	}
	if r, err := c.GetReservation(ctx, "uuid-1", "02:00:00:aa:bb:01"); err != nil || r != nil {
		t.Errorf("released reservation still reported: %+v, %v", r, err)
	}
	if len(*users) != 1 {
		t.Error("release must keep the client record")
	}
}
