# Todo

## Working Notes
- Fix prefix_length default to avoid invalid planned values.

## Plan
- [x] Replace plan modifier with computed default for prefix_length.
- [x] Remove unused helper/default plan modifier.
- [x] Run go test ./... and commit.

## Acceptance Criteria
- `unifi_network.ipv4_configuration.prefix_length` defaults to 24 when omitted.
- No invalid plan errors from Terraform.
- Tests pass.

## Results
- Defaulted `prefix_length` via computed default to avoid invalid plan errors.
- Removed unused plan modifier helper.
- Verified with `go test ./...`.
