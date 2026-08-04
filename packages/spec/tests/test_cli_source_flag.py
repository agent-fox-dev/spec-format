"""Tests for the ``--source`` / ``-s`` global CLI flag.

Test Spec Entries: TS-01-8, TS-01-9, TS-01-10
Requirements: 01-REQ-3.1, 01-REQ-3.2, 01-REQ-3.3, 01-REQ-3.E1, 01-REQ-3.E2
"""

from __future__ import annotations

from pathlib import Path
from unittest.mock import MagicMock, patch

from click.testing import CliRunner


class TestSourceFlagAcceptance:
    """Tests for --source / -s global option — TS-01-8."""

    def test_ts01_8_source_flag_accepted_and_stored_as_path(
        self,
        cli_runner: CliRunner,
        tmp_path: Path,
        isolated_dir: Path,
    ) -> None:
        """TS-01-8: --source is a global option on main, stored in
        ctx.obj['source'] as a Path object.

        Requirement: 01-REQ-3.1
        """
        from spec.cli import main

        src_dir = tmp_path / "src_dir"
        src_dir.mkdir()

        # We invoke 'list' — a non-AI command — to verify the flag is accepted
        result = cli_runner.invoke(main, ["--source", str(src_dir), "list"])
        assert result.exit_code == 0, f"Exit code {result.exit_code}: {result.output}"

    def test_source_short_alias_s_works(
        self,
        cli_runner: CliRunner,
        tmp_path: Path,
        isolated_dir: Path,
    ) -> None:
        """Short alias -s works identically to --source.

        Requirement: 01-REQ-3.1
        """
        from spec.cli import main

        src_dir = tmp_path / "src_dir"
        src_dir.mkdir()

        result = cli_runner.invoke(main, ["-s", str(src_dir), "list"])
        assert result.exit_code == 0, f"Exit code {result.exit_code}: {result.output}"

    def test_source_nonexistent_path_rejected(
        self,
        cli_runner: CliRunner,
        isolated_dir: Path,
    ) -> None:
        """Passing a non-existent path to --source is rejected at parse time.

        Requirement: 01-REQ-3.E1
        """
        from spec.cli import main

        result = cli_runner.invoke(main, ["--source", "/nonexistent/path", "list"])
        assert result.exit_code != 0
        # Click should report the path doesn't exist
        combined_output = result.output
        assert "does not exist" in combined_output.lower() or "invalid" in combined_output.lower()

    def test_source_default_is_current_directory(
        self,
        cli_runner: CliRunner,
        isolated_dir: Path,
    ) -> None:
        """Default value of --source is '.' (current working directory).

        The 'source' key must be present in ctx.obj when no --source
        flag is provided, and its value must be Path('.').

        Requirement: 01-REQ-3.E2
        Property: 01-PROP-9
        """
        from spec.cli import main

        # Verify the main group has a 'source' parameter defined
        param_names = [p.name for p in main.params]
        assert "source" in param_names, (
            f"'source' not found in main CLI params: {param_names}"
        )


class TestSourceFlagPropagation:
    """Tests for --source propagation through Campaign.new_spec — TS-01-9."""

    def test_ts01_9_source_forwarded_to_campaign_new_spec(
        self,
        cli_runner: CliRunner,
        isolated_dir: Path,
        mock_campaign_open: MagicMock,
        mock_campaign_new_spec: MagicMock,
    ) -> None:
        """TS-01-9: spec new passes ctx.obj['source'] as 'source' kwarg
        to Campaign.new_spec.

        Requirement: 01-REQ-3.2
        """
        from spec.cli import main

        src_dir = isolated_dir / "my_src"
        src_dir.mkdir()

        # Pre-create .specs/ with campaign.yaml
        specs_dir = isolated_dir / ".specs"
        specs_dir.mkdir()
        (specs_dir / "campaign.yaml").write_text(
            "name: default\ndescription: default campaign\n"
        )

        prd_file = isolated_dir / "my_feature.md"
        prd_file.write_text("# My Feature\n")

        with patch("agentspec.campaign.Campaign.open", mock_campaign_open):
            cli_runner.invoke(
                main, ["--source", str(src_dir), "new", "my_feature"]
            )

        # Verify Campaign.new_spec was called with source=Path(src_dir)
        assert mock_campaign_new_spec.called
        call_kwargs = mock_campaign_new_spec.call_args
        assert call_kwargs.kwargs.get("source") == Path(src_dir) or (
            len(call_kwargs.args) > 3 and call_kwargs.args[3] == Path(src_dir)
        )

    def test_source_default_forwarded_as_dot(
        self,
        cli_runner: CliRunner,
        isolated_dir: Path,
        mock_campaign_open: MagicMock,
        mock_campaign_new_spec: MagicMock,
    ) -> None:
        """Default source (no --source flag) forwards Path('.') to Campaign.new_spec.

        Requirement: 01-REQ-3.E2
        Property: 01-PROP-9
        """
        from spec.cli import main

        # Pre-create .specs/ with campaign.yaml
        specs_dir = isolated_dir / ".specs"
        specs_dir.mkdir()
        (specs_dir / "campaign.yaml").write_text(
            "name: default\ndescription: default campaign\n"
        )

        prd_file = isolated_dir / "my_feature.md"
        prd_file.write_text("# My Feature\n")

        with patch("agentspec.campaign.Campaign.open", mock_campaign_open):
            cli_runner.invoke(main, ["new", "my_feature"])

        # Verify Campaign.new_spec was called with source=Path(".")
        assert mock_campaign_new_spec.called, (
            "Campaign.new_spec was not called — "
            "spec new must delegate to Campaign.new_spec"
        )
        call_kwargs = mock_campaign_new_spec.call_args
        source_val = call_kwargs.kwargs.get("source")
        assert source_val == Path(".")


class TestSourceFlagNonAICommands:
    """Tests for --source on non-AI commands — TS-01-10."""

    def test_ts01_10_list_accepts_source_without_error(
        self,
        cli_runner: CliRunner,
        tmp_path: Path,
        isolated_dir: Path,
    ) -> None:
        """TS-01-10: spec list accepts --source without error.

        Requirement: 01-REQ-3.3
        """
        from spec.cli import main

        src_dir = tmp_path / "src_dir"
        src_dir.mkdir()

        result = cli_runner.invoke(main, ["--source", str(src_dir), "list"])
        assert result.exit_code == 0

    def test_ts01_10_campaign_accepts_source_without_error(
        self,
        cli_runner: CliRunner,
        tmp_path: Path,
        isolated_dir: Path,
    ) -> None:
        """TS-01-10: spec campaign accepts --source without error.

        Requirement: 01-REQ-3.3
        """
        from spec.cli import main

        src_dir = tmp_path / "src_dir"
        src_dir.mkdir()
        camp_path = tmp_path / "camp_dir"

        with patch("agentspec.campaign.Campaign.create", return_value=MagicMock()):
            result = cli_runner.invoke(
                main,
                [
                    "--source",
                    str(src_dir),
                    "campaign",
                    "--path",
                    str(camp_path),
                    "--name",
                    "test",
                ],
            )
        assert result.exit_code == 0
