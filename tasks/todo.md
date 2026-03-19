# Todo

## Working Notes
- Make provider-level `site_id` handling consistent across all site-scoped resources and data sources.
- Keep the precedence uniform: resource/data source `site_id` overrides provider `site_id`; if neither is set, return a diagnostic.
- Encode this pattern in `AGENTS.md` so future resources follow the same rule.

## Plan
- [x] Audit current site-scoped resources/data sources for `site_id` schema, resolution, and persisted state behavior.
- [x] Update `AGENTS.md` with explicit DOs and DON'Ts for provider-level `site_id` support.
- [x] Fix any resource/data source inconsistencies in runtime or state handling.
- [x] Add regression tests that enforce optional `site_id` on all site-scoped resources/data sources and verify precedence helpers.
- [x] Update docs so provider-level `site_id` behavior is described consistently.
- [x] Verify with `go test ./...`.

## Acceptance Criteria
- Every site-scoped resource and data source accepts optional `site_id`.
- All site-scoped CRUD/read paths resolve `site_id` with the same precedence.
- Resource state persists the resolved `site_id`.
- Docs describe the same behavior everywhere.
- Tests pass with `go test ./...`.

## Results
- Audited site-scoped resources and data sources and confirmed they already resolve `site_id` through the shared helper at runtime.
- Added explicit repository rules for provider-level `site_id` handling in `AGENTS.md`.
- Added regression tests for `resolveSiteID` precedence and for optional `site_id` on all site-scoped resources and data sources.
- Updated README, provider index, and site-scoped docs to describe the same `site_id` precedence consistently.
- Verified with `go test ./...`.
