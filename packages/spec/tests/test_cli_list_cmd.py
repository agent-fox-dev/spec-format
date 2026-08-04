"""Tests for the ``spec list`` command.

Test Spec Entries: TS-01-15, TS-01-16, TS-01-17, TS-01-18, TS-01-19
Requirements: 01-REQ-5.1, 01-REQ-5.2, 01-REQ-5.3, 01-REQ-5.4, 01-REQ-5.5,
              01-REQ-5.E1, 01-REQ-5.E2, 01-REQ-5.E3, 01-REQ-5.E4
Properties: 01-PROP-3, 01-PROP-4, 01-PROP-7
"""

from __future__ import annotations

import json
from pathlib import Path

from click.testing import CliRunner

# ---------------------------------------------------------------------------
# Happy-path tests (subtask 2.1)
# ---------------------------------------------------------------------------


class TestListHappyPaths:
    """Tests for ``spec list`` with valid spec directories."""

    def test_ts01_15_list_scans_spec_dir_and_outputs_json(
        self,
        cli_runner: CliRunner,
        isolated_dir: Path,
    ) -> None:
        """TS-01-15: ``spec list`` scans spec_dir, reads _session.json state,
        and outputs correct JSON with exit code 0.

        Requirement: 01-REQ-5.1
        Property: 01-PROP-3 (spec list always exits 0)
        """
        from spec.cli import main

        specs_dir = isolated_dir / ".specs"
        specs_dir.mkdir()

        # Create two valid spec directories with session files
        spec1 = specs_dir / "01_my_feature"
        spec1.mkdir()
        (spec1 / "_session.json").write_text('{"state": "generated"}')

        spec2 = specs_dir / "02_auth_flow"
        spec2.mkdir()
        (spec2 / "_session.json").write_text('{"state": "assessing"}')

        result = cli_runner.invoke(main, ["list"])
        assert result.exit_code == 0, f"Exit code {result.exit_code}: {result.output}"

        output = json.loads(result.output)
        assert output["spec_dir"] == ".specs"
        names = [s["name"] for s in output["specs"]]
        assert "01_my_feature" in names
        assert "02_auth_flow" in names
        states = {s["name"]: s["state"] for s in output["specs"]}
        assert states["01_my_feature"] == "generated"
        assert states["02_auth_flow"] == "assessing"

    def test_ts01_16_list_returns_empty_when_spec_dir_missing(
        self,
        cli_runner: CliRunner,
        isolated_dir: Path,
    ) -> None:
        """TS-01-16: ``spec list`` returns empty specs list with exit 0 when
        spec_dir does not exist.

        Requirement: 01-REQ-5.2
        Property: 01-PROP-3 (spec list always exits 0)
        """
        from spec.cli import main

        # No .specs/ directory exists at all
        result = cli_runner.invoke(main, ["list"])
        assert result.exit_code == 0, f"Exit code {result.exit_code}: {result.output}"

        output = json.loads(result.output)
        assert output["spec_dir"] == ".specs"
        assert output["specs"] == []

    def test_ts01_17_list_absent_session_json_reports_no_session(
        self,
        cli_runner: CliRunner,
        isolated_dir: Path,
    ) -> None:
        """TS-01-17: ``spec list`` sets state to ``no_session`` when
        ``_session.json`` is absent from a spec directory.

        Requirement: 01-REQ-5.3
        """
        from spec.cli import main

        specs_dir = isolated_dir / ".specs"
        specs_dir.mkdir()
        spec1 = specs_dir / "01_my_feature"
        spec1.mkdir()
        # No _session.json created

        result = cli_runner.invoke(main, ["list"])
        assert result.exit_code == 0, f"Exit code {result.exit_code}: {result.output}"

        output = json.loads(result.output)
        entry = next(s for s in output["specs"] if s["name"] == "01_my_feature")
        assert entry["state"] == "no_session"

    def test_ts01_18_list_spec_dir_is_relative_path(
        self,
        cli_runner: CliRunner,
        isolated_dir: Path,
    ) -> None:
        """TS-01-18: The ``spec_dir`` field in JSON output is the relative
        path as configured, never an absolute path.

        Requirement: 01-REQ-5.4
        Property: 01-PROP-7 (spec_dir output is always the configured relative path)
        """
        from spec.cli import main

        custom_dir = isolated_dir / "my_custom_specs"
        custom_dir.mkdir()

        result = cli_runner.invoke(
            main, ["--spec-dir", "my_custom_specs", "list"]
        )
        assert result.exit_code == 0, f"Exit code {result.exit_code}: {result.output}"

        output = json.loads(result.output)
        assert output["spec_dir"] == "my_custom_specs"
        assert not output["spec_dir"].startswith("/")

    def test_ts01_19_list_skips_non_matching_dirs_and_files(
        self,
        cli_runner: CliRunner,
        isolated_dir: Path,
    ) -> None:
        """TS-01-19: ``spec list`` silently skips directories not matching the
        spec naming pattern (including ``archive/``) and plain files.

        Requirement: 01-REQ-5.5
        Property: 01-PROP-4 (spec list never includes non-spec directories)
        """
        from spec.cli import main

        specs_dir = isolated_dir / ".specs"
        specs_dir.mkdir()

        # Create non-matching entries
        (specs_dir / "archive").mkdir()
        (specs_dir / "not_a_spec").mkdir()
        (specs_dir / "somefile.txt").write_text("hello")

        # Create one valid spec entry
        spec_valid = specs_dir / "01_valid_spec"
        spec_valid.mkdir()
        (spec_valid / "_session.json").write_text('{"state": "init"}')

        result = cli_runner.invoke(main, ["list"])
        assert result.exit_code == 0, f"Exit code {result.exit_code}: {result.output}"

        output = json.loads(result.output)
        names = [s["name"] for s in output["specs"]]
        assert "01_valid_spec" in names
        assert "archive" not in names
        assert "not_a_spec" not in names
        assert "somefile.txt" not in names


# ---------------------------------------------------------------------------
# Edge-case tests (subtask 2.2)
# ---------------------------------------------------------------------------


class TestListEdgeCases:
    """Tests for ``spec list`` edge cases and error handling."""

    def test_list_empty_spec_dir_returns_empty_specs(
        self,
        cli_runner: CliRunner,
        isolated_dir: Path,
    ) -> None:
        """``spec list`` returns empty specs when spec_dir exists but is empty.

        Requirement: 01-REQ-5.2
        """
        from spec.cli import main

        specs_dir = isolated_dir / ".specs"
        specs_dir.mkdir()
        # No subdirectories at all

        result = cli_runner.invoke(main, ["list"])
        assert result.exit_code == 0
        output = json.loads(result.output)
        assert output["specs"] == []

    def test_list_only_nonmatching_dirs_returns_empty_specs(
        self,
        cli_runner: CliRunner,
        isolated_dir: Path,
    ) -> None:
        """Only non-matching directories (e.g. ``archive/``) result in empty
        specs list.

        Requirement: 01-REQ-5.E1
        Property: 01-PROP-4
        """
        from spec.cli import main

        specs_dir = isolated_dir / ".specs"
        specs_dir.mkdir()
        (specs_dir / "archive").mkdir()
        (specs_dir / "random_dir").mkdir()

        result = cli_runner.invoke(main, ["list"])
        assert result.exit_code == 0
        output = json.loads(result.output)
        assert output["specs"] == []

    def test_list_invalid_session_json_reports_no_session(
        self,
        cli_runner: CliRunner,
        isolated_dir: Path,
    ) -> None:
        """``_session.json`` with invalid JSON sets state to ``no_session``.

        Requirement: 01-REQ-5.E2
        """
        from spec.cli import main

        specs_dir = isolated_dir / ".specs"
        specs_dir.mkdir()
        spec1 = specs_dir / "01_broken"
        spec1.mkdir()
        (spec1 / "_session.json").write_text("{invalid json!")

        result = cli_runner.invoke(main, ["list"])
        assert result.exit_code == 0
        output = json.loads(result.output)
        entry = next(s for s in output["specs"] if s["name"] == "01_broken")
        assert entry["state"] == "no_session"

    def test_list_session_json_missing_state_field_reports_no_session(
        self,
        cli_runner: CliRunner,
        isolated_dir: Path,
    ) -> None:
        """``_session.json`` valid JSON but missing ``state`` field sets
        state to ``no_session``.

        Requirement: 01-REQ-5.E3
        """
        from spec.cli import main

        specs_dir = isolated_dir / ".specs"
        specs_dir.mkdir()
        spec1 = specs_dir / "01_no_state"
        spec1.mkdir()
        (spec1 / "_session.json").write_text('{"other_field": "value"}')

        result = cli_runner.invoke(main, ["list"])
        assert result.exit_code == 0
        output = json.loads(result.output)
        entry = next(s for s in output["specs"] if s["name"] == "01_no_state")
        assert entry["state"] == "no_session"

    def test_list_custom_spec_dir_nonexistent_returns_empty(
        self,
        cli_runner: CliRunner,
        isolated_dir: Path,
    ) -> None:
        """``spec list --spec-dir <nonexistent>`` returns empty specs with
        the custom path as ``spec_dir``.

        Requirement: 01-REQ-5.E4
        """
        from spec.cli import main

        result = cli_runner.invoke(
            main, ["--spec-dir", "does_not_exist", "list"]
        )
        assert result.exit_code == 0
        output = json.loads(result.output)
        assert output["spec_dir"] == "does_not_exist"
        assert output["specs"] == []

    def test_list_output_is_valid_json(
        self,
        cli_runner: CliRunner,
        isolated_dir: Path,
    ) -> None:
        """``spec list`` output is always parseable as valid JSON.

        Property: 01-PROP-3 (spec list always exits 0 and emits valid JSON)
        """
        from spec.cli import main

        specs_dir = isolated_dir / ".specs"
        specs_dir.mkdir()
        spec1 = specs_dir / "01_test"
        spec1.mkdir()
        (spec1 / "_session.json").write_text('{"state": "init"}')

        result = cli_runner.invoke(main, ["list"])
        assert result.exit_code == 0
        # Must not raise json.JSONDecodeError
        output = json.loads(result.output)
        assert "spec_dir" in output
        assert "specs" in output
        assert isinstance(output["specs"], list)
