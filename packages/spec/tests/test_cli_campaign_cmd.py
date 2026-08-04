"""Tests for the simplified ``spec campaign`` command.

Test Spec Entries: TS-01-11, TS-01-12, TS-01-13, TS-01-14
Requirements: 01-REQ-4.1, 01-REQ-4.2, 01-REQ-4.3, 01-REQ-4.4,
              01-REQ-4.E1, 01-REQ-4.E2, 01-REQ-4.E3
"""

from __future__ import annotations

from pathlib import Path
from unittest.mock import MagicMock, patch

from click.testing import CliRunner


class TestCampaignLegacySubcommandsRemoved:
    """Tests that legacy subcommands are rejected — TS-01-11."""

    def test_ts01_11_campaign_open_rejected(
        self,
        cli_runner: CliRunner,
    ) -> None:
        """TS-01-11: ``spec campaign open`` exits with non-zero and usage error.

        The campaign command must NOT accept 'open' as a valid subcommand
        or positional argument. The error must be a Click usage error, not
        an operational failure from Campaign.open.

        Requirement: 01-REQ-4.1, 01-REQ-4.E2
        """
        from spec.cli import main

        result = cli_runner.invoke(main, ["campaign", "open"])
        assert result.exit_code != 0
        # The error must specifically be about 'open' not being recognised,
        # not about Campaign.open failing operationally.
        # Click usage errors have exit code 2.
        assert result.exit_code == 2, (
            f"Expected Click usage error (exit code 2), got {result.exit_code}. "
            f"Output: {result.output}"
        )

    def test_ts01_11_campaign_new_spec_rejected(
        self,
        cli_runner: CliRunner,
    ) -> None:
        """TS-01-11: ``spec campaign new-spec`` exits with non-zero.

        Must be a Click usage error (exit code 2), not an operational error.

        Requirement: 01-REQ-4.1, 01-REQ-4.E2
        """
        from spec.cli import main

        result = cli_runner.invoke(main, ["campaign", "new-spec"])
        assert result.exit_code == 2, (
            f"Expected Click usage error (exit code 2), got {result.exit_code}. "
            f"Output: {result.output}"
        )

    def test_campaign_no_action_argument_accepted(
        self,
        cli_runner: CliRunner,
    ) -> None:
        """Campaign command no longer accepts an ACTION positional argument.

        Passing 'create' as a positional arg must fail with a Click usage
        error (exit code 2), because the command no longer takes a
        positional ACTION.

        Requirement: 01-REQ-4.1
        """
        from spec.cli import main

        result = cli_runner.invoke(main, ["campaign", "create"])
        assert result.exit_code == 2, (
            f"Expected Click usage error (exit code 2), got {result.exit_code}. "
            f"Output: {result.output}"
        )


class TestCampaignOptions:
    """Tests for --path / --name / --description options — TS-01-12."""

    def test_ts01_12_campaign_with_path_and_name(
        self,
        cli_runner: CliRunner,
        tmp_path: Path,
    ) -> None:
        """TS-01-12: ``spec campaign --path P --name N`` calls Campaign.create
        with correct args and default empty description.

        Requirement: 01-REQ-4.2
        """
        from spec.cli import main

        camp_path = tmp_path / "camp_test"

        mock_create = MagicMock(return_value=MagicMock())
        with patch("agentspec.campaign.Campaign.create", mock_create):
            result = cli_runner.invoke(
                main, ["campaign", "--path", str(camp_path), "--name", "mycamp"]
            )

        assert result.exit_code == 0, f"Exit code {result.exit_code}: {result.output}"
        mock_create.assert_called_once_with(
            path=Path(str(camp_path)), name="mycamp", description=""
        )

    def test_campaign_short_aliases(
        self,
        cli_runner: CliRunner,
        tmp_path: Path,
    ) -> None:
        """Short aliases -p and -n work for --path and --name.

        Requirement: 01-REQ-4.2
        """
        from spec.cli import main

        camp_path = tmp_path / "camp_test"

        mock_create = MagicMock(return_value=MagicMock())
        with patch("agentspec.campaign.Campaign.create", mock_create):
            result = cli_runner.invoke(
                main, ["campaign", "-p", str(camp_path), "-n", "mycamp"]
            )

        assert result.exit_code == 0

    def test_campaign_missing_path_fails(
        self,
        cli_runner: CliRunner,
    ) -> None:
        """Missing --path exits with non-zero and Click error about missing option.

        The error must mention '--path' specifically, confirming that --path
        is a required option on the campaign command.

        Requirement: 01-REQ-4.E3
        """
        from spec.cli import main

        result = cli_runner.invoke(main, ["campaign", "--name", "mycamp"])
        assert result.exit_code != 0
        assert "--path" in result.output, (
            f"Expected error about missing '--path' option. Output: {result.output}"
        )

    def test_campaign_missing_name_fails(
        self,
        cli_runner: CliRunner,
        tmp_path: Path,
    ) -> None:
        """Missing --name exits with non-zero and Click error about missing option.

        The error must mention '--name' specifically, confirming that --name
        is a required option on the campaign command.

        Requirement: 01-REQ-4.E3
        """
        from spec.cli import main

        result = cli_runner.invoke(main, ["campaign", "--path", str(tmp_path / "camp")])
        assert result.exit_code != 0
        assert "--name" in result.output, (
            f"Expected error about missing '--name' option. Output: {result.output}"
        )


class TestCampaignCreateCalled:
    """Tests for Campaign.create invocation — TS-01-13."""

    def test_ts01_13_campaign_create_with_description(
        self,
        cli_runner: CliRunner,
        tmp_path: Path,
    ) -> None:
        """TS-01-13: ``spec campaign --path P --name N --description D``
        calls Campaign.create with all three arguments.

        Requirement: 01-REQ-4.3
        """
        from spec.cli import main

        camp_path = tmp_path / "camp_dir"

        mock_create = MagicMock(return_value=MagicMock())
        with patch("agentspec.campaign.Campaign.create", mock_create):
            result = cli_runner.invoke(
                main,
                [
                    "campaign",
                    "--path",
                    str(camp_path),
                    "--name",
                    "mycamp",
                    "--description",
                    "My campaign",
                ],
            )

        assert result.exit_code == 0, f"Exit code {result.exit_code}: {result.output}"
        mock_create.assert_called_once_with(
            path=Path(str(camp_path)), name="mycamp", description="My campaign"
        )

    def test_campaign_duplicate_path_error(
        self,
        cli_runner: CliRunner,
        tmp_path: Path,
    ) -> None:
        """Duplicate campaign at --path exits non-zero with 'already exists' message.

        Requirement: 01-REQ-4.E1
        """
        from agentspec.errors import CampaignError
        from spec.cli import main

        camp_path = tmp_path / "existing_camp"

        mock_create = MagicMock(
            side_effect=CampaignError(f"Campaign already exists at {camp_path}")
        )
        with patch("agentspec.campaign.Campaign.create", mock_create):
            result = cli_runner.invoke(
                main,
                ["campaign", "--path", str(camp_path), "--name", "mycamp"],
            )

        assert result.exit_code != 0
        combined_output = result.output
        assert "already exists" in combined_output.lower()


class TestCampaignPathIndependence:
    """Tests for --path / --spec-dir independence — TS-01-14."""

    def test_ts01_14_path_independent_of_spec_dir(
        self,
        cli_runner: CliRunner,
        tmp_path: Path,
    ) -> None:
        """TS-01-14: --path on campaign is independent of --spec-dir.

        Requirement: 01-REQ-4.4
        """
        from spec.cli import main

        camp_path = tmp_path / "camp_path"
        other_specs = tmp_path / "other_specs"

        mock_create = MagicMock(return_value=MagicMock())
        with patch("agentspec.campaign.Campaign.create", mock_create):
            result = cli_runner.invoke(
                main,
                [
                    "--spec-dir",
                    str(other_specs),
                    "campaign",
                    "--path",
                    str(camp_path),
                    "--name",
                    "mycamp",
                ],
            )

        assert result.exit_code == 0, f"Exit code {result.exit_code}: {result.output}"
        call_args = mock_create.call_args
        assert call_args.kwargs["path"] == Path(str(camp_path))
