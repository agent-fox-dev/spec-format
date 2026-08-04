"""Tests for ``load_theme_config`` hardcoded defaults and ``[theme]`` removal.

Test Spec Entry: TS-01-24
Requirements: 01-REQ-7.3, 01-REQ-7.E3
Property: 01-PROP-6 (load_theme_config never reads from the filesystem)
"""

from __future__ import annotations

from pathlib import Path

import pytest
from spec.config import load_theme_config
from spec.ui import ThemeConfig


class TestLoadThemeConfigDefaults:
    """Tests that ``load_theme_config`` returns hardcoded ThemeConfig defaults."""

    def test_ts01_24_load_theme_config_returns_hardcoded_defaults(self) -> None:
        """TS-01-24: ``load_theme_config()`` returns a ``ThemeConfig``
        instance with all fields set to their hardcoded defaults.

        Requirement: 01-REQ-7.3
        Property: 01-PROP-6
        """
        theme = load_theme_config()
        assert isinstance(theme, ThemeConfig)
        assert theme == ThemeConfig()

    def test_load_theme_config_does_not_read_config_files(
        self,
        tmp_path: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """``load_theme_config()`` does NOT read any config file from the
        filesystem. Even when config files with ``[theme]`` sections exist,
        they are ignored.

        Requirement: 01-REQ-7.3
        Property: 01-PROP-6
        """
        # Set up config files that have [theme] sections
        work_dir = tmp_path / "work"
        work_dir.mkdir()
        monkeypatch.setattr(Path, "cwd", staticmethod(lambda: work_dir))

        home_dir = tmp_path / "home"
        home_dir.mkdir()
        monkeypatch.setattr(Path, "home", staticmethod(lambda: home_dir))

        local_config_dir = work_dir / ".specs"
        local_config_dir.mkdir()
        (local_config_dir / "config.toml").write_text(
            "[theme]\nplayful = false\nheader = 'bold red'\n"
        )

        global_config_dir = home_dir / ".specs"
        global_config_dir.mkdir()
        (global_config_dir / "config.toml").write_text(
            "[theme]\nplayful = false\nerror = 'dim white'\n"
        )

        theme = load_theme_config()
        assert isinstance(theme, ThemeConfig)
        # Must return hardcoded defaults, NOT values from config files
        assert theme == ThemeConfig()

    def test_theme_config_has_expected_default_fields(self) -> None:
        """``ThemeConfig`` has all expected fields with correct defaults.

        Requirement: 01-REQ-7.3
        """
        theme = ThemeConfig()
        assert theme.playful is True
        assert theme.header == "bold cyan"
        assert theme.success == "bold green"
        assert theme.error == "bold red"
        assert theme.warning == "bold yellow"
        assert theme.info == "bold blue"
        assert theme.tool == "bold magenta"
        assert theme.muted == "dim"

    def test_load_theme_config_with_theme_section_does_not_error(
        self,
        tmp_path: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Calling ``load_theme_config()`` when config files containing a
        ``[theme]`` section exist does not raise an error; it still returns
        hardcoded defaults.

        Requirement: 01-REQ-7.E3
        """
        work_dir = tmp_path / "work"
        work_dir.mkdir()
        monkeypatch.setattr(Path, "cwd", staticmethod(lambda: work_dir))

        home_dir = tmp_path / "home"
        home_dir.mkdir()
        monkeypatch.setattr(Path, "home", staticmethod(lambda: home_dir))

        config_dir = work_dir / ".specs"
        config_dir.mkdir()
        (config_dir / "config.toml").write_text(
            "[theme]\nplayful = false\n"
            "[model]\nmodel = 'ADVANCED'\n"
        )

        # Should not raise
        theme = load_theme_config()
        assert isinstance(theme, ThemeConfig)
        assert theme == ThemeConfig()
