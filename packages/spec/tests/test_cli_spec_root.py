"""Tests for default spec root directory and --spec-dir / SPEC_DIR override.

Test Spec Entries: TS-01-1, TS-01-2, TS-01-3
Requirements: 01-REQ-1.1, 01-REQ-1.2, 01-REQ-1.3, 01-REQ-1.E1
"""

from __future__ import annotations

import json
from pathlib import Path

import pytest
from click.testing import CliRunner


class TestDefaultSpecDir:
    """Tests for _DEFAULT_SPEC_DIR constant — TS-01-1."""

    def test_ts01_1_default_spec_dir_is_dotspecs(self) -> None:
        """TS-01-1: _DEFAULT_SPEC_DIR equals '.specs' (not legacy '.spec/specs').

        Requirement: 01-REQ-1.1
        """
        from spec.cli import _DEFAULT_SPEC_DIR

        assert _DEFAULT_SPEC_DIR == ".specs"


class TestSpecDirResolution:
    """Tests for spec_dir resolution — TS-01-2, TS-01-3."""

    def test_ts01_2_default_spec_dir_in_output(
        self,
        cli_runner: CliRunner,
        isolated_dir: Path,
    ) -> None:
        """TS-01-2: spec list without --spec-dir resolves spec_dir to '.specs'.

        Requirement: 01-REQ-1.2
        """
        from spec.cli import main

        result = cli_runner.invoke(main, ["list"])
        assert result.exit_code == 0, f"Exit code {result.exit_code}: {result.output}"
        output = json.loads(result.output)
        assert output["spec_dir"] == ".specs"

    def test_ts01_3_spec_dir_flag_overrides_default(
        self,
        cli_runner: CliRunner,
        tmp_path: Path,
    ) -> None:
        """TS-01-3: --spec-dir overrides _DEFAULT_SPEC_DIR.

        Requirement: 01-REQ-1.3
        """
        from spec.cli import main

        custom_path = str(tmp_path / "custom_specs")
        result = cli_runner.invoke(main, ["--spec-dir", custom_path, "list"])
        assert result.exit_code == 0, f"Exit code {result.exit_code}: {result.output}"
        output = json.loads(result.output)
        assert output["spec_dir"] == custom_path

    def test_spec_dir_env_var_overrides_default(
        self,
        cli_runner: CliRunner,
        tmp_path: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """SPEC_DIR env var overrides _DEFAULT_SPEC_DIR.

        Requirement: 01-REQ-1.3
        """
        from spec.cli import main

        custom_path = str(tmp_path / "env_specs")
        monkeypatch.setenv("SPEC_DIR", custom_path)
        result = cli_runner.invoke(main, ["list"])
        assert result.exit_code == 0, f"Exit code {result.exit_code}: {result.output}"
        output = json.loads(result.output)
        assert output["spec_dir"] == custom_path

    def test_spec_dir_flag_takes_precedence_over_env_var(
        self,
        cli_runner: CliRunner,
        tmp_path: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """CLI --spec-dir flag takes precedence over SPEC_DIR env var.

        Requirement: 01-REQ-1.E1
        """
        from spec.cli import main

        flag_path = str(tmp_path / "flag_specs")
        env_path = str(tmp_path / "env_specs")
        monkeypatch.setenv("SPEC_DIR", env_path)
        result = cli_runner.invoke(main, ["--spec-dir", flag_path, "list"])
        assert result.exit_code == 0, f"Exit code {result.exit_code}: {result.output}"
        output = json.loads(result.output)
        assert output["spec_dir"] == flag_path
