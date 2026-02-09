# Todo

## Working Notes
- Support `unifi_client` lookup by `client_id` or `mac_address`.

## Plan
- [x] Inspect SDK client list endpoint for MAC lookup options.
- [x] Update `unifi_client` schema and read logic to allow `client_id` or `mac_address`.
- [x] Add tests for identifier validation/normalization.
- [x] Update docs/README.
- [x] Run go test ./... and commit.

## Acceptance Criteria
- `unifi_client` supports lookup by `client_id` or `mac_address` (exactly one required).
- Docs and README updated.
- Tests pass.

## Results
- `unifi_client` supports lookup by MAC address via client list search.
- Added validation tests and updated docs/README.
- Verified with `go test ./...`.
