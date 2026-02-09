# Lessons

- Before assuming provider drift bugs, verify the Terraform config actually sets the attributes being compared; locals not wired into resources won't produce diffs.
- If state does not reflect upstream data, review Read mapping and schema optional/computed flags before assuming plan is correct.
