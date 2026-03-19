# Todo

## Working Notes
- Upgrade `github.com/ilmax/unifi-client-go` from `v0.1.6` to `v0.1.7`.
- Fix only the breakage introduced by `v0.1.7` and keep the provider behavior stable otherwise.

## Plan
- [ ] Update `go.mod`/`go.sum` to `github.com/ilmax/unifi-client-go v0.1.7`.
- [ ] Fix any compile or test failures introduced by the SDK update.
- [ ] Verify with `go test ./...`.

## Acceptance Criteria
- Provider builds and tests cleanly against `github.com/ilmax/unifi-client-go@v0.1.7`.
- Existing provider behavior remains unchanged unless `v0.1.7` requires an adjustment.
- Tests pass.

## Results
- In progress.
