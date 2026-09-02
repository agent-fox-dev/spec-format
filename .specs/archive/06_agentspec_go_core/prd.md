---
spec_id: '06'
spec_name: agentspec_go_core
title: Agentspec Go Core
status: draft
created_at: '2026-08-11T11:29:15.151596+00:00'
updated_at: '2026-08-11T11:29:15.151596+00:00'
owner: ''
source: docs/prds/gospec.md
schema_version: 1
---
# Go agentspec Core Package

## Intent

The Python `agentspec` package provides the session management, campaign lifecycle, and configuration layer that sits between the `afspec` format library and the `spec` CLI. It manages the stateful process of authoring a spec — from PRD input through assessment, refinement, and artifact generation. This spec ports the non-AI portions of `agentspec` to Go: error types, configuration loading, campaign management, session state machine with persistence, and tool schema definitions. These components contain no LLM calls and can be tested without an API key.

## Goals

- Implement Go error types matching the Python `agentspec.errors` hierarchy with structured metadata for programmatic handling.
- Implement TOML configuration loading with environment variable overrides matching `agentspec.config`.
- Implement campaign directory lifecycle (create, open, list specs, provision new specs) matching `agentspec.campaign`.
- Implement the session state machine with atomic persistence, assessment history, QA exchange recording, and partial failure recovery matching `agentspec.session` (excluding AI methods).
- Implement tool schema definitions for Anthropic structured output matching `agentspec.tools`.

## Non-goals

- LLM client integration — covered by a separate spec.
- AI-powered session methods (assess, refine, generate) — covered by a separate spec.
- Prompt template system — covered by a separate spec.
- CLI commands — covered by a separate spec.

## Functional Requirements

### Error Types

- `AgentSpecError` — base error type. All agentspec errors must be matchable via `errors.As` to this type. Has a `Category() string` method returning `"internal"` by default.
- `ConfigError` — raised for invalid TOML, missing required fields, or symlinked config files. `Category()` returns `"config"`.
- `CampaignError` — raised for campaign directory operation failures (duplicate path, missing campaign.yaml, invalid spec name).
- `SessionError` — raised for illegal state transitions or invalid session state. `Category()` returns `"state"`.
- `AgentError` — the richest error type, carrying: `Detail` (string), `ErrorCategory` (string: rate_limit, auth, transient, overloaded, input, internal, validation, refusal, context_window, pause_turn), `Retryable` (bool), `HTTPStatus` (optional int). Must preserve the original error via `Unwrap()`.

### Configuration

- `AgentSpecConfig` struct with fields: `Model` (string, default `"STANDARD"`), `AuthMethod` (string), `VertexProject` (string), `VertexRegion` (string).
- `LoadConfig() (AgentSpecConfig, error)` — search for `config.toml` in `.specs/` (relative to cwd) then `~/.specs/` (global). First file found wins; files are not merged. Read the `[model]` section for the `model` key and the `[provider]` section for `auth_method`, `vertex_project`, `vertex_region`. Silently ignore unknown sections (e.g., legacy `[spec_tool]`, `[theme]`).
- If `AF_SPEC_MODEL` environment variable is set, it always overrides the `model` field from the config file.
- Reject symlinked config files — if the resolved path differs from the original path, return a `ConfigError`.
- Return `ConfigError` for invalid TOML syntax.
- When no config file is found, return a config with defaults (Model: `"STANDARD"`).

### Campaign Management

- `CampaignMetadata` struct: `Name` (string), `Description` (string), `CreatedAt` (string, ISO 8601), `UpdatedAt` (string, ISO 8601).
- `Campaign` struct holding a `Path` and `Metadata`.
- `CreateCampaign(path, name, description string) (*Campaign, error)` — validate parent directory exists, check that `campaign.yaml` does not already exist at path (return `CampaignError` if it does), write `campaign.yaml` atomically (temp file + rename) with name, description, created_at, updated_at. Return the new Campaign.
- `OpenCampaign(path string) (*Campaign, error)` — read and parse `campaign.yaml`. Return `CampaignError` if file is missing or malformed.
- `(*Campaign).Specs() ([]string, error)` — list subdirectories matching the `{NN}_{snake_case}` pattern (via `ParseSpecDirName`), sorted by numeric prefix, excluding `archive/`. Return directory names.
- `(*Campaign).NewSpec(specName, prdPath, mode, source string) (*SpecSession, error)` — validate spec name against `[a-z][a-z0-9_]*` pattern, compute the next numeric prefix by scanning both active specs and the `archive/` subdirectory, create the spec directory, write `prd.md` with YAML frontmatter (spec_id, spec_name, title, status: draft, created_at, updated_at, owner, source, schema_version: 1), copy PRD body from the provided file, and initialize a SpecSession via `CreateSession`.

### Session State Machine

- `SessionState` string type with constants: `StateInit`, `StateAssessing`, `StateRefining`, `StatePRDAccepted`, `StateGenerating`, `StateGenerated`.
- `Question` struct: `ID`, `Text`, `Context` (strings), `Options` ([]string), `Required` (bool).
- `Assessment` struct: `Quality` (string: ready/needs_refinement/incomplete), `Summary` (string), `Gaps` ([]string), `Questions` ([]Question).
- `RepairSuggestion` struct: `Artifact`, `Description`, `Patch` (strings), `AutoFixable` (bool).
- `SessionValidationResult` struct: `Valid` (bool), `SchemaErrors` ([]string), `IntegrityErrors` ([]string), `RepairSuggestions` ([]RepairSuggestion).
- `GenerateResult` struct: `Artifacts` ([]string), `Validation` (SessionValidationResult), `Warnings` ([]string).
- `QAExchange` struct: `AssessmentIndex` (int), `Answers` (map[string]string), `Timestamp` (string, ISO 8601 UTC).
- `LastError` struct: `Message`, `Category` (strings), `Retryable` (bool), `HTTPStatus` (optional int), `Cause` (string).

#### SpecSession

- `SpecSession` struct holding: `specDir` (string), `state` (SessionState), `mode` (string), `prdPath` (string), `assessmentHistory` ([]Assessment), `qaExchanges` ([]QAExchange), `generatedArtifacts` ([]string), `lastError` (*LastError).
- `CreateSession(specDir, mode, source string) (*SpecSession, error)` — create a new session in `StateInit`, persist to `_session.json`.
- `ResumeSession(specDir string) (*SpecSession, error)` — read `_session.json`, reconstruct session state. Return `SessionError` if file is missing or malformed.
- `(*SpecSession).State() SessionState` — return current state.
- `(*SpecSession).SpecDir() string` — return spec directory path.
- `(*SpecSession).Assessment() *Assessment` — return most recent assessment, or nil.
- `(*SpecSession).PendingQuestions() []map[string]any` — return questions from latest assessment as serializable maps.
- `(*SpecSession).AcceptPRD() error` — transition from `StateAssessing` or `StateRefining` to `StatePRDAccepted`. Return `SessionError` if current state is neither.
- `(*SpecSession).Validate() (SessionValidationResult, error)` — load the spec via afspec, run full validation, categorize errors as schema or integrity, return structured result. Fall back to loading individual JSON artifacts when `LoadSpec` fails.
- `(*SpecSession).Render(combined bool) (any, error)` — delegate to afspec `RenderCombined` or `RenderIndividual`. Return string for combined, map[string]string for individual. Fall back to rendering available artifacts when some are missing.

#### Session Persistence

- All state transitions persist atomically to `_session.json` via temp-file-and-rename.
- `_session.json` contains: `state`, `prd_path`, `assessment_history`, `qa_exchanges`, `generated_artifacts`, `mode`, and optionally `last_error`.
- JSON field names use snake_case to match the serialization convention.

### Tool Definitions

- `AssessmentTools() []map[string]any` — return a list containing the `submit_assessment` tool definition with input schema: quality (enum: ready/needs_refinement/incomplete), summary (string), gaps (array of strings), questions (array of objects with id, text, context, options, required).
- `RefinementTools() []map[string]any` — return `[submit_prd_update, submit_assessment]`. `submit_prd_update` has input schema: updated_prd (string).
- `ArtifactTool(artifactName string) []map[string]any` — return a list containing `submit_{artifactName}` with input schema whose `content` property embeds the afspec JSON Schema for that artifact type (requirements, test_spec, or tasks). The schema must have all `$ref`/`$defs` resolved inline. Metadata noise (title labels, default values) must be stripped while preserving description fields.
- `InlineRefs(schema map[string]any) map[string]any` — recursively resolve `$ref` references against `$defs`, producing a self-contained schema. Remove the `$defs` key from the result.
- `CleanSchema(schema map[string]any) map[string]any` — strip `title` fields from all levels, remove `default` fields, remove the top-level `$schema` alias field. Preserve `description` fields.

## Technical Boundaries

- Go 1.26.5 per current go.mod.
- TOML parsing requires a Go TOML library (e.g., `github.com/BurntSushi/toml`).
- Campaign YAML uses `github.com/goccy/go-yaml` (already in go.mod).
- Tool schemas load from the embedded JSON Schema files via `afspec.Schemas()`.
- All file writes use atomic temp-file-and-rename pattern.
- `context.Context` is the first parameter for methods that will later gain AI calls.

## Verified External API

### `afspec` (Go, in-repo at `golang/`)

| Symbol | File | Signature | Notes |
|--------|------|-----------|-------|
| `LoadSpec` | afspec.go | `LoadSpec(dir string) (*Spec, error)` | |
| `Validate` | validate.go | `(*Spec).Validate() ValidationResult` | |
| `RenderCombined` | render.go | `(*Spec).RenderCombined(opts ...RenderOption) string` | |
| `RenderIndividual` | render.go | `(*Spec).RenderIndividual(opts ...RenderOption) map[string]string` | |
| `MarshalJSON` | marshal.go | `MarshalJSON(v any) ([]byte, error)` | Deterministic JSON |
| `Schemas` | schemas.go | `Schemas() map[string][]byte` | Embedded JSON Schema files |
| `ParseSpecDirName` | dirname.go | `ParseSpecDirName(name string) (int, string, bool)` | |
| `IsSpecDirName` | dirname.go | `IsSpecDirName(name string) bool` | |

### `github.com/goccy/go-yaml` (already in go.mod)

| Symbol | Package | Signature | Notes |
|--------|---------|-----------|-------|
| `Marshal` | `yaml` | `func Marshal(v interface{}) ([]byte, error)` | For campaign.yaml |
| `Unmarshal` | `yaml` | `func Unmarshal(data []byte, v interface{}) error` | For campaign.yaml |

