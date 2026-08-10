---
spec_id: '02'
spec_name: scoped_rendering_ref_inference
title: Scoped Rendering Ref Inference
status: draft
created_at: '2026-08-10T11:10:08.879228+00:00'
updated_at: '2026-08-10T11:14:22.457873+00:00'
owner: mickume
source: https://github.com/agent-fox-dev/agent-fox/issues/752
schema_version: 1
---
# Scoped Rendering Ref Inference

## Intent

Ensure scoped rendering always activates by inferring requirement and test spec refs from traceability data and subtask text when explicit refs are missing, reducing token waste by ~15,000 tokens per session and providing validation warnings to surface the gap to users.

## Problem

When subtasks lack `requirement_refs` and `test_spec_refs` (older specs or
incomplete spec generation), `render_individual_scoped()` /
`RenderIndividualScoped()` falls back to unscoped rendering, dumping the
**entire** requirements and test spec — a swing of **15,000–20,000 tokens**
compared to properly scoped rendering.

Scoped rendering already works well when refs are populated: it filters
requirements and test specs to only those referenced by the target task
group's subtasks. The problem is that the fallback silently inflates token
usage with no signal to the user.

Additionally, the Go implementation has a pre-existing bug: its fallback path
returns fully unscoped tasks, whereas the Python implementation correctly
scopes tasks even in the fallback. This must be fixed as part of the backport.

## Origin

Backport of [agent-fox issue #752](https://github.com/agent-fox-dev/agent-fox/issues/752),
fixed in commit `8b921e984bf61463149681bab1b2aeef7dfcbda5`.

## Goals

1. Ensure scoped rendering activates whenever possible by inferring refs from
   the traceability table and subtask text when explicit refs are missing.
2. Warn users during validation when subtasks have empty refs so they can fix
   specs at generation time.
3. Strengthen generation instructions to mandate ref population on every subtask.
4. Fix the Go fallback to scope tasks like Python already does.
5. Maintain cross-implementation parity between Python and Go.

## Non-Goals

- Changing the scoped rendering logic itself (filtering, output format).
- Modifying the spec JSON schemas.
- Requiring spec regeneration for existing specs (inference handles them).
- Handling refs that are present but contain placeholder or invalid values
  (e.g., `"TBD"`, whitespace-only strings); invalid-ID cleanup is out of scope.

## Background

This PRD introduces **new** functionality (inference fallback, missing-refs
validation warning) that did not exist at the time of the `afspec_go` spec,
which covered the initial Go port of core afspec functionality (load, validate,
render, marshal). Although both this PRD and `afspec_go` touch
`golang/render.go` and `golang/validate.go`, they operate on **different
functions** with no overlap in implementation scope. These two specs are
therefore independent features on the same codebase — no dependency or
supersession relationship is required.

## Tech Stack

- **Python**: packages/afspec/afspec/ (Pydantic models, pytest tests)
- **Go**: golang/ (generated structs from JSON schema, `go test`)
- **Template**: packages/agentspec/agentspec/_templates/prompts/

## Proposed Changes

### 1. Render Inference Fallback (Python + Go)

Add two inference helpers before the unscoped fallback in
`render_individual_scoped()` / `RenderIndividualScoped()`:

**Traceability-based inference:** Filter `spec.tasks.traceability` entries
whose `task_id` starts with `"{target_group}."` and collect their
`requirement_id` and `test_spec_id` values.

**Text-based inference:** Regex-scan subtask `title` and `details` for ID
patterns matching known requirement IDs and test spec IDs in the spec. Use
compiled patterns:

- **Python** (module-level constants):
  - `_REQ_ID_RE = re.compile(r"\b(\w+-REQ-\d+(?:\.\d+|\.E\d+)?)\b")`
  - `_TS_ID_RE = re.compile(r"\b(TS-\w+-(?:\d+|P\d+|E\d+|SMOKE-\d+))\b")`

- **Go** (package-level `var` block using `regexp.MustCompile`, mirroring Python's module-level constants):
  ```go
  var reqIDRe = regexp.MustCompile(`\b(\w+-REQ-\d+(?:\.\d+|\.E\d+)?)\b`)
  var tsIDRe  = regexp.MustCompile(`\b(TS-\w+-(?:\d+|P\d+|E\d+|SMOKE-\d+))\b`)
  ```
  Defining these at package level avoids per-call compilation overhead and
  matches the Python performance idiom. Go's `regexp.Regexp` is documented as
  safe for concurrent use, so no additional synchronisation is required for
  these package-level variables.

Only include matches that exist in the actual spec (validated against known IDs).
If zero validated matches remain after filtering for both ref types, the chain
falls through to the unscoped fallback (with scoped tasks).

**Partial inference is accepted.** The inference chain proceeds using a logical
OR: if traceability yields `requirement_refs` but not `test_spec_refs` (or vice
versa), scoped rendering activates with whatever refs were inferred, and the
missing ref type falls back to full rendering for that section only. This
matches the agent-fox implementation where the condition is
`if inferred_req or inferred_ts`.

**Python inference logging:** When inference successfully activates scoped
rendering (either traceability-based or text-based), emit a log message at
`logging.INFO` via the module-level logger (`logging.getLogger(__name__)`).
This matches the agent-fox implementation and surfaces inference activation to
users who enable logging. Go omits this log (no logging infrastructure in the
Go codebase).

The inference chain: try traceability first. If empty, try text. If both empty,
fall back to full unscoped rendering (with scoped tasks).

#### Definition of "Empty" Refs

Both the inference fallback guard and the validation warning use the same
definition of "empty":

- **Python:** `if not subtask.requirement_refs` — falsy for both `None` and
  `[]`. The Pydantic model defaults to `Field(default_factory=list)`, so absent
  fields become `[]`. The check catches both.
- **Go:** `len(subtask.RequirementRefs) == 0` — catches both `nil` (the zero
  value for `[]string`) and explicitly empty slices.

This definition is consistent with the existing render code
(`if not requirement_ids and not test_spec_ids` in Python;
`len(reqRefs) == 0` in Go).

Only truly empty lists/slices trigger inference or warnings. Refs containing
placeholder strings (e.g., `"TBD"`) are not considered empty and are out of
scope for this feature.

### 2. Validation Warning (Python + Go)

Add `_check_missing_subtask_refs()` / `checkMissingSubtaskRefs()` that warns
when subtasks have empty `requirement_refs` or `test_spec_refs` (using the
emptiness definition above).

**Exclusion granularity:** Skip the **entire `TaskGroup`** when
`group.kind == WIRING_VERIFICATION` (`group.kind.value == "wiring_verification"`
in Go). The check returns early for the whole group — there is no subtask-level
marker. This matches the agent-fox implementation.

Wire into the existing `validate()` / `Validate()` function.

**Warning message format:** Emit a **single warning per subtask** with both
missing field names joined by `' and '`:

```
Subtask {id} has empty {field_names} — scoped rendering will fall back to full spec dump
```

Where `{field_names}` is:
- `"requirement_refs"` — if only requirement_refs is empty
- `"test_spec_refs"` — if only test_spec_refs is empty
- `"requirement_refs and test_spec_refs"` — if both are empty (using `' and '.join(missing)`)

This matches the agent-fox implementation's serialisation convention and ensures
consistent test assertions across Python and Go.

### 3. Generation Template Update

Add a `### Subtask refs (MANDATORY)` section to `generation_user_tasks.md`.

**Exact insertion point:** Insert after **line 52**, which reads:

> _"Reference both requirement IDs and test IDs from the previously generated
> artifacts in subtask `requirement_refs` and `test_spec_refs` fields."_

…and before `### Smoke test authoring` (line 54). Add a blank line, then the
new section, then a blank line before `### Smoke test authoring`.

**Exact content of the new section:**

```markdown
### Subtask refs (MANDATORY)
Every subtask MUST have non-empty `requirement_refs` and `test_spec_refs`. These fields control scoped rendering — when they are empty, the entire requirements and test spec are dumped into the coding session (~15,000+ extra tokens). Populate both fields for every subtask by cross-referencing the requirement IDs and test spec IDs from the previously generated artifacts.
```

### 4. Go Fallback Bug Fix

In the Go `RenderIndividualScoped()` else-branch (when no refs are found and
inference fails), replace `return s.RenderIndividual()` with code that also
scopes tasks via `s.renderScopedTasks(targetGroup)`, matching Python behavior.

## Files to Modify

| File | Change |
|------|--------|
| `packages/afspec/afspec/render.py` | Add `_infer_refs_from_traceability()`, `_infer_refs_from_subtask_text()`, modify `render_individual_scoped()` fallback |
| `packages/afspec/afspec/validation.py` | Add `_check_missing_subtask_refs()`, wire into `validate()` |
| `packages/agentspec/agentspec/_templates/prompts/generation_user_tasks.md` | Add mandatory refs section after line 52, before `### Smoke test authoring` |
| `golang/render.go` | Add `inferRefsFromTraceability()`, `inferRefsFromSubtaskText()`, package-level `reqIDRe`/`tsIDRe` vars, modify `RenderIndividualScoped()`, fix fallback task scoping |
| `golang/validate.go` | Add `checkMissingSubtaskRefs()`, wire into `Validate()` |

## New Test Files

| File | Coverage |
|------|----------|
| `packages/afspec/tests/test_render_inference.py` | Traceability inference, text inference, partial inference (OR logic), fallback behavior |
| `packages/afspec/tests/test_validation_warnings_missing_refs.py` | Missing refs warning, wiring-verification group skip, partial empty, `' and '` join serialisation |
| `golang/render_inference_test.go` | Go equivalents of Python render inference tests |
| `golang/validate_missing_refs_test.go` | Go equivalents of Python validation warning tests |

### Test Fixture Strategy

- **Python:** Use inline Pydantic model construction — build `Spec`,
  `Requirements`, `TestSpec`, `Tasks`, etc. directly in each test function or
  via `pytest` fixtures. Follow the pattern established in
  `packages/afspec/tests/test_render.py`.
- **Go:** Use inline struct literals. Follow the pattern established in
  `golang/render_test.go`.

No shared JSON fixture files are introduced. This keeps tests self-contained
and avoids fixture-maintenance overhead.

## Verified External API

### `afspec` (Python, local package)

| Symbol | Module | Signature | Notes |
|--------|--------|-----------|-------|
| `Spec` | `afspec.models` | Pydantic BaseModel with `.prd`, `.requirements`, `.test_spec`, `.tasks`, `.architecture` | Top-level spec container |
| `TaskGroup` | `afspec.models` | `.id: int`, `.kind: TaskGroupKind`, `.subtasks: list[Subtask]` | |
| `TaskGroupKind` | `afspec.models` | Enum: `TESTS`, `STANDARD`, `CHECKPOINT`, `WIRING_VERIFICATION` | |
| `Subtask` | `afspec.models` | `.id: str`, `.title: str`, `.details: list[str]`, `.requirement_refs: list[str]`, `.test_spec_refs: list[str]` | Defaults via `Field(default_factory=list)`; empty check: `if not subtask.requirement_refs` |
| `TraceabilityEntry` | `afspec.models` | `.requirement_id: str`, `.test_spec_id: str`, `.task_id: str`, `.test_path: Optional[str]` | |
| `Tasks` | `afspec.models` | `.traceability: list[TraceabilityEntry]`, `.task_groups: list[TaskGroup]` | |
| `Requirements` | `afspec.models` | `.requirements: list[Requirement]` with `.acceptance_criteria`, `.edge_cases` | |
| `TestSpec` | `afspec.models` | `.test_cases`, `.property_tests`, `.edge_case_tests`, `.smoke_tests` | |
| `ValidationWarning` | `afspec.validation` | `(message: str, entity_id: str)` | Pydantic BaseModel |
| `validate` | `afspec.validation` | `(spec: Spec) -> ValidationResult` | `.valid`, `.errors`, `.warnings` |
| `render_individual_scoped` | `afspec.render` | `(spec: Spec, target_group: int) -> dict[str, str]` | |
| `render_individual` | `afspec.render` | `(spec: Spec) -> dict[str, str]` | |

### `afspec` (Go, local package)

| Symbol | Package | Signature | Notes |
|--------|---------|-----------|-------|
| `Spec` | `afspec` | struct with `.Requirements`, `.TestSpec`, `.Tasks`, `.Architecture` | |
| `TaskGroup` | `afspec` | struct: `.Id int`, `.Kind TaskGroupKind`, `.Subtasks []Subtask` | |
| `TaskGroupKind` | `afspec` | type string; consts `TaskGroupKindWiringVerification`, etc. | |
| `Subtask` | `afspec` | struct: `.Id string`, `.Title string`, `.Details []string`, `.RequirementRefs []string`, `.TestSpecRefs []string` | Zero value is `nil`; empty check: `len(subtask.RequirementRefs) == 0` |
| `TraceabilityEntry` | `afspec` | struct: `.TaskId string`, `.RequirementId string`, `.TestSpecId string` | Generated from JSON schema |
| `ValidationResult` | `afspec` | struct: `.Valid bool`, `.Errors []ValidationEntry`, `.Warnings []ValidationEntry` | |
| `ValidationEntry` | `afspec` | struct: `.Category`, `.Check`, `.Message`, `.EntityID`, `.Artifact` | Used for both errors and warnings |
| `RenderIndividualScoped` | `afspec` | `(targetGroup int) map[string]string` | Method on `*Spec` |
| `Validate` | `afspec` | `() ValidationResult` | Method on `*Spec` |

## Impact

- **~15,000 token savings** per session when fallback currently activates
- **~43% reduction** in those cases
- Risk to quality: **low** — scoped rendering is already the intended behavior
- Backward compatible — inference handles existing specs without regeneration

## Design Decisions

1. **Traceability-first inference order:** Traceability data is authoritative
   (explicitly linked), so it is tried first. Text inference is a secondary
   heuristic for specs that also lack traceability data.
2. **Partial inference accepted (OR logic):** If traceability (or text) yields
   refs for only one ref type, scoped rendering still activates for that type.
   The missing type falls back to full rendering for that section. Condition:
   `if inferred_req or inferred_ts`. This avoids unnecessarily discarding
   partially inferrable context.
3. **Validated matches only:** Text inference matches are checked against the
   set of IDs actually present in the spec to prevent false positives from
   coincidental ID-like strings. If zero validated matches remain after
   filtering for both ref types, the chain falls through to unscoped fallback.
4. **WIRING_VERIFICATION exclusion at group level:** The missing-refs validation
   warning skips entire `TaskGroup`s whose `kind` is `WIRING_VERIFICATION`.
   There is no subtask-level marker; the check returns early for the whole group.
5. **Warning field serialisation:** Multiple missing field names are joined with
   `' and '` into a single warning per subtask (e.g.,
   `"requirement_refs and test_spec_refs"`), matching the agent-fox
   implementation's `' and '.join(missing)` convention.
6. **Go regex at package level:** Go regex patterns are defined as package-level
   `var` blocks via `regexp.MustCompile`, mirroring Python's module-level
   compiled constants for equivalent performance characteristics.
   `regexp.Regexp` is safe for concurrent use per Go documentation; no
   additional synchronisation is required.
7. **No logging in Go:** The Go codebase has no logging infrastructure, so
   the inference info log (present in Python) is omitted in Go.
8. **Python inference logging:** Successful inference activation is logged at
   `logging.INFO` via `logging.getLogger(__name__)`, matching the agent-fox
   implementation.
9. **Go fallback fix included:** The Go fallback is fixed to scope tasks,
   aligning with Python's existing behavior.
10. **Empty refs definition:** Inference and validation warnings trigger only on
    truly empty refs (absent/null or empty list). Refs with placeholder strings
    (e.g., `"TBD"`) are out of scope.
11. **Test fixture strategy:** Python tests use inline Pydantic model
    construction; Go tests use inline struct literals — consistent with the
    existing `test_render.py` and `render_test.go` conventions.
12. **Independent from `afspec_go` spec:** The `afspec_go` spec covered the
    initial Go port of core afspec functionality. This PRD adds inference and
    validation-warning features that post-date that work. Both touch the same
    files (`golang/render.go`, `golang/validate.go`) but in distinct,
    non-overlapping functions. No dependency or supersession relationship is
    declared.
