---
spec_id: '01'
spec_name: cli_campaign_config
title: CLI Campaign Simplification and Config Restructuring
status: draft
created_at: '2026-08-04T12:43:43.477619+00:00'
updated_at: '2026-08-04T12:49:32.372383+00:00'
owner: spec-format
source: docs/prds/prd1.md
schema_version: 1
---
# CLI Campaign Simplification and Config Restructuring

## Intent

The current CLI has an inconsistent interface for managing specs with and
without campaigns, and the config file layout mixes concerns (theme styling
alongside LLM provider settings). This spec unifies campaign handling by
introducing a "default campaign" that is auto-created on first use, simplifies
the `spec campaign` subcommand surface, adds a `--source` flag for controlling
the AI context directory, and restructures the config file with clear
`[provider]` and `[model]` sections.

## Background

The existing CLI grew organically, resulting in two pain points that motivate
this spec:

1. **Confusing campaign workflow.** Users creating their first spec are forced
   to run `spec campaign create` before `spec new`, even though most projects
   only ever need a single campaign. The `spec campaign open` command further
   conflates "listing specs" with "opening a campaign", causing confusion about
   what the command actually does.

2. **Mixed concerns in config.** The `[spec_tool]` section conflates LLM
   provider authentication details with model selection, and the presence of a
   `[theme]` section in a developer-facing config file is unexpected. Splitting
   these into `[provider]` and `[model]` makes the config easier to understand
   and document.

The `.spec/` → `.specs/` rename and the removal of `[theme]` from the config
file are breaking changes for existing users. No automated migration is
provided. Communication of these changes is handled exclusively via
changelog/release notes outside the codebase — the CLI itself emits no
warnings when a legacy `.spec/` directory is detected. A compatibility shim,
detection log, or any other in-process migration mechanism is explicitly out
of scope by design (not merely deferred).

## Goals

1. Make the default workflow (creating specs without explicit campaign setup)
   seamless by auto-initializing a default campaign.
2. Simplify the `spec campaign` subcommand by removing `open` and `new-spec`
   actions, and collapsing `create` into `spec campaign` directly.
3. Introduce `--source` as a global CLI flag to set the working directory for
   source code analysis during AI operations.
4. Add `spec list` to replace `spec campaign open`.
5. Restructure config from `[spec_tool]` to `[provider]` + `[model]` and remove
   the `[theme]` section.
6. Move config and spec root from `.spec/` to `.specs/`.

## Non-Goals

- Automated migration from `.spec/` to `.specs/` directory structure.
- CLI warnings, deprecation notices, compatibility shims, or detection logs
  when a legacy `.spec/` directory is detected (communication is handled via
  changelog/release notes only; no in-process migration mechanism of any kind).
- Merging project-local and global config files (first-found-wins stays).
- Changing the spec format itself (only the CLI and config are affected).
- Establishing a new test coverage threshold (existing test patterns in the
  repo are followed for all new behavior).

## Definitions

- **spec_root**: The directory where `spec` looks for and writes specs. New
  default: `.specs/` (relative to the current working directory). Overridable
  with `--spec-dir` / `-d` / `SPEC_DIR` env var.
- **context_dir**: The working directory used for source code analysis during
  AI-driven operations (`spec new`, `spec refine`, `spec generate`). Default:
  current working directory. Overridable with `--source` / `-s`.

## Tech Stack

- Python 3.11+
- Click (CLI framework)
- TOML (config format, read with `tomllib`)
- uv workspace (monorepo with packages: `afspec`, `agentspec`, `spec`)
- pytest + pytest-asyncio (test framework and async test support)
- ruff (linting and formatting)

## Changes

### 1. Default Spec Root Changes

The default spec root changes from `.spec/specs` to `.specs/`.

- `_DEFAULT_SPEC_DIR` in `packages/spec/spec/cli.py` changes from `".spec/specs"`
  to `".specs"`.
- `--spec-dir` / `-d` global option and `SPEC_DIR` env var continue to work.

### 2. Auto-Init Default Campaign

When `spec new` is called without `--spec-dir` and the `.specs/` directory does
not exist in the current working directory:

1. Create the `.specs/` directory.
2. Write a `campaign.yaml` inside it with `name: "default"` and
   `description: "default campaign"`.
3. Proceed with spec creation as normal.

This is equivalent to running:
```
spec campaign -p ".specs" -n "default" --description "default campaign"
```

**Edge-case behaviour:**

- If `.specs/` already exists **and** `campaign.yaml` is present inside it:
  skip auto-init entirely.
- If `.specs/` already exists **but** `campaign.yaml` is absent: write
  `campaign.yaml` (do not skip). This ensures the campaign file is always
  present after `spec new` completes.
- If an old `.spec/` directory exists alongside `.specs/`: do not warn or
  interfere — just create/update `.specs/` silently.
- If the `.specs/` directory (or `campaign.yaml`) cannot be written due to a
  permission error: surface a clear error message to the user and abort (do
  not silently swallow the error).
- If the write fails for any other I/O reason (e.g., disk full, partial
  write): surface the OS-level error message to the user and abort. No partial
  state cleanup is attempted — the caller is responsible for retrying.

**Acceptance criteria:**

| Scenario | Expected behaviour |
|----------|--------------------|
| `.specs/` absent | Create `.specs/` and write `campaign.yaml`, then proceed |
| `.specs/` present, `campaign.yaml` present | Skip auto-init, proceed |
| `.specs/` present, `campaign.yaml` absent | Write `campaign.yaml`, proceed |
| `.specs/` is read-only | Abort with user-facing permission error |
| Disk full / other I/O error during write | Abort with OS-level error message |
| Old `.spec/` exists alongside | Ignore `.spec/`, create `.specs/` normally |

### 3. New `--source` / `-s` Global Flag

Add `--source` / `-s` as a global option on the `main` CLI group, alongside
`--spec-dir` and `--quiet`.

- Type: `click.Path(exists=True)` — must point to an existing directory.
- Default: `"."` (current working directory).
- Stored in `ctx.obj["source"]` as a `Path` object.
- Functionally consumed by AI-driven commands: `spec new`, `spec refine`,
  `spec generate`. Other commands accept the flag but ignore it.
- Passed to `Campaign.new_spec` at call time via the new `source` parameter
  (see call chain below). `Campaign.new_spec` forwards `source` to
  `SpecSession._create`, which stores it as `self._source` on the session.
  The path is then used internally by `assess`, `refine`, and `generate` when
  invoking the AI agent — no per-method parameter change is required.

**Call chain for `spec new`:**

```
CLI (spec new)
  → Campaign.new_spec(spec_name, prd, mode, source=ctx.obj["source"])
      → SpecSession._create(spec_dir, mode, source=source)
          → self._source stored on session
```

This keeps the CLI free of `SpecSession._create` internals; the CLI only
interacts with `Campaign.new_spec`.

### 4. CLI Command Changes

#### Remove `spec campaign open` and `spec campaign new-spec`

Remove the `open` and `new-spec` actions from the `campaign` command's
`click.Choice`. The `campaign` command retains only the `create` behavior.

#### `spec campaign create` becomes `spec campaign`

The `campaign` command drops the `ACTION` argument entirely. It becomes a
simple command (not a choice dispatcher) that creates a campaign. The
signature becomes:

```
spec campaign --path PATH --name NAME [--description TEXT]
```

- `--path` / `-p`: Campaign directory path (required, no default). This is
  independent of the global `--spec-dir` flag — `--path` sets where the new
  campaign is created, while `--spec-dir` controls where `spec` reads/writes
  specs for other commands. The two flags do not interact; if a user passes
  both, `--path` governs campaign creation and `--spec-dir` governs spec
  resolution for other commands.
- `--name` / `-n`: Campaign name (required).
- `--description`: Campaign description (default: empty string).

This command is for creating campaigns at non-default locations only. The
default campaign at `.specs/` is auto-created by `spec new`.

**Conflict handling:** If `campaign.yaml` already exists at the given `--path`,
the command exits with a non-zero status and the error message
`"campaign already exists at <path>"`. This preserves the existing
`Campaign.create()` behavior — no library change is required.

#### New `spec list` command

Add a `list` subcommand to the `main` group that replaces `spec campaign open`.
It lists all specs in the spec root directory.

Output format (JSON to stdout):

```json
{
  "spec_dir": ".specs",
  "specs": [
    {"name": "01_my_feature", "state": "generated"},
    {"name": "02_auth_flow", "state": "assessing"}
  ]
}
```

Each entry shows:
- `name`: The full directory name (prefix + snake_case name).
- `state`: The current session state from `_session.json` (one of: `init`,
  `assessing`, `refining`, `prd_accepted`, `generating`, `generated`), or
  `"no_session"` if `_session.json` does not exist **or** if `_session.json`
  is present but malformed (invalid JSON or missing the `state` field).

**`spec_dir` path format:** The `spec_dir` value in the JSON output is always
the relative path as configured — matching what was passed via `--spec-dir` or
the default `.specs`. It is never resolved to an absolute path. Callers who
need an absolute path can resolve it against their working directory.

**Filtering:** Only directories whose names match the spec naming pattern
(numeric prefix + snake_case name) are included. The filter uses
`parse_spec_dir_name()` from `afspec.discovery` — the same logic used by
`Campaign.specs()`. Directories that do not match (including `archive/`,
arbitrary subdirectories, and plain files) are silently skipped.

**Error and edge-case behaviour:**

| Scenario | Output | Exit code |
|----------|--------|-----------|
| `spec_root` does not exist | `{"spec_dir": "...", "specs": []}` | 0 |
| `spec_root` exists but is empty | `{"spec_dir": "...", "specs": []}` | 0 |
| Directory does not match spec naming pattern | Silently skipped | 0 |
| `archive/` subdirectory | Silently skipped | 0 |
| `_session.json` absent | `"state": "no_session"` | 0 |
| `_session.json` malformed (invalid JSON or missing `state`) | `"state": "no_session"` | 0 |

### 5. Config Path Changes

Config file paths change from `.spec/config.toml` to `.specs/config.toml`:

| Scope | Old path | New path |
|-------|----------|----------|
| Project-local | `.spec/config.toml` | `.specs/config.toml` |
| Global | `~/.spec/config.toml` | `~/.specs/config.toml` |

`packages/agentspec/agentspec/config.py` (`load_config`) must update its
candidate paths accordingly.

### 6. Config Section Restructuring

Remove:
- `[theme]` — all theme/style options. `ThemeConfig` and `load_theme_config()`
  stay in the code but return hardcoded defaults only. The path-search logic
  is removed entirely from `load_theme_config` — it simply constructs and
  returns `ThemeConfig()` without reading any config file.
- `[spec_tool]` — replaced by `[provider]` and `[model]`.

Add:
- `[model]` — model selection:
  - `model`: Model tier or ID (default: `"STANDARD"`).
- `[provider]` — LLM provider configuration:
  - `auth_method`: Authentication method (e.g. `"vertex"`, `"bedrock"`, `""`).
  - `vertex_project`: Google Cloud project ID for Vertex AI.
  - `vertex_region`: Google Cloud region for Vertex AI.

**Provider-specific notes:**
- **Vertex AI** (`auth_method = "vertex"`): Requires `vertex_project` and
  `vertex_region` in `[provider]`.
- **Bedrock** (`auth_method = "bedrock"`): Relies entirely on ambient AWS
  credential resolution (standard AWS credential chain: env vars,
  `~/.aws/credentials`, IAM role). No extra fields are required or supported
  in `[provider]` for Bedrock.

**Config value precedence for model selection (highest to lowest):**

1. `AF_SPEC_MODEL` environment variable (always wins).
2. `[model].model` in config file.
3. Hardcoded default (`"STANDARD"`).

When both `AF_SPEC_MODEL` and `[model].model` are set, `AF_SPEC_MODEL` takes
precedence. This is the existing behavior, now stated explicitly as a formal
rule.

Example `.specs/config.toml`:

```toml
[model]
model = "ADVANCED"

[provider]
auth_method = "vertex"
vertex_project = "my-gcp-project"
vertex_region = "us-central1"
```

## Design Decisions

1. **Auto-init path is `.specs/` (plural).** Consistent with the new default
   spec root. The old `.spec/` (singular) path is not reused to avoid confusion
   with the legacy directory structure.

2. **`spec list` shows spec name and state.** Minimal output matching the most
   common need — checking which specs exist and their progress. Additional
   details (title, created_at) can be added later if needed.

3. **`--source` applies to `spec new` in addition to `refine` and `generate`.**
   `spec new` runs an AI assessment of the PRD, which benefits from source code
   context for understanding the codebase.

4. **Config field mapping: `[model]` gets `model`, `[provider]` gets the rest.**
   Clean separation of concerns — model selection is independent of provider
   authentication. This mirrors how users think about configuration: "which
   model?" vs "how do I authenticate?"

5. **`[theme]` hardcoded, path-search removed.** `ThemeConfig` and theme
   rendering stay in the codebase with hardcoded defaults. `load_theme_config`
   no longer reads from any config file and has no candidate path list — it
   simply returns `ThemeConfig()`. This avoids a larger refactor while
   achieving the config simplification goal and eliminates the logical
   inconsistency of searching for a file section that is no longer read.

6. **`AF_SPEC_MODEL` env var unchanged; precedence is env var > config > default.**
   Changing env var names is a breaking change with no benefit. The var
   continues to override `[model].model` when both are set. The precedence
   order (env var highest, then config file, then hardcoded default) is now
   an explicit contract.

7. **Old `.spec/` silently ignored, no CLI warning, no compatibility shim.**
   No warning, detection log, or migration mechanism is emitted when `.spec/`
   exists alongside the new `.specs/`. This is an intentional design decision,
   not a deferral. Users who want to migrate can rename manually. Breaking-change
   communication is handled via changelog/release notes only.

8. **`--source` threaded via `Campaign.new_spec`, not directly to `SpecSession._create`.**
   `Campaign.new_spec` gains a `source` parameter and forwards it to
   `SpecSession._create`. This keeps the CLI free of session internals and
   maintains a clean call chain. `source` is stored as `self._source` on the
   session and used by `assess`, `refine`, and `generate` without per-method
   parameter changes.

9. **`AgentSpecConfig` remains a flat dataclass.** The `[provider]` / `[model]`
   split is a config-file concern only. `load_config` reads from the new
   sections but populates the same flat `AgentSpecConfig` fields
   (`model`, `auth_method`, `vertex_project`, `vertex_region`). No structural
   change to the dataclass avoids breaking existing consumers.

10. **`spec list` returns empty JSON (exit 0) for missing/empty spec root.**
    A missing spec root is not an error condition — it simply means no specs
    have been created yet. An empty list communicates this without alarming the
    caller.

11. **`spec list` uses `parse_spec_dir_name()` for filtering.** Reusing the
    same filter function as `Campaign.specs()` ensures consistent behavior
    across the codebase. Non-matching directories (including `archive/`) are
    silently skipped with exit code 0.

12. **`spec list` `spec_dir` is always relative (as configured).** The JSON
    output reflects the path as passed via `--spec-dir` or the default `.specs`.
    Callers who need an absolute path resolve it themselves.

13. **`spec campaign --path` is independent of global `--spec-dir`.** The two
    flags serve different purposes and do not interact. `--path` governs where
    a new campaign is created; `--spec-dir` governs where specs are read/written
    for other commands.

14. **`spec campaign` error message names the conflicting path.** If
    `campaign.yaml` already exists at `--path`, the error message is
    `"campaign already exists at <path>"`. This preserves existing
    `Campaign.create()` semantics without requiring a library change.

15. **Bedrock uses ambient AWS credentials only.** No Bedrock-specific fields
    are added to `[provider]` or `AgentSpecConfig`. The Bedrock SDK resolves
    credentials via the standard AWS credential chain.

16. **No compatibility shim or migration is in scope, by design.** The absence
    of migration logic is an explicit decision, not an oversight. This is
    documented in Non-Goals and repeated here to prevent future scope creep.

17. **Auto-init I/O failures surface OS error and abort.** For failures beyond
    the permission-error case (e.g., disk full, partial write), the OS-level
    error message is surfaced to the user and the command aborts. No partial
    state cleanup is attempted — the user retries after resolving the
    underlying issue.

18. **Test coverage follows existing repo patterns.** No new coverage threshold
    is introduced. All new behavior (auto-init, `spec list`, `--source` flag,
    config restructuring) should be covered following the same unit/integration
    test patterns already established in the repository.

## Verified External API

### `agentspec` (local workspace package, Python)

| Symbol | Module | Signature | Notes |
|--------|--------|-----------|-------|
| `Campaign.create` | `agentspec.campaign` | `(path: Path, name: str, description: str) -> Campaign` | Static method. Raises an error if `campaign.yaml` already exists at `path`. Error message: `"campaign already exists at <path>"`. |
| `Campaign.open` | `agentspec.campaign` | `(path: Path) -> Campaign` | Static method |
| `Campaign.new_spec` | `agentspec.campaign` | `(self, spec_name: str, prd: str \| Path, mode: str = "interactive", source: Path = Path(".")) -> SpecSession` | **Updated:** gains `source` parameter; forwards it to `SpecSession._create`. |
| `Campaign.specs` | `agentspec.campaign` | `(self) -> list[Path]` | Excludes archive/ and non-matching directories |
| `SpecSession._create` | `agentspec.session` | `(spec_dir: Path, mode: str = "interactive", source: Path = Path(".")) -> SpecSession` | Static, private. `source` stored as `self._source`. |
| `SpecSession.assess` | `agentspec.session` | `(self) -> Assessment` | async; uses `self._source` internally |
| `SpecSession.refine` | `agentspec.session` | `(self, answers: dict[str, str]) -> Assessment` | async; uses `self._source` internally |
| `SpecSession.generate` | `agentspec.session` | `(self) -> GenerateResult` | async; uses `self._source` internally; uses spec_dir.parent as spec_root |
| `load_config` | `agentspec.config` | `() -> AgentSpecConfig` | **Updated:** reads `[model]` and `[provider]` from `.specs/config.toml` (was `[spec_tool]` from `.spec/config.toml`). Precedence: `AF_SPEC_MODEL` env var > `[model].model` > hardcoded default `"STANDARD"`. |
| `AgentSpecConfig` | `agentspec.config` | `model: str, auth_method: str, vertex_project: str, vertex_region: str` | Flat dataclass, fields unchanged. `model` now sourced from `[model]` section; `auth_method`, `vertex_project`, `vertex_region` from `[provider]` section. No Bedrock-specific fields. |

### `afspec` (local workspace package, Python)

| Symbol | Module | Signature | Notes |
|--------|--------|-----------|-------|
| `discover_specs` | `afspec.discovery` | `(root: Union[str, Path]) -> list[SpecMeta]` | |
| `parse_spec_dir_name` | `afspec.discovery` | `(name: str) -> tuple[int, str] \| None` | Used by `spec list` to filter non-spec directories |
| `load_spec_landscape` | `afspec.discovery` | `(spec_root, *, include_archive=True, current_spec_id=None) -> list[dict]` | |
| `load_dependent_interfaces` | `afspec.discovery` | `(spec_id: str, spec_root: Path) -> list[dict]` | |

### `spec` CLI (local workspace package, Python)

| Symbol | Module | Signature | Notes |
|--------|--------|-----------|-------|
| `load_theme_config` | `spec.config` | `() -> ThemeConfig` | **Updated:** returns `ThemeConfig()` with hardcoded defaults only. Path-search logic removed entirely — no config file is read. |
| `ThemeConfig` | `spec.ui` | `playful, header, success, error, warning, info, tool, muted` | Pydantic BaseModel; unchanged |
