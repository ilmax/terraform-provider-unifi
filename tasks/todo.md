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
- [x] Define the Terraform resource model for each DNS policy variant and decide what happens to the generic `unifi_dns_policy` resource.
- [x] Implement shared DNS policy CRUD helpers plus one resource per DNS policy type.
- [x] Add schema/unit tests for each new resource and regression coverage for mapping/state handling.
- [x] Update provider registration and remove or deprecate the generic resource as decided.
- [x] Update documentation and examples for each DNS policy resource.
- [x] Verify with `go test ./...`.

## Acceptance Criteria
- Each DNS policy API variant has its own Terraform resource with typed attributes.
- Resource names and Go file names match.
- The provider no longer requires users to manage DNS policy `type` manually for typed resources.
- Tests pass with `go test ./...`.

## Results
- Added shared typed DNS policy helper code and the first typed resources for `A_RECORD`, `AAAA_RECORD`, and `CNAME_RECORD`.
- Added unit coverage for the new record payload/state mapping.
- Added the remaining typed resources for `FORWARD_DOMAIN`, `MX_RECORD`, `SRV_RECORD`, and `TXT_RECORD`.
- Removed the generic `unifi_dns_policy` resource from the provider and deleted its implementation/tests.
- Added one resource document per DNS policy type and updated the provider index/README resource lists.
