"""Minimal configuration loader for the spec CLI.

Returns a ThemeConfig with hardcoded defaults.
"""

from __future__ import annotations

from spec.ui import ThemeConfig


def load_theme_config() -> ThemeConfig:
    """Return theme configuration with hardcoded defaults.

    # [theme] section removed from config; hardcoded defaults only
    """
    return ThemeConfig()
