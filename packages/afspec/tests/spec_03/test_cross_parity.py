"""Cross-implementation parity tests between Python and Go.

Test Spec: TS-03-28, TS-03-29
Requirements: 03-REQ-8

These tests verify that the Python rendering implementation produces output
structurally equivalent to what the Go implementation should produce for
the same spec and budget values. Since we cannot call Go from Python, we
validate against shared algorithmic contracts:

- estimate_tokens(text) == len(text) // 4 for any string (TS-03-29)
- render_individual with max_tokens applies the same truncation strategy:
  Level 0 (full), Level 1 (drop architecture), Level 2 (slim test spec)
- Same keys present/absent and same truncation level for same budget (TS-03-28)
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
# Helpers
# ---------------------------------------------------------------------------


def _estimate_tokens(text: str) -> int:
    """Import and call estimate_tokens at runtime."""
    from afspec import estimate_tokens

    return estimate_tokens(text)


def _make_parity_spec_with_architecture() -> Spec:
    """Build a Spec with architecture used for cross-implementation parity tests.

    Uses deterministic, known content so that both Python and Go
    implementations produce structurally comparable output.
    """
    arch_content = "# Architecture\n\n" + ("Architecture detail paragraph. " * 100)

    return Spec(
        prd=PRDDocument(
            frontmatter=PRDFrontmatter(spec_id="03", spec_name="parity_test"),
            body="# Parity Test PRD\n\nPRD body for parity testing.",
        ),
        requirements=Requirements(
            spec_id="03",
            spec_name="parity_test",
            introduction="Introduction for parity test.",
            glossary={"term": "definition"},
            requirements=[
                Requirement(
                    id="03-REQ-1",
                    title="Parity requirement",
                    user_story=UserStory(role="user", goal="test", benefit="parity"),
                    acceptance_criteria=[
                        Criterion(
                            id="03-REQ-1.1",
                            ears_pattern=EARSPattern.UBIQUITOUS,
                            system="the system",
                            action="do parity check",
                        ),
                    ],
                ),
            ],
        ),
        test_spec=TestSpec(
            spec_id="03",
            spec_name="parity_test",
            test_cases=[
                TestCase(
                    id="TS-03-1",
                    requirement_id="03-REQ-1.1",
                    kind="unit",
                    description="Parity test case",
                    preconditions=["precondition"],
                    input="test input",
                    expected="expected output",
                    assertion_pseudocode="assert result == expected",
                ),
            ],
            coverage=Coverage(requirements_covered=["03-REQ-1"]),
        ),
        tasks=Tasks(
            spec_id="03",
            spec_name="parity_test",
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


def _make_parity_spec_without_architecture() -> Spec:
    """Build a Spec without architecture for cross-implementation parity."""
    return Spec(
        prd=PRDDocument(
            frontmatter=PRDFrontmatter(spec_id="03", spec_name="parity_noarch"),
            body="# Parity No-Arch PRD\n\nPRD body.",
        ),
        requirements=Requirements(
            spec_id="03",
            spec_name="parity_noarch",
            introduction="Intro.",
            glossary={"term": "def"},
            requirements=[
                Requirement(
                    id="03-REQ-1",
                    title="Parity req",
                    user_story=UserStory(role="user", goal="test", benefit="val"),
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
            spec_name="parity_noarch",
            test_cases=[
                TestCase(
                    id="TS-03-1",
                    requirement_id="03-REQ-1.1",
                    kind="unit",
                    description="Test case",
                ),
            ],
            coverage=Coverage(requirements_covered=["03-REQ-1"]),
        ),
        tasks=Tasks(
            spec_id="03",
            spec_name="parity_noarch",
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
# TS-03-29: Python estimate_tokens and Go EstimateTokens return the same
#           integer for any string input.
# Requirement: 03-REQ-8.2
# ---------------------------------------------------------------------------


def test_estimate_tokens_parity_empty() -> None:
    """Both Python and Go: estimate_tokens('') == 0."""
    assert _estimate_tokens("") == len("") // 4 == 0


def test_estimate_tokens_parity_single_char() -> None:
    """Both Python and Go: estimate_tokens('a') == 0."""
    assert _estimate_tokens("a") == len("a") // 4 == 0


def test_estimate_tokens_parity_four_chars() -> None:
    """Both Python and Go: estimate_tokens('abcd') == 1."""
    assert _estimate_tokens("abcd") == len("abcd") // 4 == 1


def test_estimate_tokens_parity_sixteen_chars() -> None:
    """Both Python and Go: estimate_tokens of 16 chars == 4."""
    text = "abcdefghijklmnop"
    assert _estimate_tokens(text) == len(text) // 4 == 4


def test_estimate_tokens_parity_thousand_chars() -> None:
    """Both Python and Go: estimate_tokens of 1000 chars == 250."""
    text = "x" * 1000
    assert _estimate_tokens(text) == len(text) // 4 == 250


def test_estimate_tokens_parity_formula_match() -> None:
    """Verify the chars/4 formula for a range of string lengths.

    Both Python estimate_tokens and Go EstimateTokens must return
    len(text) // 4 for any input string.
    """
    test_strings = [
        "",
        "a",
        "ab",
        "abc",
        "abcd",
        "abcde",
        "abcdefghijklmnop",
        "x" * 100,
        "y" * 999,
        "z" * 1000,
    ]
    for text in test_strings:
        result = _estimate_tokens(text)
        expected = len(text) // 4
        assert result == expected, f"estimate_tokens({text!r:.20}) returned {result}, expected {expected}"


# ---------------------------------------------------------------------------
# TS-03-28: Python and Go produce structurally equivalent output
#           (same keys present/absent, same truncation level).
# Requirement: 03-REQ-8.1
# ---------------------------------------------------------------------------


def _truncation_level(result: dict[str, str], spec: Spec) -> int:
    """Determine the truncation level of a render result.

    Level 0: Full render (architecture present if spec has it).
    Level 1: Architecture dropped, test spec is full.
    Level 2: Architecture dropped, test spec is slim
             (assertion_pseudocode absent).
    """
    has_architecture_in_spec = spec.architecture is not None

    if has_architecture_in_spec and "architecture" in result:
        return 0

    # Architecture absent (either dropped or never present)
    ts_content = result.get("test_spec", "")

    # If spec has test cases with assertion_pseudocode, check if it's present
    if spec.test_spec and spec.test_spec.test_cases:
        for tc in spec.test_spec.test_cases:
            if tc.assertion_pseudocode and tc.assertion_pseudocode in ts_content:
                # Full test spec is present — Level 1
                return 1 if has_architecture_in_spec else 0

    # assertion_pseudocode is absent — Level 2
    return 2 if has_architecture_in_spec else (1 if ts_content else 0)


def test_parity_same_keys_level0() -> None:
    """At Level 0 (no truncation), Python produces expected keys.

    Go must produce the same keys: prd, requirements, test_spec, tasks,
    architecture.
    """
    spec = _make_parity_spec_with_architecture()
    result = render_individual(spec, max_tokens=None)

    expected_keys = {"prd", "requirements", "test_spec", "tasks", "architecture"}
    assert set(result.keys()) == expected_keys


def test_parity_same_keys_level1() -> None:
    """At Level 1, architecture is absent. Same must hold for Go.

    Both implementations should drop only the architecture key when
    Level 1 truncation is triggered.
    """
    spec = _make_parity_spec_with_architecture()

    # Compute budget that triggers Level 1
    level0 = render_individual(spec, max_tokens=None)
    level0_tokens = sum(_estimate_tokens(v) for v in level0.values())
    arch_tokens = _estimate_tokens(level0["architecture"])
    budget = level0_tokens - arch_tokens + 5

    result = render_individual(spec, max_tokens=budget)

    expected_keys = {"prd", "requirements", "test_spec", "tasks"}
    assert set(result.keys()) == expected_keys
    assert "architecture" not in result


def test_parity_same_keys_level2() -> None:
    """At Level 2, architecture absent and test_spec is slim.

    Both Python and Go must produce the same key set.
    """
    spec = _make_parity_spec_with_architecture()
    result = render_individual(spec, max_tokens=1)

    expected_keys = {"prd", "requirements", "test_spec", "tasks"}
    assert set(result.keys()) == expected_keys


def test_parity_truncation_level_matches() -> None:
    """Both Python and Go apply the same truncation level for same budget.

    For a spec with architecture:
    - High budget -> Level 0 (full)
    - Medium budget (architecture removed) -> Level 1
    - Low budget -> Level 2
    """
    spec = _make_parity_spec_with_architecture()

    level0 = render_individual(spec, max_tokens=None)
    level0_tokens = sum(_estimate_tokens(v) for v in level0.values())
    arch_tokens = _estimate_tokens(level0["architecture"])

    # Level 0: sufficient budget
    result_l0 = render_individual(spec, max_tokens=level0_tokens + 1000)
    assert _truncation_level(result_l0, spec) == 0

    # Level 1: budget triggers architecture removal
    budget_l1 = level0_tokens - arch_tokens + 5
    result_l1 = render_individual(spec, max_tokens=budget_l1)
    assert _truncation_level(result_l1, spec) == 1

    # Level 2: very small budget
    result_l2 = render_individual(spec, max_tokens=1)
    assert _truncation_level(result_l2, spec) == 2


def test_parity_no_architecture_same_keys() -> None:
    """Without architecture, Python produces same keys as Go should.

    Both should have: prd, requirements, test_spec, tasks.
    No architecture key.
    """
    spec = _make_parity_spec_without_architecture()
    result = render_individual(spec, max_tokens=None)

    expected_keys = {"prd", "requirements", "test_spec", "tasks"}
    assert set(result.keys()) == expected_keys
    assert "architecture" not in result


def test_parity_no_architecture_truncation_skips_level1() -> None:
    """Without architecture, Level 1 is skipped, proceeding to Level 2.

    Both Python and Go must exhibit this behavior.
    """
    spec = _make_parity_spec_without_architecture()

    # Force truncation — should go straight from Level 0 to Level 2
    result = render_individual(spec, max_tokens=1)

    assert "architecture" not in result
    assert "prd" in result
    assert "test_spec" in result


# ---------------------------------------------------------------------------
# 03-REQ-8.E1: Same spec with architecture, Level 1 sufficient in both
# ---------------------------------------------------------------------------


def test_parity_level1_both_match_architecture_absent() -> None:
    """When Level 1 is sufficient, both Python and Go return results without
    architecture and with test spec intact.
    """
    spec = _make_parity_spec_with_architecture()

    level0 = render_individual(spec, max_tokens=None)
    level0_tokens = sum(_estimate_tokens(v) for v in level0.values())
    arch_tokens = _estimate_tokens(level0["architecture"])

    # Budget so that Level 1 (removing architecture) fits
    budget = level0_tokens - arch_tokens + 5
    result = render_individual(spec, max_tokens=budget)

    # Architecture absent
    assert "architecture" not in result

    # Test spec intact (assertion_pseudocode still present)
    ts = result.get("test_spec", "")
    assert "assert result == expected" in ts
