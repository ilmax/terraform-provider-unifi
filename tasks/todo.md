# Todo

## Working Notes
- Remove `network_ids` from the managed `unifi_firewall_zone` resource.
- Keep `network_ids` on firewall zone data sources.
- Preserve existing zone assignment behavior by continuing to read current assignments during updates.

## Plan
- [x] Remove `network_ids` from `unifi_firewall_zone` schema and state handling.
- [x] Update resource schema tests and docs to reflect the narrower resource surface.
- [x] Verify with `go test ./...`.

## Acceptance Criteria
- `unifi_firewall_zone` no longer exposes `network_ids`.
- `unifi_firewall_zone` updates still preserve current assignments server-side.
- Firewall zone data sources still expose `network_ids`.
- Tests pass with `go test ./...`.

## Results
- Removed `network_ids` from the `unifi_firewall_zone` resource schema and state handling.
- Kept assignment preservation in update logic by still reusing current server-side network assignments.
- Updated firewall zone resource docs and schema tests.
- Verified with `go test ./...`.
