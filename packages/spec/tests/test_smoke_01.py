"""End-to-end smoke tests for CLI flows in spec 01_cli_campaign_config.

These smoke tests exercise full CLI invocations through the Click runner,
verifying the end-to-end wiring of commands that span multiple subsystems.

Test Spec Entries: TS-01-4, TS-01-11, TS-01-15
Requirements: 01-REQ-2, 01-REQ-4, 01-REQ-5
"""

from __future__ import annotations

import json
from pathlib import Path
from unittest.mock import MagicMock, patch

from click.testing import CliRunner


class TestSmokeTests:
    """End-to-end smoke tests for spec CLI flows."""

    def test_smoke_1_spec_list_with_real_spec_dirs(
        self,
        cli_runner: CliRunner,
        isolated_dir: Path,
    ) -> None:
        """SMOKE-1: ``spec list`` against a real temp spec root with one
        valid spec dir produces valid JSON output.

        Exercises: spec list command, parse_spec_dir_name filtering,
        _session.json reading, JSON output formatting.
        """
        from spec.cli import main

        specs_dir = isolated_dir / ".specs"
        specs_dir.mkdir()

        # Create a valid spec directory with session state
        spec_dir = specs_dir / "01_smoke_test"
        spec_dir.mkdir()
        (spec_dir / "_session.json").write_text('{"state": "generated"}')

        # Also create non-matching entries that should be skipped
        (specs_dir / "archive").mkdir()
        (specs_dir / "notes.txt").write_text("random file")

        result = cli_runner.invoke(main, ["list"])
        assert result.exit_code == 0, (
            f"SMOKE-1: spec list failed with exit code {result.exit_code}: "
            f"{result.output}"
        )

        output = json.loads(result.output)
        assert output["spec_dir"] == ".specs"
        assert len(output["specs"]) == 1
        assert output["specs"][0]["name"] == "01_smoke_test"
        assert output["specs"][0]["state"] == "generated"

    def test_smoke_2_spec_campaign_creates_campaign(
        self,
        cli_runner: CliRunner,
        isolated_dir: Path,
    ) -> None:
        """SMOKE-2: ``spec campaign --path P --name N`` creates a campaign
        at the specified path.

        Exercises: campaign command option parsing, Campaign.create call.
        """
        from spec.cli import main

        camp_path = isolated_dir / "my_campaign"

        mock_camp = MagicMock()
        mock_camp.path = camp_path
        mock_camp.metadata.name = "my_camp"
        mock_create = MagicMock(return_value=mock_camp)

        with patch("agentspec.campaign.Campaign.create", mock_create):
            result = cli_runner.invoke(
                main,
                [
                    "campaign",
                    "--path",
                    str(camp_path),
                    "--name",
                    "my_camp",
                ],
            )

        assert result.exit_code == 0, (
            f"SMOKE-2: spec campaign failed with exit code {result.exit_code}: "
            f"{result.output}"
        )
        mock_create.assert_called_once_with(
            path=camp_path, name="my_camp", description=""
        )

    def test_smoke_3_spec_new_auto_creates_campaign_yaml(
        self,
        cli_runner: CliRunner,
        isolated_dir: Path,
        mock_campaign_open: MagicMock,
    ) -> None:
        """SMOKE-3: ``spec new`` in a temp dir without ``.specs/``
        auto-creates ``.specs/campaign.yaml`` before campaign creation
        proceeds.

        Exercises: auto-init campaign logic, campaign.yaml writing,
        Campaign.open wiring.
        """
        from spec.cli import main

        # Create a dummy PRD file
        prd_file = isolated_dir / "smoke_feature.md"
        prd_file.write_text("# Smoke Feature\nA description for smoke test.")

        with patch("agentspec.campaign.Campaign.open", mock_campaign_open):
            result = cli_runner.invoke(main, ["new", str(prd_file)])

        # Auto-init should have created .specs/ and campaign.yaml
        assert (isolated_dir / ".specs").is_dir(), (
            "SMOKE-3: .specs/ directory not created by auto-init"
        )
        assert (isolated_dir / ".specs" / "campaign.yaml").exists(), (
            "SMOKE-3: .specs/campaign.yaml not created by auto-init"
        )
        assert result.exit_code == 0, (
            f"SMOKE-3: spec new failed with exit code {result.exit_code}: "
            f"{result.output}"
        )

    def test_smoke_campaign_legacy_rejected(
        self,
        cli_runner: CliRunner,
    ) -> None:
        """SMOKE bonus: Legacy subcommands ``open`` and ``new-spec`` are
        rejected with Click usage errors.

        Exercises: campaign command no longer accepts ACTION argument.
        """
        from spec.cli import main

        result_open = cli_runner.invoke(main, ["campaign", "open"])
        assert result_open.exit_code == 2, (
            f"Expected Click usage error (exit code 2) for 'campaign open', "
            f"got {result_open.exit_code}: {result_open.output}"
        )

        result_new_spec = cli_runner.invoke(main, ["campaign", "new-spec"])
        assert result_new_spec.exit_code == 2, (
            f"Expected Click usage error (exit code 2) for 'campaign new-spec', "
            f"got {result_new_spec.exit_code}: {result_new_spec.output}"
        )
