# Todo

## Working Notes
- Require `ipv4_configuration.prefix_length` when ipv4_configuration is set.

## Plan
- [x] Remove computed default and enforce prefix_length.
- [x] Update tests and docs.
- [x] Run go test ./... and commit.

## Acceptance Criteria
- `unifi_network.ipv4_configuration.prefix_length` is required when ipv4_configuration is set.
- Tests pass.

## Results
- Enforced explicit `prefix_length` and removed defaults.
- Updated tests and docs.
- Verified with `go test ./...`.
