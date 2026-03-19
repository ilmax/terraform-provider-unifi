# Todo

## Working Notes
- Upgrade `github.com/ilmax/unifi-client-go` from `v0.1.3` to `v0.1.6`.
- Expect generated type/package changes and revisit `unifi_site` against the latest SDK surface.

## Plan
- [ ] Upgrade the Go client to `v0.1.6`.
- [ ] Fix compile/test failures introduced by SDK breaking changes.
- [ ] Rework `unifi_site` to match the latest site listing API shape.
- [ ] Verify with `go test ./...` and update docs if required.

## Acceptance Criteria
- Provider builds and tests cleanly against `github.com/ilmax/unifi-client-go@v0.1.6`.
- `unifi_site` still resolves `site_id` by site name with clear missing/duplicate handling.
- Tests pass.

## Results
- In progress.
