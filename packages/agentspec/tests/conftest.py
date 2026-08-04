"""Shared test fixtures for agentspec tests."""

from __future__ import annotations

import os
import sys
from pathlib import Path

import pytest

# Add this test directory to sys.path so conftest_agent can be imported
# as a top-level module. This avoids needing __init__.py in the test
# directory, which would cause conftest name collisions across packages
# in the monorepo layout.
_test_dir = str(Path(__file__).resolve().parent)
if _test_dir not in sys.path:
    sys.path.insert(0, _test_dir)

from conftest_agent import (
    mock_ai_call,  # noqa: F401
    mock_client,  # noqa: F401
    sample_assessment,  # noqa: F401
    sample_questions,  # noqa: F401
)


@pytest.fixture()
def clean_env(monkeypatch: pytest.MonkeyPatch) -> None:
    """Remove all AF_SPEC_* and ANTHROPIC_API_KEY env vars.

    Ensures test isolation from host environment variables that could
    affect configuration loading or client creation.
    """
    keys_to_remove = [
        k for k in os.environ if k.startswith("AF_SPEC_") or k == "ANTHROPIC_API_KEY"
    ]
    for key in keys_to_remove:
        monkeypatch.delenv(key, raising=False)


@pytest.fixture()
def mock_home(tmp_path: Path, monkeypatch: pytest.MonkeyPatch) -> Path:
    """Patch Path.home() to return tmp_path for config isolation."""
    monkeypatch.setattr(Path, "home", lambda: tmp_path)
    return tmp_path


@pytest.fixture()
def config_toml(tmp_path: Path, monkeypatch: pytest.MonkeyPatch) -> Path:
    """Create .spec/ directory in a temp working dir and return path to config.toml.

    Patches cwd() so load_config() finds the local config file.
    """
    spec_dir = tmp_path / ".spec"
    spec_dir.mkdir(exist_ok=True)
    monkeypatch.setattr(Path, "cwd", lambda: tmp_path)
    return spec_dir / "config.toml"
