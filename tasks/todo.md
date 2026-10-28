# Legacy-API temp provider (Teleport, port PoE, EtherLighting)

## Decisions (from owner)
- Separate temp provider binary under `cmd/` (working name `unifi-legacy`), own schema; official provider untouched.
- Mini SDK lives in this repo (`internal/legacy`), minimal surface, deleted per-feature as the official API covers it.
- Auth: **API key only** (`X-API-Key`). Endpoints that reject the key are out of scope.
- First slice: Teleport / WireGuard VPN config, switch port PoE / port overrides, EtherLighting colors.

## Acceptance criteria
- Each resource has CRUD against the Docker console, unit tests (model <-> API), and an env-gated acceptance test.
- Every legacy call is documented with the captured request/response in `internal/legacy/testdata/`.
- Each resource doc page states which official endpoint will replace it (removal trigger).

## Steps (status as of 2026-10-28)
1. [x] **Spike: key auth on legacy paths.** API key works on `/proxy/network/api/s/{site}/*` (bogus/no key = 401). Pitfall: judge only after checking a bogus key; a stale key looked valid once.
2. [~] **Capture traffic.** Done: port PoE (`PUT /api/s/default/rest/device/{_id}`, `port_overrides`), Teleport setting. Missing: PoE request body (network tool shows URL/method only), 24V passive value, VPN server creation UI (picker would not open), EtherLighting (no emulated device supports it). Ground truth for field names is in the controller jar (`api/fields/*.json`).
3. [x] **Mini client** `internal/legacy` (thin layer over `internal/unifi.Client`: `{meta,data}` envelope, site UUID -> name via `/api/self/sites`, `GetSetting`/`UpdateSetting`). Unit tested with httptest.
4. [~] **Resource 1: Teleport site toggle** (`unifilegacy_teleport` resource + data source, setting `{enabled, subnet_cidr}`). NOT the many-field form: that is the WireGuard/OpenVPN VPN server entry (see Findings). Awaiting owner answer on which one "Teleport" meant. Unit + acceptance tests pass (3/3 repeated runs, full `make ci-test` 75 pass, 0 fail).
5. [ ] **Resource 2: switch port PoE** (`port_overrides` read-merge-write; one resource per port). Console has an adopted USM8P to test against. Need the real UI body first (or diff device state).
6. [x] **Resource 3: EtherLighting** (`unifilegacy_ether_lighting_settings` site overrides, `unifilegacy_device_ether_lighting` per switch; acceptance tests pass 3/3). UI exposure still being worked on (see below). Original note: Site setting `ether_lighting` (network/speed defaults + overrides) is implementable now; per-device `ether_lighting` is blocked (see below).
7. [~] **Provider shell** `cmd/terraform-provider-unifi-legacy` + `internal/legacyprovider` (type name `unifilegacy`, so both providers can coexist). Done; README/dev_overrides instructions not written.
8. [ ] **Docs + CI.** No tfplugindocs for the legacy provider yet; CI already runs its acceptance test via `make ci-test`.
9. [~] **Verify:** build, unit tests, `make ci-test` from a clean console all pass. Advisor audit of the diff not done; nothing committed.

## Risks / open questions
- Legacy controller APIs may reject API keys (step 1 gates the whole plan).
- `port_overrides` is a whole-device array: concurrent writers can clobber it, so use read-merge-PUT and document it.
- Fields captured from Network 11 may differ on older firmware; pin to the console version in `UNIFI_API_VERSION`.

## Working notes
- Local main fast-forwarded to origin/main 0e2164a1; sdk v0.1.12, Network 11 test infra and tfplugindocs are already present.

## Spike result (2026-10-28) - CONFIRMED
- Valid key: legacy `/proxy/network/api/s/default/*` returns 200; bogus key and no key return 401. API-key-only is viable.
- Earlier false positive: the stale `.env` key (rejected by the integrations API) still got 200 on legacy paths. Cause unknown; do not trust a legacy 200 without also checking a bogus key. `bootstrap.sh` regenerated the key.
- `rest/setting` has `teleport` and `ether_lighting` keys (per-site settings).
- Fleet adopted with `make adopt-fleet` (2 emulated devices online).

## Capture progress
- Port PoE: captured (`internal/legacy/testdata/port_poe.md`). Write = `PUT /api/s/default/rest/device/{_id}` with `port_overrides`.
- Teleport: UI greyed out (needs an adopted gateway); setting key `teleport` is writable via `PUT /rest/setting/teleport/{id}` with the API key. Capture via API diff.
- EtherLighting: not captured yet (setting key `ether_lighting`; find the UI page).
- The in-page fetch hook caught nothing; `read_network_requests` gives URL+method but no body.
- Gateway added (ad-hoc container `uos-test-emu-gw`, UXGENT) and adopted; Teleport UI is now enabled. Teleport = `rest/setting/teleport` {enabled} only (fixture `testdata/teleport.md`); guest invites POST `/v2/api/site/default/teleport/invitation-history` returns 501 on the emulator, so out of scope.
- EtherLighting: no UI on the emulated USM8P/UXGENT/U7PRO (needs a supported switch model); setting `ether_lighting` is readable/writable via `rest/setting`.
- Gateway adoption likely dropped the firewall zones: run `make console-reset && make console-up` before the regular acceptance suite.

## Findings 2026-10-28 (emulator limits, via javap in eclipse-temurin:21-jdk on ace.jar)
- ZBF gate: adopting a gateway runs `fguutAwDwHmdVd` -> `deviceService.LRaLS(site)`; ZBF is disabled when no gateway has `supportsZoneBasedFirewall()` = `udapi_version.versionDetail["firewall/filter"] >= 4 && ["firewall/mangle"] >= 6`. unifi-emu 0.5.5 sends `udapi_version` as a plain string, so versionDetail is always empty. Fix = patch/fork unifi-emu to emit `udapi_version: {version, versionDetail:{...}}`. Tried and failed: UXGPRO, UXGENT with fw 5.1.26.33914 + fwcaps -1.
- EtherLighting: `supportsEtherLighting()` = `switch_caps & 4194304` (bit 22); per-device `ether_lighting {led_mode standard|etherlighting, mode speed|network, brightness 1-100, behavior breath|steady}`; `etherlight_caps` int for sub-capabilities. unifi-emu emits neither `switch_caps` nor `etherlight_caps` (adopted USPM24P shows nothing). Same fix: let DeviceSpec carry extra inform fields.
- Ground-truth legacy field specs live in the console jar: `api/fields/{Setting,NetworkConf,Device,PortConf,...}.json` (extracted at /tmp/claude-501/ace/y/i/api/fields). Teleport setting = `{enabled, subnet_cidr}`; Teleport VPN server entry = networkconf `purpose=remote-user-vpn`, `vpn_type=uid-server`, `uid_vpn_type=openvpn|wireguard`.
- UDMB in unifi-emu is a wireless bridge (type uap), not a gateway.

## ZBF with a gateway in the shared fleet (done)
- `restore_zbf.sh` + `make restore-zbf` write the removed `config.firewall_zone` / `config.site_feature_migration` rows back in the Network 11 Postgres (port 5434) and restart unifi.service; `ci-test` = adopt-fleet, restore-zbf, testacc. Fleet is now UXG,USM8P,U7PRO. Verified: clean `console-reset` + `make console-up` + `make ci-test` passes; `restore-zbf` re-run is a no-op.
- No emulator fork needed for ZBF. EtherLighting still needs `switch_caps` bit 22 from the emulator (open).

## EtherLighting per-device (investigated, blocked)
- Controller reads `switch_caps.feature_caps & 4194304` and `etherlight_caps` from the live (in-memory) device built from each inform. unifi-emu 0.5.5 sends no `switch_caps` (it only sends `udapi_version: {version}` and the `*_caps` ints). A Postgres trigger can make the DB row carry the bits, and `etherlight_caps` survives, but every inform replaces `switch_caps` in memory, so `supportsEtherLighting()` stays false. Needs an emulator change (add `switch_caps.feature_caps`, and `udapi_version.versionDetail`). Without forking: implement only the site-level `ether_lighting` setting resource (testable) and defer per-device `ether_lighting {led_mode,...}`.

## Pitfalls learned (add to lessons.md if a correction repeats)
- Teleport `subnet_cidr: ""` is stored by the controller and makes `GET /integration/v1/sites/{id}/networks` fail with HTTP 500 ("Could not parse []"), which broke unrelated gateway network tests. Never send an empty subnet; `null` is rejected; the key disappears on a disabled -> enabled transition.
- PUT responses on `rest/setting` echo values the controller then discards; always read back after writing.
- `make console-reset` is the only way back after experiments with device rows/restarts.

## Candidate resources (official API cannot do these; ranked by value)
1. **Client IP reservations (fixed IP)** - `rest/user` (`mac`, `name`, `note`, `use_fixedip`, `fixed_ip`, `network_id`, `local_dns_record`, `fixed_ap_*`). Official API only has read-only `clients` plus actions. Highest value, simple CRUD, no gateway needed.
2. **Port profiles** (`rest/portconf`: native/tagged VLANs, PoE mode, speed, op_mode) - pairs naturally with port PoE; port_overrides can reference them.
3. **Port forwards** (`rest/portforward`: name, proto, dst_port, fwd, fwd_port, src limiting).
4. **Static routes** (`rest/routing`) and **Dynamic DNS** (`rest/dynamicdns`).
5. **User groups / client QoS profiles** (`rest/usergroup`: name, qos_rate_max_up/down) - needed to assign rate limits.
6. **VPN servers and clients** (WireGuard/OpenVPN/L2TP as `rest/networkconf`: `purpose=remote-user-vpn`, `vpn_type`, `wireguard_*`). Largest field set; needs a gateway that advertises the feature (UI picker only offered L2TP/OpenVPN with the emulated UXG).
7. **Site settings singletons** from `rest/setting` (same pattern as Teleport, almost free once the helper exists): `ntp`, `mgmt` (SSH, LED), `country`, `locale`, `snmp`, `rsyslogd`, `usg` (gateway flags), `igmp_snooping`, `mdns`, `doh`, `global_nat`, `netflow`, `ips`, `lcm`, `ether_lighting`.
8. **AP groups / hotspot operators / firewall groups** - lower priority (traffic matching lists cover most firewall groups in the official API).

## Findings after advisor review (2026-10-28)
- **VPN server (the many-field form)** is creatable through `POST rest/networkconf`: `{name, purpose: remote-user-vpn, vpn_type: wireguard-server, enabled, local_port, ip_subnet, wireguard_interface: wan, x_wireguard_private_key}`; controller derives the public key and adds `wireguard_id`, `firewall_zone_id` (VPN zone). Probed through the API with a throwaway entry (deleted). Fixture: `internal/legacy/testdata/wireguard_server.md`. The `uid-server`/`uid_vpn_type` guess was unverified (looks like UniFi Identity, not Teleport).
- **Per-device EtherLighting is testable today**: `PUT rest/device/{_id}` with `ether_lighting {led_mode, mode, brightness, behavior}` is stored on the emulated USM8P without the capability bit. The USM8P on the current console was left with `led_mode=etherlighting` (reset restores it). Fixture: `testdata/device_ether_lighting.md`.
- `resolveSiteID` now lives in `internal/siteid` (main provider helper delegates).
- `restore_zbf.sh` now waits for every device to be ONLINE (state 1) after the restart.
- Verified: `make docs` leaves `docs/` clean; no golangci/staticcheck available or configured; two clean `console-reset` + `make ci-test` runs: 75 pass, 0 fail each.
- Still open: docs for the legacy provider (AGENTS.md requires per-resource docs; tfplugindocs only covers the main provider), README section for the legacy provider, nothing committed.

## EtherLighting in the UI (2026-10-28)
- Resources done: site `ether_lighting` setting (only the override lists are writable; defaults are controller-owned) and per-device `ether_lighting {led_mode, mode, brightness, behavior}` via `PUT rest/device/{_id}`; devices are listed from `stat/device` (`GET rest/device` is 404).
- Capability gate is `switch_caps.feature_caps & 4194304` (+ `etherlight_caps`) on the in-memory device. Every inform replaces it, so the DB trigger approach failed. Working route: stop the `uos-test-emu` container (devices go offline, no informs), set `dynamic_part.switch_caps.feature_caps=4194304` and `etherlight_caps=7` on the USM8P row in Postgres, restart `unifi.service`; the API then shows the bits and `ether_lighting` on the (offline) switch. Whether the UI renders the EtherLighting controls for an offline device is being checked next (needs a signed-in Chrome tab).

## Patched emulator result (2026-10-28)
- `test-infra/unifi/emu/{extra-fields.patch,build.sh}` + `docker-compose.extra.yml`: unifi-emu 0.5.5 plus a per-device `extra` map merged into the inform payload. Not a fork; the patch is applied to a fresh clone at build time and has its own unit test.
- Gateway with `udapi_version.versionDetail {firewall/filter:4, firewall/mangle:6}` keeps ZBF on adoption: no restore step; `make testacc` 79 pass, 0 fail on the patched fleet.
- Switch (USPM24P) with `switch_caps {vlan_caps:7, feature_caps:4194304, etherlight_caps:7}` makes the UI show an "Etherlighting" section in the device settings (Speed/Network, colors, brightness, breathing). Writing the same values the resources write (GbE FF0000, brightness 40, steady) through the legacy client was reflected in the UI after reload.
- Caveats: `etherlight_caps` lives inside `switch_caps` (top-level was ignored). Restarting only the emulator container breaks adoption (state 9); reset the console. CI still uses the stock image + restore_zbf.
- Open: publish the patched image or build it in the workflow before switching CI; upstream the patch to jamesbraid/unifi-emu.
