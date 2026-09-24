# DNS Resource Acceptance Tests Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `TF_ACC`-gated acceptance tests (using `terraform-plugin-testing`) for the 6 DNS record resources, the DNS forward-domain-policy resource, and the 2 DNS data sources, so they can be run against a real UniFi console the developer already has credentials for.

**Architecture:** Standard HashiCorp `resource.Test`/`resource.TestCase` acceptance tests, gated by `TF_ACC` (built into `resource.Test` itself) and a `testAccPreCheck` that additionally skips when `UNIFI_API_KEY`/`UNIFI_SITE_ID` env vars are unset. No test-owned container lifecycle: `docker-compose.yml` remains a manual, local-dev-only convenience (a developer runs it, creates an account, and generates an API key by hand via the UI once) because API keys can only be minted at `unifi.ui.com` against a real, cloud-linked console — this cannot be scripted, in CI or locally. Each resource gets its own `_test.go` acceptance test file living in `internal/provider` (same package as the resource, matching the existing unit-test convention in this package).

**Tech Stack:** Go, `github.com/hashicorp/terraform-plugin-testing` (new dependency), `github.com/hashicorp/terraform-plugin-framework/providerserver`, existing `internal/provider` package.

**Spec:** This document (no separate spec doc — requirements were established through conversation: see `AGENTS.md`'s new "UniFi Network API Reference" section for the credential-provisioning constraints that shape this design).

## Global Constraints

- Acceptance tests must not fail when credentials are absent — they must `t.Skip`, so `go test ./...` in CI (no `TF_ACC`, no `UNIFI_API_KEY`) stays green exactly as it is today.
- No new runtime dependency on Docker/testcontainers from Go test code — the container is a human-operated convenience only.
- Test files live in `package provider`, alongside the resource/data source they test, following the existing `resource_dns_records_test.go` unit-test convention.
- Every acceptance test must exercise: create with initial values, update to changed values, and import (`ImportStateVerify: true`) using the resource's documented `<site_id>/<id>` import format.
- Use `resource.ComposeAggregateTestCheckFunc` (not the deprecated non-aggregate form) so a single run surfaces every mismatch, not just the first.

## Review Focus

- **Credentials unset (the default/CI case):** `go test ./...` must still pass with zero new failures — `testAccPreCheck` must skip, not fail. Covered by Task 1's own verification (running the suite with no env vars set).
- **`site_id` resolution:** tests must prove the provider-level `site_id` (not a per-resource override) is what's exercised, since that's the common case documented in the schema. Covered by omitting `site_id` from every resource block in every task's config and asserting the resource ends up with a non-empty `site_id` in state.
- **Import round-trip correctness:** the import ID format is `<site_id>/<id>`, not just the raw resource ID — a naive `ImportStateVerify` using only the resource's `id` attribute would silently pass while masking a broken import for anyone following the docs. Covered by the shared `testAccSiteScopedImportStateIdFunc` helper in Task 1, used by every subsequent task.
- **Optional/pointer fields round-tripping through updates:** e.g. `ttl_seconds`, `priority`, `weight` are optional int64 fields backed by `*int32` on the wire — an update that changes them must be asserted, not just the create. Covered by each resource task's second (update) step asserting the new values.
- **Origin/computed metadata surviving updates:** `origin` is API-set metadata that must remain populated (not reset to null) after an update, since a regression here would look fine on create but break on the far more common update path. Covered by asserting `origin` is set (`TestCheckResourceAttrSet`) in both the create and update steps of every resource task.

---

## File Structure

- Create: `internal/provider/acctest_test.go` — shared acceptance-test scaffolding (provider factories, precheck, provider config template, site-scoped import-ID helper). Nothing resource-specific lives here.
- Create: `internal/provider/resource_dns_a_record_acctest_test.go`
- Create: `internal/provider/resource_dns_aaaa_record_acctest_test.go`
- Create: `internal/provider/resource_dns_cname_record_acctest_test.go`
- Create: `internal/provider/resource_dns_mx_record_acctest_test.go`
- Create: `internal/provider/resource_dns_srv_record_acctest_test.go`
- Create: `internal/provider/resource_dns_txt_record_acctest_test.go`
- Create: `internal/provider/resource_dns_forward_domain_policy_acctest_test.go`
- Create: `internal/provider/data_source_dns_policy_acctest_test.go` (covers both `unifi_dns_policy` and `unifi_dns_policies`)
- Modify: `Makefile` (new file — none exists yet) — add a `testacc` target
- Modify: `README.md` — document the acceptance-test env vars and the manual, local-only container workflow

---

### Task 1: Shared acceptance-test scaffolding

**Files:**
- Create: `internal/provider/acctest_test.go`
- Modify: `go.mod`, `go.sum` (via `go get`)

**Interfaces:**
- Produces: `testAccProtoV6ProviderFactories map[string]func() (tfprotov6.ProviderServer, error)`, `testAccPreCheck(t *testing.T)`, `testAccProviderConfig() string`, `testAccSiteScopedImportStateIdFunc(resourceName string) resource.ImportStateIdFunc` — every later task consumes all four.

- [ ] **Step 1: Add the terraform-plugin-testing dependency**

Run: `go get github.com/hashicorp/terraform-plugin-testing@v1.16.0`

- [ ] **Step 2: Write the shared scaffolding file**

```go
package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testAccProtoV6ProviderFactories wires the real provider into
// terraform-plugin-testing's acceptance test runner.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"unifi": providerserver.NewProtocol6WithError(New("acctest")()),
}

// testAccPreCheck gates every acceptance test on TF_ACC (handled by
// resource.Test itself) plus the credentials this provider actually needs.
// Unlike a hard failure, this skips: acceptance tests are opt-in and must
// not break `go test ./...` in environments (including CI) that have no
// UniFi console to talk to. See AGENTS.md's "UniFi Network API Reference"
// section for why these credentials can't be provisioned automatically.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	for _, name := range []string{"UNIFI_API_KEY", "UNIFI_SITE_ID"} {
		if os.Getenv(name) == "" {
			t.Skipf("%s must be set to run acceptance tests", name)
		}
	}
}

// testAccProviderConfig renders the `provider "unifi" {}` block from
// environment variables. It deliberately sets site_id at the provider
// level (not per-resource) so every acceptance test also exercises the
// provider -> resource site_id fallback documented in docs/index.md.
func testAccProviderConfig() string {
	allowInsecure := os.Getenv("UNIFI_ALLOW_INSECURE") == "true"
	return fmt.Sprintf(`
provider "unifi" {
  api_key        = %q
  api_url        = %q
  site_id        = %q
  allow_insecure = %t
}
`, os.Getenv("UNIFI_API_KEY"), os.Getenv("UNIFI_API_URL"), os.Getenv("UNIFI_SITE_ID"), allowInsecure)
}

// testAccSiteScopedImportStateIdFunc builds the "<site_id>/<id>" import
// identifier documented for every DNS resource. Falls back to the
// provider-level UNIFI_SITE_ID when the resource itself omits site_id
// (the common case exercised by these tests), matching resolveSiteID's
// own fallback behavior in internal/provider/helpers.go.
func testAccSiteScopedImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found in state: %s", resourceName)
		}
		siteID := rs.Primary.Attributes["site_id"]
		if siteID == "" {
			siteID = os.Getenv("UNIFI_SITE_ID")
		}
		if siteID == "" {
			return "", fmt.Errorf("no site_id available for import (resource %s)", resourceName)
		}
		return fmt.Sprintf("%s/%s", siteID, rs.Primary.ID), nil
	}
}
```

- [ ] **Step 3: Verify the credentials-unset path stays green**

Run: `go test ./... 2>&1 | tail -20`
Expected: `ok  	github.com/ilmax/terraform-provider-unifi/internal/provider	...` — this file adds no `Test*` functions itself, so it should only affect compilation. Confirms the new dependency and helpers compile cleanly with no env vars set.

- [ ] **Step 4: Run go vet and go mod tidy**

Run: `go mod tidy && go vet ./...`
Expected: no output from `go vet`; `go.mod`/`go.sum` gain `terraform-plugin-testing` as a direct dependency (and promote `terraform-plugin-go` from indirect to direct, since it's now imported directly).

- [ ] **Step 5: Commit**

```bash
git add internal/provider/acctest_test.go go.mod go.sum
git commit -m "test: add DNS acceptance test scaffolding"
```

---

### Task 2: `unifi_dns_a_record` acceptance test

**Files:**
- Create: `internal/provider/resource_dns_a_record_acctest_test.go`

**Interfaces:**
- Consumes: `testAccProtoV6ProviderFactories`, `testAccPreCheck`, `testAccProviderConfig()`, `testAccSiteScopedImportStateIdFunc` from Task 1.

- [ ] **Step 1: Write the acceptance test**

```go
package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDNSARecord_basic(t *testing.T) {
	resourceName := "unifi_dns_a_record.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "unifi_dns_a_record" "test" {
  enabled      = true
  domain       = "acctest-a.example.com"
  ipv4_address = "192.0.2.10"
  ttl_seconds  = 300
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "domain", "acctest-a.example.com"),
					resource.TestCheckResourceAttr(resourceName, "ipv4_address", "192.0.2.10"),
					resource.TestCheckResourceAttr(resourceName, "ttl_seconds", "300"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "site_id"),
					resource.TestCheckResourceAttrSet(resourceName, "origin"),
				),
			},
			{
				Config: testAccProviderConfig() + `
resource "unifi_dns_a_record" "test" {
  enabled      = false
  domain       = "acctest-a.example.com"
  ipv4_address = "192.0.2.20"
  ttl_seconds  = 600
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "ipv4_address", "192.0.2.20"),
					resource.TestCheckResourceAttr(resourceName, "ttl_seconds", "600"),
					resource.TestCheckResourceAttrSet(resourceName, "origin"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccSiteScopedImportStateIdFunc(resourceName),
			},
		},
	})
}
```

- [ ] **Step 2: Verify it compiles and skips cleanly with no credentials**

Run: `go test ./internal/provider/... -run TestAccDNSARecord_basic -v`
Expected: `--- SKIP: TestAccDNSARecord_basic` (acceptance tests self-skip without `TF_ACC=1`; with `TF_ACC=1` but no `UNIFI_API_KEY`/`UNIFI_SITE_ID` it skips via `testAccPreCheck` instead).

- [ ] **Step 3: (Optional, requires real credentials) run it for real**

Run: `TF_ACC=1 UNIFI_API_KEY=... UNIFI_SITE_ID=... UNIFI_API_URL=... go test ./internal/provider/... -run TestAccDNSARecord_basic -v -timeout 5m`
Expected: `--- PASS: TestAccDNSARecord_basic`

- [ ] **Step 4: Commit**

```bash
git add internal/provider/resource_dns_a_record_acctest_test.go
git commit -m "test: add unifi_dns_a_record acceptance test"
```

---

### Task 3: `unifi_dns_aaaa_record` acceptance test

**Files:**
- Create: `internal/provider/resource_dns_aaaa_record_acctest_test.go`

**Interfaces:**
- Consumes: same four helpers from Task 1.

- [ ] **Step 1: Write the acceptance test**

```go
package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDNSAAAARecord_basic(t *testing.T) {
	resourceName := "unifi_dns_aaaa_record.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "unifi_dns_aaaa_record" "test" {
  enabled      = true
  domain       = "acctest-aaaa.example.com"
  ipv6_address = "2001:db8::10"
  ttl_seconds  = 300
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "domain", "acctest-aaaa.example.com"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address", "2001:db8::10"),
					resource.TestCheckResourceAttr(resourceName, "ttl_seconds", "300"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "site_id"),
					resource.TestCheckResourceAttrSet(resourceName, "origin"),
				),
			},
			{
				Config: testAccProviderConfig() + `
resource "unifi_dns_aaaa_record" "test" {
  enabled      = false
  domain       = "acctest-aaaa.example.com"
  ipv6_address = "2001:db8::20"
  ttl_seconds  = 600
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address", "2001:db8::20"),
					resource.TestCheckResourceAttr(resourceName, "ttl_seconds", "600"),
					resource.TestCheckResourceAttrSet(resourceName, "origin"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccSiteScopedImportStateIdFunc(resourceName),
			},
		},
	})
}
```

- [ ] **Step 2: Verify it compiles and skips cleanly**

Run: `go test ./internal/provider/... -run TestAccDNSAAAARecord_basic -v`
Expected: `--- SKIP: TestAccDNSAAAARecord_basic`

- [ ] **Step 3: Commit**

```bash
git add internal/provider/resource_dns_aaaa_record_acctest_test.go
git commit -m "test: add unifi_dns_aaaa_record acceptance test"
```

---

### Task 4: `unifi_dns_cname_record` acceptance test

**Files:**
- Create: `internal/provider/resource_dns_cname_record_acctest_test.go`

**Interfaces:**
- Consumes: same four helpers from Task 1.

- [ ] **Step 1: Write the acceptance test**

```go
package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDNSCNAMERecord_basic(t *testing.T) {
	resourceName := "unifi_dns_cname_record.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "unifi_dns_cname_record" "test" {
  enabled       = true
  domain        = "acctest-cname.example.com"
  target_domain = "target-a.example.com"
  ttl_seconds   = 300
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "domain", "acctest-cname.example.com"),
					resource.TestCheckResourceAttr(resourceName, "target_domain", "target-a.example.com"),
					resource.TestCheckResourceAttr(resourceName, "ttl_seconds", "300"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "site_id"),
					resource.TestCheckResourceAttrSet(resourceName, "origin"),
				),
			},
			{
				Config: testAccProviderConfig() + `
resource "unifi_dns_cname_record" "test" {
  enabled       = false
  domain        = "acctest-cname.example.com"
  target_domain = "target-b.example.com"
  ttl_seconds   = 600
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "target_domain", "target-b.example.com"),
					resource.TestCheckResourceAttr(resourceName, "ttl_seconds", "600"),
					resource.TestCheckResourceAttrSet(resourceName, "origin"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccSiteScopedImportStateIdFunc(resourceName),
			},
		},
	})
}
```

- [ ] **Step 2: Verify it compiles and skips cleanly**

Run: `go test ./internal/provider/... -run TestAccDNSCNAMERecord_basic -v`
Expected: `--- SKIP: TestAccDNSCNAMERecord_basic`

- [ ] **Step 3: Commit**

```bash
git add internal/provider/resource_dns_cname_record_acctest_test.go
git commit -m "test: add unifi_dns_cname_record acceptance test"
```

---

### Task 5: `unifi_dns_mx_record` acceptance test

**Files:**
- Create: `internal/provider/resource_dns_mx_record_acctest_test.go`

**Interfaces:**
- Consumes: same four helpers from Task 1.

- [ ] **Step 1: Write the acceptance test**

```go
package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDNSMXRecord_basic(t *testing.T) {
	resourceName := "unifi_dns_mx_record.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "unifi_dns_mx_record" "test" {
  enabled             = true
  domain              = "acctest-mx.example.com"
  mail_server_domain  = "mail-a.example.com"
  priority            = 10
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "domain", "acctest-mx.example.com"),
					resource.TestCheckResourceAttr(resourceName, "mail_server_domain", "mail-a.example.com"),
					resource.TestCheckResourceAttr(resourceName, "priority", "10"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "site_id"),
					resource.TestCheckResourceAttrSet(resourceName, "origin"),
				),
			},
			{
				Config: testAccProviderConfig() + `
resource "unifi_dns_mx_record" "test" {
  enabled            = false
  domain             = "acctest-mx.example.com"
  mail_server_domain = "mail-b.example.com"
  priority           = 20
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "mail_server_domain", "mail-b.example.com"),
					resource.TestCheckResourceAttr(resourceName, "priority", "20"),
					resource.TestCheckResourceAttrSet(resourceName, "origin"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccSiteScopedImportStateIdFunc(resourceName),
			},
		},
	})
}
```

- [ ] **Step 2: Verify it compiles and skips cleanly**

Run: `go test ./internal/provider/... -run TestAccDNSMXRecord_basic -v`
Expected: `--- SKIP: TestAccDNSMXRecord_basic`

- [ ] **Step 3: Commit**

```bash
git add internal/provider/resource_dns_mx_record_acctest_test.go
git commit -m "test: add unifi_dns_mx_record acceptance test"
```

---

### Task 6: `unifi_dns_srv_record` acceptance test

**Files:**
- Create: `internal/provider/resource_dns_srv_record_acctest_test.go`

**Interfaces:**
- Consumes: same four helpers from Task 1.

- [ ] **Step 1: Write the acceptance test**

```go
package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDNSSRVRecord_basic(t *testing.T) {
	resourceName := "unifi_dns_srv_record.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "unifi_dns_srv_record" "test" {
  enabled       = true
  domain        = "acctest-srv.example.com"
  service       = "_ldap"
  protocol      = "_tcp"
  server_domain = "srv-a.example.com"
  port          = 389
  priority      = 10
  weight        = 20
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "domain", "acctest-srv.example.com"),
					resource.TestCheckResourceAttr(resourceName, "service", "_ldap"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "_tcp"),
					resource.TestCheckResourceAttr(resourceName, "server_domain", "srv-a.example.com"),
					resource.TestCheckResourceAttr(resourceName, "port", "389"),
					resource.TestCheckResourceAttr(resourceName, "priority", "10"),
					resource.TestCheckResourceAttr(resourceName, "weight", "20"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "site_id"),
					resource.TestCheckResourceAttrSet(resourceName, "origin"),
				),
			},
			{
				Config: testAccProviderConfig() + `
resource "unifi_dns_srv_record" "test" {
  enabled       = false
  domain        = "acctest-srv.example.com"
  service       = "_ldap"
  protocol      = "_tcp"
  server_domain = "srv-b.example.com"
  port          = 636
  priority      = 5
  weight        = 30
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "server_domain", "srv-b.example.com"),
					resource.TestCheckResourceAttr(resourceName, "port", "636"),
					resource.TestCheckResourceAttr(resourceName, "priority", "5"),
					resource.TestCheckResourceAttr(resourceName, "weight", "30"),
					resource.TestCheckResourceAttrSet(resourceName, "origin"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccSiteScopedImportStateIdFunc(resourceName),
			},
		},
	})
}
```

- [ ] **Step 2: Verify it compiles and skips cleanly**

Run: `go test ./internal/provider/... -run TestAccDNSSRVRecord_basic -v`
Expected: `--- SKIP: TestAccDNSSRVRecord_basic`

- [ ] **Step 3: Commit**

```bash
git add internal/provider/resource_dns_srv_record_acctest_test.go
git commit -m "test: add unifi_dns_srv_record acceptance test"
```

---

### Task 7: `unifi_dns_txt_record` acceptance test

**Files:**
- Create: `internal/provider/resource_dns_txt_record_acctest_test.go`

**Interfaces:**
- Consumes: same four helpers from Task 1.

- [ ] **Step 1: Write the acceptance test**

```go
package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDNSTXTRecord_basic(t *testing.T) {
	resourceName := "unifi_dns_txt_record.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "unifi_dns_txt_record" "test" {
  enabled = true
  domain  = "acctest-txt.example.com"
  text    = "v=spf1 include:example.com ~all"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "domain", "acctest-txt.example.com"),
					resource.TestCheckResourceAttr(resourceName, "text", "v=spf1 include:example.com ~all"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "site_id"),
					resource.TestCheckResourceAttrSet(resourceName, "origin"),
				),
			},
			{
				Config: testAccProviderConfig() + `
resource "unifi_dns_txt_record" "test" {
  enabled = false
  domain  = "acctest-txt.example.com"
  text    = "v=spf1 -all"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "text", "v=spf1 -all"),
					resource.TestCheckResourceAttrSet(resourceName, "origin"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccSiteScopedImportStateIdFunc(resourceName),
			},
		},
	})
}
```

- [ ] **Step 2: Verify it compiles and skips cleanly**

Run: `go test ./internal/provider/... -run TestAccDNSTXTRecord_basic -v`
Expected: `--- SKIP: TestAccDNSTXTRecord_basic`

- [ ] **Step 3: Commit**

```bash
git add internal/provider/resource_dns_txt_record_acctest_test.go
git commit -m "test: add unifi_dns_txt_record acceptance test"
```

---

### Task 8: `unifi_dns_forward_domain_policy` acceptance test

**Files:**
- Create: `internal/provider/resource_dns_forward_domain_policy_acctest_test.go`

**Interfaces:**
- Consumes: same four helpers from Task 1.

- [ ] **Step 1: Write the acceptance test**

```go
package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDNSForwardDomainPolicy_basic(t *testing.T) {
	resourceName := "unifi_dns_forward_domain_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "unifi_dns_forward_domain_policy" "test" {
  enabled    = true
  domain     = "acctest-forward.example.com"
  ip_address = "192.0.2.53"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "domain", "acctest-forward.example.com"),
					resource.TestCheckResourceAttr(resourceName, "ip_address", "192.0.2.53"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "site_id"),
					resource.TestCheckResourceAttrSet(resourceName, "origin"),
				),
			},
			{
				Config: testAccProviderConfig() + `
resource "unifi_dns_forward_domain_policy" "test" {
  enabled    = false
  domain     = "acctest-forward.example.com"
  ip_address = "192.0.2.54"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "ip_address", "192.0.2.54"),
					resource.TestCheckResourceAttrSet(resourceName, "origin"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccSiteScopedImportStateIdFunc(resourceName),
			},
		},
	})
}
```

- [ ] **Step 2: Verify it compiles and skips cleanly**

Run: `go test ./internal/provider/... -run TestAccDNSForwardDomainPolicy_basic -v`
Expected: `--- SKIP: TestAccDNSForwardDomainPolicy_basic`

- [ ] **Step 3: Commit**

```bash
git add internal/provider/resource_dns_forward_domain_policy_acctest_test.go
git commit -m "test: add unifi_dns_forward_domain_policy acceptance test"
```

---

### Task 9: `unifi_dns_policy` and `unifi_dns_policies` data source acceptance tests

**Files:**
- Create: `internal/provider/data_source_dns_policy_acctest_test.go`

**Interfaces:**
- Consumes: same four helpers from Task 1. Depends on a resource (`unifi_dns_txt_record`) existing in the same config so the data sources have something real to look up — this is the standard "create a resource, then look it up via data source" acceptance-test shape.

- [ ] **Step 1: Write the acceptance tests**

```go
package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDNSPolicyDataSource_basic(t *testing.T) {
	dataSourceName := "data.unifi_dns_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "unifi_dns_txt_record" "lookup" {
  enabled = true
  domain  = "acctest-ds-policy.example.com"
  text    = "acctest-lookup"
}

data "unifi_dns_policy" "test" {
  policy_id = unifi_dns_txt_record.lookup.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "type", "TXT_RECORD"),
					resource.TestCheckResourceAttr(dataSourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "domain", "acctest-ds-policy.example.com"),
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "origin"),
				),
			},
		},
	})
}

func TestAccDNSPoliciesDataSource_basic(t *testing.T) {
	dataSourceName := "data.unifi_dns_policies.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "unifi_dns_txt_record" "lookup" {
  enabled = true
  domain  = "acctest-ds-policies.example.com"
  text    = "acctest-lookup"
}

data "unifi_dns_policies" "test" {
  domain = unifi_dns_txt_record.lookup.domain

  depends_on = [unifi_dns_txt_record.lookup]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "domain", "acctest-ds-policies.example.com"),
					resource.TestCheckResourceAttr(dataSourceName, "policies.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "policies.0.type", "TXT_RECORD"),
					resource.TestCheckResourceAttr(dataSourceName, "policies.0.domain", "acctest-ds-policies.example.com"),
				),
			},
		},
	})
}
```

- [ ] **Step 2: Verify both compile and skip cleanly**

Run: `go test ./internal/provider/... -run TestAccDNSPolic -v`
Expected: `--- SKIP: TestAccDNSPolicyDataSource_basic` and `--- SKIP: TestAccDNSPoliciesDataSource_basic`

- [ ] **Step 3: Commit**

```bash
git add internal/provider/data_source_dns_policy_acctest_test.go
git commit -m "test: add unifi_dns_policy and unifi_dns_policies data source acceptance tests"
```

---

### Task 10: Makefile target and README documentation

**Files:**
- Create: `Makefile`
- Modify: `README.md`

**Interfaces:**
- None (docs/tooling only; no code consumes or is consumed here).

- [ ] **Step 1: Create the Makefile**

```makefile
TEST         ?= ./...
TESTARGS     ?=
TEST_COUNT   ?= 1
TEST_TIMEOUT ?= 10m

.PHONY: build
build:
	go build ./...

.PHONY: test
test:
	go test -count $(TEST_COUNT) $(TEST) $(TESTARGS)

.PHONY: testacc
testacc:
	TF_ACC=1 go test -count $(TEST_COUNT) -timeout $(TEST_TIMEOUT) -run TestAcc -v $(TEST) $(TESTARGS)
```

- [ ] **Step 2: Document acceptance testing in README.md**

Add a new section right after the existing "Development" / debugging section (check `README.md` for the right insertion point — after the `TF_LOG_PATH` block, before `## Resources`):

```markdown
## Acceptance Testing

Acceptance tests exercise the DNS resources and data sources against a
real UniFi console. They are opt-in and skip automatically unless all of
the following are set:

```sh
export TF_ACC=1
export UNIFI_API_KEY=...      # required
export UNIFI_SITE_ID=...      # required
export UNIFI_API_URL=...      # optional, defaults to https://api.ui.com
export UNIFI_ALLOW_INSECURE=true  # optional, for self-signed local consoles
```

API keys can only be generated at [unifi.ui.com](https://unifi.ui.com)
against a real, cloud-linked UniFi console — there is no way to script
this. `docker-compose.yml` in this repo starts a local UniFi controller
for manual development convenience only (`docker compose up`, then
create an account through the web UI); it cannot produce a working
`UNIFI_API_KEY` and is not used by the acceptance tests themselves.

Run the acceptance tests with:

```sh
make testacc
```
```

- [ ] **Step 3: Verify the Makefile target works in skip mode**

Run: `make testacc TEST=./internal/provider/...`
Expected: every `TestAcc*` test reports `SKIP` (no `UNIFI_API_KEY`/`UNIFI_SITE_ID` set in this environment), overall exit code 0.

- [ ] **Step 4: Run the full existing suite one more time to confirm no regressions**

Run: `go build ./... && go vet ./... && go test ./...`
Expected: build succeeds, vet is silent, all packages report `ok`.

- [ ] **Step 5: Commit**

```bash
git add Makefile README.md
git commit -m "docs: document DNS acceptance testing workflow"
```
