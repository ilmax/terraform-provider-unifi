# UniFi Network API — Versioning, Promotion & Download Findings

_A complete record of the investigation into the UniFi Network Application API surface,
how Ubiquiti promotes capabilities into the public "Integrations API," and where each
artifact can be downloaded. Compiled 2026-10-04._

## 1. The two API surfaces

The UniFi Network Application exposes **two distinct HTTP APIs** inside a single running
console. They differ in base path, auth, and stability contract.

| Surface | Base path | Auth | In the public OpenAPI spec? | Stable |
|---|---|---|---|---|
| **Integrations API** | `/v1` | API key (`X-API-Key`) | ✅ Yes (44 paths in 10.6.106) | Yes — versioned, documented at developer.ui.com |
| **Internal API** | `/api/site/...` + `/v2` | Session cookie (username/password or SSO) | ❌ No | No — undocumented, churning, UI-only |

- The **Integrations API** is what `unifi-client-go` generates from, and what this
  Terraform provider targets. It's the public, versioned contract.
- The **Internal API** is what the Web UI drives. In the 10.6.106 build it has **384
  routes** (336 of them `/api/site/{siteName}/*`) covering QoS, SSL-inspection, OSPF,
  per-SSID radio tuning, WAN load-balancing, client-experience overrides, etc. — none of
  which are in the public spec.

`/docs/*` (5 routes) are the springdoc OpenAPI doc endpoints (`/v1/api-docs`,
`/integration/api-docs`, `/v2/api-docs`).

## 2. Spec growth across versions (the "promotion" trajectory)

The shipped `integration.json` inside each Network app build is **byte-identical** to the
`unifi-client-go` `openapi/unifi-network/<ver>.json` for the same version (verified for
10.6.106). So the Go client and the app always agree.

| Version | Paths | Ops | Schemas | What changed |
|---|---:|---:|---:|---|
| 10.1.85 | 38 | 67 | 360 | baseline |
| 10.1.89 | 38 | 67 | 360 | **byte-identical** — bugfix cut, no promotion |
| 10.3.55 | 44 | 73 | 378 | +6 paths (all `switching/*`), +21 schemas (`*Switch*`/`*Lag*`/`*McLag*`/`*SwitchStack*` DTOs) |
| 10.6.106 | 44 | 73 | 380 | +2 schemas only (small `Wifi broadcast` property additions), **no new paths** |

**Trajectory:** ~5 months, 3 spec cuts, and the *only* new capability promoted to the
public API was **read-only Switching** (LAGs, MC-LAG domains, switch stacks). Everything
else in the UI stayed on the internal surface.

## 3. How a capability gets "promoted" (from decompiled source)

The integrations API lives in the `com/ubnt/net/d/a/` package tree of the Network app.
Each promoted resource = **one sub-package** with a fixed internal shape:

```
d/a/<X>/
├── BXbf.java           ← the @RestController + @Tag(name="<Resource>")  (the controller)
├── <something>Dto.java ← the top-level @Schema(name="<Resource>") sealed/abstract DTO
├── <something>Service  ← business logic (obfuscated class name)
├── <something>Mapper   ← entity ↔ DTO (obfuscated)
└── .../
    ├── <Variant>A.java  ← @Schema(name="<Resource> <variant>") — discriminator variants
    └── <Variant>B.java  ← @Schema(name="<Resource> <variant>")
```

The OpenAPI spec is **generated from the `@Tag`/`@Schema`/`@*Mapping` annotations by
springdoc**, with `apiUrlPrefix=/integration`. So a promotion is literally: *add a
sub-package, re-cut the spec*.

### The 19 integrations sub-packages in 10.6.106

| pkg | `@Tag` | paths | added |
|---|---|---:|---|
| `a` | Access Control (ACL Rules) | 3 | 10.1.85 |
| `b` | Clients | 3 | 10.1.85 |
| `d` | UniFi Devices | 7 | 10.1.85 |
| `f` | DNS Policies | 3 | 10.1.85 |
| `h` | Firewall | 6 | 10.1.85 |
| `i` | Hotspot (vouchers) | 2 | 10.1.85 |
| `j` | Application Info | 1 | 10.1.85 |
| `k` | Networks | 4 | 10.1.85 |
| `n` | Sites | 2 | 10.1.85 |
| `o` | **Switching** | **6** | **10.3.55** ← the only new family |
| `p` | Traffic Matching Lists | 3 | 10.1.85 |
| `s` | WiFi Broadcasts | 3 | 10.1.85 |
| `c e g l q r` | Supporting Resources | 12 (scattered) | 10.1.85 |

### Key observations

1. **No feature flags.** There are no `@ConditionalOnProperty` / `@EnabledIf` / A/B
   toggles on any integrations controller. A resource either ships or it doesn't.
   Promotion is a deliberate product decision, not a gradual rollout.
2. **Most promoted paths are read-only.** Even within the 44: all of `switching/*`,
   `vpn/servers`, `radius/profiles`, `device-tags`, `networks/{id}/references`,
   `/v1/sites`, `/v1/pending-devices`, `/v1/dpi/*` are GET-only. Write ops are limited to
   the core CRUD set (networks, wifi, firewall, dns, acl, tml, vouchers, device actions).
3. **The bar is high.** The internal surface has *far* more write ops than the 13 write
   endpoints in the public spec — so promotion isn't even "add the read path."

## 4. Version coupling (ucore vs Network app)

A self-hosted **UniFi OS Server** is assembled from Debian packages, not a monolith:

| Package | Version (10.6.106-era console) | Role |
|---|---|---|
| `unifi-core` | **5.1.132** | ucore = the OS core (the `ace.jar` launcher) |
| `uos` | 5.1.5 | UOS OS tooling (`uos` CLI) |
| `unifi` | **10.6.106-36011-1** | the Network app (all the API) |
| `unifi-directory`, `ucs-agent`, `unifi-identity-update`, `unifi-assets-uosserver`, `mongodb-server`, … | — | supporting services |

- The `unifi` **package** is what changed 10.6.106 → 11.0.81.
- **ucore and the app are version-decoupled at the OS level** — the current console runs a
  10.6.106 / Java-25 Network app under a 5.1.5 ucore. But the `unifi` package `Depends`
  on `temurin-25-jre` / a 6.x ucore for 11.x, so a 11.x app needs a 6.x ucore to install
  and serve. You **cannot** jump Network 10.6 → 11.0 in place on a 5.x ucore; the ucore
  must update to 6.0 first.

## 5. Where to download each artifact

### Public, no-auth, confirmed working

| Artifact | URL pattern | How to find the exact build |
|---|---|---|
| **Self-host Network app `.deb`** | `https://dl.ui.com/unifi/<version>/unifi-uos_sysvinit.deb` | `download.svc.ui.com` catalog (10.6.106 → `dl.ui.com/unifi/10.6.106/unifi-uos_sysvinit.deb`) |
| **UniFi OS Server installer** (embeds OCI image) | `https://fw-download.ubnt.com/data/unifi-os-server/<hash>-linux-<arch>-<version>-<commit>.<ext>` | `fw-update.ubnt.com/api/firmware-latest?filter=eq~~product~~unifi-os-server&...` |
| **Network app `unifi` .deb** (uos variant) | `https://fw-download.ubnt.com/data/unifi/<hash>-uos-deb11-<arch>-<version>-<rev>-<commit>.deb` | `uos runnable latest-versions unifi` (from inside a console) |
| **AP/switch/gateway firmware** | `https://dl.ubnt.com/unifi/firmware/<MODEL>/<ver>/<file>` | the device's own inform / the console's firmware.json |

### The public "version catalog" APIs (the real source of truth)

These are the no-auth endpoints Ubiquiti's own download page and the community
`install.sh`/extraction scripts use:

1. **`https://download.svc.ui.com/v1/downloads/products/slugs/<slug>`** — lists a product's
   downloads with exact file URLs.
   - Slugs: `unifi-os-server`, `unifi-network-application-<ver>-<platform>`.
   - `unifi-os-server` → newest is **5.1.42** (Linux x64/arm64, macOS, Windows).
2. **`https://download.svc.ui.com/v1/downloads?product=unifi`** — full catalog (3698
   entries, 37 pages). Newest self-host Network app `.deb` = **10.6.106**.
3. **`https://fw-update.ubnt.com/api/firmware-latest?filter=eq~~product~~<product>&filter=eq~~platform~~<platform>&filter=eq~~channel~~<channel>`** — the
   firmware-update API devices query. `unifi` product across **all** channels
   (release/RC/beta/alpha/qa) tops out at **10.6.106**.
4. **`https://community.svc.ui.com/graphql`** — `release(id: "...")` query (needs
   `Content-Type: application/json` to pass CSRF). Exposes release metadata but **hides
   self-host download links for 11.x** (`links: []`, `content: null`).

## 6. The 11.0.81 / UOS 6.0 wall (why we can't get it yet)

11.0.81 requires **UOS 6.0**. Verified from every public angle that **neither is
downloadable for self-hosting** as of 2026-10-04:

- **Latest UniFi OS Server** = **5.1.42**. There is **no 6.0.x OS Server** in the public
  catalog — UOS 6.0.10/6.0.11 is published **only for physical consoles** (UDR/UDR7/UNVR
  etc. as `.bin`). Community images top out: hieutq **5.1.42**, unihosted **5.1.40**.
- **Latest self-host Network app** = **10.6.106** (across all firmware channels). No 11.x
  `.deb` is published; the 11.0.81 community release has its links stripped.
- The 5.1.132 ucore's update service (`uos runnable check-version unifi 11.0.81`) returns
  `{}` — it won't serve 11.x.

**Consequence:** a self-hosted 5.1 console (the only publicly-buildable self-host) **cannot
self-update to 11.0** — it needs a 6.x ucore first, and that's not published for self-host.
The in-UI-updater on a 5.1 console offers 10.x Network at best.

### To actually get a 6.0 / 11.0 console, one of:

1. **Wait for a community 6.0 image** (the chosen path) — hieutq or unihosted will add a
   6.0.x tag once Ubiquiti ships a self-host 6.0 build. A recurring watch is set up
   (see §7).
2. **Ubiquiti support ticket** (support@ui.com) requesting the UniFi OS Server 6.0.x Linux
   installer for self-hosting.
3. **11.0.81 self-host `.deb`** via support → extract `ace.jar` → `internal-dependencies.jar`,
   diff the `com/ubnt/net/d/a/*` tree + `webapps/ROOT/api-docs/integration.json` vs
   10.6.106. **No console boot needed** to answer "what new integrations API endpoints."
4. **A cloud (ui.com) console** at 6.0/11.0.81 with an API key → diff its live
   `/integration/api-docs` vs 10.6.106.

## 7. Current state & open work

- **Running console** (Docker): `unifi-os-server` (hieutq:latest = UOS 5.1.42, ucore
  5.1.132, Network 10.6.106) + `unifi-emu`, up at `https://localhost:11443`. Existing
  admin: `admin` / `unifi-test-password-1` (persisted in the volume).
- **Done:** a recurring watch for a new public API surface —
  `tools/unifi-network-api-compare/watch/` (cron-able). Its reliable signal polls
  `unifi-client-go`'s `unifi-network-version` marker + newest `openapi/unifi-network/*.json`
  spec; bonus signals (best-effort) watch the `download.svc.ui.com` catalog +
  `unihosted/unifi-os-server-docker` for a 6.0/11.x artifact.
- **Reusable tooling** (in `tools/unifi-network-api-compare/`): `compare_network_api.py`
  diffs a new build vs the 10.6.106 baseline on spec (exact), routes, and promoted
  packages; the `baseline-10.6.106/` dir holds the 10.6.106 `integration.json` spec +
  `all-routes.json` (433 routes). See its README for the spec-only (fast) vs full
  (decompile) modes.

## 8. Bottom line for the Terraform provider

The provider's resource set is well-aligned with what Ubiquiti is actually willing to
expose. The 44-path ceiling on the Integrations API is a **deliberate product choice**
(promote a handful of stable, high-value resources; keep everything else internal), not a
spec bug. Don't expect the public spec to grow much. If a UI feature the spec lacks is
needed, the realistic levers are: wait for promotion, request promotion, or (breaking the
api-key-only design + accepting firmware-breakage risk) build a second client for the
internal `/api/site/*` API.
