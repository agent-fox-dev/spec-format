"""Tests for generation_user_prompt with empty optional blocks (issue #56).

Covers NS-REQ-1 through NS-REQ-5 from the issue #56 spec:
- TS-NS-5: Python generation_user_prompt with spec_landscape=None and
  dependent_interfaces=None produces no consecutive blank lines.
"""

from __future__ import annotations

import pytest
from agentspec.prompts import generation_user_prompt

# ===========================================================================
# TS-NS-5 (issue #56): generation_user_prompt with None optional blocks
# produces no '\n\n\n' and no landscape/interface section headings.
# (NS-REQ-5)
# ===========================================================================


class TestGenerationUserPromptEmptyBlocks:
    """TS-NS-5: generation_user_prompt with None blocks is clean."""

    def test_nil_landscape_no_triple_newline(self) -> None:
        """NS-REQ-1,NS-REQ-5: nil spec_landscape → no '\n\n\n' in prompt."""
        result = generation_user_prompt(
            "prd text",
            "requirements",
            spec_landscape=None,
            dependent_interfaces=None,
        )
        assert "\n\n\n" not in result

    def test_nil_landscape_no_landscape_heading(self) -> None:
        """NS-REQ-1,NS-REQ-5: nil spec_landscape → no 'Existing Spec Landscape' heading."""
        result = generation_user_prompt(
            "prd text",
            "requirements",
            spec_landscape=None,
            dependent_interfaces=None,
        )
        assert "Existing Spec Landscape" not in result

    def test_nil_interfaces_no_interfaces_heading(self) -> None:
        """NS-REQ-2,NS-REQ-5: nil dependent_interfaces → no 'Dependent Spec Interfaces' heading."""
        result = generation_user_prompt(
            "prd text",
            "requirements",
            spec_landscape=None,
            dependent_interfaces=None,
        )
        assert "Dependent Spec Interfaces" not in result

    @pytest.mark.parametrize("artifact_name", ["requirements", "test_spec", "tasks"])
    def test_all_artifact_types_no_triple_newline(self, artifact_name: str) -> None:
        """NS-REQ-3: all artifact types produce clean prompts with no '\n\n\n'."""
        result = generation_user_prompt(
            "prd text",
            artifact_name,
            spec_landscape=None,
            dependent_interfaces=None,
        )
        assert "\n\n\n" not in result, (
            f"generation_user_prompt({artifact_name!r}) contains three or more "
            "consecutive newlines; want no excess blank lines"
        )

    def test_non_empty_landscape_and_interfaces_present(self) -> None:
        """NS-REQ-4 regression guard: non-empty blocks still appear in prompt."""
        landscape = [
            {
                "spec_id": "01",
                "spec_name": "core_foundation",
                "title": "Core Foundation",
                "status": "implemented",
                "intent": "Base layer",
                "archived": False,
            }
        ]
        interfaces = [
            {
                "spec_id": "01",
                "spec_name": "core_foundation",
                "glossary": {"Widget": "A UI element"},
                "external_apis": [],
                "interface_symbols": [],
            }
        ]

        result = generation_user_prompt(
            "prd text",
            "requirements",
            spec_landscape=landscape,
            dependent_interfaces=interfaces,
        )

        assert "Existing Spec Landscape" in result
        assert "Core Foundation" in result
        assert "Dependent Spec Interfaces" in result
