# Todo

## Working Notes
- Update `github.com/ilmax/unifi-client-go` from `v0.1.6` to `v0.1.7`.
- `v0.1.7` changes DNS policies from flat structs to discriminator-based union types.
- Keep the provider schema stable; only adapt the internal SDK mapping.

## Plan
- [x] Update `go.mod`/`go.sum` to `github.com/ilmax/unifi-client-go v0.1.7`.
- [x] Fix the DNS policy compile breakage introduced by the SDK update.
- [x] Verify with `go test ./...`.

## Acceptance Criteria
- Provider builds and tests cleanly against `github.com/ilmax/unifi-client-go@v0.1.7`.
- Existing provider schema stays unchanged for this SDK bump.
- Tests pass.

## Results
- Updated `github.com/ilmax/unifi-client-go` to `v0.1.7`.
- Added DNS policy union helpers so the provider can read shared DNS policy fields and build create/update payloads against the new SDK model.
- Verified with `go test ./...`.
