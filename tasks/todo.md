# Todo

## Working Notes
- Remove `network_ids` from the managed `unifi_firewall_zone` resource.
- Keep `network_ids` on firewall zone data sources.
- Preserve existing zone assignment behavior by continuing to read current assignments during updates.

## Plan
- [ ] Remove `network_ids` from `unifi_firewall_zone` schema and state handling.
- [ ] Update resource schema tests and docs to reflect the narrower resource surface.
- [ ] Verify with `go test ./...`.

## Acceptance Criteria
- `unifi_firewall_zone` no longer exposes `network_ids`.
- `unifi_firewall_zone` updates still preserve current assignments server-side.
- Firewall zone data sources still expose `network_ids`.
- Tests pass with `go test ./...`.

## Results
- In progress.
