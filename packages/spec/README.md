# spec

CLI for AI-powered spec creation. Drives a multi-step workflow that turns a
plain-language PRD into a complete, validated specification pack (requirements,
test contracts, tasks, and architecture). Output is JSON on stdout; progress and
errors go to stderr, making the tool suitable for both human and agent use.

## Installation

```bash
pip install "spec @ git+https://github.com/agent-fox-dev/spec-format.git@v1.3.0#subdirectory=packages/spec"
```

## Quick Start

```bash
# 1. Create a spec from a PRD
spec new docs/my-feature.md --name my-feature

# 2. Refine the PRD through iterative AI review
spec refine 01_my_feature

# 3. Generate JSON artifacts (requirements, tests, tasks)
spec generate 01_my_feature
```

## Commands

| Command | Description |
|---------|-------------|
| `spec new PRD_FILE [--name TEXT]` | Create a new spec from a PRD file |
| `spec refine SPEC [--answers TEXT] [--force]` | Assess PRD, submit answers, refine iteratively |
| `spec generate SPEC [--force]` | Generate JSON artifacts from accepted PRD |
| `spec validate [SPEC] [--cross]` | Run schema and cross-file validation checks |
| `spec lint [--all]` | Lint all spec packs for validation errors |
| `spec render SPEC [--combined] [--json]` | Render spec as markdown (or JSON with `--json`) |
| `spec status SPEC` | Query session state (read-only) |
| `spec campaign create\|open\|new-spec` | Manage spec campaigns |

## Global Options

| Option | Description |
|--------|-------------|
| `-d, --spec-dir PATH` | Override spec directory (default: `.spec/specs`) |
| `-q, --quiet` | Suppress progress output on stderr |
| `--version` | Show version and exit |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `ANTHROPIC_API_KEY` | API key for Anthropic (required for `new`, `refine`, `generate`) |
| `AF_SPEC_MODEL` | Override the default model used for AI calls |

## Requirements

- Python 3.12+
- [afspec](../afspec) -- spec format library
- [agentspec](../agentspec) -- AI-powered spec creation library
- click 8.1+, rich 13.0+, pydantic 2.13+
