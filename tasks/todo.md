# Todo

## Working Notes
- Default `ipv4_configuration.prefix_length` to 24.

## Plan
- [x] Update schema/defaults and resolve logic to allow default prefix length.
- [x] Update tests and docs.
- [x] Run go test ./... and commit.

## Acceptance Criteria
- `unifi_network.ipv4_configuration.prefix_length` defaults to 24 when omitted.
- Docs updated.
- Tests pass.

## Results
- Defaulted `prefix_length` to 24 in schema and resolution logic.
- Updated docs and tests.
- Verified with `go test ./...`.
