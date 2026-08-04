"""Tests for agentspec configuration loading."""

from __future__ import annotations

from pathlib import Path

import pytest
from hypothesis import HealthCheck, given, settings
from hypothesis import strategies as st


class TestConfigLoading:
    """Tests for load_config and AgentSpecConfig."""

    def test_load_from_toml(
        self,
        clean_env: None,
        mock_home: Path,
        config_toml: Path,
    ) -> None:
        """Config loads from .specs/config.toml."""
        config_toml.write_text('[model]\nmodel = "claude-opus-4-6"\n')
        from agentspec.config import load_config

        config = load_config()
        assert config.model == "claude-opus-4-6"

    def test_env_overrides_toml(
        self,
        clean_env: None,
        mock_home: Path,
        config_toml: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Environment variables override config file values."""
        config_toml.write_text('[model]\nmodel = "claude-opus-4-6"\n')
        monkeypatch.setenv("AF_SPEC_MODEL", "claude-haiku-4-5")
        from agentspec.config import load_config

        config = load_config()
        assert config.model == "claude-haiku-4-5"

    def test_config_fields(self) -> None:
        """AgentSpecConfig has required fields."""
        from agentspec.config import AgentSpecConfig

        config = AgentSpecConfig()
        assert hasattr(config, "model")

    def test_defaults(
        self,
        clean_env: None,
        mock_home: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Default config values when no config file or env vars."""
        monkeypatch.setattr(Path, "cwd", lambda: mock_home)
        from agentspec.config import load_config

        config = load_config()
        assert config.model == "STANDARD"

    def test_all_config_fields(
        self,
        clean_env: None,
        mock_home: Path,
        config_toml: Path,
    ) -> None:
        """All [model] and [provider] fields are loaded."""
        config_toml.write_text(
            "[model]\n"
            'model = "ADVANCED"\n\n'
            "[provider]\n"
            'auth_method = "vertex"\n'
            'vertex_project = "my-project"\n'
            'vertex_region = "us-east5"\n'
        )
        from agentspec.config import load_config

        config = load_config()
        assert config.model == "ADVANCED"
        assert config.auth_method == "vertex"
        assert config.vertex_project == "my-project"
        assert config.vertex_region == "us-east5"

    def test_global_config_fallback(
        self,
        clean_env: None,
        mock_home: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Falls back to ~/.specs/config.toml when no local config."""
        monkeypatch.setattr(Path, "cwd", lambda: mock_home / "no-local")
        global_dir = mock_home / ".specs"
        global_dir.mkdir()
        (global_dir / "config.toml").write_text('[model]\nmodel = "SIMPLE"\n')
        from agentspec.config import load_config

        config = load_config()
        assert config.model == "SIMPLE"

    def test_local_takes_precedence_over_global(
        self,
        clean_env: None,
        tmp_path: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Local .specs/config.toml wins over ~/.specs/config.toml."""
        home_dir = tmp_path / "home"
        home_dir.mkdir()
        project_dir = tmp_path / "project"
        project_dir.mkdir()
        monkeypatch.setattr(Path, "home", lambda: home_dir)
        monkeypatch.setattr(Path, "cwd", lambda: project_dir)

        local_dir = project_dir / ".specs"
        local_dir.mkdir()
        (local_dir / "config.toml").write_text('[model]\nmodel = "local-model"\n')

        global_dir = home_dir / ".specs"
        global_dir.mkdir()
        (global_dir / "config.toml").write_text('[model]\nmodel = "global-model"\n')

        from agentspec.config import load_config

        config = load_config()
        assert config.model == "local-model"


class TestConfigEdgeCases:
    """Edge case tests for configuration loading."""

    def test_invalid_toml(
        self,
        clean_env: None,
        mock_home: Path,
        config_toml: Path,
    ) -> None:
        """Invalid TOML raises ConfigError with file path."""
        config_toml.write_text(":::bad toml")
        from agentspec.config import load_config
        from agentspec.errors import ConfigError

        with pytest.raises(ConfigError) as exc_info:
            load_config()
        assert str(config_toml) in str(exc_info.value)

    def test_missing_section(
        self,
        clean_env: None,
        mock_home: Path,
        config_toml: Path,
    ) -> None:
        """Missing model/provider sections uses defaults."""
        config_toml.write_text("[other_tool]\nkey = 'value'\n")
        from agentspec.config import load_config

        config = load_config()
        assert config.model == "STANDARD"

    def test_unknown_keys(
        self,
        clean_env: None,
        mock_home: Path,
        config_toml: Path,
    ) -> None:
        """Unknown keys in model section are silently ignored."""
        config_toml.write_text(
            '[model]\nunknown_key = "value"\nmodel = "test-model"\n'
        )
        from agentspec.config import load_config

        config = load_config()
        assert config.model == "test-model"


class TestConfigProperties:
    """Property-based tests for configuration."""

    @settings(suppress_health_check=[HealthCheck.function_scoped_fixture])
    @given(
        env_model=st.text(
            min_size=1,
            alphabet=st.characters(categories=("L", "N")),
        ),
    )
    def test_env_always_overrides_toml(
        self,
        env_model: str,
        clean_env: None,
        mock_home: Path,
        config_toml: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Env vars always override config file values."""
        monkeypatch.setenv("AF_SPEC_MODEL", env_model)
        config_toml.write_text('[model]\nmodel = "different-value"\n')
        from agentspec.config import load_config

        config = load_config()
        assert config.model == env_model

    @settings(suppress_health_check=[HealthCheck.function_scoped_fixture])
    @given(iteration=st.integers(min_value=0, max_value=10))
    def test_defaults_consistent(
        self,
        iteration: int,
        clean_env: None,
        mock_home: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Defaults are consistent across invocations."""
        monkeypatch.setattr(Path, "cwd", lambda: mock_home)
        from agentspec.config import load_config

        config = load_config()
        assert config.model == "STANDARD"


class TestConfigSmoke:
    """Integration smoke tests for configuration."""

    def test_config_load_end_to_end(
        self,
        clean_env: None,
        mock_home: Path,
        config_toml: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Full config load from TOML + env var override."""
        config_toml.write_text('[model]\nmodel = "toml-model"\n')
        monkeypatch.setenv("AF_SPEC_MODEL", "env-model")
        from agentspec.config import load_config

        config = load_config()
        assert config.model == "env-model"
