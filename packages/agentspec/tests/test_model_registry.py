"""Tests for model registry, resolve_model(), and model_variant config."""

from __future__ import annotations

from pathlib import Path

import pytest
from agentspec.client import (
    _CACHE_TOKEN_THRESHOLDS,
    MODEL_REGISTRY,
    TIER_DEFAULTS,
    resolve_model,
)
from agentspec.errors import ConfigError


class TestModelRegistryEntries:
    """Verify registry contains expected entries."""

    def test_extended_model_registered(self) -> None:
        assert "claude-opus-4-6[1m]" in MODEL_REGISTRY
        entry = MODEL_REGISTRY["claude-opus-4-6[1m]"]
        assert entry.model_id == "claude-opus-4-6[1m]"
        assert entry.variant == "extended"

    def test_tier_defaults_unchanged(self) -> None:
        assert TIER_DEFAULTS["SIMPLE"] == "claude-haiku-4-5"
        assert TIER_DEFAULTS["STANDARD"] == "claude-sonnet-4-6"
        assert TIER_DEFAULTS["ADVANCED"] == "claude-opus-4-6"


class TestResolveModelDirect:
    """Direct model ID lookup."""

    def test_resolve_extended_model_direct(self) -> None:
        assert resolve_model("claude-opus-4-6[1m]") == "claude-opus-4-6[1m]"

    def test_resolve_standard_models(self) -> None:
        assert resolve_model("claude-haiku-4-5") == "claude-haiku-4-5"
        assert resolve_model("claude-sonnet-4-6") == "claude-sonnet-4-6"
        assert resolve_model("claude-opus-4-6") == "claude-opus-4-6"


class TestResolveModelTier:
    """Tier name resolution."""

    def test_tier_defaults(self) -> None:
        assert resolve_model("SIMPLE") == "claude-haiku-4-5"
        assert resolve_model("STANDARD") == "claude-sonnet-4-6"
        assert resolve_model("ADVANCED") == "claude-opus-4-6"

    def test_case_insensitive_tier(self) -> None:
        assert resolve_model("advanced") == "claude-opus-4-6"
        assert resolve_model("Advanced") == "claude-opus-4-6"
        assert resolve_model("aDvAnCeD") == "claude-opus-4-6"
        assert resolve_model("simple") == "claude-haiku-4-5"
        assert resolve_model("standard") == "claude-sonnet-4-6"

    def test_unknown_model_raises(self) -> None:
        with pytest.raises(ConfigError, match="Unknown model"):
            resolve_model("gpt-4-turbo")


class TestVariantAwareTierResolution:
    """Variant-aware tier resolution."""

    def test_advanced_extended_returns_1m(self) -> None:
        assert resolve_model("ADVANCED", variant="extended") == "claude-opus-4-6[1m]"

    def test_advanced_no_variant_returns_default(self) -> None:
        assert resolve_model("ADVANCED") == "claude-opus-4-6"
        assert resolve_model("ADVANCED", variant=None) == "claude-opus-4-6"

    def test_case_insensitive_tier_with_variant(self) -> None:
        assert resolve_model("advanced", variant="extended") == "claude-opus-4-6[1m]"

    def test_unknown_variant_falls_back_to_default(self) -> None:
        assert resolve_model("ADVANCED", variant="nonexistent") == "claude-opus-4-6"
        assert resolve_model("SIMPLE", variant="extended") == "claude-haiku-4-5"


class TestCacheThresholds:
    """Cache token thresholds include the extended model."""

    def test_extended_model_threshold_is_4096(self) -> None:
        assert _CACHE_TOKEN_THRESHOLDS["claude-opus-4-6[1m]"] == 4096


class TestModelVariantConfig:
    """model_variant config field parsing."""

    def test_model_variant_parsed_from_toml(
        self,
        clean_env: None,
        mock_home: Path,
        config_toml: Path,
    ) -> None:
        config_toml.write_text(
            '[model]\nmodel = "ADVANCED"\nmodel_variant = "extended"\n'
        )
        from agentspec.config import load_config

        config = load_config()
        assert config.model_variant == "extended"

    def test_model_variant_defaults_to_none(self) -> None:
        from agentspec.config import AgentSpecConfig

        config = AgentSpecConfig()
        assert config.model_variant is None

    def test_model_variant_not_in_toml_remains_none(
        self,
        clean_env: None,
        mock_home: Path,
        config_toml: Path,
    ) -> None:
        config_toml.write_text('[model]\nmodel = "ADVANCED"\n')
        from agentspec.config import load_config

        config = load_config()
        assert config.model_variant is None
