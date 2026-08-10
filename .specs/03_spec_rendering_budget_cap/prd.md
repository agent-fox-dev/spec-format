---
spec_id: '03'
spec_name: spec_rendering_budget_cap
title: Spec Rendering Budget Cap
status: draft
created_at: '2026-08-10T13:36:52.079489+00:00'
updated_at: '2026-08-10T13:38:37.268179+00:00'
owner: ''
source: https://github.com/agent-fox-dev/agent-fox/issues/754
schema_version: 1
---
# Spec Rendering Budget Cap

## Intent

Add budget-aware rendering to the `afspec` library (Python and Go) that
progressively truncates rendered spec content when it exceeds a configurable
token budget. Spec content (rendered requirements, test_spec, tasks) accounts
for 70-90% of system prompt tokens in agent sessions. This feature gives
downstream consumers (e.g., agent-fox's `assemble_context()`) a rendering-layer
knob to cap spec content at a token target without manual truncation.

## Goals

1. Add a `max_tokens` optional parameter to `render_individual_scoped`,
   `render_individual`, and `render_combined` that caps total rendered output
   using progressive truncation.
   **Acceptance criteria:** Given a spec whose Level-0 render totals N estimated
   tokens and `max_tokens` is set to M < N, the returned output must have
   estimated tokens ≤ M when achievable by Level-1 or Level-2 truncation. When
   `max_tokens` is `None`, output must be byte-identical to current behavior.

2. Implement a token estimation utility using the chars/4 heuristic.
   **Acceptance criteria:** `estimate_tokens('abcd') == 1`;
   `estimate_tokens('') == 0`; `estimate_tokens(text) == len(text) // 4` for
   any input.

3. Define a two-level progressive truncation strategy: drop architecture first,
   then slim test spec assertions.
   **Acceptance criteria:** Given a spec with an architecture section present,
   Level-1 output must not contain the `architecture` key (dict) or architecture
   section (combined string). Given Level-2 output, test case entries must
   contain only `id`, `description`, `type`, and requirement linkage fields —
   `assertion_pseudocode`, `input`, and `expected` must be absent.

4. Implement equivalent functionality in the Go library for cross-implementation
   parity.
   **Acceptance criteria:** For any `Spec` value, Python
   `render_individual(spec, max_tokens=N)` and Go
   `RenderIndividual(WithMaxTokens(N))` must produce structurally equivalent
   output (same keys present/absent, same truncation level applied).

## Non-Goals

- Token estimation using tiktoken or model-specific tokenizers.
- Budget orchestration at the agent context assembly layer (lives in agent-fox).
- Truncation of PRD body content (always included in full).
- Returning truncation metadata (caller cannot distinguish truncated from full output in either the Python or Go API — both return the same types regardless of whether truncation occurred).
- Truncation beyond Level 2 (test spec slimming).

## Background

Spec content — rendered requirements, test specifications, and tasks — accounts
for 70–90% of system prompt tokens in agent-fox sessions. Without a
rendering-layer budget knob, downstream consumers such as `assemble_context()`
must implement ad-hoc manual truncation, leading to inconsistent behavior across
callers. This feature centralises the truncation policy inside `afspec` itself.

The chars/4 token heuristic is fast, dependency-free, and accurate enough for
budget decisions. Model-specific tokenizers (e.g., tiktoken) are explicitly
excluded to keep `afspec` free of heavy dependencies.

## Dependencies

- **`scoped_rendering_ref_inference` (spec 02) — required predecessor.**
  This spec modifies `render_individual_scoped` / `RenderIndividualScoped`, the
  same functions changed by `scoped_rendering_ref_inference`. Spec 02 is fully
  implemented and merged into main. `spec_rendering_budget_cap` must be
  implemented after spec 02; it builds on the ref-inference work already in
  place.

- **`afspec_go` (spec 01) — required predecessor.**
  This spec adds budget-aware rendering to the Go `afspec` package. The Go
  rendering methods (`RenderCombined`, `RenderIndividual`,
  `RenderIndividualScoped`) were established by `afspec_go`. Spec 01 is fully
  implemented and merged into main. This spec's Go work builds on top of those
  foundations.

## Tech Stack

- Python 3.12+ (afspec package, Pydantic models)
- Go 1.22+ (golang/ module, JSON Schema generated types)
- pytest, hypothesis (Python tests)
- go test (Go tests)

## Design

### Token Estimation

A public utility function that estimates token count from text length:

```
estimate_tokens(text) -> len(text) // 4
```

This is a fast, dependency-free heuristic. The function is public so consumers
can independently estimate their own content before calling budget-aware
rendering.

- Python: `estimate_tokens(text: str) -> int` in `afspec.render`
- Go: `EstimateTokens(text string) int` in `afspec` package

### Progressive Truncation Strategy

When `max_tokens` is set and the initial render exceeds the budget, truncation
levels are applied in order until the output fits:

**Level 0 — Full render.** No truncation. This is the default when `max_tokens`
is `None` / `0`.

**Level 1 — Drop architecture.** Remove the `architecture` key from the result
dict (for `render_individual` / `render_individual_scoped`) or omit the
architecture section entirely (for `render_combined`). Architecture is the
lowest-signal artifact for implementation tasks. PRD body is never dropped.

If the spec has no `architecture` section, Level 1 is a no-op. The evaluation
loop checks whether Level 1 actually changed anything (i.e., whether an
architecture section was present). If architecture was already absent, the loop
skips Level 1's budget re-evaluation and proceeds directly to Level 2.

**Level 2 — Slim test spec.** Re-render the test specification without
`assertion_pseudocode`, `input`, and `expected` blocks on test cases and
edge case tests. Each test entry is reduced to its ID, description, type, and
requirement linkage. Property tests drop `for_any_strategy` and
`invariant_check`. Smoke tests drop `expected_effects`. This preserves
traceability while removing the bulk of test spec verbosity.

If the output still exceeds the budget after Level 2, return the Level 2
output as-is. No further truncation is applied.

### Budget Evaluation

For dict-returning functions (`render_individual`, `render_individual_scoped`),
the budget is evaluated against the **sum of all values** in the result dict.
This matches how consumers concatenate artifacts into a prompt. Keys whose
values are empty strings contribute 0 tokens; missing keys are not included in
the sum. A spec with no architecture section therefore contributes no
architecture tokens at Level 0, and Level 1 is a no-op (see above).

For `render_combined`, the budget is evaluated against the single returned
string.

The evaluation loop:
1. Render at Level 0. Sum token estimates across all artifacts.
2. If within budget, return.
3. Apply Level 1. If architecture was already absent, skip to step 5.
4. Re-sum. If within budget, return.
5. Apply Level 2. Return regardless of budget.

Architecture is dropped by removing it from the result (cheap — no re-render).
Test spec slimming requires a re-render using slim helpers.

### Python API Changes

Existing function signatures gain an optional `max_tokens` parameter:

```python
def render_individual_scoped(
    spec: Spec, target_group: int, max_tokens: int | None = None
) -> dict[str, str]: ...

def render_individual(
    spec: Spec, max_tokens: int | None = None
) -> dict[str, str]: ...

def render_combined(
    spec: Spec, max_tokens: int | None = None
) -> str: ...

def estimate_tokens(text: str) -> int: ...
```

`estimate_tokens` is added to `__init__.py` exports and `__all__`.

### Go API Changes

Go uses a functional options pattern since it lacks optional parameters:

```go
type RenderOption func(*renderConfig)

func WithMaxTokens(n int) RenderOption

func (s *Spec) RenderCombined(opts ...RenderOption) string
func (s *Spec) RenderIndividual(opts ...RenderOption) map[string]string
func (s *Spec) RenderIndividualScoped(targetGroup int, opts ...RenderOption) map[string]string

func EstimateTokens(text string) int
```

Adding variadic parameters is backward-compatible in Go — existing callers
continue to work without modification. The `renderConfig` struct and
`RenderOption` type are package-level. `renderConfig` is unexported;
`RenderOption` and `WithMaxTokens` are exported.

Go callers receive no truncation metadata — `RenderIndividual` and
`RenderIndividualScoped` return `map[string]string` and `RenderCombined` returns
`string`, identical to today. This is the same contract as the Python API.

### Internal Helpers

New private functions for slim (Level 2) rendering:

**Python:**
- `_render_test_spec_slim(ts: TestSpec) -> str` — full test spec without
  assertion pseudocode, input, or expected blocks.
- `_render_test_spec_scoped_slim(ts: TestSpec, test_spec_ids: set[str]) -> str`
  — scoped variant of slim rendering.

**Go:**
- `renderTestSpecSlim(ts *TestSpecV1Json) string`
- `renderTestSpecScopedSlim(ts *TestSpecV1Json, ids map[string]bool) string`

These mirror the structure of the existing `render_test_spec` /
`render_test_spec_scoped` functions but omit verbose fields.

## Verified External API

### `afspec` (current Python — functions being modified)

| Symbol | Module | Signature | Notes |
|--------|--------|-----------|-------|
| `render_individual_scoped` | `afspec.render` | `(spec: Spec, target_group: int) -> dict[str, str]` | Adding optional `max_tokens` param |
| `render_individual` | `afspec.render` | `(spec: Spec) -> dict[str, str]` | Adding optional `max_tokens` param |
| `render_combined` | `afspec.render` | `(spec: Spec) -> str` | Adding optional `max_tokens` param |
| `render_test_spec` | `afspec.render` | `(ts: TestSpec) -> str` | Referenced by slim variants |
| `render_test_spec_scoped` | `afspec.render` | `(ts: TestSpec, test_spec_ids: set[str]) -> str` | Referenced by slim variants |

### `afspec` (current Go — methods being modified)

| Symbol | Package | Signature | Notes |
|--------|---------|-----------|-------|
| `(*Spec).RenderCombined` | `afspec` | `() string` | Adding variadic `opts ...RenderOption` |
| `(*Spec).RenderIndividual` | `afspec` | `() map[string]string` | Adding variadic `opts ...RenderOption` |
| `(*Spec).RenderIndividualScoped` | `afspec` | `(targetGroup int) map[string]string` | Adding variadic `opts ...RenderOption` |

## Design Decisions

1. **Python and Go:** Implementing in both for cross-implementation parity,
   even though only Python is currently consumed by agent-fox.
2. **Token estimation:** chars/4 heuristic — fast, zero dependencies, good
   enough for budget decisions.
3. **API style:** Optional parameter on existing functions. Go uses functional
   options since Go lacks optional parameters. This preserves the existing API
   surface — callers that don't pass `max_tokens` get identical behavior.
4. **No truncation metadata:** Budget-aware rendering returns the same types as
   today in both Python and Go. Simplicity over observability. Go callers (CLIs,
   servers, CI tools) have the same contract: no signal that truncation occurred.
5. **PRD body is never truncated:** The PRD is human-authored intent. It is
   always included in full.
6. **Budget overflow:** If Level 2 still exceeds the budget, output is returned
   as-is. No further degradation — the caller handles overflow.
7. **Task scoping is not a budget level:** The `render_individual_scoped`
   function already collapses non-target task groups to one-line summaries.
   This is a rendering mode, not a budget truncation level. Budget truncation
   operates on top of whatever rendering mode was chosen.
8. **Missing architecture is a no-op at Level 1:** If a spec has no architecture
   section, Level 1 changes nothing. The evaluation loop detects this and
   proceeds directly to Level 2 without re-evaluating the budget.
9. **Dependency ordering:** This spec is implemented after spec 01 (`afspec_go`)
   and spec 02 (`scoped_rendering_ref_inference`), both of which are fully
   merged. No in-flight conflicts exist with those specs.

## Owner

mkuehl

## Source

https://github.com/agent-fox-dev/agent-fox/issues/754
