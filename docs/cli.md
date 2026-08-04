# spec CLI Reference

The `spec` command is the CLI entry point for AI-powered spec creation. It manages the full lifecycle of creating, refining, generating, validating, and rendering specification packs.

## Global Options

| Option | Description |
|--------|-------------|
| `-d, --spec-dir PATH` | Spec directory (default: `.specs`). Can also be set via `SPEC_DIR` env var; CLI flag takes precedence. |
| `-s, --source PATH` | Source code directory for AI context analysis (default: `.`). Passed to AI-driven commands (`new`, `refine`, `generate`) as the codebase context directory. Non-AI commands accept the flag without error but do not use the value. The path must exist on the filesystem. |
| `-q, --quiet` | Suppress progress output |
| `--version` | Show the version and exit |
| `--help` | Show help and exit |

```
spec [OPTIONS] [COMMAND] [ARGS]...
```

## Commands

### new

Create a new spec. Auto-initialises the spec root directory and a default campaign if they do not already exist, then delegates to Campaign.new_spec.

When the spec root (`.specs/` by default) does not exist, it is created along with a `campaign.yaml` containing `name: default` and `description: default campaign`. If the directory exists but `campaign.yaml` is absent, only the YAML file is written. If both exist, auto-init is skipped (idempotent).

```
spec new [OPTIONS] SPEC_NAME
```

| Option | Description |
|--------|-------------|
| `--prd PATH` | PRD file path (optional) |

**Example:**

```bash
spec new my_feature
spec new my_feature --prd docs/my-feature.md
spec new my_feature --source /path/to/source
```

### list

List all specs in the spec root with their session states. Outputs a JSON object containing `spec_dir` and a `specs` array. Each entry has the directory name and the session state read from `_session.json` (or `"no_session"` if absent or malformed). Always exits with status 0.

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
| `--force` | Discard previous assessments and start a fresh refine cycle |

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

Run schema and cross-file checks on spec packs.

When `SPEC` is given, validates that single spec pack. When omitted, discovers and validates all spec packs in the spec directory.

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

Lint spec packs for validation errors.

Discovers all spec packs in the spec directory and runs afspec validation on each. By default, skips fully-implemented specs.

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

Render a spec pack as markdown for human review.

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
spec status [OPTIONS] SPEC
```

No additional options.

**Example:**

```bash
spec status 01_auth_redesign
```

### campaign

Manage spec campaigns -- groups of related specs in a shared directory with a `campaign.yaml` manifest.

```
spec campaign [OPTIONS] {create|open|new-spec}
```

**Actions:**

| Action | Description |
|--------|-------------|
| `create` | Create a new campaign directory |
| `open` | Open an existing campaign and list its specs |
| `new-spec` | Add a new spec to a campaign |

**Options:**

| Option | Description |
|--------|-------------|
| `-p, --path PATH` | Campaign directory path |
| `-n, --name TEXT` | Campaign or spec name |
| `--description TEXT` | Campaign description |
| `--prd PATH` | PRD file path (for `new-spec`) |

**Example:**

```bash
# Create a campaign
spec campaign create -n "Q3 Auth Overhaul" --description "Auth system redesign" -p campaigns/q3-auth

# List specs in a campaign
spec campaign open -p campaigns/q3-auth

# Add a spec to a campaign
spec campaign new-spec -p campaigns/q3-auth -n session-tokens --prd docs/session-tokens.md
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
| `SPEC_DIR` | Override the default spec root directory (`.specs`). The `--spec-dir` CLI flag takes precedence over this env var. |
