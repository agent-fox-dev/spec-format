# spec CLI Reference

The `spec` command is the CLI entry point for AI-powered spec creation. It manages the full lifecycle of creating, refining, generating, validating, and rendering specs.

## Global Options

| Option | Description |
|--------|-------------|
| `-d, --spec-dir PATH` | Spec directory (default: `.specs`). Can also be set via `SPEC_DIR` env var; CLI flag takes precedence. |
| `-s, --source PATH` | Source code directory for AI context during spec creation (used by `new`). Default: `.`. The path must exist on the filesystem. |
| `-q, --quiet` | Suppress progress output |
| `--version` | Show the version and exit |
| `--help` | Show help and exit |

```
spec [OPTIONS] [COMMAND] [ARGS]...
```

## Commands

### new

Create a new spec from a PRD file. Auto-initialises the spec root directory and a default campaign if they do not already exist.

```
spec new [OPTIONS] SPEC_PATH
```

| Argument / Option | Description |
|-------------------|-------------|
| `SPEC_PATH` | Path to an existing PRD file (required positional argument). The file must exist. |
| `--name TEXT` | Snake-case spec name. When omitted, derived automatically from the PRD filename. Must match `[a-z][a-z0-9_]*`. |

**Example:**

```bash
spec new docs/my-feature.md
spec new docs/my-feature.md --name my_feature
spec new docs/my-feature.md --source /path/to/source
```

### list

List all specs in the spec root with their session states. Outputs a JSON object containing `spec_dir` and a `specs` array. Each entry has the directory name and the session state (or `"no_session"` if absent or malformed). Always exits with status 0.

```
spec list
```

No additional options.

**Example:**

```bash
spec list
spec --spec-dir /path/to/specs list
```

**Output format:**

```json
{
  "spec_dir": ".specs",
  "specs": [
    {"name": "01_my_feature", "state": "generated"},
    {"name": "02_auth_flow", "state": "assessing"}
  ]
}
```

### refine

Assess a PRD for quality and completeness, then iteratively refine it by answering questions.

Without `--answers`, runs the initial assessment and outputs pending questions as JSON. With `--answers`, submits answers, updates the PRD, and outputs the new assessment as JSON. Loop until the assessment quality reaches "ready", then run `generate`.

```
spec refine [OPTIONS] SPEC
```

| Option | Description |
|--------|-------------|
| `--answers TEXT` | JSON file with answers, or `-` to read from stdin |
| `--force` | Reset session to initial state, discarding all assessments, answers, and generated artifacts |

**Example:**

```bash
# Run initial assessment
spec refine 01_auth_redesign

# Submit answers from a file
spec refine --answers answers.json 01_auth_redesign

# Submit answers from stdin
echo '{"q1": "Yes, OAuth2 only"}' | spec refine --answers - 01_auth_redesign

# Force a fresh assessment
spec refine --force 01_auth_redesign
```

### generate

Generate JSON artifacts (requirements.json, test_spec.json, tasks.json) from an accepted PRD.

The PRD must have passed the refine cycle (quality "ready") before generation can proceed.

```
spec generate [OPTIONS] SPEC
```

| Option | Description |
|--------|-------------|
| `--force` | Delete existing artifacts and regenerate from scratch |

**Example:**

```bash
spec generate 01_auth_redesign
spec generate --force 01_auth_redesign
```

### validate

Run schema and cross-file checks on specs.

When `SPEC` is given, validates that single spec. When omitted, discovers and validates all specs in the spec directory.

```
spec validate [OPTIONS] [SPEC]
```

| Option | Description |
|--------|-------------|
| `--cross` | Run cross-spec interface consistency checks |

**Example:**

```bash
# Validate all specs
spec validate

# Validate a single spec
spec validate 01_auth_redesign

# Include cross-spec checks
spec validate --cross
```

### lint

Lint specs for validation errors.

Discovers all specs in the spec directory and runs validation on each. By default, skips fully-implemented specs.

```
spec lint [OPTIONS]
```

| Option | Description |
|--------|-------------|
| `--all` | Include fully-implemented specs |

**Example:**

```bash
spec lint
spec lint --all
```

### render

Render a spec as markdown for human review.

```
spec render [OPTIONS] SPEC
```

| Option | Description |
|--------|-------------|
| `--combined` | Render as single combined document |
| `--json` | Output JSON envelope |

**Example:**

```bash
spec render 01_auth_redesign
spec render --combined 01_auth_redesign
spec render --json 01_auth_redesign
```

### status

Query session state for a spec (read-only). Reports the current lifecycle state of the spec session.

```
spec status SPEC
```

No additional options.

**Example:**

```bash
spec status 01_auth_redesign
```

### campaign

Create a new campaign directory at the specified path, independent of the global `--spec-dir` option.

```
spec campaign [OPTIONS]
```

| Option | Description |
|--------|-------------|
| `-p, --path PATH` | Campaign directory path (required) |
| `-n, --name TEXT` | Campaign name (required) |
| `--description TEXT` | Campaign description (default: empty string) |

**Example:**

```bash
spec campaign --path campaigns/q3-auth --name "Q3 Auth Overhaul" --description "Auth system redesign"
spec campaign -p campaigns/q3-auth -n "Q3 Auth Overhaul"
```

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success (or lint found no errors) |
| `1` | Error or findings exist (e.g., lint/validate found problems) |
| `2` | Usage error (invalid arguments or missing required options) |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `ANTHROPIC_API_KEY` | **Required** for `new`, `refine`, and `generate` commands. Anthropic API key used to call Claude for PRD assessment, refinement, and artifact generation. |
| `AF_SPEC_MODEL` | Override the default Claude model used for AI operations. |
| `AF_AGENT` | Set to `1` to enable agent mode. Suppresses the banner and forces JSON output for `render`. |
| `SPEC_DIR` | Override the default spec root directory (`.specs`). The `--spec-dir` CLI flag takes precedence over this env var. |
| `CLAUDE_CODE_USE_VERTEX` | Set to `1` to use Google Vertex AI as the provider (requires `google-auth` and `anthropic[vertex]`). |
| `CLAUDE_CODE_USE_BEDROCK` | Set to `1` to use AWS Bedrock as the provider (requires `boto3` and `anthropic[bedrock]`). |

---

# afspec CLI Reference

The `afspec` command manages spec artifacts directly. It provides low-level operations for updating subtask states within a spec's `tasks.json`.

```
afspec [COMMAND] [ARGS]...
```

## Commands

### update-subtask

Transition a subtask to a new state. Validates the transition against the subtask lifecycle state machine and persists the updated `tasks.json`.

```
afspec update-subtask SPEC_DIR SUBTASK_ID TARGET_STATE
```

| Argument | Description |
|----------|-------------|
| `SPEC_DIR` | Path to the spec directory containing `tasks.json` |
| `SUBTASK_ID` | Subtask identifier (e.g. `1.1`, `3.2`) |
| `TARGET_STATE` | Target state: `pending`, `queued`, `in_progress`, `done`, `pending_reevaluation`, `dropped` |

**Example:**

```bash
# Mark subtask 1.1 as in progress
afspec update-subtask .specs/01_auth_redesign 1.1 in_progress

# Mark subtask 2.3 as done
afspec update-subtask .specs/01_auth_redesign 2.3 done
```

**Exit codes:**

| Code | Meaning |
|------|---------|
| `0` | Transition succeeded |
| `1` | Error (invalid subtask ID, invalid state transition, load/save failure) |
