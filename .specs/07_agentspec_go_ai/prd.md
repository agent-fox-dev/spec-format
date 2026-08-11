---
spec_id: '07'
spec_name: agentspec_go_ai
title: Agentspec Go Ai
status: draft
created_at: '2026-08-11T11:29:18.770138+00:00'
updated_at: '2026-08-11T11:29:18.770138+00:00'
owner: ''
source: docs/prds/gospec.md
schema_version: 1
---
# Go agentspec AI Layer

## Intent

The AI-powered portion of the `agentspec` package drives the core value of the `spec` tool: assessing PRD quality, refining PRDs through iterative Q&A, and generating specification artifacts (requirements, test_spec, tasks) from accepted PRDs. This spec ports the LLM client, prompt template system, and AI agent pipeline from Python to Go, wiring them into the session state machine built by the agentspec core spec.

## Goals

- Implement a Go LLM client with model resolution, platform-aware Anthropic client creation, streaming, retry with exponential backoff, and prompt caching.
- Implement the prompt template system with embedded templates, two-tier fallback loading, variable substitution, and language detection.
- Implement the AI agent pipeline: PRD assessment, PRD refinement, and sequential artifact generation with repair loops.
- Wire AI methods into SpecSession: `Assess()`, `Refine()`, `Generate()`.

## Non-goals

- Supporting LLM providers other than Anthropic (direct API, Vertex AI, Bedrock).
- Changing prompt templates or AI behavior — the Go version uses the same templates and pipeline as Python.
- Interactive user prompting — the AI layer is a library; the CLI handles user interaction.

## Functional Requirements

### LLM Client

- `ModelTier` string type with constants: `TierSimple` ("SIMPLE"), `TierStandard` ("STANDARD"), `TierAdvanced` ("ADVANCED").
- `ModelEntry` struct: `ModelID` (string), `Tier` (ModelTier), `Variant` (string).
- `ModelRegistry` — a package-level map from model ID strings to `ModelEntry`. Initial entries: claude-haiku-4-5 (SIMPLE), claude-sonnet-4-6 (STANDARD), claude-opus-4-6 (ADVANCED).
- `TierDefaults` — a package-level map from `ModelTier` to default model ID string.
- `ResolveModel(name string) (string, error)` — if `name` matches a tier constant (case-insensitive), return the tier default. If `name` matches a model ID in the registry, return it directly. Otherwise return an error.
- `CachePolicy` string type with constants: `CacheNone`, `CacheDefault`, `CacheExtended`.
- `AICall(ctx context.Context, opts AICallOptions) (string, any, error)` — the central LLM interface. Options include: ModelTier, MaxTokens (default 65536), Messages, System, Context (string for logging), CachePolicy (default CacheDefault), Temperature (optional), Tools, ToolChoice. Returns (response_text_or_empty, raw_response, error).
  - Resolve model tier to model ID.
  - Create a platform-aware client: check `CLAUDE_CODE_USE_VERTEX` env var for Vertex AI, `CLAUDE_CODE_USE_BEDROCK` for Bedrock, otherwise use direct API with `ANTHROPIC_API_KEY`.
  - Use streaming API to collect the response.
  - Retry on rate limit errors, 5xx server errors, and connection errors with exponential backoff (delays: 2s, 30s, 60s; 4 total attempts).
  - When `CachePolicy` is not `CacheNone` and the system prompt text exceeds a model-specific token threshold (2048 tokens for sonnet models, 4096 for opus/haiku, estimated via `len(text)/4`), inject `cache_control` on the last system message block. Use `{"type": "ephemeral"}` for `CacheDefault`, add `{"ttl": "1h"}` for `CacheExtended`. If the API rejects the cache_control, retry without it.

### Prompt Template System

- 10 markdown prompt templates embedded at compile time via `//go:embed`: `assessment_system`, `assessment_user`, `refinement_system`, `refinement_user`, `generation_system`, `generation_user_base`, `generation_user_requirements`, `generation_user_test_spec`, `generation_user_tasks`, `repair_user`.
- `LoadPrompt(name string, projectDir string) (string, error)` — load a prompt template. Validate `name` against `[a-zA-Z0-9_-]+` to prevent path traversal. Check `<projectDir>/.spec/prompts/<name>.md` first; if it exists and is not a symlink, use it. Otherwise use the embedded default. Strip YAML frontmatter (content between `---` delimiters at the start of the file). Return the content.
- `LoadPromptTemplate(name string, projectDir string, vars map[string]string) (string, error)` — load the template via `LoadPrompt`, then substitute `$variable` references with values from `vars`. Unmatched variables pass through unchanged (safe substitute behavior).
- `AssessmentSystemPrompt(projectDir string) (string, error)` — load `assessment_system` template.
- `AssessmentUserPrompt(prdText, specName, projectDir string, specLandscape []map[string]any) (string, error)` — load `assessment_user` template with variables: `$spec_name`, `$spec_landscape_block`, `$prd_text`.
- `RefinementSystemPrompt(projectDir string) (string, error)` — load `refinement_system` template.
- `RefinementUserPrompt(prdText string, answers map[string]string, prevAssessment Assessment, projectDir string, specLandscape []map[string]any) (string, error)` — load `refinement_user` template with variables: `$prd_text`, `$assessment_block`, `$qa_block`, `$spec_landscape_block`.
- `GenerationSystemPrompt(projectDir string) (string, error)` — load `generation_system` template.
- `GenerationUserPrompt(prdText, artifactName, specID, projectDir string, priorArtifacts map[string]any, dependentInterfaces []map[string]any, specLandscape []map[string]any) (string, error)` — load `generation_user_base` template and the artifact-specific template (`generation_user_requirements`, `generation_user_test_spec`, or `generation_user_tasks`), compose with variables.
- `RepairUserPrompt(artifactName string, originalContent map[string]any, errors []string, projectDir string) (string, error)` — load `repair_user` template with variables.
- `DetectProjectLanguage(projectDir string) (language, toolHints string)` — scan for manifest files (go.mod, Cargo.toml, package.json, pyproject.toml, Gemfile, build.gradle, pom.xml) and return the detected language name and tooling hints string. Return empty strings if no manifest is found.
- `FormatSpecLandscape(landscape []map[string]any) string` — format spec landscape entries into markdown tables, split into active and archived sections.

### AI Agent Pipeline

- `SpecAgent` struct holding `modelTier` string.
- `NewSpecAgent(modelTier string) *SpecAgent`.
- `(*SpecAgent).AssessPRD(ctx context.Context, prdText, specName string, opts ...AgentOption) (Assessment, error)` — send PRD with assessment system prompt and `submit_assessment` tool (via `AssessmentTools()`). Force tool use via `tool_choice: {"type": "any"}`. Check stop reason before extraction. Extract tool call, parse and validate Assessment. Wrap Anthropic SDK errors as `AgentError` with classified category.
- `(*SpecAgent).RefinePRD(ctx context.Context, prdText string, answers map[string]string, prevAssessment Assessment, opts ...AgentOption) (string, Assessment, error)` — send with refinement tools. Extract `submit_prd_update` for updated PRD text and `submit_assessment` for new assessment. If assessment tool call is missing, make a second API call with only the assessment tool to get it. Return (updatedPRD, newAssessment, error).
- `(*SpecAgent).GenerateArtifacts(ctx context.Context, prdText, specID, specName string, opts ...AgentOption) (map[string]any, error)` — generate three artifacts sequentially: requirements, test_spec, tasks. For each:
  - Build the prompt with prior artifacts context (prior artifacts become progressively richer as each is generated).
  - Call the API with the artifact-specific tool, temperature=0.2.
  - Extract the tool call result.
  - Validate the result by constructing the afspec model type (via JSON unmarshal into the Go struct).
  - Run afspec schema and cross-file validation on the result.
  - If validation fails, run a repair loop (up to 2 attempts): send validation errors and the original content via `RepairUserPrompt`, re-extract, re-validate.
  - Call the `OnArtifact` callback (if set) after each successful artifact.
  - Return a map of artifact name to validated content.
- `AgentOption` functional option type for agent methods. Options include: `WithSpecLandscape([]map[string]any)`, `WithDependentInterfaces([]map[string]any)`, `WithProjectDir(string)`, `WithOnArtifact(func(name string, content any))`.
- Stop reason handling: before extracting tool calls from a response, check the stop reason. If `refusal`, raise `AgentError` with category `refusal`. If `context_window_exceeded`, raise with category `context_window`. If `pause_turn`, raise with category `pause_turn`.
- Error classification: catch Anthropic SDK errors and wrap as `AgentError` with appropriate category (rate_limit, auth, transient, overloaded, input, internal), retryable flag, and HTTP status code.

### Session AI Methods

- `(*SpecSession).Assess(ctx context.Context) (Assessment, error)` — transition to `StateAssessing`, create a `SpecAgent`, load spec landscape from sibling specs, call `AssessPRD`, append result to assessment history, persist state. On error, persist the error as `lastError`.
- `(*SpecSession).Refine(ctx context.Context, answers map[string]string) (Assessment, error)` — transition to `StateRefining`, call `RefinePRD` with current PRD text, user answers, and latest assessment. Update the PRD file on disk with the refined text. Append new assessment to history. Record a QA exchange with assessment_index, answers, and UTC timestamp. Persist state. On error, persist the error but do not record the QA exchange.
- `(*SpecSession).Generate(ctx context.Context) (GenerateResult, error)` — transition to `StateGenerating` immediately. Create a `SpecAgent`. Check for existing artifact files (partial failure recovery): only generate missing artifacts. Call `GenerateArtifacts` with an `OnArtifact` callback that writes each artifact to disk via `afspec.MarshalJSON` and records it in `generatedArtifacts`. After all artifacts are generated, transition to `StateGenerated`. Run `Validate()` and return `GenerateResult` with artifact list, validation result, and any warnings.

## Technical Boundaries

- Go 1.26.5 per current go.mod.
- The Anthropic Go SDK (`github.com/anthropics/anthropic-sdk-go`) is the LLM dependency.
- Prompt templates are embedded via `//go:embed templates/*.md` in the agentspec package.
- All AI methods accept `context.Context` as the first parameter for cancellation and timeout.
- Streaming happens internally within `AICall`; callers receive the complete response.

## Dependencies

| Spec | From Group | To Group | Relationship |
|------|-----------|----------|--------------|
| agentspec_go_core | last | 1 | Uses SpecSession, Assessment, AgentError, tool definitions, config |

