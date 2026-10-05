# Local UniFi OS Server for acceptance testing

Runs a genuine, locally-administered UniFi OS Server (no cloud account) via
[`hieutq/unifi-os-server`](https://github.com/toquanghieu/unifi-os-server-docker),
which flattens the official installer's extracted appliance rootfs into a
single Docker image and runs systemd directly as PID 1 — no nested
Podman-in-Docker layer. This is what actually has the Network Integrations
API (`/network/default/integrations` → "Create New API Key"); the legacy
standalone Network Application does not.

## Quick start

```sh
docker compose up -d
```

Then, **once** (state persists in the `unifi_data` volume across restarts
and full container recreation — already verified):

1. Open `https://localhost:11443` (accept the self-signed cert warning).
2. Name the server, click **Next**.
3. Click **Proceed Without a UI Account** → **Continue Anyway** (skips
   cloud account creation entirely — stays fully local).
4. Set a console password, agree to terms, **Finish**. Save the password
   into `.env` as `UOS_CONSOLE_PASSWORD` (copy `.env.example` first).
5. Go to `https://localhost:11443/network/default/integrations` →
   **Create New API Key**. Save the shown key into `.env` as
   `UNIFI_API_KEY` — **it is shown only once**. Use `pbpaste` after
   clicking "Copy Key" rather than reading it off the screen: this key's
   font renders capital `I` and lowercase `l` identically.

The Network application version is pinned automatically post-boot by
`fix-model-and-start.sh` (default `10.6.106`, override via `NETWORK_VERSION`
env var / `.env`) — no manual step needed.

## Running the acceptance suite

```sh
make testacc
```

This reads `UNIFI_API_KEY` from `.env`, resolves the console's real site
UUID automatically (see "Known issues" below for why that matters), and
runs everything under `internal/acceptance/` — the package all acceptance
tests live in, separate from the unit tests in `internal/provider/`.

To wire up a provider config or run `go test` directly instead:

```sh
export TF_ACC=1
export UNIFI_API_KEY="$(grep UNIFI_API_KEY .env | cut -d= -f2)"
export UNIFI_API_URL="https://localhost:11443/proxy/network/integrations"
export UNIFI_SITE_ID="<the real site UUID, e.g. from GET /v1/sites — see below>"
export UNIFI_ALLOW_INSECURE=true
```

Note the `/proxy/network/integrations` suffix on `UNIFI_API_URL` — this
console's Integrations API lives at that path, not bare `/v1/...`.

**`UNIFI_SITE_ID` must be the real UUID, not `"default"`.** The console
displays the default site as "Default" and its `internalReference` is the
string `"default"`, but the Integrations API rejects that literal string as
a `siteId` path parameter (`400 'default' is not a valid 'siteId' value`).
Get the real UUID with:

```sh
curl -sk -H "X-API-Key: $UNIFI_API_KEY" \
  https://localhost:11443/proxy/network/integrations/v1/sites
```

`make testacc` does this lookup for you automatically.

## CI/CD reproducibility

`.github/workflows/acctest.yml` runs the full suite on every push/PR. It
targets **one long-lived console**, not a fresh one per run: a GitHub-hosted
runner starts a clean VM every job with nothing listening on
`localhost:11443`, and fully automating this image's own onboarding wizard
(name the server, skip cloud account, set a password, click "Create New API
Key") was deliberately not pursued — see "Known issues" below for exactly
what's manual and why. Instead:

1. **One-time, by a human:** bring the console up (`docker compose up -d`)
   somewhere a self-hosted runner can reach it, complete the manual setup in
   "Quick start" above, and save `UOS_CONSOLE_PASSWORD`/`UNIFI_API_KEY` as
   the `UNIFI_TEST_CONSOLE_PASSWORD`/`UNIFI_TEST_API_KEY` repository secrets
   the workflow reads. Point the workflow's `runs-on` at that runner (it
   ships pointing at a placeholder `[self-hosted, unifi-test-console]`
   label — rename it to match).
2. **Every CI run, fully scripted:** `make ci-test` — `adopt-fleet` (adopts
   whatever the `unifi-emu` fleet reports pending), `enable-zbf`
   (force-enables Zone Based Firewall; see below), then `testacc`. All three
   steps are idempotent: re-running them against a console that's already in
   the target state is a fast no-op, so the same long-lived console can run
   this on every job indefinitely without manual intervention.

This means the console's local state (which devices are adopted, whether
Zone Based Firewall is enabled, `.env`'s credentials) is exactly the kind of
long-lived fixture state a real network appliance test suite has to live
with — CI reproducibility here means *idempotent setup steps against
persistent state*, not *stateless from-scratch boot*.

## Known issues in this image (already worked around here)

- **Missing `/usr/lib/app_model` / `/usr/lib/product_name`**: the image's
  own `/entrypoint.sh` never writes these, so `unifi-core` crash-loops with
  `Unsupported console model: ""` (it shells out to `/sbin/ubnt-tools`,
  which reads those two files). Fixed by `fix-model-and-start.sh`, which
  writes them before handing off to the real entrypoint.
- **Browser automation and the self-signed cert**: Chrome's native SSL
  interstitial cannot be scripted by browser extensions (by design) — a
  human has to click through "Advanced → Proceed" once per fresh
  certificate. A real CI browser tool (Playwright/Puppeteer with
  `ignoreHTTPSErrors: true`) does not have this problem.
- **Stale frontend session after a volume wipe**: if you tear down with
  `docker compose down -v` while a browser tab from a previous instance is
  still open, "Create New API Key" fails with `403` (the tab's cached user
  ID no longer matches the fresh console's admin user). Reload the page
  (or open a new tab) after any full volume reset.

## Adopted device fleet

A bare console has no adopted devices, which blocks GATEWAY-managed networks
and everything gated on Zone Based Firewall. There's no API/UI way around
that — but the `unifi-emu` service runs
[`jamesbraid/unifi-emu`](https://github.com/jamesbraid/unifi-emu), which
speaks the real UniFi "inform" protocol well enough to appear as a whole
fleet of adoptable devices: one gateway, one switch, one AP (see
`docker-compose.yml`'s `SIM_MODELS`). It replaced an earlier one-device
emulator (`amd989/unifi-gateway`) that only covered a decade-old USG3;
`unifi-emu` covers the current device lineup and is purpose-built for
exactly this "give a test controller real, adoptable devices" job — its own
docs name this provider as one of the projects it exists for.

Bring it up and adopt everything it reports as pending, in one step:

```sh
docker compose up -d unifi-emu
make adopt-fleet   # from the repo root; needs test-infra/unifi/.env
```

`adopt-fleet` (`adopt_fleet.py`) waits for pending devices, then adopts each
one with nothing but the API key, via the modern
`POST /v1/sites/{siteId}/devices` (`adoptDevice`) endpoint — no admin
username/password, no UI click, no CLI dance. All three devices typically
reach `ONLINE` within about a minute. `adopt-fleet` is safe to re-run: it's a
no-op once everything is already adopted.

**Removing a device.** The real API's `DELETE
/v1/sites/{siteId}/devices/{deviceId}` only completes once the device itself
acknowledges a "forget" command over its next inform — exactly like
adoption, it's a live round-trip, not a database delete. If you stop the
`unifi-emu` container (or reconfigure `SIM_MODELS`) while a device is
mid-delete, it gets stuck showing `state: DELETING` forever, because nothing
is left informing to receive the forget command. If that happens, temporarily
run a single `unifi-emu` container reporting that same MAC again (`docker run
--rm --network <network> ghcr.io/jamesbraid/unifi-emu:latest -inform
http://unifi-os-server:8080/inform -mac <mac> -model <model>`) until the
device disappears from `GET /v1/sites/{siteId}/devices`, then stop it.

## Zone Based Firewall

Real Zone Based Firewall CRUD — `unifi_firewall_zone`,
`unifi_firewall_zone_networks`, `unifi_firewall_policy`,
`unifi_firewall_policy_ordering`, and the zone data sources — is force-enabled
on this console, via a direct database write rather than any device or API
trick:

```sh
make enable-zbf
```

**Why a database write, and not just a more capable emulated gateway.** Zone
Based Firewall requires a newer gateway model (UDM/UDR/UXG-class), so the
first approach tried was giving the `unifi-emu` fleet a `UXG`/`UXGENT`
("Gateway Lite"/"Gateway Enterprise") gateway instead of a classic one. That
was investigated in real depth, not just tried once, and ruled out
conclusively:

- `unifi-emu`'s `model_profiles.json` has a real, hardware-captured
  firmware-capability bitmap (`fw_caps`) only for the classic `ugw` line
  (`UGW3`/`UGW4`); every `uxg` model falls back to a placeholder, and `UXG`
  specifically has no `udapi_version` at all, so the controller skips its
  entire capability-negotiation pass regardless of `fw_caps` (`Skip updating
  capability for device [..] due to empty udapi_version`).
- Switching to `UXGENT` (which does have `udapi_version`) with `unifi-emu`'s
  per-device `fwcaps: -1` override (every bit set — added by its maintainer
  for exactly this kind of experiment) was tested directly: every `UDAPI
  feature [..] is not supported by firmware` warning disappeared from the
  controller's log — full capability negotiation succeeded — and Zone Based
  Firewall was **still** reported as not configured.
- `unifi-emu`'s own reverse-engineered `capability_bits.json` (bit names
  decompiled from a real Network application `.jar`) confirms why: none of
  `fw_caps`, `udapi_caps`, `switch_caps`, or `hw_caps` — the only four
  bitmap fields it sends on the wire at all — contain anything resembling a
  zone/firewall bit.

That ruled out every device-capability lever. Decompiling
`/unifi/app/lib/ace.jar` **directly from the running console** (`docker cp`
it out, then [CFR](https://github.com/leibnitz27/cfr) against the
Spring-Boot-nested `BOOT-INF/lib/internal-dependencies.jar`) found the real
gate: a Spring `HandlerInterceptor`
(`com.ubnt.net.d.a.h.BXbf.preHandle`) checks
`SiteFeatureMigrationService.hasFeature(siteId, ZONE_BASED_FIREWALL)` — a
**per-site feature flag**, not a device check at all, stored as a MongoDB
document (`ace.site_feature_migration`: `{site_id, feature, timestamp}`;
console mongo lives on `127.0.0.1:27117`, database `ace`, reachable only via
`docker exec`, not a published port). Setting that flag alone isn't quite
enough: the real migration this stands in for also bootstraps one
system-default zone per `zone_key` (`internal`/`external`/`gateway`/`vpn`/
`hotspot`/`dmz`/`mgmt` → Internal/External/Gateway/VPN/Hotspot/DMZ/Management,
verified against `com.ubnt.service.firewallzone.rCTEZFVxHHljb`) before
marking the site migrated — skip that and the first policy create throws
`Could not find hotspot firewall zone`, since policy creation unconditionally
looks up the site's Hotspot zone for network-purpose bookkeeping. `enable_zbf.sh`
does both steps, then restarts the Network application (`systemctl restart
unifi.service`) to clear its 1-hour zone-list cache
(`com.ubnt.service.firewallzone.YfrfkymPjtCnsw`'s Caffeine cache, only ever
invalidated by the service layer this bypasses). It's idempotent — see
"CI/CD reproducibility" above.

**Two real, unrelated bugs surfaced once this was actually testable** — Zone
Based Firewall had never been reachable before, so this code had never run
against a real API:

- `unifi_firewall_policy`'s `ip_protocol_scope` sent a bare string; the real
  API wants `{"ipVersion": "..."}` (a discriminated object, confirmed against
  the OpenAPI spec's `Firewall policy IP protocol scope` schema). Fixed.
- `unifi_firewall_zone_networks`'s `Delete` only forgot the resource in
  Terraform state, never actually unassigning its networks via the API. A
  zone whose `networkIds` still named a network Terraform went on to delete
  next (the common case, since a network usually depends on its zone) then
  refused to delete itself (`Configured network does not exist`). Fixed:
  `Delete` now unassigns every network first.

**Still not unlocked by any of this:** `unifi_client` (needs a real connected
wireless/wired client, which nothing here emulates) and
`unifi_switch_lag`/`unifi_switch_mc_lag_domain`/`unifi_switch_stack` (need a
LAG/stack/MC-LAG actually configured on real switch hardware — the API has no
way to create one, and a freshly adopted switch has none by default).
`unifi_device`/`unifi_wan` have real data to read (gateway, switch, and AP all
show up for `unifi_device`). `unifi_traffic_matching_list` was always a
standalone endpoint (not nested under `/firewall/`), unaffected by any of
this either way.

## GET-only endpoints exposed as data sources

Several real API endpoints have no create/update/delete — they're read-only
reference or inventory data, so they're implemented as data sources instead
of resources: `unifi_vpn_server`, `unifi_vpn_site_to_site_tunnel`,
`unifi_radius_profile`, `unifi_device_tag`, `unifi_pending_devices`,
`unifi_dpi_application`, `unifi_dpi_category`, `unifi_switch_lag`,
`unifi_switch_mc_lag_domain`, and `unifi_switch_stack`. On this console:

- `unifi_radius_profile` and the two `unifi_dpi_*` data sources have real
  data to read with no setup — every site ships a SYSTEM_DEFINED "Default"
  RADIUS profile, and DPI application/category IDs are global reference
  data (2000+ entries) unrelated to any adopted hardware.
- `unifi_vpn_server`, `unifi_vpn_site_to_site_tunnel`, and
  `unifi_device_tag` have no UI/API path to populate on a bare console, so
  only their not-found error path is tested here.
- `unifi_pending_devices` always returns `200` with a (possibly empty)
  list — its acceptance test asserts the empty-list case, which is this
  console's steady state once the fleet is adopted (see above).
- `unifi_switch_lag`/`unifi_switch_mc_lag_domain`/`unifi_switch_stack` need a
  LAG/MC-LAG/stack actually *configured*, not just an adopted switch — a
  freshly adopted switch has none by default, and the real API has no way
  to create one (`/switching/lags` etc. are GET-only). Only their not-found
  path is tested here, even with a real switch in the fleet.

## Pinning versions

- **Network application**: `NETWORK_VERSION` in `.env` / compose
  environment (default `10.6.106`, matching `UNIFI_API_VERSION` at the
  repo root). Installed post-boot via the real `uos runnable install unifi
  --version <version>` CLI — confirmed to survive both a simple restart
  and a full `docker compose down && up` (volume-only persistence).
- **UniFi OS Server** itself: pin the image tag, e.g.
  `hieutq/unifi-os-server:5.1.42` instead of `:latest`, in
  `docker-compose.yml`.
