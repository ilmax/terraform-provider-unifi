# Todo

## Working Notes
- Split firewall zone concerns:
- `unifi_firewall_zone` manages zone lifecycle/metadata only.
- `unifi_firewall_zone_networks` manages network assignments for a zone.

## Plan
- [ ] Refactor `unifi_firewall_zone` schema/CRUD so `network_ids` is no longer configurable.
- [ ] Add `unifi_firewall_zone_networks` resource (assign one or more networks to a zone).
- [ ] Register new resource and add/adjust unit/schema tests.
- [ ] Update docs for both zone resources and provider index/README.
- [ ] Verify with `go test ./...`.

## Acceptance Criteria
- `unifi_firewall_zone` no longer takes assignment input (`network_ids`) and does not overwrite assignments on rename/update.
- `unifi_firewall_zone_networks` can manage zone assignments by `zone_id` + `network_ids`.
- Docs clearly describe separation of concerns between resources.
- Tests pass.

## Results
- In progress.
