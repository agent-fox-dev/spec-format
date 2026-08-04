"""Tests for config file path changes: ``.spec/`` to ``.specs/``.

Test Spec Entries: TS-01-20, TS-01-21
Requirements: 01-REQ-6.1, 01-REQ-6.2, 01-REQ-6.E1, 01-REQ-6.E2
"""

from __future__ import annotations

from pathlib import Path

import pytest
from agentspec.config import load_config


class TestConfigPathSearch:
    """Tests for load_config searching .specs/config.toml paths."""

    def test_ts01_20_load_config_reads_global_when_local_absent(
        self,
        tmp_path: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """TS-01-20: ``load_config()`` reads from ``~/.specs/config.toml``
        when ``.specs/config.toml`` does not exist locally.

        Requirement: 01-REQ-6.1
        """
        # Remove AF_SPEC_MODEL so it doesn't override
        monkeypatch.delenv("AF_SPEC_MODEL", raising=False)

        # Set up global config at ~/.specs/config.toml
        home_dir = tmp_path / "home"
        home_dir.mkdir()
        monkeypatch.setattr(Path, "home", staticmethod(lambda: home_dir))

        global_config_dir = home_dir / ".specs"
        global_config_dir.mkdir()
        (global_config_dir / "config.toml").write_text(
            '[model]\nmodel = "ADVANCED"\n'
        )

        # No local .specs/config.toml
        work_dir = tmp_path / "work"
        work_dir.mkdir()
        monkeypatch.setattr(Path, "cwd", staticmethod(lambda: work_dir))

        config = load_config()
        assert config.model == "ADVANCED"

    def test_ts01_21_load_config_returns_defaults_when_no_config(
        self,
        tmp_path: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """TS-01-21: ``load_config()`` returns hardcoded defaults when
        neither ``.specs/config.toml`` nor ``~/.specs/config.toml`` exists.

        Requirement: 01-REQ-6.2
        """
        monkeypatch.delenv("AF_SPEC_MODEL", raising=False)

        home_dir = tmp_path / "home"
        home_dir.mkdir()
        monkeypatch.setattr(Path, "home", staticmethod(lambda: home_dir))

        work_dir = tmp_path / "work"
        work_dir.mkdir()
        monkeypatch.setattr(Path, "cwd", staticmethod(lambda: work_dir))

        config = load_config()
        assert config.model == "STANDARD"
        assert config.auth_method == ""
        assert config.vertex_project == ""
        assert config.vertex_region == ""

    def test_local_config_takes_precedence_over_global(
        self,
        tmp_path: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """When both ``.specs/config.toml`` and ``~/.specs/config.toml``
        exist, the project-local file wins (first-found-wins).

        Requirement: 01-REQ-6.E1
        """
        monkeypatch.delenv("AF_SPEC_MODEL", raising=False)

        # Global config
        home_dir = tmp_path / "home"
        home_dir.mkdir()
        monkeypatch.setattr(Path, "home", staticmethod(lambda: home_dir))
        global_config_dir = home_dir / ".specs"
        global_config_dir.mkdir()
        (global_config_dir / "config.toml").write_text(
            '[model]\nmodel = "GLOBAL_MODEL"\n'
        )

        # Local config
        work_dir = tmp_path / "work"
        work_dir.mkdir()
        monkeypatch.setattr(Path, "cwd", staticmethod(lambda: work_dir))
        local_config_dir = work_dir / ".specs"
        local_config_dir.mkdir()
        (local_config_dir / "config.toml").write_text(
            '[model]\nmodel = "LOCAL_MODEL"\n'
        )

        config = load_config()
        assert config.model == "LOCAL_MODEL"

    def test_legacy_spec_dir_config_is_not_read(
        self,
        tmp_path: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Legacy ``.spec/config.toml`` is not read; defaults are returned
        instead.

        Requirement: 01-REQ-6.E2
        """
        monkeypatch.delenv("AF_SPEC_MODEL", raising=False)

        home_dir = tmp_path / "home"
        home_dir.mkdir()
        monkeypatch.setattr(Path, "home", staticmethod(lambda: home_dir))

        work_dir = tmp_path / "work"
        work_dir.mkdir()
        monkeypatch.setattr(Path, "cwd", staticmethod(lambda: work_dir))

        # Only legacy config exists — should be ignored
        legacy_dir = work_dir / ".spec"
        legacy_dir.mkdir()
        (legacy_dir / "config.toml").write_text(
            '[spec_tool]\nmodel = "LEGACY_MODEL"\n'
        )

        config = load_config()
        assert config.model == "STANDARD"

    def test_load_config_reads_project_local_config(
        self,
        tmp_path: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """``load_config()`` reads project-local config from
        ``.specs/config.toml``.

        Requirement: 01-REQ-6.1
        """
        monkeypatch.delenv("AF_SPEC_MODEL", raising=False)

        home_dir = tmp_path / "home"
        home_dir.mkdir()
        monkeypatch.setattr(Path, "home", staticmethod(lambda: home_dir))

        work_dir = tmp_path / "work"
        work_dir.mkdir()
        monkeypatch.setattr(Path, "cwd", staticmethod(lambda: work_dir))

        local_config_dir = work_dir / ".specs"
        local_config_dir.mkdir()
        (local_config_dir / "config.toml").write_text(
            '[model]\nmodel = "LOCAL_STANDARD"\n'
            '[provider]\nauth_method = "vertex"\n'
        )

        config = load_config()
        assert config.model == "LOCAL_STANDARD"
        assert config.auth_method == "vertex"
