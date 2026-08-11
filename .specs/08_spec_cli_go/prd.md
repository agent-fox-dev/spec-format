---
spec_id: 08
spec_name: spec_cli_go
title: Spec Cli Go
status: draft
created_at: '2026-08-11T11:29:22.708003+00:00'
updated_at: '2026-08-11T11:29:22.708003+00:00'
owner: ''
source: docs/prds/gospec.md
schema_version: 1
---
# Go spec CLI

## Intent

The `spec` CLI is the user-facing tool for creating and managing specifications. It is currently implemented in Python using Click, requiring Python 3.12+ and `uv` for installation. This spec ports the CLI to Go as a single static binary using cobra, producing a drop-in replacement with the same commands, flags, JSON output shapes, and exit codes. Once validated, the Go binary replaces the Python CLI and the Python packages are deprecated.

## Goals

- Implement all 9 CLI commands (`new`, `list`, `refine`, `generate`, `render`, `validate`, `lint`, `status`, `campaign`) with identical flags and behavior.
- Match the JSON output contract: `{"ok": true, ...}` for success, `{"ok": false, "error": "..."}` for errors in agent mode.
- Support agent mode (`AF_AGENT=1`) with suppressed banner, forced quiet, and JSON error wrapping.
- Provide the same progress feedback (banner on stderr, spinner during AI operations).
- Produce a single static binary distributable for darwin/arm64, darwin/amd64, linux/arm64, linux/amd64.

## Non-goals

- Adding new CLI commands or flags not present in the Python version.
- Rich interactive TUI — the CLI uses simple banner + spinner, not a full TUI framework.
- Supporting package manager installation (brew, apt) — binary distribution only for initial release.
- Backward compatibility with Python CLI sessions — users re-initialize sessions with the Go CLI.

## Functional Requirements

### Root Command and Global Flags

- The binary is named `spec`.
- Global flags: `--spec-dir/-d` (default `.specs`, overridable via `SPEC_DIR` env var), `--source/-s` (default `.`, must be an existing directory), `--quiet/-q` (suppress progress output), `--version` (print version and exit), `--help`.
- When invoked without a subcommand, display help.
- On startup (before subcommand execution), display an ASCII art banner on stderr showing the tool name, version (embedded at build time via `-ldflags`), and current working directory. Suppress the banner when `--quiet` is set, when `AF_AGENT=1` is set, or when the subcommand is one of `validate`, `status`, `list` (JSON-producing commands), or when `--json` appears in the arguments.

### Agent Mode

- When `AF_AGENT=1` environment variable is set: suppress banner, force `--quiet`, and wrap any unhandled error as `{"ok": false, "error": "<message>"}` on stdout before exiting with code 1.
- In agent mode, all structured output goes to stdout as JSON. Progress and diagnostic messages never appear on stdout.

### JSON Output Helpers

- `emit(data any)` — write pretty-printed JSON to stdout. Handle `BrokenPipeError` silently.
- `emitOK(fields ...any)` — wrap data with `{"ok": true, ...}` and emit.
- All commands except `render` (without `--json`) produce JSON on stdout.

### Progress Feedback

- `StatusSpinner` — a context-manager-style type that shows an animated spinner with a message on stderr during long-running operations. Uses a simple spinner animation (e.g., dots or braille characters). Falls back to plain text lines when stderr is not a TTY. All methods are no-ops when `--quiet` is set. Supports `Update(message)` to change the spinner text and `Log(message)` to print a permanent line above the spinner.

### Spec Resolution

- Several commands accept a spec argument that can be a name or a numeric prefix. Implement a resolver that scans the spec directory for matching entries. If the argument is purely numeric, match by prefix number. Otherwise match by directory name. Return an error if zero or multiple matches are found.

### `spec new` Command

- Positional argument: `SPEC_PATH` (required, path to PRD file, must exist).
- Flag: `--name TEXT` (optional, must match `[a-z][a-z0-9_]*`). If omitted, derive snake_case name from the filename.
- Auto-initialize spec directory and default `campaign.yaml` (name: "default", description: "default campaign") if they do not exist.
- Open the campaign, create a new spec via `Campaign.NewSpec`, emit JSON with `spec_dir` and `state`.

### `spec list` Command

- No positional arguments. Always exit 0.
- Scan the spec directory for subdirectories matching the spec naming pattern. For each, read `_session.json` to get the `state` field (default to `"no_session"` if missing or malformed).
- Emit JSON: `{"ok": true, "spec_dir": "<configured dir>", "specs": [{"name": "<dir_name>", "state": "<state>"}]}`.

### `spec refine` Command

- Positional argument: `SPEC` (required, name or number).
- Flag: `--answers TEXT` (JSON file path or `-` for stdin).
- Flag: `--force` (reset session: delete artifact files, reset state to INIT, clear assessment history, QA exchanges, and generated artifacts).
- Without `--answers`: if session needs assessment, run `session.Assess()` with spinner; otherwise return `session.PendingQuestions()`.
- With `--answers`: parse JSON from file or stdin (unwrap `"answers"` key if present), run `session.Refine(answers)` with spinner.
- Emit the assessment result as JSON.

### `spec generate` Command

- Positional argument: `SPEC` (required, name or number).
- Flag: `--force` (delete existing artifact files and regenerate).
- If session state is `ASSESSING` or `REFINING`, auto-accept via `session.AcceptPRD()`.
- Run `session.Generate()` with spinner.
- Emit JSON with list of generated artifacts.

### `spec render` Command

- Positional argument: `SPEC` (required, name or number).
- Flag: `--combined` (single concatenated document).
- Flag: `--json` (JSON envelope). Auto-enabled in agent mode.
- Without `--json`: print raw markdown to stdout. For per-artifact mode, separate artifacts with `---`.
- With `--json`: for combined mode emit `{"ok": true, "format": "combined", "content": "<markdown>"}`. For per-artifact mode emit `{"ok": true, "format": "individual", "artifacts": {"requirements": "...", ...}}`.
- Fall back to rendering available artifacts individually when some files are missing.

### `spec validate` Command

- Optional positional argument: `SPEC` (name or number for single-spec mode; omit for multi-spec).
- Flag: `--cross` (run cross-spec interface consistency checks).
- Flag: `--short` (condensed output: only valid/error_count/warning_count).
- Single-spec mode: check required files exist (prd.md, requirements.json, test_spec.json, tasks.json), verify JSON readability, load spec, run `ValidateStructured`. Emit result.
- Multi-spec mode: discover all specs, validate each, aggregate results.
- Cross-spec mode: discover specs, build dependency graph, run `ValidateCrossSpec`, merge results.
- Exit code 1 if any spec has validation errors.

### `spec lint` Command

- No positional arguments.
- Flag: `--all` (include fully-implemented specs).
- Run `RunLintSpecs(specDir, lintAll)` from the afspec lint module with spinner.
- Emit findings as JSON. Exit code 1 if exit_code is non-zero.

### `spec status` Command

- Positional argument: `SPEC` (required, name or number).
- Resume the session (read-only).
- Emit JSON: `{"ok": true, "state": "...", "has_assessment": bool, "generated_artifacts": [...]}`. Include `last_error` and `quality` when present.

### `spec campaign` Command

- Flags: `--path/-p` (required), `--name/-n` (required), `--description` (optional, default empty).
- Create campaign via `CreateCampaign`. Print confirmation to stderr on success.
- Exit 1 with error to stderr on `CampaignError`.

### Build and Distribution

- The binary version is set at build time via `-ldflags "-X main.version=..."`.
- `CGO_ENABLED=0` for static compilation.
- Build targets: darwin/arm64, darwin/amd64, linux/arm64, linux/amd64.
- The `install.sh` script is updated to download and install the Go binary instead of using `uv tool install`.

## Technical Boundaries

- Go 1.26.5 per current go.mod.
- cobra is the CLI framework (already in go.mod).
- The binary lives at `golang/cmd/spec/`.
- JSON output uses `encoding/json` with `json.MarshalIndent` (2-space indent).
- Spinner uses a lightweight library or custom implementation — no bubbletea dependency.
- All AI-calling commands pass `context.Context` with signal handling for graceful cancellation on SIGINT/SIGTERM.

## Dependencies

| Spec | From Group | To Group | Relationship |
|------|-----------|----------|--------------|
| afspec_go_validation | last | 1 | CLI validate command uses full validation pipeline |
| afspec_go_mutation_lint | last | 1 | CLI lint command uses lint module |
| agentspec_go_core | last | 1 | CLI uses Campaign, SpecSession, config |
| agentspec_go_ai | last | 1 | CLI AI commands use session Assess/Refine/Generate |

