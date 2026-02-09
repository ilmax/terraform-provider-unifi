# Todo

## Working Notes
- Rename network attribute `multicast_dns_enable` -> `multicast_dns_enabled`.

## Plan
- [x] Update network schema/model and mapping to new attribute name.
- [x] Update tests and docs.
- [x] Run go test ./... and commit.

## Acceptance Criteria
- `multicast_dns_enabled` is the only exposed Terraform attribute.
- Tests and docs updated.
- go test ./... passes.

## Results
- Renamed network attribute to `multicast_dns_enabled` across schema, state mapping, tests, and docs.
- go test ./... passes.
