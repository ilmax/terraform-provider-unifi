# Todo

## Working Notes
- Upgrade `github.com/ilmax/unifi-client-go` from `v0.1.3` to `v0.1.6`.
- Expect generated type/package changes and revisit `unifi_site` against the latest SDK surface.

## Plan
- [x] Upgrade the Go client to `v0.1.6`.
- [x] Fix compile/test failures introduced by SDK breaking changes.
- [x] Rework `unifi_site` to match the latest site listing API shape.
- [x] Verify with `go test ./...` and update docs if required.

## Acceptance Criteria
- Provider builds and tests cleanly against `github.com/ilmax/unifi-client-go@v0.1.6`.
- `unifi_site` still resolves `site_id` by site name with clear missing/duplicate handling.
- Tests pass.

## Results
- Upgraded `github.com/ilmax/unifi-client-go` to `v0.1.6`.
- Reworked the provider HTTP client to stop depending on removed SDK packages (`pkg/config`, `pkg/sitemanager`).
- Updated `unifi_site` to use the generated `SiteOverview` page with `limit`/`offset` pagination.
- Adjusted `unifi_site` outputs to match the latest site model (`site_id`, `name`, `internal_reference`).
- Verified with `go test ./...`.
