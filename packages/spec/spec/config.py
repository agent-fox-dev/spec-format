"""Minimal configuration loader for the spec CLI.

Reads [theme] from .agent-fox/config.toml (local) or
~/.agent-fox/config.toml (global) and returns a ThemeConfig.
"""

from __future__ import annotations

import logging
import tomllib
from pathlib import Path

from spec.ui import ThemeConfig

logger = logging.getLogger(__name__)


def load_theme_config() -> ThemeConfig:
    """Load theme configuration from TOML config files.

    Checks local (.agent-fox/config.toml) then global
    (~/.agent-fox/config.toml). Returns defaults if neither exists.
    """
    candidates = [
        Path.cwd() / ".agent-fox" / "config.toml",
        Path.home() / ".agent-fox" / "config.toml",
    ]

    for path in candidates:
        if not path.exists() or path.is_symlink():
            continue
        try:
            raw = path.read_text(encoding="utf-8")
            data = tomllib.loads(raw)
            theme_data = data.get("theme", {})
            if theme_data:
                return ThemeConfig(**theme_data)
        except Exception:
            logger.debug("Failed to load config from %s", path, exc_info=True)

    return ThemeConfig()
