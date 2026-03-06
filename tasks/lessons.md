# Lessons

- Before assuming provider drift bugs, verify the Terraform config actually sets the attributes being compared; locals not wired into resources won't produce diffs.
- If state does not reflect upstream data, review Read mapping and schema optional/computed flags before assuming plan is correct.
- Plugin Framework defaults require `Computed=true`; prefer plan modifiers for optional defaults.
- Plan modifiers cannot inject non-computed defaults without triggering invalid plan errors.
- When consuming list endpoints, do not assume one request returns all results; verify pagination metadata (`offset`, `limit`, `totalCount`) and add regression tests for multi-page reads.
