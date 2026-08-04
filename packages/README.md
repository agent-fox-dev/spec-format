# Packages

This repository contains the following packages, managed as a
[uv workspace](https://docs.astral.sh/uv/concepts/workspaces/).

| Package | Description | Install |
|---------|-------------|---------|
| **[afspec](afspec/)** | Standalone library for the agent-fox specification format (v1.3). Loads, validates, renders, and mutates spec directories. | `uv pip install -e packages/afspec` |
| **[agentspec](agentspec/)** | AI-powered spec creation library. Drives PRD assessment, refinement, and artifact generation via Claude. | `uv pip install -e packages/agentspec` |
| **[spec](spec/)** | CLI for AI-powered spec creation. Provides the `spec` command. Agent-friendly JSON output. | `uv pip install -e packages/spec` |

TBD
