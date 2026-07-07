# Spec Format

## Why a spec format?

Natural-language requirements are ambiguous, untestable, and drift from
implementation. The spec format solves this by defining a structured,
machine-readable package that turns design intent into verifiable contracts.
Every requirement maps to a test, every test maps to a task — nothing is
specified without verification and nothing is built without a requirement.

The format is the foundation of the [Agent Fox](https://github.com/agent-fox-dev/agent-fox)
toolchain: it is what the `spec` CLI produces, what the libraries validate, and
what AI agents consume to generate and verify code.

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
**[spec-format.md](spec-format.md)**.

## Implementation

The spec format is implemented by three Python packages in the
[`agent-fox`](https://github.com/agent-fox-dev/agent-fox) monorepo (under
`packages/`):

- **[afspec](https://github.com/agent-fox-dev/agent-fox/tree/main/packages/afspec)** —
  standalone library for loading, validating, rendering, and mutating spec
  packages. No AI dependencies.
- **[agentspec](https://github.com/agent-fox-dev/agent-fox/tree/main/packages/agentspec)** —
  AI-powered spec creation. Drives PRD assessment, refinement, and artifact
  generation using Claude models. Depends on afspec.
- **[spec CLI](https://github.com/agent-fox-dev/agent-fox/tree/main/packages/spec_cli)** —
  command-line wrapper around agentspec for working with specs locally.

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
spec status                        # show session state for all specs
```
