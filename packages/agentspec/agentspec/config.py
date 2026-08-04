"""Configuration loading for agentspec.

Resolves model and auth settings from TOML config files, with
environment variable overrides.

Precedence (highest to lowest):

1. ``AF_SPEC_MODEL`` environment variable
2. ``[model]`` section from ``.specs/config.toml`` (local or global)
3. Hardcoded default ``STANDARD`` (resolved to a model ID by ``resolve_model()``)
"""

from __future__ import annotations

import logging
import os
import tomllib
from dataclasses import dataclass
from pathlib import Path

from agentspec.errors import ConfigError

logger = logging.getLogger(__name__)

_DEFAULT_MODEL = "STANDARD"


@dataclass
class AgentSpecConfig:
    """Resolved configuration for agentspec.

    Attributes:
        model: The Anthropic model to use for spec generation.
        auth_method: Authentication method (e.g. ``"api_key"``, ``"vertex"``).
        vertex_project: Google Cloud project ID for Vertex AI.
        vertex_region: Google Cloud region for Vertex AI.
    """

    model: str = _DEFAULT_MODEL
    auth_method: str = ""
    vertex_project: str = ""
    vertex_region: str = ""


def load_config() -> AgentSpecConfig:
    """Load configuration with environment variable override.

    Checks for a TOML config file at ``.specs/config.toml`` (relative to
    the working directory) or ``~/.specs/config.toml`` (global).  The first
    file found is used — files are not merged.

    Reads ``model`` from the ``[model]`` section and ``auth_method``,
    ``vertex_project``, ``vertex_region`` from the ``[provider]`` section.
    Legacy ``[spec_tool]`` and ``[theme]`` sections are ignored.

    The ``AF_SPEC_MODEL`` environment variable overrides the model from
    any config file.

    Returns:
        A fully resolved :class:`AgentSpecConfig`.

    Raises:
        ConfigError: If a config file contains invalid TOML.
    """
    config = AgentSpecConfig()

    toml_data = _load_toml()
    if toml_data is not None:
        # Read model from [model] section
        model_section = toml_data.get("model")
        if isinstance(model_section, dict) and "model" in model_section:
            config.model = str(model_section["model"])

        # Read provider fields from [provider] section
        provider_section = toml_data.get("provider")
        if isinstance(provider_section, dict):
            if "auth_method" in provider_section:
                config.auth_method = str(provider_section["auth_method"])
            if "vertex_project" in provider_section:
                config.vertex_project = str(provider_section["vertex_project"])
            if "vertex_region" in provider_section:
                config.vertex_region = str(provider_section["vertex_region"])

    env_model = os.environ.get("AF_SPEC_MODEL")
    if env_model:
        config.model = env_model

    return config


def _load_toml() -> dict | None:
    """Load the first config.toml found, or None."""
    candidates = [
        Path.cwd() / ".specs" / "config.toml",
        Path.home() / ".specs" / "config.toml",
    ]
    for path in candidates:
        if not path.exists() or path.is_symlink():
            continue
        try:
            return tomllib.loads(path.read_text(encoding="utf-8"))
        except Exception as exc:
            msg = f"Invalid TOML in {path}: {exc}"
            raise ConfigError(msg) from exc
    return None
