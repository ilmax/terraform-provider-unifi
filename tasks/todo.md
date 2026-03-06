# Todo

## Working Notes
- Split firewall zone concerns:
- `unifi_firewall_zone` manages zone lifecycle/metadata only.
- `unifi_firewall_zone_networks` manages network assignments for a zone.

## Plan
- [x] Refactor `unifi_firewall_zone` schema/CRUD so `network_ids` is no longer configurable.
- [x] Add `unifi_firewall_zone_networks` resource (assign one or more networks to a zone).
- [x] Register new resource and add/adjust unit/schema tests.
- [x] Update docs for both zone resources and provider index/README.
- [x] Verify with `go test ./...`.

## Acceptance Criteria
- `unifi_firewall_zone` no longer takes assignment input (`network_ids`) and does not overwrite assignments on rename/update.
- `unifi_firewall_zone_networks` can manage zone assignments by `zone_id` + `network_ids`.
- Docs clearly describe separation of concerns between resources.
- Tests pass.

## Results
- Refactored `unifi_firewall_zone` so `network_ids` is computed-only and not a managed input.
- Updated firewall zone update flow to preserve existing assignments while allowing metadata changes (name updates).
- Added `unifi_firewall_zone_networks` resource to manage assignments by `zone_id` + `network_ids`.
- Registered the new resource and added unit/schema coverage for both resources.
- Updated docs and provider listings in `README.md` and `docs/index.md`.
- Verified with `go test ./...`.
