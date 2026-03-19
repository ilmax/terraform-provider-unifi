# Todo

## Working Notes
- Change `unifi_firewall_zone` data source lookup from zone ID to zone name.
- Keep `zone_id` as a computed output so callers can still use the identifier downstream.
- Return a clear error when no zone matches or multiple zones share the same name.

## Plan
- [x] Update `unifi_firewall_zone` data source schema and read logic to look up by `name`.
- [x] Add/update tests and docs for the new lookup behavior.
- [x] Verify with `go test ./...`.

## Acceptance Criteria
- `unifi_firewall_zone` data source requires `name` instead of `zone_id`.
- The data source returns `zone_id` as a computed output.
- Duplicate or missing zone names produce clear diagnostics.
- Tests pass with `go test ./...`.

## Results
- Updated `unifi_firewall_zone` to look up zones by name using the list endpoint.
- Kept `zone_id` as a computed output for downstream use.
- Added case-insensitive name matching and duplicate-name protection.
- Updated schema tests and data source docs.
- Verified with `go test ./...`.
