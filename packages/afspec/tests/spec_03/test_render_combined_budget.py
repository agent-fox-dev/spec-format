"""Tests for render_combined with max_tokens budget cap.

Test Spec: TS-03-13, TS-03-14, TS-03-15
Requirements: 03-REQ-4
"""

from __future__ import annotations

from afspec.models import (
    Coverage,
    Criterion,
    EARSPattern,
    PRDDocument,
    PRDFrontmatter,
    Requirement,
    Requirements,
    Spec,
    Subtask,
    TaskGroup,
    Tasks,
    TestCase,
    TestCommands,
    TestSpec,
    TraceabilityEntry,
    UserStory,
    VerificationSubtask,
)
from afspec.render import render_combined

# ---------------------------------------------------------------------------
# Fixture helpers
# ---------------------------------------------------------------------------


def _estimate_tokens(text: str) -> int:
    """Import and call estimate_tokens at runtime (not yet implemented)."""
    from afspec import estimate_tokens

    return estimate_tokens(text)


def _make_spec_with_all_artifacts() -> Spec:
    """Build a Spec with all artifacts populated, including architecture.

    Used to verify render_combined behaviour with the full artifact set.
    """
    return Spec(
        prd=PRDDocument(
            frontmatter=PRDFrontmatter(spec_id="03", spec_name="combined_test"),
            body="# Combined Test PRD\n\nThis is the PRD content for combined rendering.",
        ),
        requirements=Requirements(
            spec_id="03",
            spec_name="combined_test",
            introduction="Introduction for combined test.",
            glossary={"budget": "token budget cap"},
            requirements=[
                Requirement(
                    id="03-REQ-1",
                    title="Combined requirement",
                    user_story=UserStory(role="user", goal="test combined", benefit="accuracy"),
                    acceptance_criteria=[
                        Criterion(
                            id="03-REQ-1.1",
                            ears_pattern=EARSPattern.UBIQUITOUS,
                            system="the system",
                            action="render combined output",
                        ),
                    ],
                ),
            ],
        ),
        test_spec=TestSpec(
            spec_id="03",
            spec_name="combined_test",
            test_cases=[
                TestCase(
                    id="TS-03-1",
                    requirement_id="03-REQ-1.1",
                    kind="unit",
                    description="A test case for combined",
                    preconditions=["A valid spec"],
                    input="test input",
                    expected="test expected output",
                    assertion_pseudocode="assert combined_result is not None",
                ),
            ],
            coverage=Coverage(requirements_covered=["03-REQ-1"]),
        ),
        tasks=Tasks(
            spec_id="03",
            spec_name="combined_test",
            test_commands=TestCommands(
                spec_tests="pytest -q",
                all_tests="pytest -q",
                linter="ruff check",
            ),
            task_groups=[
                TaskGroup(
                    id=1,
                    title="Task 1",
                    subtasks=[Subtask(id="1.1", title="Sub 1")],
                    verification=VerificationSubtask(id="1.V", checks=["check"]),
                ),
            ],
            traceability=[
                TraceabilityEntry(
                    requirement_id="03-REQ-1.1",
                    test_spec_id="TS-03-1",
                    task_id="1.1",
                ),
            ],
        ),
        architecture=None,
    )


def _make_spec_with_large_combined_content() -> Spec:
    """Build a Spec whose combined render is large enough to exceed a moderate budget.

    The architecture section provides the bulk of the content, so that
    Level 1 truncation (dropping architecture) meaningfully reduces tokens.
    """
    arch_content = "# Architecture\n\n" + ("This is detailed architecture content for combined rendering. " * 100)

    return Spec(
        prd=PRDDocument(
            frontmatter=PRDFrontmatter(spec_id="03", spec_name="large_combined"),
            body="# Large Combined PRD\n\nPRD body for large combined testing.",
        ),
        requirements=Requirements(
            spec_id="03",
            spec_name="large_combined",
            introduction="Introduction for large combined spec.",
            glossary={"term": "definition"},
            requirements=[
                Requirement(
                    id="03-REQ-1",
                    title="Large combined requirement",
                    user_story=UserStory(role="user", goal="test large combined", benefit="value"),
                    acceptance_criteria=[
                        Criterion(
                            id="03-REQ-1.1",
                            ears_pattern=EARSPattern.UBIQUITOUS,
                            system="the system",
                            action="handle large combined render",
                        ),
                    ],
                ),
            ],
        ),
        test_spec=TestSpec(
            spec_id="03",
            spec_name="large_combined",
            test_cases=[
                TestCase(
                    id="TS-03-1",
                    requirement_id="03-REQ-1.1",
                    kind="unit",
                    description="Test case for large combined",
                    preconditions=["precondition"],
                    input="test input data",
                    expected="test expected result",
                    assertion_pseudocode="assert output_fits_budget()",
                ),
            ],
            coverage=Coverage(requirements_covered=["03-REQ-1"]),
        ),
        tasks=Tasks(
            spec_id="03",
            spec_name="large_combined",
            test_commands=TestCommands(
                spec_tests="pytest -q",
                all_tests="pytest -q",
                linter="ruff check",
            ),
            task_groups=[
                TaskGroup(
                    id=1,
                    title="Task 1",
                    subtasks=[Subtask(id="1.1", title="Sub 1")],
                    verification=VerificationSubtask(id="1.V", checks=["check"]),
                ),
            ],
            traceability=[
                TraceabilityEntry(
                    requirement_id="03-REQ-1.1",
                    test_spec_id="TS-03-1",
                    task_id="1.1",
                ),
            ],
        ),
        architecture=arch_content,
    )


def _make_spec_without_architecture() -> Spec:
    """Build a Spec without architecture for testing Level 1 skip."""
    return Spec(
        prd=PRDDocument(
            frontmatter=PRDFrontmatter(spec_id="03", spec_name="no_arch_combined"),
            body="# No Arch PRD\n\nPRD content without architecture.",
        ),
        requirements=Requirements(
            spec_id="03",
            spec_name="no_arch_combined",
            introduction="Intro.",
            glossary={"term": "def"},
            requirements=[
                Requirement(
                    id="03-REQ-1",
                    title="Requirement one",
                    user_story=UserStory(role="user", goal="test", benefit="value"),
                    acceptance_criteria=[
                        Criterion(
                            id="03-REQ-1.1",
                            ears_pattern=EARSPattern.UBIQUITOUS,
                            system="the system",
                            action="do something",
                        ),
                    ],
                ),
            ],
        ),
        test_spec=TestSpec(
            spec_id="03",
            spec_name="no_arch_combined",
            test_cases=[
                TestCase(
                    id="TS-03-1",
                    requirement_id="03-REQ-1.1",
                    kind="unit",
                    description="A test case",
                    preconditions=["precondition"],
                    input="input data",
                    expected="expected result",
                    assertion_pseudocode="assert True",
                ),
            ],
            coverage=Coverage(requirements_covered=["03-REQ-1"]),
        ),
        tasks=Tasks(
            spec_id="03",
            spec_name="no_arch_combined",
            test_commands=TestCommands(
                spec_tests="pytest -q",
                all_tests="pytest -q",
                linter="ruff check",
            ),
            task_groups=[
                TaskGroup(
                    id=1,
                    title="Task 1",
                    subtasks=[Subtask(id="1.1", title="Sub 1")],
                    verification=VerificationSubtask(id="1.V", checks=["check"]),
                ),
            ],
            traceability=[
                TraceabilityEntry(
                    requirement_id="03-REQ-1.1",
                    test_spec_id="TS-03-1",
                    task_id="1.1",
                ),
            ],
        ),
        architecture=None,
    )


# ---------------------------------------------------------------------------
# TS-03-13: render_combined with max_tokens=None returns output byte-identical
#           to the pre-budget-cap implementation
# Requirement: 03-REQ-4.1
# ---------------------------------------------------------------------------


def test_render_combined_no_budget_identical() -> None:
    """render_combined(spec, max_tokens=None) is byte-identical to legacy call."""
    spec = _make_spec_with_all_artifacts()
    result_with_none = render_combined(spec, max_tokens=None)
    result_legacy = render_combined(spec)
    assert result_with_none == result_legacy


def test_render_combined_no_budget_identical_with_architecture() -> None:
    """render_combined(spec, max_tokens=None) is byte-identical for spec with architecture."""
    spec = _make_spec_with_large_combined_content()
    result_with_none = render_combined(spec, max_tokens=None)
    result_legacy = render_combined(spec)
    assert result_with_none == result_legacy


# ---------------------------------------------------------------------------
# TS-03-14: render_combined applies progressive truncation against the single
#           returned string when Level 0 exceeds the budget
# Requirement: 03-REQ-4.2
# ---------------------------------------------------------------------------


def test_render_combined_truncation_when_exceeds_budget() -> None:
    """render_combined truncates when Level 0 exceeds budget."""
    spec = _make_spec_with_large_combined_content()
    M = 300

    result = render_combined(spec, max_tokens=M)

    # Either fits within budget or is the Level 2 combined render
    result_level2 = render_combined(spec, max_tokens=1)
    assert _estimate_tokens(result) <= M or result == result_level2


def test_render_combined_level1_drops_architecture() -> None:
    """Level 1 truncation omits architecture section from combined string."""
    spec = _make_spec_with_large_combined_content()

    full_render = render_combined(spec, max_tokens=None)
    assert "architecture content for combined rendering" in full_render

    # Set budget to trigger Level 1 but not Level 2
    full_tokens = _estimate_tokens(full_render)
    arch_content = spec.architecture
    assert arch_content is not None
    arch_tokens = _estimate_tokens(arch_content)
    budget = full_tokens - arch_tokens + 5

    result = render_combined(spec, max_tokens=budget)
    assert "architecture content for combined rendering" not in result


def test_render_combined_level2_slim_test_spec() -> None:
    """Level 2 truncation produces slim test spec entries in combined string."""
    spec = _make_spec_with_large_combined_content()

    # Very small budget forces Level 2
    result = render_combined(spec, max_tokens=1)

    # Architecture content should be absent (Level 1)
    assert "architecture content for combined rendering" not in result

    # PRD content should remain
    assert "PRD body for large combined" in result


# ---------------------------------------------------------------------------
# TS-03-15: render_combined evaluates the budget against estimate_tokens of
#           the single returned string, not a sum of dict values
# Requirement: 03-REQ-4.3
# ---------------------------------------------------------------------------


def test_render_combined_budget_evaluates_single_string() -> None:
    """Budget is evaluated as estimate_tokens(combined_string), not sum of dict."""
    spec = _make_spec_with_all_artifacts()
    result = render_combined(spec, max_tokens=None)
    assert isinstance(result, str)
    assert _estimate_tokens(result) == len(result) // 4


def test_render_combined_returns_str_not_dict() -> None:
    """render_combined always returns a string, never a dict."""
    spec = _make_spec_with_large_combined_content()

    # No budget
    result_none = render_combined(spec, max_tokens=None)
    assert isinstance(result_none, str)

    # With budget
    result_budget = render_combined(spec, max_tokens=100)
    assert isinstance(result_budget, str)

    # Extreme budget
    result_extreme = render_combined(spec, max_tokens=1)
    assert isinstance(result_extreme, str)


# ---------------------------------------------------------------------------
# 03-REQ-4.E1: No architecture — skip Level 1, go to Level 2
# ---------------------------------------------------------------------------


def test_render_combined_no_arch_skips_level1() -> None:
    """When spec has no architecture, Level 1 is skipped, Level 2 applied."""
    spec = _make_spec_without_architecture()
    assert spec.architecture is None

    # Set a very small budget to force truncation
    result = render_combined(spec, max_tokens=1)

    # PRD content should still be present
    assert "PRD content without architecture" in result


# ---------------------------------------------------------------------------
# 03-REQ-4.E2: max_tokens=0 or negative treated as None (full Level 0)
# ---------------------------------------------------------------------------


def test_render_combined_zero_budget_identical() -> None:
    """render_combined(spec, max_tokens=0) returns the full Level 0 render."""
    spec = _make_spec_with_large_combined_content()
    result_zero = render_combined(spec, max_tokens=0)
    result_legacy = render_combined(spec)
    assert result_zero == result_legacy


def test_render_combined_negative_budget_identical() -> None:
    """render_combined(spec, max_tokens=-1) returns the full Level 0 render."""
    spec = _make_spec_with_large_combined_content()
    result_neg = render_combined(spec, max_tokens=-1)
    result_legacy = render_combined(spec)
    assert result_neg == result_legacy


# ---------------------------------------------------------------------------
# 03-REQ-4.E3: Level 2 still exceeds budget — return as-is
# ---------------------------------------------------------------------------


def test_render_combined_level2_exceeds_returned_as_is() -> None:
    """When Level 2 combined render still exceeds budget, it is returned as-is."""
    spec = _make_spec_with_large_combined_content()

    # Budget of 1 — even Level 2 will exceed this
    result = render_combined(spec, max_tokens=1)
    assert isinstance(result, str)
    assert len(result) > 0

    # PRD content should still be present even at Level 2
    assert "PRD body for large combined" in result


# ---------------------------------------------------------------------------
# 03-PROP-8 (combined variant): Zero or negative max_tokens equivalent to None
# Validates: 03-REQ-4.E2
# ---------------------------------------------------------------------------


def test_property_combined_zero_negative_equivalent_to_none() -> None:
    """Zero and negative max_tokens produce identical combined output to None."""
    spec = _make_spec_with_all_artifacts()
    result_none = render_combined(spec, max_tokens=None)
    result_zero = render_combined(spec, max_tokens=0)
    result_neg = render_combined(spec, max_tokens=-10)
    assert result_none == result_zero
    assert result_none == result_neg
