# Todo

## Working Notes
- Move `docs/` from hand-written Markdown to tfplugindocs output so the docs cannot drift from the schema.
- tfplugindocs builds the provider from the repository root, but `main` is in `cmd/terraform-provider-unifi`, so `scripts/generate-docs.sh` builds it, exports the schema with the Terraform CLI (dev override) and passes it with `--providers-schema`. tfplugindocs runs at a pinned version.
- 300 of 314 attributes already had text in the hand-written pages (the Go schemas had descriptions for only 40); that text moved into the schemas. 2 attributes were missing from the hand-written pages and appear now.
- Examples moved to `examples/`; notes and the DNS policy types list to per-page templates in `templates/`; the provider page to `templates/index.md.tmpl`.
- Import sections held only the ID format, so each became a real `terraform import` command in `import.sh`.

## Plan
- [x] Add `scripts/generate-docs.sh` and `make docs`.
- [x] Put the attribute text from the pages into the Go schemas (shared DNS helpers included).
- [x] Move examples, imports and notes into `examples/` and `templates/`.
- [x] Regenerate `docs/` and compare every attribute text with the old pages.
- [x] CI check that fails when `docs/` is stale.
- [x] Update AGENTS.md.
- [x] Verify: build, vet, tests, deterministic regeneration.

## Acceptance Criteria
- `make docs` regenerates all of `docs/` (except `api-versioning-and-promotion.md` and `releasing.md`) with no diff when run twice.
- No attribute text is lost compared with the hand-written pages.
- CI fails when a schema change is not reflected in `docs/`.

## Results
- Regeneration is deterministic. Of 343 attribute texts, 7 differ from the old pages: the Go schemas already had slightly different wording (richer or with fewer backticks), and the generator adds links to nested schemas.
- Dropped on purpose: the hand-written lists of resources and data sources on the provider page, which only duplicated what the registry shows and went stale.
