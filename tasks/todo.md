# Todo

## Working Notes
- Replace the generic `unifi_dns_policy` resource with focused resources per API DNS policy type.
- `unifi-client-go@v0.1.7` exposes 7 create/update variants:
  - `A_RECORD`
  - `AAAA_RECORD`
  - `CNAME_RECORD`
  - `FORWARD_DOMAIN`
  - `MX_RECORD`
  - `SRV_RECORD`
  - `TXT_RECORD`
- Keep resource naming aligned with the exposed Terraform type names and Go files.
- Prefer the smallest stable scope: split the managed resources first, then adapt docs/tests around them.

## Plan
- [ ] Define the Terraform resource model for each DNS policy variant and decide what happens to the generic `unifi_dns_policy` resource.
- [ ] Implement shared DNS policy CRUD helpers plus one resource per DNS policy type.
- [ ] Add schema/unit tests for each new resource and regression coverage for mapping/state handling.
- [ ] Update provider registration and remove or deprecate the generic resource as decided.
- [ ] Update documentation and examples for each DNS policy resource.
- [ ] Verify with `go test ./...`.

## Acceptance Criteria
- Each DNS policy API variant has its own Terraform resource with typed attributes.
- Resource names and Go file names match.
- The provider no longer requires users to manage DNS policy `type` manually for typed resources.
- Tests pass with `go test ./...`.

## Results
- In planning.
