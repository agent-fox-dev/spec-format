# Configuration

The `spec` and `agentspec` packages use the Anthropic Claude API for LLM-powered
spec generation. This document covers how to set your API credentials, choose a
model, switch between LLM providers, and manage configuration files.

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

## LLM Providers

### Anthropic API (Direct)

This is the default provider. It requires only the `ANTHROPIC_API_KEY`
environment variable:

```sh
export ANTHROPIC_API_KEY="sk-ant-..."
```

No additional libraries or configuration are needed.

### Google Vertex AI

The Vertex AI SDK dependencies are included with `agentspec` automatically.
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
is handled through `google-auth`.

Note: The `[provider]` section in the TOML config file accepts
`vertex_project` and `vertex_region` fields, but these are not currently
passed to the Anthropic SDK client. Use the `CLOUD_ML_REGION` environment
variable and Application Default Credentials instead.

### AWS Bedrock

The Bedrock SDK dependencies are included with `agentspec` automatically.
To use Claude through AWS Bedrock:

1. Set the environment variable to enable Bedrock:

   ```sh
   export CLAUDE_CODE_USE_BEDROCK=1
   ```

2. Configure your AWS credentials using any standard method (environment
   variables, `~/.aws/credentials`, IAM role, etc.).

When Bedrock is enabled, `ANTHROPIC_API_KEY` is not needed -- authentication
is handled through `boto3` and your AWS credentials.

### Provider precedence

If multiple provider environment variables are set, the first match wins:

1. `CLAUDE_CODE_USE_VERTEX=1` -- use Vertex AI
2. `CLAUDE_CODE_USE_BEDROCK=1` -- use Bedrock
3. Neither set -- use the direct Anthropic API

## Configuration Files

### TOML config

The tool reads TOML configuration from two locations, checked in order:

| Location | Scope |
|----------|-------|
| `.specs/config.toml` (relative to working directory) | Project-local |
| `~/.specs/config.toml` | Global (user-wide) |

The first file found is used. Project-local settings take precedence over
global settings.

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
section controls authentication and cloud provider settings.

#### Available `[model]` fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `model` | string | `"STANDARD"` | Model tier or model ID |

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

## Environment Variables

| Variable | Purpose |
|----------|---------|
| `ANTHROPIC_API_KEY` | API key for direct Anthropic API access |
| `AF_SPEC_MODEL` | Override model tier (`SIMPLE`, `STANDARD`, `ADVANCED`) or model ID |
| `CLAUDE_CODE_USE_VERTEX` | Set to `1` to route requests through Google Vertex AI |
| `CLOUD_ML_REGION` | Google Cloud region for Vertex AI (required when Vertex is enabled) |
| `CLAUDE_CODE_USE_BEDROCK` | Set to `1` to route requests through AWS Bedrock |
| `AF_AGENT` | Set to `1` for agent mode (suppresses banner, routes errors to JSON on stdout) |

