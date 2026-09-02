"""Tests for Level 1 truncation behavior (architecture removal).

Test Spec: TS-03-16, TS-03-17, TS-03-18, TS-03-19
Requirements: 03-REQ-5
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
from afspec.render import render_combined, render_individual

# ---------------------------------------------------------------------------
# Fixture helpers
# ---------------------------------------------------------------------------

# Known PRD body content used for assertion in TS-03-19
_PRD_BODY = "# Budget Cap PRD\n\nThis is the PRD content for testing."


def _estimate_tokens(text: str) -> int:
    """Import and call estimate_tokens at runtime (not yet implemented)."""
    from afspec import estimate_tokens

    return estimate_tokens(text)


def _make_spec_with_architecture() -> Spec:
    """Build a Spec with a sizeable architecture section.

    Architecture content is large enough that removing it meaningfully
    reduces the token count. Other artifacts are kept relatively small.
    """
    arch_content = "# Architecture\n\n" + ("Architecture detail paragraph. " * 80)

    return Spec(
        prd=PRDDocument(
            frontmatter=PRDFrontmatter(spec_id="03", spec_name="trunc_test"),
            body=_PRD_BODY,
        ),
        requirements=Requirements(
            spec_id="03",
            spec_name="trunc_test",
            introduction="Short intro.",
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
                            action="perform action",
                        ),
                    ],
                ),
            ],
        ),
        test_spec=TestSpec(
            spec_id="03",
            spec_name="trunc_test",
            test_cases=[
                TestCase(
                    id="TS-03-1",
                    requirement_id="03-REQ-1.1",
                    kind="unit",
                    description="A test case",
                    preconditions=["precondition A"],
                    input="test input",
                    expected="test expected",
                    assertion_pseudocode="assert result == expected",
                ),
            ],
            coverage=Coverage(requirements_covered=["03-REQ-1"]),
        ),
        tasks=Tasks(
            spec_id="03",
            spec_name="trunc_test",
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
        architecture=arch_content,
    )


def _make_spec_without_architecture() -> Spec:
    """Build a Spec without an architecture section."""
    return Spec(
        prd=PRDDocument(
            frontmatter=PRDFrontmatter(spec_id="03", spec_name="no_arch"),
            body=_PRD_BODY,
        ),
        requirements=Requirements(
            spec_id="03",
            spec_name="no_arch",
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
            spec_name="no_arch",
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
            spec_name="no_arch",
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
# TS-03-16: Level 1 removes 'architecture' key from render_individual dict
# Requirement: 03-REQ-5.1
# ---------------------------------------------------------------------------


def test_level1_removes_architecture_key() -> None:
    """Level 1 truncation removes the 'architecture' key from result dict."""
    spec = _make_spec_with_architecture()

    # Compute Level 0 total
    level0 = render_individual(spec, max_tokens=None)
    assert "architecture" in level0
    level0_tokens = sum(_estimate_tokens(v) for v in level0.values())
    arch_tokens = _estimate_tokens(level0["architecture"])

    # Set budget that Level 0 exceeds but Level 1 fits
    # (total minus architecture, plus a small margin)
    budget = level0_tokens - arch_tokens + 5

    result = render_individual(spec, max_tokens=budget)
    assert "architecture" not in result


# ---------------------------------------------------------------------------
# TS-03-17: Level 1 omits architecture from render_combined string
# Requirement: 03-REQ-5.2
# ---------------------------------------------------------------------------


def test_level1_omits_architecture_from_combined() -> None:
    """Level 1 truncation omits architecture section from combined output."""
    spec = _make_spec_with_architecture()

    # Full render should contain architecture content
    full_render = render_combined(spec, max_tokens=None)
    assert "Architecture detail paragraph" in full_render

    # Compute budget to trigger Level 1 but not Level 2
    full_tokens = _estimate_tokens(full_render)
    arch_content = spec.architecture
    arch_tokens = _estimate_tokens(arch_content)
    budget = full_tokens - arch_tokens + 5

    result = render_combined(spec, max_tokens=budget)
    assert "Architecture detail paragraph" not in result


# ---------------------------------------------------------------------------
# TS-03-18: Level 1 is no-op when no architecture; loop proceeds to Level 2
# Requirement: 03-REQ-5.3
# ---------------------------------------------------------------------------


def test_level1_noop_without_architecture() -> None:
    """Without architecture, Level 1 is skipped, Level 2 applied directly."""
    spec = _make_spec_without_architecture()
    assert spec.architecture is None

    # Use budget of 1 to force past all levels
    result = render_individual(spec, max_tokens=1)

    # Architecture was never present
    assert "architecture" not in result

    # The test_spec should be slimmed (Level 2) — verbose fields like
    # assertion_pseudocode, input, expected should be omitted
    ts_content = result.get("test_spec", "")
    # In Level 2, assertion_pseudocode should be absent
    assert "assert True" not in ts_content


# ---------------------------------------------------------------------------
# TS-03-19: PRD body content is never dropped at any truncation level
# Requirement: 03-REQ-5.4
# ---------------------------------------------------------------------------


def test_prd_body_never_dropped_individual() -> None:
    """PRD body content is present in render_individual at every truncation level."""
    spec = _make_spec_with_architecture()

    # Level 0 (no cap)
    result_l0 = render_individual(spec, max_tokens=None)
    assert "prd" in result_l0
    assert _PRD_BODY in result_l0["prd"] or "This is the PRD content" in result_l0["prd"]

    # Level 1 (moderate cap)
    level0_tokens = sum(_estimate_tokens(v) for v in result_l0.values())
    arch_tokens = _estimate_tokens(result_l0["architecture"])
    result_l1 = render_individual(spec, max_tokens=level0_tokens - arch_tokens + 5)
    assert "prd" in result_l1
    assert "This is the PRD content" in result_l1["prd"]

    # Level 2 (extreme cap)
    result_l2 = render_individual(spec, max_tokens=1)
    assert "prd" in result_l2
    assert "This is the PRD content" in result_l2["prd"]


def test_prd_body_never_dropped_combined() -> None:
    """PRD body content is present in render_combined at every truncation level."""
    spec = _make_spec_with_architecture()

    # Level 0
    result_l0 = render_combined(spec, max_tokens=None)
    assert "This is the PRD content" in result_l0

    # Level 2 (extreme cap forcing maximum truncation)
    result_l2 = render_combined(spec, max_tokens=1)
    assert "This is the PRD content" in result_l2


# ---------------------------------------------------------------------------
# 03-REQ-5.E1: architecture key with empty string value — remove key,
#              Level 1 is effectively a no-op
# ---------------------------------------------------------------------------


def test_level1_removes_empty_architecture_key() -> None:
    """Level 1 removes the 'architecture' key even when its value is empty."""
    spec = _make_spec_without_architecture()
    # Set architecture to empty string (not None) — Python includes the key
    spec = Spec(
        prd=spec.prd,
        requirements=spec.requirements,
        test_spec=spec.test_spec,
        tasks=spec.tasks,
        architecture="",
    )

    level0 = render_individual(spec, max_tokens=None)
    # With empty architecture, key should still be present at Level 0
    assert "architecture" in level0
    assert level0["architecture"] == ""

    # Force truncation — Level 1 should remove the empty architecture key
    result = render_individual(spec, max_tokens=1)
    assert "architecture" not in result


# ---------------------------------------------------------------------------
# 03-PROP-4: Level 1 removes architecture key when present
# Validates: 03-REQ-5.1
# ---------------------------------------------------------------------------


def test_property_level1_removes_architecture() -> None:
    """For any spec with architecture, Level 1 truncation removes the key."""
    spec = _make_spec_with_architecture()

    level0 = render_individual(spec, max_tokens=None)
    level0_tokens = sum(_estimate_tokens(v) for v in level0.values())
    arch_tokens = _estimate_tokens(level0["architecture"])

    # Budget that triggers Level 1 but not Level 2
    budget = level0_tokens - arch_tokens + 5
    result = render_individual(spec, max_tokens=budget)
    assert "architecture" not in result


# ---------------------------------------------------------------------------
# 03-PROP-7: PRD body is never absent from any render
# Validates: 03-REQ-5.4
# ---------------------------------------------------------------------------


def test_property_prd_always_present() -> None:
    """PRD body is present in every render regardless of max_tokens."""
    for spec in [_make_spec_with_architecture(), _make_spec_without_architecture()]:
        for budget in [None, 0, 1, 10, 1000, -5]:
            result = render_individual(spec, max_tokens=budget)
            assert "prd" in result
            assert len(result["prd"]) > 0

            result_combined = render_combined(spec, max_tokens=budget)
            assert "This is the PRD content" in result_combined
