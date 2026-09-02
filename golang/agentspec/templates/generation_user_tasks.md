---
name: generation_user_tasks
description: Additional instructions for tasks artifact generation
---
Generate the implementation tasks artifact. Include task groups, subtasks, dependencies, verification steps, and traceability matrix.

## Task Group Ordering Rules

The following ordering rules are schema-enforced and must be respected exactly:

1. **The first task group (id: 1) must have `kind: "tests"`** — the first group always writes failing spec tests (TDD). No exceptions.
2. **The final task group must have `kind: "wiring_verification"`** — the last group always verifies integration wiring. No exceptions.
3. **No more than one `"wiring_verification"` group per spec** — only the final group may carry this kind.

Allowed `kind` values: `"tests"`, `"standard"`, `"checkpoint"`, `"wiring_verification"`.

## Subtask and Verification IDs

- Subtask ID format: `{group_id}.{N}` — e.g. `2.1`, `2.2`, `3.1`
- Verification subtask ID format: `{group_id}.V` — e.g. `1.V`, `2.V`, `3.V`
- Every task group must have exactly one verification subtask with ID `{group_id}.V`
- Target 3–6 regular subtasks per group (excluding the verification subtask)
