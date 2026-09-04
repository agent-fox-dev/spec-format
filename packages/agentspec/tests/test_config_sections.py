"""Tests for ``[provider]`` and ``[model]`` config section restructuring.

Test Spec Entries: TS-01-22, TS-01-23, TS-01-24, TS-01-25
Requirements: 01-REQ-7.1, 01-REQ-7.2, 01-REQ-7.4,
              01-REQ-7.E1, 01-REQ-7.E2, 01-REQ-7.E3, 01-REQ-7.E4
"""

from __future__ import annotations

import dataclasses
from pathlib import Path

import pytest
from agentspec.config import AgentSpecConfig, load_config


class TestModelAndProviderSections:
    """Tests for ``[model]`` and ``[provider]`` TOML section reads."""

    def test_ts01_22_model_and_provider_sections_read_correctly(
        self,
        tmp_path: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """TS-01-22: ``load_config()`` reads ``model`` from ``[model]``
        section and provider fields from ``[provider]`` section.

        Requirement: 01-REQ-7.1
        """
        monkeypatch.delenv("AF_SPEC_MODEL", raising=False)

        home_dir = tmp_path / "home"
        home_dir.mkdir()
        monkeypatch.setattr(Path, "home", staticmethod(lambda: home_dir))

        work_dir = tmp_path / "work"
        work_dir.mkdir()
        monkeypatch.setattr(Path, "cwd", staticmethod(lambda: work_dir))

        config_dir = work_dir / ".specs"
        config_dir.mkdir()
        (config_dir / "config.toml").write_text(
            '[model]\nmodel = "ADVANCED"\n\n'
            "[provider]\n"
            'auth_method = "vertex"\n'
            'vertex_project = "my-project"\n'
            'vertex_region = "us-central1"\n'
        )

        config = load_config()
        assert config.model == "ADVANCED"
        assert config.auth_method == "vertex"
        assert config.vertex_project == "my-project"
        assert config.vertex_region == "us-central1"

    def test_ts01_23_af_spec_model_overrides_config_file(
        self,
        tmp_path: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """TS-01-23: ``AF_SPEC_MODEL`` env var overrides ``[model].model``
        in the config file.

        Requirement: 01-REQ-7.2
        Property: 01-PROP-5 (AF_SPEC_MODEL always wins)
        """
        home_dir = tmp_path / "home"
        home_dir.mkdir()
        monkeypatch.setattr(Path, "home", staticmethod(lambda: home_dir))

        work_dir = tmp_path / "work"
        work_dir.mkdir()
        monkeypatch.setattr(Path, "cwd", staticmethod(lambda: work_dir))

        config_dir = work_dir / ".specs"
        config_dir.mkdir()
        (config_dir / "config.toml").write_text('[model]\nmodel = "ADVANCED"\n')

        monkeypatch.setenv("AF_SPEC_MODEL", "PREMIUM")

        config = load_config()
        assert config.model == "PREMIUM"

    def test_ts01_24_hardcoded_default_standard(
        self,
        tmp_path: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """TS-01-24 (partial): Hardcoded default ``STANDARD`` is used when
        neither env var nor config key is present.

        Requirement: 01-REQ-7.2
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

    def test_ts01_25_agent_spec_config_has_expected_fields(self) -> None:
        """TS-01-25: ``AgentSpecConfig`` is a flat dataclass with the
        expected model, variant, per-phase, and provider fields.

        Requirement: 01-REQ-7.4
        """
        field_names = {f.name for f in dataclasses.fields(AgentSpecConfig)}
        assert field_names == {
            "model",
            "model_variant",
            "assess_model",
            "refine_model",
            "generate_model",
            "auth_method",
            "vertex_project",
            "vertex_region",
        }


class TestLegacySectionsIgnored:
    """Tests that legacy ``[spec_tool]`` and ``[theme]`` sections are ignored."""

    def test_spec_tool_section_ignored(
        self,
        tmp_path: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """``[spec_tool]`` section in config is ignored; defaults returned.

        Requirement: 01-REQ-7.E2
        """
        monkeypatch.delenv("AF_SPEC_MODEL", raising=False)

        home_dir = tmp_path / "home"
        home_dir.mkdir()
        monkeypatch.setattr(Path, "home", staticmethod(lambda: home_dir))

        work_dir = tmp_path / "work"
        work_dir.mkdir()
        monkeypatch.setattr(Path, "cwd", staticmethod(lambda: work_dir))

        config_dir = work_dir / ".specs"
        config_dir.mkdir()
        # Config with only legacy [spec_tool] section, no [model] or [provider]
        (config_dir / "config.toml").write_text(
            '[spec_tool]\nmodel = "LEGACY_MODEL"\nauth_method = "vertex"\n'
        )

        config = load_config()
        assert config.model == "STANDARD"
        assert config.auth_method == ""

    def test_theme_section_ignored_in_load_config(
        self,
        tmp_path: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """``[theme]`` section in config is ignored by ``load_config``.

        Requirement: 01-REQ-7.E3
        """
        monkeypatch.delenv("AF_SPEC_MODEL", raising=False)

        home_dir = tmp_path / "home"
        home_dir.mkdir()
        monkeypatch.setattr(Path, "home", staticmethod(lambda: home_dir))

        work_dir = tmp_path / "work"
        work_dir.mkdir()
        monkeypatch.setattr(Path, "cwd", staticmethod(lambda: work_dir))

        config_dir = work_dir / ".specs"
        config_dir.mkdir()
        (config_dir / "config.toml").write_text(
            '[theme]\nplayful = false\n\n[model]\nmodel = "ADVANCED"\n'
        )

        config = load_config()
        # model from [model] section is read, [theme] is silently ignored
        assert config.model == "ADVANCED"

    def test_af_spec_model_overrides_when_spec_tool_present(
        self,
        tmp_path: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """``AF_SPEC_MODEL`` env var wins even when ``[spec_tool]`` is present.

        Requirement: 01-REQ-7.E1, 01-REQ-7.E2
        Property: 01-PROP-5
        """
        home_dir = tmp_path / "home"
        home_dir.mkdir()
        monkeypatch.setattr(Path, "home", staticmethod(lambda: home_dir))

        work_dir = tmp_path / "work"
        work_dir.mkdir()
        monkeypatch.setattr(Path, "cwd", staticmethod(lambda: work_dir))

        config_dir = work_dir / ".specs"
        config_dir.mkdir()
        (config_dir / "config.toml").write_text('[spec_tool]\nmodel = "LEGACY_MODEL"\n')

        monkeypatch.setenv("AF_SPEC_MODEL", "ENV_OVERRIDE")

        config = load_config()
        assert config.model == "ENV_OVERRIDE"

    def test_bedrock_auth_method(
        self,
        tmp_path: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """``auth_method = "bedrock"`` in ``[provider]`` is read correctly,
        with ``vertex_project`` and ``vertex_region`` left as empty strings.

        Requirement: 01-REQ-7.E4
        """
        monkeypatch.delenv("AF_SPEC_MODEL", raising=False)

        home_dir = tmp_path / "home"
        home_dir.mkdir()
        monkeypatch.setattr(Path, "home", staticmethod(lambda: home_dir))

        work_dir = tmp_path / "work"
        work_dir.mkdir()
        monkeypatch.setattr(Path, "cwd", staticmethod(lambda: work_dir))

        config_dir = work_dir / ".specs"
        config_dir.mkdir()
        (config_dir / "config.toml").write_text('[provider]\nauth_method = "bedrock"\n')

        config = load_config()
        assert config.auth_method == "bedrock"
        assert config.vertex_project == ""
        assert config.vertex_region == ""
