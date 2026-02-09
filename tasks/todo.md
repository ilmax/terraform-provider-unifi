# Todo

## Working Notes
- Requests: remove IPv4 `cidr` attribute, rename mDNS field to `multicastDnsEnable`, and implement two actions (execute port/device action).
- Plugin framework upgraded to v1.16.0 for actions support.

## Plan
- [x] Upgrade terraform-plugin-framework to version with actions support.
- [x] Remove `cidr` attribute from network schema/model and update docs/tests.
- [x] Rename mDNS forwarding field to multicast DNS and update mapping/docs/tests.
- [x] Implement actions for Execute Port Action and Execute Adopted Device Action with docs/tests.
- [x] Run go test ./... and verify.

## Acceptance Criteria
- `cidr` attribute removed from provider schema, docs, and tests.
- mDNS forwarding field renamed per API docs and mapped correctly.
- Actions available for port and adopted device actions.
- Tests pass.

## Results
- Upgraded terraform-plugin-framework to v1.16.0.
- Removed IPv4 `cidr` attribute and simplified host/prefix handling.
- Renamed mDNS field to `multicast_dns_enable` with raw response mapping.
- Added `unifi_execute_port_action` and `unifi_execute_adopted_device_action` actions with docs and tests.
- go test ./... passes.
