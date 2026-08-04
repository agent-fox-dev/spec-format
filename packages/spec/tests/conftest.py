"""Shared test fixtures for spec CLI tests.

Provides isolated filesystem and Click runner fixtures used across
the test_cli_*.py and test_auto_init_*.py test modules.
"""

from __future__ import annotations

from pathlib import Path
from unittest.mock import MagicMock

import pytest
from click.testing import CliRunner


@pytest.fixture()
def cli_runner() -> CliRunner:
    """Return a Click CliRunner with isolated filesystem disabled.

    Tests that need filesystem isolation should use ``tmp_path`` and
    ``monkeypatch.chdir()`` instead — this keeps the runner simple and
    avoids double-isolation surprises.
    """
    return CliRunner()


@pytest.fixture()
def isolated_dir(tmp_path: Path, monkeypatch: pytest.MonkeyPatch) -> Path:
    """Change CWD to a temporary directory and return it.

    Ensures tests don't pollute the real working directory.
    """
    monkeypatch.chdir(tmp_path)
    return tmp_path


@pytest.fixture()
def mock_campaign_new_spec() -> MagicMock:
    """Return a mock for Campaign.new_spec that returns a fake session.

    The mock prevents actual AI calls and filesystem side-effects from
    Campaign.new_spec / SpecSession._create.
    """
    mock_session = MagicMock()
    mock_session.state.value = "init"
    mock_session._spec_dir = Path(".specs/01_my_feature")
    mock = MagicMock(return_value=mock_session)
    return mock


@pytest.fixture()
def mock_campaign_open(mock_campaign_new_spec: MagicMock) -> MagicMock:
    """Return a mock for Campaign.open that returns a Campaign with mocked new_spec."""
    mock_camp = MagicMock()
    mock_camp.new_spec = mock_campaign_new_spec
    mock_camp.specs.return_value = []
    mock_open = MagicMock(return_value=mock_camp)
    return mock_open
