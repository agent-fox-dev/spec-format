"""Tests for render_individual with max_tokens budget cap.

Test Spec: TS-03-6, TS-03-7, TS-03-8, TS-03-9
Requirements: 03-REQ-2
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
from afspec.render import render_individual

# ---------------------------------------------------------------------------
# Fixture helpers
# ---------------------------------------------------------------------------


def _estimate_tokens(text: str) -> int:
    """Import and call estimate_tokens at runtime (not yet implemented)."""
    from afspec import estimate_tokens

    return estimate_tokens(text)


def _make_small_spec() -> Spec:
    """Build a minimal Spec whose Level 0 render fits a small token budget.

    Total rendered content is small enough to stay well within typical budgets.
    """
    return Spec(
        prd=PRDDocument(
            frontmatter=PRDFrontmatter(spec_id="03", spec_name="small"),
            body="# Small PRD\n\nMinimal content.",
        ),
        requirements=Requirements(
            spec_id="03",
            spec_name="small",
            introduction="Short intro.",
            glossary={"term": "definition"},
            requirements=[
                Requirement(
                    id="03-REQ-1",
                    title="Test requirement",
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
            spec_name="small",
            test_cases=[
                TestCase(
                    id="TS-03-1",
                    requirement_id="03-REQ-1.1",
                    kind="unit",
                    description="A test case",
                ),
            ],
            coverage=Coverage(requirements_covered=["03-REQ-1"]),
        ),
        tasks=Tasks(
            spec_id="03",
            spec_name="small",
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
                        Subtask(id="1.1", title="Sub 1"),
                    ],
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


def _make_large_spec_with_architecture() -> Spec:
    """Build a Spec with substantial architecture content.

    The architecture section adds significant token count so that Level 0
    exceeds a moderate budget, but removing architecture (Level 1) fits.
    """
    # Use a large architecture string to inflate token count
    arch_content = "# Architecture\n\n" + ("This is detailed architecture content. " * 100)

    return Spec(
        prd=PRDDocument(
            frontmatter=PRDFrontmatter(spec_id="03", spec_name="large"),
            body="# Large PRD\n\nPRD body content here.",
        ),
        requirements=Requirements(
            spec_id="03",
            spec_name="large",
            introduction="Introduction text.",
            glossary={"term": "definition"},
            requirements=[
                Requirement(
                    id="03-REQ-1",
                    title="Test requirement",
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
            spec_name="large",
            test_cases=[
                TestCase(
                    id="TS-03-1",
                    requirement_id="03-REQ-1.1",
                    kind="unit",
                    description="Test case",
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
            spec_name="large",
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
                        Subtask(id="1.1", title="Sub 1"),
                    ],
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


# ---------------------------------------------------------------------------
# TS-03-6: render_individual with max_tokens=None is byte-identical to
#          the pre-budget-cap implementation
# Requirement: 03-REQ-2.1
# ---------------------------------------------------------------------------


def test_render_individual_no_budget_identical() -> None:
    """render_individual(spec, max_tokens=None) is byte-identical to legacy call."""
    spec = _make_large_spec_with_architecture()
    result_with_none = render_individual(spec, max_tokens=None)
    result_legacy = render_individual(spec)
    assert result_with_none == result_legacy
    for key in result_legacy:
        assert result_with_none[key] == result_legacy[key]


# ---------------------------------------------------------------------------
# 03-REQ-2.E1: max_tokens=0 is equivalent to None (no budget cap)
# ---------------------------------------------------------------------------


def test_render_individual_zero_budget_identical() -> None:
    """render_individual(spec, max_tokens=0) returns Level 0 (no truncation)."""
    spec = _make_large_spec_with_architecture()
    result_zero = render_individual(spec, max_tokens=0)
    result_legacy = render_individual(spec)
    assert result_zero == result_legacy


# ---------------------------------------------------------------------------
# 03-REQ-2.E4: negative max_tokens treated as None
# ---------------------------------------------------------------------------


def test_render_individual_negative_budget_identical() -> None:
    """render_individual(spec, max_tokens=-1) returns Level 0 (no truncation)."""
    spec = _make_large_spec_with_architecture()
    result_neg = render_individual(spec, max_tokens=-1)
    result_legacy = render_individual(spec)
    assert result_neg == result_legacy


# ---------------------------------------------------------------------------
# 03-PROP-8: Zero or negative max_tokens equivalent to no budget
# Validates: 03-REQ-2.E1
# ---------------------------------------------------------------------------


def test_property_zero_negative_equivalent_to_none() -> None:
    """Zero and negative max_tokens produce identical output to None."""
    spec = _make_small_spec()
    result_none = render_individual(spec, max_tokens=None)
    result_zero = render_individual(spec, max_tokens=0)
    result_neg = render_individual(spec, max_tokens=-10)
    assert result_none == result_zero
    assert result_none == result_neg


# ---------------------------------------------------------------------------
# TS-03-7: Progressive truncation when Level 0 exceeds budget
# Requirement: 03-REQ-2.2
# ---------------------------------------------------------------------------


def test_render_individual_truncation_when_exceeds_budget() -> None:
    """render_individual applies truncation when Level 0 exceeds budget."""
    spec = _make_large_spec_with_architecture()

    # First, compute the Level 0 token count
    level0 = render_individual(spec, max_tokens=None)
    level0_tokens = sum(_estimate_tokens(v) for v in level0.values())

    # Set a budget that is less than Level 0 but should be achievable
    # after dropping architecture (Level 1)
    arch_tokens = _estimate_tokens(level0["architecture"])
    budget = level0_tokens - arch_tokens  # Just enough for Level 1

    result = render_individual(spec, max_tokens=budget)
    total = sum(_estimate_tokens(v) for v in result.values())

    # Either fits within budget or is the Level 2 render
    assert total <= budget or "architecture" not in result
    # Level 1 at minimum was applied (architecture dropped)
    assert "architecture" not in result


# ---------------------------------------------------------------------------
# TS-03-8: No truncation when budget is sufficient
# Requirement: 03-REQ-2.3
# ---------------------------------------------------------------------------


def test_render_individual_no_truncation_when_budget_sufficient() -> None:
    """render_individual returns Level 0 without truncation when budget is enough."""
    spec = _make_large_spec_with_architecture()

    # Get the full render
    result_full = render_individual(spec, max_tokens=None)
    level0_tokens = sum(_estimate_tokens(v) for v in result_full.values())

    # Set budget above the Level 0 total
    generous_budget = level0_tokens + 1000

    result = render_individual(spec, max_tokens=generous_budget)
    assert result == result_full
    assert "architecture" in result


# ---------------------------------------------------------------------------
# TS-03-9: Budget evaluation sums estimate_tokens(v) for all values,
#          empty strings contribute 0
# Requirement: 03-REQ-2.4
# ---------------------------------------------------------------------------


def test_budget_evaluation_sums_all_values() -> None:
    """Budget evaluation counts empty-string values as 0 tokens."""
    spec = _make_small_spec()
    result = render_individual(spec, max_tokens=None)

    total = sum(_estimate_tokens(v) for v in result.values())

    # Empty string contributes 0
    assert _estimate_tokens("") == 0

    # Total equals sum of non-empty values only
    non_empty_total = sum(_estimate_tokens(v) for v in result.values() if v != "")
    assert total == non_empty_total


# ---------------------------------------------------------------------------
# 03-REQ-2.E2: No architecture section — skip Level 1, go to Level 2
# ---------------------------------------------------------------------------


def test_render_individual_no_architecture_skips_level1() -> None:
    """When spec has no architecture, Level 1 is skipped, Level 2 applied."""
    spec = _make_small_spec()
    assert spec.architecture is None

    # Set a very small budget to force truncation
    result = render_individual(spec, max_tokens=1)

    # Architecture was never present, so it shouldn't be in result
    assert "architecture" not in result


# ---------------------------------------------------------------------------
# 03-REQ-2.E3: Level 2 still exceeds budget — return as-is
# ---------------------------------------------------------------------------


def test_render_individual_level2_exceeds_returned_as_is() -> None:
    """When Level 2 render still exceeds budget, it is returned as-is."""
    spec = _make_large_spec_with_architecture()

    # Use a budget of 1 — even Level 2 will exceed this
    result = render_individual(spec, max_tokens=1)

    # Should still have the required keys (prd, requirements, test_spec, tasks)
    assert "prd" in result
    assert "requirements" in result
    assert "test_spec" in result
    assert "tasks" in result
    # Architecture should be dropped (Level 1 at minimum)
    assert "architecture" not in result


# ---------------------------------------------------------------------------
# 03-PROP-2: No-budget render is byte-identical to legacy render
# Validates: 03-REQ-2.1
# ---------------------------------------------------------------------------


def test_property_no_budget_byte_identical() -> None:
    """No-budget render is byte-identical to legacy for specs with architecture."""
    spec = _make_large_spec_with_architecture()
    result_default = render_individual(spec)
    result_none = render_individual(spec, max_tokens=None)
    assert result_default == result_none

    # Also test with a spec without architecture
    spec_no_arch = _make_small_spec()
    result_default_no_arch = render_individual(spec_no_arch)
    result_none_no_arch = render_individual(spec_no_arch, max_tokens=None)
    assert result_default_no_arch == result_none_no_arch


# ---------------------------------------------------------------------------
# 03-PROP-3: Budget-capped render token sum is within budget when achievable
# Validates: 03-REQ-2.2, 03-REQ-2.3
# ---------------------------------------------------------------------------


def test_property_budget_cap_respected() -> None:
    """When budget is achievable, total tokens <= budget.

    If not achievable, the result is the Level 2 render.
    """
    spec = _make_large_spec_with_architecture()

    level0 = render_individual(spec, max_tokens=None)
    level0_tokens = sum(_estimate_tokens(v) for v in level0.values())

    # Choose a budget that Level 1 can achieve (remove architecture)
    arch_tokens = _estimate_tokens(level0["architecture"])
    budget = level0_tokens - arch_tokens + 10  # Slightly above Level 1

    result = render_individual(spec, max_tokens=budget)
    total = sum(_estimate_tokens(v) for v in result.values())
    assert total <= budget
