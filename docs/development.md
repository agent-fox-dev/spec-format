# Development

## Prerequisites

- Python 3.12 or later
- Go 1.26.5 or later
- [uv](https://docs.astral.sh/uv/) (Python package manager)

## Repository layout

```
specification/schemas/    # Canonical JSON schemas (source of truth)
packages/
  afspec/                 # Core spec-format library (models, validation, EARS, discovery)
  agentspec/              # AI-powered spec creation (session state machine, Claude API)
  spec/                   # CLI entry point (click + rich)
golang/                   # Go implementation (generated types, validation, lifecycle)
testdata/                 # Shared test fixtures
skills/                   # Agent skills (Claude Code SKILL.md files)
docs/                     # Documentation and ADRs
```

The root `pyproject.toml` defines a uv workspace with all three Python packages
as members. The Go module lives in `golang/` under module path
`github.com/agent-fox-dev/spec-format`.

## Setup

```bash
# Clone the repo
git clone <repo-url> && cd spec-format

# Install Python dependencies (creates a virtualenv automatically)
uv sync

# Go dependencies are fetched on first build
cd golang && go mod download
```

## Common tasks

All tasks are driven through `make`. Run from the repository root.

| Command              | Description                                                              |
|----------------------|--------------------------------------------------------------------------|
| `make check`         | Run lint and all tests (run this before committing)                      |
| `make test`          | Run all tests: `uv run pytest -q` and `go test ./... -count=1`          |
| `make test-fast`     | Run non-slow Python tests (`-m "not slow"`) plus Go tests                |
| `make lint`          | Lint Python (ruff check + ruff format --check) and Go (gofmt + go vet)  |
| `make format`        | Auto-format Python (ruff format) and Go (gofmt -w)                      |
| `make clean`         | Remove caches and build artifacts                                        |
| `make json-gen`      | Copy schemas to packages and regenerate Go types (see below)             |
| `make install-skills`| Install agent skills from `skills/` to `~/.claude/skills/`               |
| `make uninstall-skills`| Remove installed agent skills                                          |

## Python development

### Package overview

- **afspec** -- core library. Pydantic models for every spec-format file
  (`prd.md` frontmatter, `requirements.json`, `tasks.json`, `test_spec.json`).
  Includes validation, EARS criterion builders, rendering, discovery, lifecycle
  state machines, and coverage analysis.
- **agentspec** -- AI session layer. Manages multi-turn spec creation campaigns
  using the Claude API, with prompt templates and tool definitions.
- **spec** -- CLI. Built with click and rich. Thin wrapper that wires afspec and
  agentspec together behind user-facing commands.

### Running tests

```bash
# All Python tests
uv run pytest -q

# Skip slow tests (property tests, integration tests)
uv run pytest -m "not slow" -q

# Single test file
uv run pytest packages/afspec/tests/test_models.py -q

# Verbose output for a failing test
uv run pytest packages/afspec/tests/test_validation.py -v --tb=short
```

Tests use pytest with:
- **hypothesis** for property-based testing
- **pytest-asyncio** for async test functions

Test fixtures are in each package's `tests/` directory. Shared test data
(sample specs) lives in `testdata/` at the repository root.

### Linting and formatting

```bash
# Check for lint errors (does not modify files)
uv run ruff check packages/
uv run ruff format --check packages/

# Auto-fix formatting
uv run ruff format packages/
```

## Go development

The Go package (`golang/`) provides the same spec-format functionality for Go
consumers: discovery, validation, lifecycle state machines, dependency graphs,
coverage, and EARS parsing.

### Running tests

```bash
cd golang && go test ./... -count=1
```

### Linting

```bash
cd golang && gofmt -l .       # List unformatted files
cd golang && go vet ./...     # Static analysis
cd golang && gofmt -w .       # Auto-format
```

## Schema workflow

The four JSON schemas in `specification/schemas/` are the single source of truth:

- `prd-frontmatter.v1.json`
- `requirements.v1.json`
- `tasks.v1.json`
- `test_spec.v1.json`

After modifying any schema, run:

```bash
make json-gen
```

This copies the schemas into both `packages/afspec/afspec/schemas/` and
`golang/schemas/`, then regenerates Go struct types from the schemas using
`go-jsonschema`. The generated files are `golang/*.v1.go`.

Always run `make check` after regenerating to verify nothing broke.

## Git workflow

- Branch from `main` using `feature/<descriptive-name>`.
- Never commit directly to `main`.
- Use conventional commit messages: `feat:`, `fix:`, `refactor:`, `docs:`,
  `test:`, `chore:`.
- Feature branches are local-only. Only `main` is pushed to the remote.
- Run `make check` before every commit.
