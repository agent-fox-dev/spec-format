"""CLI IO utilities for the spec tool.

Provides JSON output helpers, a simplified Click group with agent-mode
detection and error routing, and an animated status spinner.
"""

from __future__ import annotations

import json
import os
import sys
from typing import Any, Self

import click
from rich.console import Console
from rich.live import Live
from rich.spinner import Spinner
from rich.text import Text

# ---------------------------------------------------------------------------
# JSON output helpers
# ---------------------------------------------------------------------------


def emit(data: dict[str, Any]) -> None:
    """Write a dict as pretty-printed JSON to stdout."""
    try:
        click.echo(json.dumps(data, indent=2, default=str))
    except BrokenPipeError:
        pass


def emit_ok(data: dict[str, Any] | None = None, **kwargs: Any) -> None:
    """Emit a JSON envelope with ``"ok": True``."""
    if data is None:
        data = kwargs
    merged = {**data, "ok": True}
    emit(merged)


# ---------------------------------------------------------------------------
# Click group with agent-mode detection
# ---------------------------------------------------------------------------


class SpecGroup(click.Group):
    """Click group with agent-mode detection and error routing.

    When ``AF_AGENT=1`` is set, forces quiet mode and wraps errors
    as JSON on stdout. Otherwise behaves as a standard Click group.
    """

    def invoke(self, ctx: click.Context) -> None:
        try:
            ctx.ensure_object(dict)
        except (RuntimeError, TypeError):
            ctx.obj = {}

        if not isinstance(ctx.obj, dict):
            ctx.obj = {}

        agent_mode = os.environ.get("AF_AGENT") == "1"
        ctx.obj["agent_mode"] = agent_mode

        quiet = ctx.params.get("quiet", False) or False
        if agent_mode:
            quiet = True
        ctx.obj.setdefault("quiet", quiet)

        try:
            super().invoke(ctx)
        except SystemExit:
            raise
        except click.exceptions.Exit:
            raise
        except click.ClickException as exc:
            if agent_mode:
                emit({"ok": False, "error": exc.format_message()})
                sys.exit(exc.exit_code)
            raise
        except Exception as exc:
            if agent_mode:
                emit({"ok": False, "error": str(exc)})
                sys.exit(1)
            raise


# ---------------------------------------------------------------------------
# Status spinner
# ---------------------------------------------------------------------------


class StatusSpinner:
    """Animated spinner for stderr feedback during long-running operations.

    Use as a context manager. When *quiet* is True, all methods are no-ops.
    When stderr is not a TTY, prints plain text lines instead of animating.
    """

    def __init__(self, message: str, *, quiet: bool = False) -> None:
        self._message = message
        self._quiet = quiet
        self._live: Live | None = None
        self._spinner: Spinner | None = None
        self._console: Console | None = None
        self._is_tty: bool = False

    def __enter__(self) -> Self:
        if self._quiet:
            return self

        self._console = Console(file=sys.stderr, force_terminal=False, no_color=False)
        self._is_tty = self._console.is_terminal

        if self._is_tty:
            self._spinner = Spinner("dots", text=Text(self._message))
            self._live = Live(
                self._spinner,
                console=self._console,
                transient=True,
                refresh_per_second=10,
            )
            self._live.start()
        else:
            self._console.print(self._message, highlight=False)

        return self

    def __exit__(self, *exc: object) -> None:
        if self._quiet:
            return
        if self._live is not None:
            self._live.stop()
            self._live = None
            self._spinner = None
        self._console = None

    def update(self, message: str) -> None:
        """Change the spinner's status message."""
        if self._quiet:
            return
        self._message = message
        if self._is_tty and self._spinner is not None:
            self._spinner.update(text=Text(message))
        elif not self._is_tty and self._console is not None:
            self._console.print(message, highlight=False)

    def log(self, message: str) -> None:
        """Print a permanent line above the spinner."""
        if self._quiet:
            return
        if self._is_tty and self._live is not None:
            self._live.console.print(message, highlight=False)
        elif not self._is_tty and self._console is not None:
            self._console.print(message, highlight=False)
