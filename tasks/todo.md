# Todo

## Working Notes
- Upgrade SDK to `v0.1.3` and expose new WiFi fields from release notes.

## Plan
- [x] Upgrade dependency to `github.com/ilmax/unifi-client-go@v0.1.3`.
- [x] Expose new WiFi fields (`basic_data_rate_kbps_by_frequency_ghz`, `client_filtering_policy`, `blackout_schedule_configuration`) in schema, payload, and state mapping.
- [x] Add tests and update docs.
- [x] Run go test ./... and commit.

## Acceptance Criteria
- Provider uses `github.com/ilmax/unifi-client-go@v0.1.3`.
- `unifi_wifi` exposes fields introduced in release notes where applicable.
- Tests pass.

## Results
- Upgraded dependency to `v0.1.3`.
- Exposed new WiFi fields with read/create/state support.
- Added WiFi tests and docs coverage for new fields.
- Verified with `go test ./...`.
