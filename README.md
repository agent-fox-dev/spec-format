# Spec Format

## Why a spec format?

Natural-language requirements are ambiguous, untestable, and drift from
implementation. The spec format solves this by defining a structured,
machine-readable package that turns design intent into verifiable contracts.
Every requirement maps to a test, every test maps to a task — nothing is
specified without verification and nothing is built without a requirement.

## Overview

A specification package ("spec") is the durable artifact that captures design
intent, acceptance criteria, verification contracts, and implementation plans
for one cohesive feature. Every spec lives in a numbered directory
(`{NN}_{snake_case_name}/`) and contains four required artifacts plus one
optional artifact:

| Artifact | Format | Purpose |
| --- | --- | --- |
| `prd.md` | Markdown + YAML frontmatter | Narrative intent — the "why" and "what." Human-authored. Contains a hashed `## Intent` section that is protected after approval. |
| `requirements.json` | JSON (schema-validated) | What the system must do: EARS acceptance criteria, correctness properties, execution paths, and error handling. |
| `test_spec.json` | JSON (schema-validated) | How each requirement is verified: unit tests, property tests, edge-case tests, and smoke tests with computed coverage. |
| `tasks.json` | JSON (schema-validated) | What work to do, in what order: task groups, subtasks with a state machine, cross-spec dependencies, and requirement-to-test traceability. |
| `architecture.md` | Markdown (free-form) | Optional. Architectural context — modules, interfaces, data models, technology choices. No schema, not cross-validated. |

**Key properties:**

- **EARS patterns.** Requirements use the Easy Approach to Requirements Syntax — six
  structured patterns (`ubiquitous`, `event_driven`, `complex_event`,
  `state_driven`, `unwanted`, `optional`) that produce testable, unambiguous
  acceptance criteria from decomposed fields.
- **Two-layer validation.** Schema validation (per-file, sub-millisecond) plus
  cross-file integrity checks (referential integrity of IDs, requirement-to-test
  coverage, glossary completeness) run on every mutation.
- **Lifecycle.** A spec progresses through `draft → active → sealed`, with
  optional `superseded` and `archived` terminal states. The `## Intent` section
  is hashed at the `draft → active` transition and protected thereafter.
- **Traceability.** Bidirectional links connect every requirement through its
  test spec and task to an executable test, ensuring nothing is specified without
  verification and nothing is built without a requirement.

The full specification — field-level schemas, EARS pattern definitions, ID
formats, validation rules, subtask state machine, and rendering — is at
**[spec-format.md](docs/spec-format.md)**.

## Creating a Spec Package

The `spec` CLI creates a spec package from a PRD in three steps. Run all
commands from within a campaign directory (one containing `campaign.yaml`).

```bash
# 1. Start a new spec from a PRD file.
#    The agent assesses the PRD and returns questions for refinement.
spec new path/to/prd.md --name my_feature

# 2. Refine (repeatable). Answer the agent's questions as JSON.
#    The agent re-assesses until the PRD is ready.
spec refine 01_my_feature --answers answers.json

# 3. Generate the JSON artifacts (requirements, test_spec, tasks).
spec generate 01_my_feature
```

After generation, validate and inspect the result:

```bash
spec validate 01_my_feature        # schema + cross-file integrity checks
spec render 01_my_feature --combined   # render as a single markdown document
spec status 01_my_feature          # show session state
```

## Installation

Install the `spec` CLI:

```bash
curl -fsSL https://raw.githubusercontent.com/agent-fox-dev/spec-format/refs/heads/main/install.sh | sh
```

## Development

The repository is a [uv workspace](https://docs.astral.sh/uv/concepts/workspaces/)
with three packages:

| Package | Description |
|---------|-------------|
| `packages/afspec/` | Standalone library for the spec format (v1.3) |
| `packages/agentspec/` | AI-powered spec creation library |
| `packages/spec/` | CLI for AI-powered spec creation (`spec` command) |


```bash
uv sync                      # install all packages in editable mode
```

| Command | What it does |
|---------|-------------|
| `make check` | Lint + all tests (use before committing) |
| `make test` | All tests |
| `make lint` | Check lint + formatting |
| `make format` | Auto-format code |

Changes are immediately reflected via editable install. To run the local
version explicitly (rather than a globally installed release):

```bash
uv run spec <command>
```

## Using packages as standalone libraries

`afspec` and `agentspec` are designed for reuse outside the CLI. Install either
package directly from git:

```bash
pip install "afspec @ git+https://github.com/agent-fox-dev/spec-format.git@v1.3.0#subdirectory=packages/afspec"
pip install "agentspec @ git+https://github.com/agent-fox-dev/spec-format.git@v1.3.0#subdirectory=packages/agentspec"
```

- **afspec** — load, validate, mutate, and render spec packs. See
  [`packages/afspec/README.md`](packages/afspec/README.md) for the full API
  reference.
- **agentspec** — AI-powered spec creation library. Drives PRD assessment,
  refinement, and artifact generation via Claude (Anthropic API). Provides
  `SpecSession` for managing the full spec lifecycle and `Campaign` for
  organizing related specs. Depends on afspec and the Anthropic SDK.

## Documentation

- [Spec Format Reference](docs/spec-format.md) — field-level schemas, EARS patterns, validation rules, and rendering
- [CLI Reference](docs/cli.md) — commands, flags, and usage
- [Configuration](docs/configuration.md) — LLM provider setup, model selection, and config files
- [afspec API](packages/afspec/README.md) — library API for loading and manipulating spec packs
