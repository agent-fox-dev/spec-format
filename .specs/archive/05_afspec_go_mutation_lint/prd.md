---
spec_id: '05'
spec_name: afspec_go_mutation_lint
title: Afspec Go Mutation Lint
status: draft
created_at: '2026-08-11T11:29:11.716374+00:00'
updated_at: '2026-08-11T11:29:11.716374+00:00'
owner: ''
source: docs/prds/gospec.md
schema_version: 1
---
# Go afspec Mutation Functions and Lint Module

## Intent

The Go `afspec` library currently supports subtask state transitions but lacks the full set of collection mutation functions needed for programmatic spec editing (adding requirements, test cases, glossary entries, etc.) and the lint module needed for the `spec lint` CLI command. These are prerequisites for the `agentspec` Go package (which builds specs programmatically) and the `spec` CLI (which runs linting).

## Goals

- Implement all collection mutation functions (add, get, remove) for every entity type in the spec format, following the existing immutable-copy pattern.
- Implement all sequential ID generation helpers for every entity type.
- Implement the lint module for discovering, validating, and reporting on spec folders.
- Implement `LoadDependentInterfaces` for cross-spec context loading.

## Non-goals

- Changing mutation semantics — all mutations return new copies, never modify in place.
- Adding mutation functions not present in the Python `afspec.mutate` module.
- Exposing mutation functions via CLI — they are a library API consumed by `agentspec`.

## Functional Requirements

### Collection Mutation Functions

- All mutation functions must accept the target collection (or parent struct) and the item to add, and return a new copy with the addition applied. The original must not be modified.
- All add functions must check for duplicate IDs and return an error if a duplicate is found.
- Get functions return a pointer to the found item and a boolean indicating whether it was found.
- Remove functions return a new copy with the item removed and a boolean indicating whether the item existed.

#### Requirements Mutations
- `AddRequirement(req RequirementsV1Json, r Requirement) (RequirementsV1Json, error)` — add a requirement; error on duplicate ID.
- `GetRequirement(req RequirementsV1Json, id string) (*Requirement, bool)` — find by ID.
- `RemoveRequirement(req RequirementsV1Json, id string) (RequirementsV1Json, bool)` — remove by ID.
- `SetGlossaryEntry(req RequirementsV1Json, term, definition string) RequirementsV1Json` — insert or overwrite.
- `RemoveGlossaryEntry(req RequirementsV1Json, term string) (RequirementsV1Json, bool)` — remove by term.
- `AddCorrectnessProperty(req RequirementsV1Json, p CorrectnessProperty) (RequirementsV1Json, error)` — error on duplicate ID.
- `AddExecutionPath(req RequirementsV1Json, p ExecutionPath) (RequirementsV1Json, error)` — error on duplicate ID.
- `AddErrorHandling(req RequirementsV1Json, e ErrorHandlingEntry) (RequirementsV1Json, error)` — error on duplicate ID.

#### Criterion Mutations
- `AddCriterion(r Requirement, c Criterion) (Requirement, error)` — add to acceptance_criteria; error on duplicate ID.
- `AddEdgeCase(r Requirement, c Criterion) (Requirement, error)` — add to edge_cases; error on duplicate ID.
- `GetCriterion(r Requirement, id string) (*Criterion, bool)` — search both acceptance_criteria and edge_cases.

#### Test Spec Mutations
- `AddTestCase(ts TestSpecV1Json, tc TestCase) (TestSpecV1Json, error)` — error on duplicate ID.
- `AddPropertyTest(ts TestSpecV1Json, pt PropertyTest) (TestSpecV1Json, error)` — error on duplicate ID.
- `AddEdgeCaseTest(ts TestSpecV1Json, et EdgeCaseTest) (TestSpecV1Json, error)` — error on duplicate ID.
- `AddSmokeTest(ts TestSpecV1Json, st SmokeTest) (TestSpecV1Json, error)` — error on duplicate ID.

#### Tasks Mutations
- `AddTaskGroup(t TasksV1Json, g TaskGroup) (TasksV1Json, error)` — error on duplicate ID.
- `AddSubtask(g TaskGroup, s Subtask) (TaskGroup, error)` — error on duplicate ID.
- `AddTraceabilityEntry(t TasksV1Json, e TraceabilityEntry) (TasksV1Json, error)` — error on duplicate `(requirement_id, test_spec_id)` pair.
- `AddDependency(t TasksV1Json, d TaskDependency) TasksV1Json` — append dependency.

### Sequential ID Generation

- Each `Next*ID` function must scan the existing collection, extract the trailing numeric suffix from each ID using the entity's ID format pattern, find the maximum, and return `max + 1` formatted into the correct ID pattern.
- Empty collections must return the ID with suffix `1`.
- Non-contiguous IDs must still return `max + 1` (not fill gaps).

#### ID Format Patterns
- `NextRequirementID(req RequirementsV1Json) string` — `{spec_id}-REQ-{N}`
- `NextCriterionID(r Requirement) string` — `{requirement_id}.{N}`
- `NextEdgeCaseID(r Requirement) string` — `{requirement_id}.E{N}`
- `NextCorrectnessPropertyID(req RequirementsV1Json) string` — `{spec_id}-PROP-{N}`
- `NextExecutionPathID(req RequirementsV1Json) string` — `{spec_id}-PATH-{N}`
- `NextErrorHandlingID(req RequirementsV1Json) string` — `{spec_id}-ERR-{N}`
- `NextTestCaseID(ts TestSpecV1Json) string` — `TS-{spec_id}-{N}`
- `NextPropertyTestID(ts TestSpecV1Json) string` — `TS-{spec_id}-P{N}`
- `NextEdgeCaseTestID(ts TestSpecV1Json) string` — `TS-{spec_id}-E{N}`
- `NextSmokeTestID(ts TestSpecV1Json) string` — `TS-{spec_id}-SMOKE-{N}`

### Lint Module

- `LintFinding` struct with fields: `SpecName`, `File`, `Rule`, `Severity` (error/warning/hint), `Message`, `Line` (optional).
- `LintSpecInfo` struct with fields: `Name`, `Prefix` (int), `Path`, `HasTasks`, `HasPRD`.
- `LintResult` struct with fields: `Findings []LintFinding`, `ExitCode int`.
- `DiscoverLintSpecs(specsDir string, filterSpec string) ([]LintSpecInfo, error)` — discover spec folders that contain `requirements.json`. Filter to a single spec by name if `filterSpec` is non-empty. Return error if the directory does not exist or contains no specs.
- `SortFindings(findings []LintFinding) []LintFinding` — sort by spec name, then file, then severity (error < warning < hint).
- `ComputeExitCode(findings []LintFinding) int` — return 1 if any finding has severity `error`, else 0.
- `RunLintSpecs(specsDir string, lintAll bool) (LintResult, error)` — discover specs, skip fully-implemented specs unless `lintAll` is true (a spec is fully implemented when all subtasks in all groups have state `done` or `dropped`), validate each remaining spec, collect findings, sort them, compute exit code.

### LoadDependentInterfaces

- `LoadDependentInterfaces(specID string, specRoot string) []map[string]any` — load interface summaries (glossary, external APIs, key symbols) from upstream dependency specs. Return an empty slice on any error (graceful degradation). For each upstream spec, extract its glossary entries, external API symbols, and any interface-defining criterion return contracts.

## Technical Boundaries

- Go 1.26.5 per current go.mod.
- All mutation functions are package-level functions (not methods on Spec) matching Go convention for operations that return new copies.
- Lint module uses existing `LoadSpec`, `Validate`, and `ParseSpecDirName` from the afspec package.
- ID extraction uses compiled regex patterns at package level.

## Verified External API

### `afspec` (Go, in-repo at `golang/`)

| Symbol | File | Signature | Notes |
|--------|------|-----------|-------|
| `LoadSpec` | afspec.go | `LoadSpec(dir string) (*Spec, error)` | |
| `Validate` | validate.go | `(*Spec).Validate() ValidationResult` | Used by lint |
| `ParseSpecDirName` | dirname.go | `ParseSpecDirName(name string) (int, string, bool)` | Returns prefix, name, ok |
| `IsSpecDirName` | dirname.go | `IsSpecDirName(name string) bool` | |
| `TransitionSubtask` | subtask.go | `(*TasksV1Json).TransitionSubtask(id string, target SubtaskState) (*TasksV1Json, error)` | Existing mutation pattern |
| `CompleteSubtaskStates` | subtask.go | `(*TasksV1Json).CompleteSubtaskStates(groupIDs []int) *TasksV1Json` | Existing mutation pattern |
| `BuildDependencyGraph` | graph.go | `BuildDependencyGraph(metas []SpecMeta, root string) (*DependencyGraph, error)` | Used by LoadDependentInterfaces |
| `DiscoverSpecs` | discover.go | `DiscoverSpecs(root string) ([]SpecMeta, error)` | |

