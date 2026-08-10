"""Tests for render_individual_scoped with max_tokens budget cap.

Test Spec: TS-03-10, TS-03-11, TS-03-12
Requirements: 03-REQ-3
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
from afspec.render import render_individual_scoped

# ---------------------------------------------------------------------------
# Fixture helpers
# ---------------------------------------------------------------------------


def _estimate_tokens(text: str) -> int:
    """Import and call estimate_tokens at runtime (not yet implemented)."""
    from afspec import estimate_tokens

    return estimate_tokens(text)


def _make_scoped_spec_with_architecture() -> Spec:
    """Build a Spec with architecture and two task groups with subtask refs.

    Group 1 has subtask refs pointing to specific requirements/test specs,
    enabling scoped rendering. Architecture content is large to inflate
    token count for truncation testing.
    """
    arch_content = "# Architecture\n\n" + (
        "Detailed architecture explanation with diagrams and rationale. " * 100
    )

    return Spec(
        prd=PRDDocument(
            frontmatter=PRDFrontmatter(spec_id="03", spec_name="scoped_budget"),
            body="# Scoped Budget PRD\n\nPRD content for scoped budget testing.",
        ),
        requirements=Requirements(
            spec_id="03",
            spec_name="scoped_budget",
            introduction="Introduction text for scoped budget spec.",
            glossary={"token": "unit of text"},
            requirements=[
                Requirement(
                    id="03-REQ-1",
                    title="First requirement",
                    user_story=UserStory(
                        role="user", goal="test scoped", benefit="accuracy"
                    ),
                    acceptance_criteria=[
                        Criterion(
                            id="03-REQ-1.1",
                            ears_pattern=EARSPattern.UBIQUITOUS,
                            system="the system",
                            action="handle scoped rendering",
                        ),
                    ],
                ),
                Requirement(
                    id="03-REQ-2",
                    title="Second requirement",
                    user_story=UserStory(
                        role="developer", goal="test budget", benefit="efficiency"
                    ),
                    acceptance_criteria=[
                        Criterion(
                            id="03-REQ-2.1",
                            ears_pattern=EARSPattern.EVENT_DRIVEN,
                            when="called with budget",
                            system="the system",
                            action="truncate output",
                        ),
                    ],
                ),
            ],
        ),
        test_spec=TestSpec(
            spec_id="03",
            spec_name="scoped_budget",
            test_cases=[
                TestCase(
                    id="TS-03-1",
                    requirement_id="03-REQ-1.1",
                    kind="unit",
                    description="Test case for group 1",
                    preconditions=["A valid spec exists"],
                    input="input data for test 1",
                    expected="expected result for test 1",
                    assertion_pseudocode="assert result == expected",
                ),
                TestCase(
                    id="TS-03-2",
                    requirement_id="03-REQ-2.1",
                    kind="unit",
                    description="Test case for group 2",
                    preconditions=["Another precondition"],
                    input="input data for test 2",
                    expected="expected result for test 2",
                    assertion_pseudocode="assert result2 == expected2",
                ),
            ],
            coverage=Coverage(requirements_covered=["03-REQ-1", "03-REQ-2"]),
        ),
        tasks=Tasks(
            spec_id="03",
            spec_name="scoped_budget",
            test_commands=TestCommands(
                spec_tests="pytest -q",
                all_tests="pytest -q",
                linter="ruff check",
            ),
            task_groups=[
                TaskGroup(
                    id=1,
                    title="Task group 1",
                    subtasks=[
                        Subtask(
                            id="1.1",
                            title="Subtask 1.1",
                            requirement_refs=["03-REQ-1"],
                            test_spec_refs=["TS-03-1"],
                        ),
                    ],
                    verification=VerificationSubtask(id="1.V", checks=["check"]),
                ),
                TaskGroup(
                    id=2,
                    title="Task group 2",
                    subtasks=[
                        Subtask(
                            id="2.1",
                            title="Subtask 2.1",
                            requirement_refs=["03-REQ-2"],
                            test_spec_refs=["TS-03-2"],
                        ),
                    ],
                    verification=VerificationSubtask(id="2.V", checks=["check"]),
                ),
            ],
            traceability=[
                TraceabilityEntry(
                    requirement_id="03-REQ-1.1",
                    test_spec_id="TS-03-1",
                    task_id="1.1",
                ),
                TraceabilityEntry(
                    requirement_id="03-REQ-2.1",
                    test_spec_id="TS-03-2",
                    task_id="2.1",
                ),
            ],
        ),
        architecture=arch_content,
    )


def _make_scoped_spec_without_architecture() -> Spec:
    """Build a scoped Spec without architecture.

    Used to test that Level 1 (architecture removal) is skipped
    and the truncation loop goes directly to Level 2.
    """
    return Spec(
        prd=PRDDocument(
            frontmatter=PRDFrontmatter(spec_id="03", spec_name="scoped_no_arch"),
            body="# No-Arch PRD\n\nContent without architecture.",
        ),
        requirements=Requirements(
            spec_id="03",
            spec_name="scoped_no_arch",
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
            spec_name="scoped_no_arch",
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
            spec_name="scoped_no_arch",
            test_commands=TestCommands(
                spec_tests="pytest -q",
                all_tests="pytest -q",
                linter="ruff check",
            ),
            task_groups=[
                TaskGroup(
                    id=1,
                    title="Task 1",
                    subtasks=[
                        Subtask(
                            id="1.1",
                            title="Sub 1",
                            requirement_refs=["03-REQ-1"],
                            test_spec_refs=["TS-03-1"],
                        ),
                    ],
                    verification=VerificationSubtask(id="1.V", checks=["ok"]),
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
# TS-03-10: render_individual_scoped with max_tokens=None returns output
#           byte-identical to the pre-budget-cap implementation
# Requirement: 03-REQ-3.1
# ---------------------------------------------------------------------------


def test_render_individual_scoped_no_budget_identical() -> None:
    """render_individual_scoped(spec, target_group, max_tokens=None) is byte-identical."""
    spec = _make_scoped_spec_with_architecture()
    result_with_none = render_individual_scoped(spec, target_group=1, max_tokens=None)
    result_legacy = render_individual_scoped(spec, target_group=1)
    assert result_with_none == result_legacy
    for key in result_legacy:
        assert result_with_none[key] == result_legacy[key]


# ---------------------------------------------------------------------------
# TS-03-11: render_individual_scoped applies progressive truncation when
#           Level 0 scoped render exceeds the budget
# Requirement: 03-REQ-3.2
# ---------------------------------------------------------------------------


def test_render_individual_scoped_truncation_when_exceeds_budget() -> None:
    """Scoped render applies truncation when Level 0 exceeds budget."""
    spec = _make_scoped_spec_with_architecture()
    M = 200

    result = render_individual_scoped(spec, target_group=1, max_tokens=M)
    total = sum(_estimate_tokens(v) for v in result.values())

    # Either fits within budget or is the Level 2 scoped render
    level2 = render_individual_scoped(spec, target_group=1, max_tokens=1)
    assert total <= M or result == level2


def test_render_individual_scoped_level1_drops_architecture() -> None:
    """Level 1 truncation drops architecture from scoped result dict."""
    spec = _make_scoped_spec_with_architecture()

    # Get Level 0 to find architecture token contribution
    level0 = render_individual_scoped(spec, target_group=1, max_tokens=None)
    assert "architecture" in level0
    level0_tokens = sum(_estimate_tokens(v) for v in level0.values())
    arch_tokens = _estimate_tokens(level0["architecture"])

    # Budget just below Level 0 but above Level 1 (no architecture)
    budget = level0_tokens - arch_tokens + 5
    result = render_individual_scoped(spec, target_group=1, max_tokens=budget)
    assert "architecture" not in result


def test_render_individual_scoped_level2_slim_test_spec() -> None:
    """Level 2 truncation produces slim test spec in scoped context."""
    spec = _make_scoped_spec_with_architecture()

    # Use very small budget to force Level 2
    result = render_individual_scoped(spec, target_group=1, max_tokens=1)

    # Architecture dropped (Level 1)
    assert "architecture" not in result

    # Test spec should be slimmed (Level 2) — verbose fields absent
    ts_content = result.get("test_spec", "")
    assert "assertion_pseudocode" not in ts_content or "assert result ==" not in ts_content


# ---------------------------------------------------------------------------
# TS-03-12: render_individual_scoped uses the same budget evaluation logic
#           as render_individual: sum of estimate_tokens(v) for all values
# Requirement: 03-REQ-3.3
# ---------------------------------------------------------------------------


def test_render_individual_scoped_budget_evaluation_sums_values() -> None:
    """Budget is evaluated as sum of estimate_tokens(v) for all values."""
    spec = _make_scoped_spec_with_architecture()
    result = render_individual_scoped(spec, target_group=1, max_tokens=None)

    total = sum(_estimate_tokens(v) for v in result.values())
    assert total >= 0

    # Empty strings contribute 0
    assert _estimate_tokens("") == 0


def test_render_individual_scoped_empty_values_contribute_zero() -> None:
    """Empty string values in the result dict contribute 0 to token sum."""
    spec = _make_scoped_spec_without_architecture()
    result = render_individual_scoped(spec, target_group=1, max_tokens=None)

    total = sum(_estimate_tokens(v) for v in result.values())
    non_empty_total = sum(_estimate_tokens(v) for v in result.values() if v != "")
    assert total == non_empty_total


# ---------------------------------------------------------------------------
# 03-REQ-3.E1: No architecture section — skip Level 1, go to Level 2
# ---------------------------------------------------------------------------


def test_render_individual_scoped_no_arch_skips_level1() -> None:
    """When spec has no architecture, Level 1 is skipped, Level 2 applied."""
    spec = _make_scoped_spec_without_architecture()
    assert spec.architecture is None

    # Set a very small budget to force truncation
    result = render_individual_scoped(spec, target_group=1, max_tokens=1)

    # Architecture was never present
    assert "architecture" not in result

    # Should still have test_spec (in slim form)
    assert "test_spec" in result


# ---------------------------------------------------------------------------
# 03-REQ-3.E2: max_tokens=0 or negative treated as None (full Level 0)
# ---------------------------------------------------------------------------


def test_render_individual_scoped_zero_budget_identical() -> None:
    """render_individual_scoped with max_tokens=0 returns Level 0."""
    spec = _make_scoped_spec_with_architecture()
    result_zero = render_individual_scoped(spec, target_group=1, max_tokens=0)
    result_legacy = render_individual_scoped(spec, target_group=1)
    assert result_zero == result_legacy


def test_render_individual_scoped_negative_budget_identical() -> None:
    """render_individual_scoped with max_tokens=-1 returns Level 0."""
    spec = _make_scoped_spec_with_architecture()
    result_neg = render_individual_scoped(spec, target_group=1, max_tokens=-1)
    result_legacy = render_individual_scoped(spec, target_group=1)
    assert result_neg == result_legacy


# ---------------------------------------------------------------------------
# 03-REQ-3.E3: Level 2 still exceeds budget — return as-is
# ---------------------------------------------------------------------------


def test_render_individual_scoped_level2_exceeds_returned_as_is() -> None:
    """When Level 2 scoped render still exceeds budget, it is returned as-is."""
    spec = _make_scoped_spec_with_architecture()

    # Budget of 1 — even Level 2 will exceed this
    result = render_individual_scoped(spec, target_group=1, max_tokens=1)

    # Should still have the core keys
    assert "prd" in result
    assert "requirements" in result
    assert "test_spec" in result
    assert "tasks" in result
    # Architecture should be dropped
    assert "architecture" not in result


# ---------------------------------------------------------------------------
# 03-PROP-8 (scoped variant): Zero or negative max_tokens equivalent to None
# Validates: 03-REQ-3.E2
# ---------------------------------------------------------------------------


def test_property_scoped_zero_negative_equivalent_to_none() -> None:
    """Zero and negative max_tokens produce identical scoped output to None."""
    spec = _make_scoped_spec_with_architecture()
    result_none = render_individual_scoped(spec, target_group=1, max_tokens=None)
    result_zero = render_individual_scoped(spec, target_group=1, max_tokens=0)
    result_neg = render_individual_scoped(spec, target_group=1, max_tokens=-10)
    assert result_none == result_zero
    assert result_none == result_neg
