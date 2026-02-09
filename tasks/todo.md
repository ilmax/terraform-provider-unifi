# Todo

## Working Notes
- Change request: remove `Computed` and related `UseStateForUnknown` plan modifiers from nested network subresources so omissions in config plan changes.
- Provider currently marks nested IPv4/IPv6/DHCP objects as Optional+Computed to preserve state; reverted per user request.

## Plan
- [x] Update network schema to drop Computed from nested subresources and remove UseStateForUnknown modifiers.
- [x] Run go test ./... to verify.
- [x] Commit with semantic message.

## Acceptance Criteria
- Nested network subresources are no longer Computed.
- Plan will show drift when nested values differ but are omitted in config.
- Tests pass.

## Results
- Updated network schema to remove Computed/UseStateForUnknown on nested IPv4/IPv6/DHCP blocks.
- go test ./... passes.
