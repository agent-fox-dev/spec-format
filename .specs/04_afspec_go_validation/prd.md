---
spec_id: '04'
spec_name: afspec_go_validation
title: Afspec Go Validation
status: draft
created_at: '2026-08-11T11:29:07.720015+00:00'
updated_at: '2026-08-11T11:29:07.720015+00:00'
owner: ''
source: docs/prds/gospec.md
schema_version: 1
---
# Go afspec Validation Completion

## Intent

The Go `afspec` library has a working validation pipeline but only implements 4 of 10 cross-file integrity rules, 1 of 5 cross-spec rules, and 1 of 7 warning categories. This means `spec validate` cannot catch the same classes of errors the Python version catches — glossary gaps, missing test coverage for properties and execution paths, traceability issues, and structural problems in task groups. Completing validation parity is a prerequisite for replacing the Python CLI.

## Goals

- Implement all 10 cross-file integrity rules in Go, matching the Python `afspec.validation.validate_cross_file` behavior.
- Implement all 5 cross-spec rules in Go, matching `afspec.validation.validate_cross_spec`.
- Implement all 7 warning categories, matching `afspec.validation.validate` warning checks.
- Align `ValidateStructured` output shape with Python's `validate_structured` for structural equivalence.

## Non-goals

- Byte-identical error messages — structural parity (same rule name, same affected entity, same category) is sufficient.
- Changing validation semantics — the Go version must flag the same specs as valid/invalid as Python for the same inputs.
- Adding new validation rules not present in Python.

## Functional Requirements

### Cross-File Integrity Rules

- **Property test coverage (rule 3):** Every `correctness_property` in requirements must have a corresponding `property_test` in test_spec (matched by `property_id`). Missing coverage produces an error with the uncovered property ID.
- **Execution path smoke test coverage (rule 4):** Every `execution_path` in requirements must have a corresponding `smoke_test` in test_spec (matched by `execution_path_id`). Missing coverage produces an error.
- **test_spec_id resolution (rule 5):** Every `test_spec_id` in traceability entries and every entry in subtask `test_spec_refs` must resolve to a known test spec entry (test_case, property_test, edge_case_test, or smoke_test). Unresolvable IDs produce errors.
- **Glossary backtick term check (rule 6):** Extract backtick-wrapped terms from criterion fields (`action`, `trigger`, `condition`, `error_condition`, `state`, `feature`) and correctness property fields (`for_any`, `invariant`). Each extracted term must have a glossary entry. Exclude terms that match any of: numeric values (including negatives and decimals, e.g. `-1`, `3.14`), single characters, quoted strings (wrapped in single or double quotes), or strings longer than 80 characters.
- **Traceability deduplication (rule 8):** No two traceability entries may share the same `(requirement_id, test_spec_id)` pair. Duplicates produce errors.
- **Subtask requirement_refs resolution (rule 9):** Every entry in subtask `requirement_refs` must resolve to a known requirement ID or criterion ID (including edge case IDs). Unresolvable refs produce errors.
- **Unwanted return_contract (rule 10):** Every criterion with `ears_pattern: unwanted` must have a non-null `return_contract`. Missing return contracts produce errors.

### Task Group Structure Validation

- The first task group must have `kind: tests`. If it does not, produce a schema-category error.
- The last task group must have `kind: wiring_verification`. If it does not, produce a schema-category error.

### Wiring Verification Semantics

- The wiring verification group (last group) must contain at least one subtask with non-empty `test_spec_refs`. If all subtasks have empty refs, produce an error.
- At least one `test_spec_refs` entry in the wiring group must match the pattern `TS-*-SMOKE-*` (a smoke test reference). If none match, produce an error.
- At least one subtask in the wiring group must mention stub or dead-code audit in its title or details (case-insensitive search for "stub" or "dead"). If none mention it, produce an error.

### Cross-Spec Validation Rules

- **Duplicate API symbol (rule 1):** When two specs connected by a dependency edge both declare an external API symbol with the same `name`, their `signature` fields must match. Mismatches produce errors.
- **Unknown dependency (rule 3):** Every `depends_on_spec` in a spec's task dependencies must reference a spec_id that exists in the provided spec set. Unknown references produce errors.
- **Interface contract mismatch (rule 4):** Extract backtick-wrapped terms from criterion `action` fields and pair them with `return_contract` values. Along dependency edges, if a downstream spec references a symbol that an upstream spec also defines with a `return_contract`, the contracts must match. Mismatches produce errors.
- **Missing boundary coverage (rule 5):** For each execution path, extract actors. If an actor names a spec that appears in the dependency graph, there must be a corresponding smoke test covering that boundary. Missing boundary coverage produces errors.

### Warning Rules

- **Group test_spec_refs ceiling:** When the total count of `test_spec_refs` across all subtasks in a task group exceeds 15, produce a warning. Apply to all group kinds.
- **Group subtask count:** When a task group has more than 6 non-verification subtasks, produce a warning. Apply to all group kinds.
- **Subtask test_spec_refs ceiling:** When a single subtask has more than 8 entries in `test_spec_refs`, produce a warning.
- **Error path return_contract:** When a criterion has an `error_condition` field or uses error-indicating keywords in its `action` and has a null `return_contract`, produce a warning.
- **Vague language:** When criterion fields contain vague words (e.g., "appropriate", "properly", "correctly", "adequate", "sufficient"), produce a warning identifying the term and field.
- **Scope limit:** When a spec has more than 10 requirements, produce a warning suggesting the spec may be too large.

### ValidateStructured Output Alignment

- `ValidateStructured` must return a map with keys `errors` and `warnings`.
- Each error must include a `category` field: `schema` for JSON Schema and EARS pattern errors, `integrity` for cross-file and cross-spec rule violations.
- Each error must include `rule`, `message`, `file`, and `path` fields.
- Each warning must include `message` and `entity_id` fields.
- The overall result must include a `valid` boolean (true when errors list is empty).

## Technical Boundaries

- Go 1.26.5 per current go.mod.
- All validation functions operate on the existing `Spec` struct and its sub-types.
- Regex patterns for backtick extraction and vague language detection must be compiled at package level for efficiency.
- Cross-spec validation receives `map[string]*Spec` and `*DependencyGraph` matching the existing Go API.

## Verified External API

### `afspec` (Go, in-repo at `golang/`)

| Symbol | File | Signature | Notes |
|--------|------|-----------|-------|
| `ValidateSchema` | validate.go | `(*Spec).ValidateSchema() []ValidationEntry` | Already implemented |
| `ValidateCrossFile` | validate.go | `(*Spec).ValidateCrossFile() []ValidationEntry` | Partially implemented — needs 6 more rules |
| `ValidateCrossSpec` | validate.go | `ValidateCrossSpec(specs map[string]*Spec, graph *DependencyGraph) []ValidationEntry` | Partially implemented — only glossary conflict |
| `Validate` | validate.go | `(*Spec).Validate() ValidationResult` | Combines schema + cross-file + warnings |
| `ValidateStructured` | validate.go | `(*Spec).ValidateStructured() map[string]any` | Needs output shape alignment |
| `ValidationResult` | validate.go | `struct { Valid bool; Errors []ValidationEntry; Warnings []ValidationEntry }` | |
| `ValidationEntry` | validate.go | `struct { Category, Message, Artifact, Path, Keyword, Check, RequirementID, EntityID string; Value any }` | |
| `checkMissingSubtaskRefs` | validate_missing_refs.go | `checkMissingSubtaskRefs(spec *Spec) []ValidationEntry` | Already implemented |
| `Schemas` | schemas.go | `Schemas() map[string][]byte` | Embedded JSON Schema files |

