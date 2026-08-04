"""Theme and banner for the spec CLI."""

from __future__ import annotations

import importlib.metadata
from dataclasses import dataclass, field
from pathlib import Path

from pydantic import BaseModel, ConfigDict
from rich.console import Console
from rich.style import Style
from rich.text import Text
from rich.theme import Theme

# ---------------------------------------------------------------------------
# Theme configuration (loaded from TOML)
# ---------------------------------------------------------------------------


class ThemeConfig(BaseModel):
    model_config = ConfigDict(extra="ignore")

    playful: bool = True
    header: str = "bold cyan"
    success: str = "bold green"
    error: str = "bold red"
    warning: str = "bold yellow"
    info: str = "bold blue"
    tool: str = "bold magenta"
    muted: str = "dim"


# ---------------------------------------------------------------------------
# AppTheme
# ---------------------------------------------------------------------------

_DEFAULT_STYLES = {
    "header": "bold cyan",
    "success": "bold green",
    "error": "bold red",
    "warning": "bold yellow",
    "info": "bold blue",
    "tool": "bold magenta",
    "muted": "dim",
}


def _validate_style(name: str) -> Style:
    try:
        return Style.parse(name)
    except Exception:
        return Style.parse(_DEFAULT_STYLES.get(name, ""))


@dataclass
class AppTheme:
    config: ThemeConfig = field(default_factory=ThemeConfig)
    console: Console = field(init=False)

    def __post_init__(self) -> None:
        styles = {
            "header": _validate_style(self.config.header),
            "success": _validate_style(self.config.success),
            "error": _validate_style(self.config.error),
            "warning": _validate_style(self.config.warning),
            "info": _validate_style(self.config.info),
            "tool": _validate_style(self.config.tool),
            "muted": _validate_style(self.config.muted),
        }
        self.console = Console(
            stderr=True,
            theme=Theme(styles),
            highlight=False,
        )


def create_theme(config: ThemeConfig | None = None) -> AppTheme:
    if config is None:
        config = ThemeConfig()
    return AppTheme(config=config)


# ---------------------------------------------------------------------------
# Banner
# ---------------------------------------------------------------------------

SPEC_ART = r"""
   ___ _ __   ___  ___
  / __| '_ \ / _ \/ __|
  \__ \ |_) |  __/ (__
  |___/ .__/ \___|\___|
      |_|
"""


def render_banner(theme: AppTheme | None = None, *, quiet: bool = False) -> None:
    """Print the spec CLI startup banner to stderr."""
    if quiet:
        return

    try:
        version = importlib.metadata.version("spec")
    except importlib.metadata.PackageNotFoundError:
        version = "dev"

    console = theme.console if theme else Console(stderr=True)

    art_text = Text(SPEC_ART.rstrip(), style="bold cyan")
    console.print(art_text)

    info_parts = [f"spec v{version}"]
    cwd = Path.cwd()
    info_parts.append(str(cwd))

    console.print("  ".join(info_parts), style="dim")
    console.print()
