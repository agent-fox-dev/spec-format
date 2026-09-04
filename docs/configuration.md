# Configuration

The spec CLI uses the Anthropic Claude API for LLM-powered spec generation.
This document covers how to set your API credentials, choose a model, switch
between LLM providers, and manage configuration files.

## Quick Start

Set your Anthropic API key and you are ready to go:

```sh
export ANTHROPIC_API_KEY="sk-ant-..."
```

No configuration files are required. With the API key in your environment, the
tool uses `claude-sonnet-4-6` (the STANDARD tier) by default.

## Model Selection

Three model tiers are available, each mapping to a specific Claude model:

| Tier | Model ID | Use case |
|------|----------|----------|
| `SIMPLE` | `claude-haiku-4-5` | Fast, low-cost tasks |
| `STANDARD` | `claude-sonnet-4-6` | Default -- good balance of quality and speed |
| `ADVANCED` | `claude-opus-4-6` | Highest quality, slower and more expensive |

The default tier is `STANDARD`.

### Override via environment variable

Set `AF_SPEC_MODEL` to a tier name or a direct model ID:

```sh
# Use the ADVANCED tier
export AF_SPEC_MODEL=ADVANCED

# Or specify a model ID directly
export AF_SPEC_MODEL=claude-opus-4-6
```

This environment variable has the highest precedence and overrides any
file-based configuration.

### Override via config file

Add a `[model]` section to your TOML config file:

```toml
[model]
model = "ADVANCED"
```

The `model` field accepts either a tier name (`SIMPLE`, `STANDARD`, `ADVANCED`)
or a model ID (`claude-haiku-4-5`, `claude-sonnet-4-6`, `claude-opus-4-6`).
Tier names are matched case-insensitively.

## LLM Providers

### Anthropic API (Direct)

This is the default provider. It requires only the `ANTHROPIC_API_KEY`
environment variable:

```sh
export ANTHROPIC_API_KEY="sk-ant-..."
```

No additional configuration is needed.

### Google Vertex AI

To use Claude through Google Cloud Vertex AI:

1. Set the environment variable to enable Vertex:

   ```sh
   export CLAUDE_CODE_USE_VERTEX=1
   ```

2. Set the `CLOUD_ML_REGION` environment variable (required by the Anthropic
   SDK):

   ```sh
   export CLOUD_ML_REGION="us-east5"
   ```

3. Authenticate with Google Cloud using Application Default Credentials
   (e.g., `gcloud auth application-default login`). The project ID is
   auto-detected from your credentials.

When Vertex AI is enabled, `ANTHROPIC_API_KEY` is not needed -- authentication
is handled through Google Cloud Application Default Credentials.

Note: The `[provider]` section in the TOML config file accepts
`vertex_project` and `vertex_region` fields, but these are not currently
passed to the Anthropic SDK client. Use the `CLOUD_ML_REGION` environment
variable and Application Default Credentials instead.

### AWS Bedrock

To use Claude through AWS Bedrock:

1. Set the environment variable to enable Bedrock:

   ```sh
   export CLAUDE_CODE_USE_BEDROCK=1
   ```

2. Configure your AWS credentials using any standard method (environment
   variables, `~/.aws/credentials`, IAM role, etc.).

When Bedrock is enabled, `ANTHROPIC_API_KEY` is not needed -- authentication
is handled through your AWS credentials.

### Provider precedence

If multiple provider environment variables are set, the first match wins:

1. `CLAUDE_CODE_USE_VERTEX` set (any non-empty value) -- use Vertex AI
2. `CLAUDE_CODE_USE_BEDROCK` set (any non-empty value) -- use Bedrock
3. Neither set -- use the direct Anthropic API

## Configuration Files

### TOML config

The tool reads TOML configuration from two locations, checked in order:

| Location | Scope |
|----------|-------|
| `.specs/config.toml` (relative to working directory) | Project-local |
| `~/.specs/config.toml` | Global (user-wide) |

The first file found is used. Project-local settings take precedence over
global settings. Symlinked config files are rejected for security.

#### Example config file

```toml
[model]
model = "ADVANCED"

[provider]
auth_method = ""
vertex_project = "my-gcp-project"
vertex_region = "us-east5"
```

The `[model]` section controls which Claude model is used. The `[provider]`
section stores authentication and cloud provider settings.

#### Available `[model]` fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `model` | string | `"STANDARD"` | Default model tier or model ID for all phases |
| `assess_model` | string | *(inherits `model`)* | Override for the PRD assessment phase |
| `refine_model` | string | *(inherits `model`)* | Override for the PRD refinement phase |
| `generate_model` | string | *(inherits `model`)* | Override for the artifact generation phase (and repair) |

Each per-phase field accepts either a tier name (`SIMPLE`, `STANDARD`, `ADVANCED`) or a direct
model ID (e.g. `claude-haiku-4-5`).  When a per-phase field is absent, the phase uses the
top-level `model` value.

##### Cost-optimized example

Assessment calls are short and cheap; generation calls are long and expensive.  Use per-phase
overrides to save cost without sacrificing generation quality:

```toml
[model]
model = "STANDARD"          # default for all phases

assess_model = "SIMPLE"     # assessment is a quick quality check — use the fast model
generate_model = "ADVANCED" # generation produces the spec artifacts — use the best model
```

With this configuration:
- `spec assess` → `claude-haiku-4-5`
- `spec refine` → `claude-sonnet-4-6` (inherits `model`)
- `spec generate` → `claude-opus-4-6`

#### Available `[provider]` fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `auth_method` | string | `""` | Authentication method hint |
| `vertex_project` | string | `""` | Google Cloud project ID (loaded but not currently used by the client) |
| `vertex_region` | string | `""` | Google Cloud region (loaded but not currently used by the client) |

## Configuration Precedence

Model resolution follows a strict 3-step precedence chain. The first source
that provides a value wins:

1. **`AF_SPEC_MODEL` environment variable** -- highest priority. If set, file-based
   model resolution is skipped entirely (auth fields from config files are still
   applied if available).

2. **`[model]` section in TOML config** -- if the section exists in
   `.specs/config.toml` or `~/.specs/config.toml`, its `model` value is used.
   Auth fields are read from the `[provider]` section of the same file.

3. **Hardcoded default** -- `model = "STANDARD"`, which resolves to
   `claude-sonnet-4-6`.

## Prompt Caching

The AI layer supports prompt caching with configurable policies to reduce
latency and cost for repeated system prompts. Three policies are available:

| Policy | Behavior |
|--------|----------|
| `none` | Caching disabled; no `cache_control` metadata is injected |
| `default` | Injects ephemeral `cache_control` when the system prompt exceeds the token threshold (this is the default) |
| `extended` | Same as `default` but with a 1-hour TTL on cached content |

Cache injection is conditional on the estimated token count of the system
prompt exceeding a model-specific threshold:

- **Sonnet models**: 2048 tokens
- **Haiku and Opus models**: 4096 tokens

If a provider rejects `cache_control` metadata (e.g., a 400 error mentioning
`cache_control`), the AI layer automatically retries the request without
caching.

## Environment Variables

| Variable | Purpose |
|----------|---------|
| `ANTHROPIC_API_KEY` | API key for direct Anthropic API access |
| `AF_SPEC_MODEL` | Override model tier (`SIMPLE`, `STANDARD`, `ADVANCED`) or model ID |
| `CLAUDE_CODE_USE_VERTEX` | Set to any non-empty value to route requests through Google Vertex AI |
| `CLOUD_ML_REGION` | Google Cloud region for Vertex AI (required when Vertex is enabled) |
| `CLAUDE_CODE_USE_BEDROCK` | Set to any non-empty value to route requests through AWS Bedrock |
| `AF_AGENT` | Set to `1` for agent mode (suppresses banner, routes errors to JSON on stdout) |

## See Also

- [Model Usage](model-usage.md) — how the spec pipeline selects and resolves
  Claude models across the assess, refine, and generate phases, with details on
  prompt caching thresholds, retry behavior, and known limitations.
