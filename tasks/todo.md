# Todo

## Working Notes
- Rename ACL resource to `unifi_firewall_rule` and add firewall zone data sources.

## Plan
- [x] Rename resource type to `unifi_firewall_rule` and remove `unifi_firewall`.
- [x] Add `unifi_firewall_zone` and `unifi_firewall_zones` data sources.
- [x] Update tests and docs.
- [x] Run go test ./... and commit.

## Acceptance Criteria
- Provider exposes `unifi_firewall_rule` and no longer exposes `unifi_firewall`.
- Firewall zone data sources are available for single and list lookups.
- Tests pass.

## Results
- Resource renamed to `unifi_firewall_rule`.
- Added `unifi_firewall_zone` and `unifi_firewall_zones` data sources.
- Updated docs and schema tests for new names and data sources.
- Verified with `go test ./...`.
