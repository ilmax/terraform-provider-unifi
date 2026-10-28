# Local UniFi OS Server for acceptance testing

Runs a genuine, locally-administered UniFi OS Server (no cloud account) with
UniFi Network 11, an admin account, an API key and a small emulated device
fleet, started by one command and with no manual steps. The Network
Integrations API this provider targets lives on a UniFi OS console; the legacy
standalone Network Application does not have it.

## Quick start

```sh
make console-up    # first run builds the image (about 800 MB download), then ~3 min
make testacc       # or: make ci-test, which also adopts the device fleet first
make console-reset # discard the console and its credentials
```

`make console-up` does, in order:

1. **Builds the image** (`image/build.sh`, once). UniFi OS Server only ships as
   a self-extracting installer that wraps a systemd-based OCI image and drives
   it with Podman. `image/Dockerfile` pulls that image out of the installer
   and flattens it (`image/flatten-oci.sh`), so it runs under plain Docker with
   Ubiquiti's own entrypoint. UniFi Network 11 needs UniFi OS 6, which no
   community image ships yet.
2. **Starts the console** (`docker-compose.yml`) and waits until it is healthy.
3. **Runs `bootstrap.sh`**, which
   - creates the admin account (`POST /api/setup`; no cloud account),
   - installs the Network application (`install-network.sh`),
   - creates an API key (`POST /proxy/users/api/v2/user/{id}/keys`, the call the
     web UI makes on Integrations > Create New API Key) and
   - writes `UOS_CONSOLE_PASSWORD` and `UNIFI_API_KEY` to the git-ignored
     `.env`.

   It is idempotent: it does nothing when `.env` already holds a key the console
   accepts. It does one login (to create the key) and uses the API key for
   everything else, because UniFi OS rate-limits `POST /api/auth/login`.
4. **Starts the emulated fleet** (`unifi-emu`). `make adopt-fleet` adopts it.

## Versions

Both versions are set by environment variables:

| Variable | Used by | Default |
|---|---|---|
| `NETWORK_VERSION` | `install-network.sh` | the `UNIFI_API_VERSION` file at the repo root |
| `UOS_VERSION`, `UOS_INSTALLER_URL` | `image/build.sh` | `6.0.19` and its installer URL |
| `UOS_IMAGE` | `build.sh`, `docker-compose.yml` | `unifi-os-server-test:dev` |
| `UNIFI_EMU_VERSION` | `docker-compose.yml` | `0.5.5` |

- **Network 11 is not in the OS updater's catalog yet**, so
  `uos runnable install unifi --version 11.0.97` fails with "Required package
  was not found". `install-network.sh` looks the `.deb` up in Ubiquiti's firmware
  API (`fw-update.ubnt.com`) for the container's architecture and installs it by
  URL. `NETWORK_URL` skips the lookup; `NETWORK_CHANNEL` picks the channel.
- **Ubiquiti's installer URLs contain a hash**, so an OS version alone cannot be
  resolved. A different `UOS_VERSION` needs its `UOS_INSTALLER_URL`.
- **UniFi OS Server 6.0.19 is an Early Access build.** It is linked from the
  community release page and is not on the public release channel yet; it bundles
  Network 10.6.106, and Network 11 is installed on top.

## Running the acceptance suite

```sh
make testacc
```

This reads `UNIFI_API_KEY` from `.env`, resolves the console's real site UUID
automatically, and runs everything under `internal/acceptance/` — the package
all acceptance tests live in, separate from the unit tests in
`internal/provider/`.

To run `go test` directly instead:

```sh
export TF_ACC=1
export UNIFI_API_KEY="$(grep UNIFI_API_KEY test-infra/unifi/.env | cut -d= -f2)"
export UNIFI_API_URL="https://localhost:11443/proxy/network/integrations"
export UNIFI_SITE_ID="<the real site UUID, from GET /v1/sites>"
export UNIFI_ALLOW_INSECURE=true
```

**`UNIFI_SITE_ID` must be the real UUID, not `"default"`.** The console shows
the default site as "Default" and its `internalReference` is the string
`"default"`, but the Integrations API rejects that literal as a `siteId` path
parameter (`400 'default' is not a valid 'siteId' value`). Get it with:

```sh
curl -sk -H "X-API-Key: $UNIFI_API_KEY" \
  https://localhost:11443/proxy/network/integrations/v1/sites
```

`make testacc` does this lookup for you.

## CI

`.github/workflows/acctest.yml` boots a fresh console on a GitHub-hosted runner
for every run (`make console-up`, then `make ci-test`). Nothing is
long-lived, so there are no secrets to maintain and no self-hosted runner.

Only arm64 (Apple silicon) has been exercised locally. The amd64 installer URL
resolves and an amd64 Network package exists, but the first run on a GitHub
runner is the first run on amd64: try it with the workflow's `workflow_dispatch`
trigger.

## Known issues

- **Your old console is untouched.** This setup uses its own Compose project
  (`unifi-os-test`), container names (`uos-test`, `uos-test-emu`) and volumes, so
  it does not reuse or remove a console that an earlier version of this directory
  set up.
- **Browser automation and the self-signed cert**: Chrome's native SSL
  interstitial cannot be scripted by browser extensions (by design) — a human
  has to click through "Advanced → Proceed" once per fresh certificate.
- **Stale frontend session after a reset**: if a browser tab from a previous
  console is still open after `make console-reset`, "Create New API Key" in the UI
  fails with `403`. Reload the page. The scripted setup is not affected.

## Adopted device fleet

A bare console has no adopted devices, which blocks GATEWAY-managed networks
and everything gated on Zone Based Firewall. There's no API/UI way around
that — but the `unifi-emu` service runs
[`jamesbraid/unifi-emu`](https://github.com/jamesbraid/unifi-emu), which
speaks the real UniFi "inform" protocol well enough to appear as a whole
fleet of adoptable devices: one switch and one AP (see
`docker-compose.yml`'s `SIM_MODELS`). It replaced an earlier one-device
emulator (`amd989/unifi-gateway`) that only covered a decade-old USG3;
`unifi-emu` covers the current device lineup and is purpose-built for
exactly this "give a test controller real, adoptable devices" job — its own
docs name this provider as one of the projects it exists for.

`make console-up` starts it; adopt everything it reports as pending with:

```sh
make adopt-fleet   # from the repo root; needs test-infra/unifi/.env
```

**The fleet has a gateway, and ZBF has to be restored after adopting it.** On
Network 11 a fresh console already has the six default firewall zones. Adopting
the emulated gateway makes Network 11 drop them (`Zone-Based Firewall is being
disabled for Site[..] because it is not supported` in
`/data/unifi/logs/server.log`), after which every zone call answers `Zone Based
Firewall is not configured`. The cause, found by disassembling `ace.jar` with
`javap`: a gateway counts as ZBF-capable only when its inform payload carries
`udapi_version.versionDetail["firewall/filter"] >= 4` and
`["firewall/mangle"] >= 6`, and `unifi-emu` sends `udapi_version` as a plain
string, so no emulated gateway can pass (`fwcaps`, firmware version and model
choice make no difference). `make restore-zbf` (`restore_zbf.sh`, run by
`ci-test` right after `adopt-fleet`) writes back what the controller removed
in the console's Postgres (`unifi-network` database, port 5434): the six
`config.firewall_zone` rows and the `config.site_feature_migration`
`ZONE_BASED_FIREWALL` row, then restarts the Network application to clear its
zone cache. It is idempotent. If `unifi-emu` ever emits `versionDetail`, the
script becomes a no-op and can be deleted.

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

> **Network 11:** nothing here is needed. A fresh console has the default zones,
> and `make enable-zbf` cannot run there, because it writes into MongoDB and
> Network 11 moves its configuration out of MongoDB at startup. `make ci-test`
> therefore no longer runs it. The rest of this section describes UniFi Network
> 10.x consoles.

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
