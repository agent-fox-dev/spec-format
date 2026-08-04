# spec CLI Reference

The `spec` command is the CLI entry point for AI-powered spec creation. It manages the full lifecycle of creating, refining, generating, validating, and rendering specification packs.

## Global Options

| Option | Description |
|--------|-------------|
| `-d, --spec-dir PATH` | Spec directory (default: `.agent-fox/specs`) |
| `-q, --quiet` | Suppress progress output |
| `--version` | Show the version and exit |
| `--help` | Show help and exit |

```
spec [OPTIONS] [COMMAND] [ARGS]...
```

## Commands

### new

Create a new spec from a PRD file. Initializes a numbered spec directory and copies the PRD into it.

```
spec new [OPTIONS] PRD_FILE
```

| Option | Description |
|--------|-------------|
| `--name TEXT` | Spec name (default: derived from filename) |

**Example:**

```bash
spec new docs/my-feature.md
spec new --name auth-redesign docs/auth-prd.md
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
