# Todo

## Working Notes
- Goal: move the provider and its acceptance tests to UniFi Network 11 (`11.0.97`). Network 11 requires UniFi OS 6, so the test console has to move too.
- `unifi-client-go` v0.1.12 is generated from API 11.0.97. Compared with v0.1.11 (10.6.106) no endpoints changed; only `reportFlowEnabled` (firewall policy) and `txPowerMode` (radio overview) were added.
- Network 11 is not in the OS updater's catalog yet: `uos runnable install unifi --version 11.0.97` fails with "Required package was not found". It installs with `--direct-url` using the `.deb` URL from `fw-update.ubnt.com` (per architecture).
- UniFi OS Server 6.0.19 is Early Access. Its installer is not on the public release channel; it is linked from the community release page.
- Headless setup works: `POST /api/setup` creates the admin without a cloud account, and `POST /proxy/users/api/v2/user/{id}/keys` creates an API key. This contradicts the old note in AGENTS.md ("no scriptable way to provision a key") and the README ("deliberately not pursued").
- Network 11 has the zone-based firewall by default, so `enable_zbf.sh` should not be needed there (verify).
- UniFi OS rate-limits `POST /api/auth/login`; use the API key for everything except the one login that creates the key.
- Open question to verify: the integrations path (`/proxy/network/integrations` here vs `/proxy/network/integration` seen on 11.0.97).

## Plan
- [x] Bump `unifi-client-go` to v0.1.12 (builds, unit tests pass).
- [x] Bump `UNIFI_API_VERSION` to 11.0.97 and the AGENTS.md reference links.
- [x] Test console on UniFi OS 6 + Network 11: image recipe, Network install by URL, automated admin + API key, fleet.
- [x] Run `internal/acceptance` against Network 11 and triage failures.
- [x] Fix provider issues the run finds: none found in provider code; the failures were test-environment assumptions (see Results).
- [x] CI: boot a fresh console per run on a GitHub-hosted runner (not yet run on GitHub; amd64 unverified).
- [x] Update README/AGENTS.md where they describe the console as manual and long-lived.
- [x] Verify: `go build ./...`, `go vet ./...`, `go test ./...`, full `make console-reset console-up ci-test` from a cold start.

## Acceptance Criteria
- `make console-up` produces a ready console (admin, API key, Network 11.0.97, adopted fleet) with no manual steps.
- `make testacc` passes (or skips with an explicit reason) against Network 11.0.97 from a cold start.
- Network and OS versions are set by environment variables.
- Every behavior change is covered by a test.

## Results
- `unifi-client-go` v0.1.12 and `UNIFI_API_VERSION` 11.0.97; no provider code changes were needed to build.
- `make console-up` builds a UniFi OS 6.0.19 image from Ubiquiti's installer, installs Network 11.0.97 by URL, creates the admin and an API key, and starts the fleet. `NETWORK_VERSION` (default `UNIFI_API_VERSION`), `UOS_VERSION` and `UOS_INSTALLER_URL` set the versions.
- Acceptance suite on Network 11.0.97: 78 pass, 0 fail, 10 skip (6 "requires ZBF" error-path guards that skip when ZBF is on; 4 that need an adopted gateway).
- The 14 first-run failures were not provider bugs: adopting the emulated gateway makes Network 11 drop the site's firewall zones. The fleet is now a switch and an AP; `enable_zbf.sh` is not needed on 11 and is no longer in `ci-test`.
- Test fixes: `hasAdoptedGateway` now detects a gateway (it counted any device); gateway-dependent tests skip without one; the zone count assertion is "at least 6".
- Not verified: the GitHub-hosted amd64 run (try `workflow_dispatch`), and Network 10.x against the new console (the old console remains in place).
- Follow-ups: exercise the new API fields (`reportFlowEnabled`, `txPowerMode`) once the provider exposes them; find an emulated gateway that keeps ZBF so the four gateway tests can run.
