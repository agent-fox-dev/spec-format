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

Add a `[spec_tool]` section to your TOML config file:

```toml
[spec_tool]
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

2. Configure your Google Cloud project and region in your TOML config file:

   ```toml
   [spec_tool]
   vertex_project = "my-gcp-project"
   vertex_region = "us-east5"
   ```

3. Authenticate with Google Cloud using Application Default Credentials
   (e.g., `gcloud auth application-default login`).

When Vertex AI is enabled, `ANTHROPIC_API_KEY` is not needed -- authentication
is handled through `google-auth`.

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

### TOML config (current)

The tool reads TOML configuration from two locations, checked in order:

| Location | Scope |
|----------|-------|
| `.agent-fox/config.toml` (relative to working directory) | Project-local |
| `~/.agent-fox/config.toml` | Global (user-wide) |

The first file found is used. Project-local settings take precedence over
global settings.

#### Example config file

```toml
[spec_tool]
model = "ADVANCED"
auth_method = ""
vertex_project = "my-gcp-project"
vertex_region = "us-east5"

[theme]
# CLI display settings (banner, colors)
```

The `[spec_tool]` section controls LLM configuration. The `[theme]` section
controls CLI banner and display settings and is unrelated to the LLM provider.

#### Available `[spec_tool]` fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `model` | string | `"STANDARD"` | Model tier or model ID |
| `auth_method` | string | `""` | Authentication method hint |
| `vertex_project` | string | `""` | Google Cloud project ID for Vertex AI |
| `vertex_region` | string | `""` | Google Cloud region for Vertex AI |

## Configuration Precedence

Model resolution follows a strict 4-step precedence chain. The first source
that provides a value wins:

1. **`AF_SPEC_MODEL` environment variable** -- highest priority. If set, file-based
   model resolution is skipped entirely (auth fields from config files are still
   applied if available).

2. **Explicit `[spec_tool]` section in TOML config** -- if the section exists in
   `.agent-fox/config.toml` or `~/.agent-fox/config.toml`, its values are used
   with no further fallback.

3. **Legacy `~/.af/settings.yaml`** -- checked only when no TOML config provides
   a value. Prints a deprecation warning to stderr.

4. **Hardcoded default** -- `model = "STANDARD"`, which resolves to
   `claude-sonnet-4-6`.

## Environment Variables

| Variable | Purpose |
|----------|---------|
| `ANTHROPIC_API_KEY` | API key for direct Anthropic API access |
| `AF_SPEC_MODEL` | Override model tier (`SIMPLE`, `STANDARD`, `ADVANCED`) or model ID |
| `CLAUDE_CODE_USE_VERTEX` | Set to `1` to route requests through Google Vertex AI |
| `CLAUDE_CODE_USE_BEDROCK` | Set to `1` to route requests through AWS Bedrock |
| `AF_AGENT` | Set to `1` for agent mode (suppresses banner, routes errors to JSON on stdout) |

## Legacy Configuration

Earlier versions used `~/.af/settings.yaml` for configuration. This file is
still read as a fallback, but it is deprecated. If the tool finds settings in
this file, it prints a warning:

```
Deprecation warning: ~/.af/settings.yaml is no longer the preferred config
location. Migrate your [spec_tool] settings to $HOME/.agent-fox/config.toml.
```

To migrate, move your `spec_tool` settings from the YAML format:

```yaml
# ~/.af/settings.yaml (deprecated)
spec_tool:
  model: ADVANCED
  vertex_project: my-gcp-project
  vertex_region: us-east5
```

to the equivalent TOML format:

```toml
# ~/.agent-fox/config.toml
[spec_tool]
model = "ADVANCED"
vertex_project = "my-gcp-project"
vertex_region = "us-east5"
```

After migrating, you can safely delete `~/.af/settings.yaml`.
