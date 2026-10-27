# Todo

## Working Notes
- Expose `reportFlowEnabled` (new in Network 11.0.97) on `unifi_firewall_policy` as `report_flow_enabled`.
- Official spec: optional on create/update ("defaults to true when omitted"), required in responses, and always false for the predefined Allow Return Traffic and multicast / IPv6 neighbor discovery policies. Patch ignores it for those.
- So: optional + computed, no schema default (a fixed `true` default would show a permanent diff on an imported predefined policy), `UseStateForUnknown`, and send it only when set.

## Plan
- [x] Add `ReportFlowEnabled` to the `sdkcompat` policy response model.
- [x] Add the attribute, payload and state mapping to `unifi_firewall_policy`.
- [x] Unit tests: payload (unset/unknown omitted, true/false sent), state mapping, schema shape.
- [x] Acceptance test: unset -> true, false, unset keeps false, true, import.
- [x] Update `docs/resources/firewall_policy.md`.
- [x] Verify: build, vet, unit tests, firewall policy acceptance tests on Network 11.0.97.

## Acceptance Criteria
- `report_flow_enabled` can be set to true or false and is read back.
- Leaving it out never sends a value, and never resets one the controller already stores.
- Tests cover the behavior.

## Results
- Added `report_flow_enabled` to `unifi_firewall_policy` (optional, computed, state-preserving).
- Unit tests pass and fail when the payload logic is broken; `TestAccFirewallPolicy_reportFlow` and the other firewall policy acceptance tests pass on Network 11.0.97.
- Not covered: the PATCH endpoint (the provider does not use it) and predefined policies (no fixture creates one).
