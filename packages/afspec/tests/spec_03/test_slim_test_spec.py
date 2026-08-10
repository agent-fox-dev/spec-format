"""Tests for Level 2 slim test spec rendering.

Test Spec: TS-03-20, TS-03-21, TS-03-22
Requirements: 03-REQ-6
"""

from __future__ import annotations

from afspec.models import (
    Coverage,
    EdgeCaseTest,
    PropertyTest,
    SmokeTest,
    TestCase,
    TestSpec,
)

# ---------------------------------------------------------------------------
# Deferred imports for functions that don't exist yet
# ---------------------------------------------------------------------------


def _render_test_spec_slim(ts: TestSpec) -> str:
    """Import and call _render_test_spec_slim at runtime (not yet implemented)."""
    from afspec.render import _render_test_spec_slim as fn

    return fn(ts)


def _render_test_spec_scoped_slim(ts: TestSpec, test_spec_ids: set[str]) -> str:
    """Import and call _render_test_spec_scoped_slim at runtime (not yet implemented)."""
    from afspec.render import _render_test_spec_scoped_slim as fn

    return fn(ts, test_spec_ids)


# ---------------------------------------------------------------------------
# Fixture helpers
# ---------------------------------------------------------------------------


def _make_test_spec_with_verbose_entries() -> TestSpec:
    """Build a TestSpec with all entry types populated with verbose fields.

    Includes test cases with assertion_pseudocode/input/expected,
    property tests with for_any_strategy/invariant_check,
    edge case tests with assertion_pseudocode/input/expected,
    and smoke tests with expected_effects.
    """
    return TestSpec(
        spec_id="03",
        spec_name="slim_test",
        test_cases=[
            TestCase(
                id="TS-03-1",
                requirement_id="03-REQ-1.1",
                kind="unit",
                description="Test case one description",
                preconditions=["A valid spec exists"],
                input="detailed input data for test one",
                expected="detailed expected output for test one",
                assertion_pseudocode="assert result == expected_output_one",
            ),
            TestCase(
                id="TS-03-2",
                requirement_id="03-REQ-2.1",
                kind="integration",
                description="Test case two description",
                preconditions=["System is running"],
                input="detailed input data for test two",
                expected="detailed expected output for test two",
                assertion_pseudocode="assert response.status == 200",
            ),
        ],
        property_tests=[
            PropertyTest(
                id="PT-03-1",
                property_id="03-PROP-1",
                validates=["03-REQ-1.1"],
                description="Property test one description",
                for_any_strategy="any integer x in range(0, 1000)",
                invariant_check="assert f(x) >= 0",
            ),
        ],
        edge_case_tests=[
            EdgeCaseTest(
                id="EC-03-1",
                requirement_id="03-REQ-1.1",
                kind="unit",
                description="Edge case one description",
                preconditions=["Empty input provided"],
                input="",
                expected="empty result",
                assertion_pseudocode="assert handle_empty() == default_value",
            ),
        ],
        smoke_tests=[
            SmokeTest(
                id="SM-03-1",
                execution_path_id="EP-03-1",
                description="Smoke test one description",
                trigger="POST /api/render",
                real_components=["render_engine"],
                mockable=["database"],
                expected_effects=["renders output", "logs event", "updates cache"],
            ),
        ],
        coverage=Coverage(
            requirements_covered=["03-REQ-1", "03-REQ-2"],
        ),
    )


def _make_test_spec_with_multiple_entries() -> TestSpec:
    """Build a TestSpec with multiple entries for scoped filtering tests."""
    return TestSpec(
        spec_id="03",
        spec_name="scoped_slim_test",
        test_cases=[
            TestCase(
                id="TS-03-1",
                requirement_id="03-REQ-1.1",
                kind="unit",
                description="First test case for scoping",
                preconditions=["precond 1"],
                input="input one",
                expected="expected one",
                assertion_pseudocode="assert first_result == 1",
            ),
            TestCase(
                id="TS-03-2",
                requirement_id="03-REQ-1.2",
                kind="unit",
                description="Second test case excluded from scope",
                preconditions=["precond 2"],
                input="input two",
                expected="expected two",
                assertion_pseudocode="assert second_result == 2",
            ),
            TestCase(
                id="TS-03-3",
                requirement_id="03-REQ-2.1",
                kind="integration",
                description="Third test case for scoping",
                preconditions=["precond 3"],
                input="input three",
                expected="expected three",
                assertion_pseudocode="assert third_result == 3",
            ),
        ],
        property_tests=[
            PropertyTest(
                id="PT-03-1",
                property_id="03-PROP-1",
                validates=["03-REQ-1.1"],
                description="Property test for scope",
                for_any_strategy="any string s of length > 0",
                invariant_check="assert len(process(s)) > 0",
            ),
        ],
        edge_case_tests=[
            EdgeCaseTest(
                id="EC-03-1",
                requirement_id="03-REQ-1.1",
                kind="unit",
                description="Edge case in scope",
                input="edge input",
                expected="edge expected",
                assertion_pseudocode="assert edge_case_handled()",
            ),
        ],
        smoke_tests=[
            SmokeTest(
                id="SM-03-1",
                execution_path_id="EP-03-1",
                description="Smoke test in scope",
                trigger="GET /health",
                expected_effects=["returns 200"],
            ),
        ],
        coverage=Coverage(requirements_covered=["03-REQ-1", "03-REQ-2"]),
    )


# ---------------------------------------------------------------------------
# TS-03-20: _render_test_spec_slim omits verbose fields from all entry types
# Requirement: 03-REQ-6.1
# ---------------------------------------------------------------------------


def test_slim_omits_assertion_pseudocode_from_test_cases() -> None:
    """Slim render omits assertion_pseudocode content from test case entries."""
    ts = _make_test_spec_with_verbose_entries()
    result = _render_test_spec_slim(ts)

    # Verbose assertion_pseudocode content should be absent
    assert "assert result == expected_output_one" not in result
    assert "assert response.status == 200" not in result


def test_slim_omits_input_from_test_cases() -> None:
    """Slim render omits input field content from test case entries."""
    ts = _make_test_spec_with_verbose_entries()
    result = _render_test_spec_slim(ts)

    # Input field content should be absent
    assert "detailed input data for test one" not in result
    assert "detailed input data for test two" not in result


def test_slim_omits_expected_from_test_cases() -> None:
    """Slim render omits expected field content from test case entries."""
    ts = _make_test_spec_with_verbose_entries()
    result = _render_test_spec_slim(ts)

    # Expected field content should be absent
    assert "detailed expected output for test one" not in result
    assert "detailed expected output for test two" not in result


def test_slim_omits_for_any_strategy_from_property_tests() -> None:
    """Slim render omits for_any_strategy from property test entries."""
    ts = _make_test_spec_with_verbose_entries()
    result = _render_test_spec_slim(ts)

    assert "any integer x in range(0, 1000)" not in result


def test_slim_omits_invariant_check_from_property_tests() -> None:
    """Slim render omits invariant_check from property test entries."""
    ts = _make_test_spec_with_verbose_entries()
    result = _render_test_spec_slim(ts)

    assert "assert f(x) >= 0" not in result


def test_slim_omits_expected_effects_from_smoke_tests() -> None:
    """Slim render omits expected_effects from smoke test entries."""
    ts = _make_test_spec_with_verbose_entries()
    result = _render_test_spec_slim(ts)

    assert "renders output" not in result
    assert "logs event" not in result
    assert "updates cache" not in result


def test_slim_omits_verbose_fields_from_edge_case_tests() -> None:
    """Slim render omits assertion_pseudocode, input, expected from edge cases."""
    ts = _make_test_spec_with_verbose_entries()
    result = _render_test_spec_slim(ts)

    assert "assert handle_empty() == default_value" not in result
    assert "empty result" not in result


def test_slim_preserves_all_test_entry_ids() -> None:
    """Slim render preserves all test entry IDs."""
    ts = _make_test_spec_with_verbose_entries()
    result = _render_test_spec_slim(ts)

    assert "TS-03-1" in result
    assert "TS-03-2" in result
    assert "PT-03-1" in result
    assert "EC-03-1" in result
    assert "SM-03-1" in result


def test_slim_preserves_all_test_entry_descriptions() -> None:
    """Slim render preserves all test entry descriptions."""
    ts = _make_test_spec_with_verbose_entries()
    result = _render_test_spec_slim(ts)

    assert "Test case one description" in result
    assert "Test case two description" in result
    assert "Property test one description" in result
    assert "Edge case one description" in result
    assert "Smoke test one description" in result


# ---------------------------------------------------------------------------
# TS-03-21: _render_test_spec_scoped_slim renders scoped TestSpec with
#           the same field omissions, filtered to relevant test spec IDs
# Requirement: 03-REQ-6.2
# ---------------------------------------------------------------------------


def test_scoped_slim_includes_in_scope_entries() -> None:
    """Scoped slim render includes only in-scope test entries."""
    ts = _make_test_spec_with_multiple_entries()
    ids = {"TS-03-1", "TS-03-3"}
    result = _render_test_spec_scoped_slim(ts, ids)

    assert "TS-03-1" in result
    assert "TS-03-3" in result


def test_scoped_slim_excludes_out_of_scope_entries() -> None:
    """Scoped slim render excludes out-of-scope test entries."""
    ts = _make_test_spec_with_multiple_entries()
    ids = {"TS-03-1", "TS-03-3"}
    result = _render_test_spec_scoped_slim(ts, ids)

    # TS-03-2 is not in scope — its unique description should be absent
    assert "Second test case excluded from scope" not in result


def test_scoped_slim_omits_verbose_fields() -> None:
    """Scoped slim render omits verbose fields from in-scope entries."""
    ts = _make_test_spec_with_multiple_entries()
    ids = {"TS-03-1", "TS-03-3"}
    result = _render_test_spec_scoped_slim(ts, ids)

    # assertion_pseudocode content from in-scope entries should be absent
    assert "assert first_result == 1" not in result
    assert "assert third_result == 3" not in result

    # expected content should be absent
    assert "expected one" not in result
    assert "expected three" not in result

    # input content should be absent
    assert "input one" not in result
    assert "input three" not in result


def test_scoped_slim_preserves_ids_and_descriptions() -> None:
    """Scoped slim render preserves id, description for in-scope entries."""
    ts = _make_test_spec_with_multiple_entries()
    ids = {"TS-03-1", "TS-03-3"}
    result = _render_test_spec_scoped_slim(ts, ids)

    assert "TS-03-1" in result
    assert "First test case for scoping" in result
    assert "TS-03-3" in result
    assert "Third test case for scoping" in result


# ---------------------------------------------------------------------------
# TS-03-22: Level 2 slim rendering preserves test entry id, description,
#           type, and requirement linkage fields
# Requirement: 03-REQ-6.3
# ---------------------------------------------------------------------------


def test_slim_preserves_requirement_linkage() -> None:
    """Slim render preserves requirement_id for each test case entry."""
    ts = _make_test_spec_with_verbose_entries()
    result = _render_test_spec_slim(ts)

    for tc in ts.test_cases:
        assert tc.id in result
        assert tc.description in result
        assert tc.requirement_id in result


def test_slim_preserves_type_field() -> None:
    """Slim render preserves type/kind for test case and edge case entries."""
    ts = _make_test_spec_with_verbose_entries()
    result = _render_test_spec_slim(ts)

    # Check that kind values appear (they are rendered as **Type:** kind)
    assert "unit" in result
    assert "integration" in result


def test_slim_preserves_property_test_linkage() -> None:
    """Slim render preserves property test's property_id and validates."""
    ts = _make_test_spec_with_verbose_entries()
    result = _render_test_spec_slim(ts)

    assert "PT-03-1" in result
    assert "Property test one description" in result
    assert "03-PROP-1" in result
    assert "03-REQ-1.1" in result


def test_slim_preserves_edge_case_requirement_linkage() -> None:
    """Slim render preserves edge case test requirement_id."""
    ts = _make_test_spec_with_verbose_entries()
    result = _render_test_spec_slim(ts)

    assert "EC-03-1" in result
    assert "Edge case one description" in result
    assert "03-REQ-1.1" in result


def test_slim_preserves_smoke_test_execution_path() -> None:
    """Slim render preserves smoke test execution_path_id."""
    ts = _make_test_spec_with_verbose_entries()
    result = _render_test_spec_slim(ts)

    assert "SM-03-1" in result
    assert "Smoke test one description" in result
    assert "EP-03-1" in result


# ---------------------------------------------------------------------------
# 03-REQ-6.E1: No TestSpec — section absent or empty, no error
# ---------------------------------------------------------------------------


def test_slim_empty_test_spec() -> None:
    """Slim render of empty TestSpec produces output without error."""
    ts = TestSpec(
        spec_id="03",
        spec_name="empty_slim",
        test_cases=[],
        property_tests=[],
        edge_case_tests=[],
        smoke_tests=[],
        coverage=Coverage(),
    )
    result = _render_test_spec_slim(ts)
    # Should not raise — result should be a string (possibly empty or minimal)
    assert isinstance(result, str)


def test_scoped_slim_empty_test_spec() -> None:
    """Scoped slim render of empty TestSpec produces output without error."""
    ts = TestSpec(
        spec_id="03",
        spec_name="empty_scoped_slim",
        test_cases=[],
        property_tests=[],
        edge_case_tests=[],
        smoke_tests=[],
        coverage=Coverage(),
    )
    result = _render_test_spec_scoped_slim(ts, {"TS-03-1"})
    assert isinstance(result, str)


# ---------------------------------------------------------------------------
# 03-REQ-6.E2: Fields already absent — renders without error
# ---------------------------------------------------------------------------


def test_slim_fields_already_absent() -> None:
    """Slim render handles entries where verbose fields are already absent/empty."""
    ts = TestSpec(
        spec_id="03",
        spec_name="sparse_slim",
        test_cases=[
            TestCase(
                id="TS-03-SPARSE",
                requirement_id="03-REQ-1.1",
                kind="unit",
                description="Sparse test case with no verbose fields",
                # No preconditions, input, expected, or assertion_pseudocode
            ),
        ],
        property_tests=[
            PropertyTest(
                id="PT-03-SPARSE",
                property_id="03-PROP-1",
                validates=["03-REQ-1.1"],
                description="Sparse property test",
                # No for_any_strategy or invariant_check
            ),
        ],
        edge_case_tests=[
            EdgeCaseTest(
                id="EC-03-SPARSE",
                requirement_id="03-REQ-1.1",
                kind="unit",
                description="Sparse edge case test",
                # No input, expected, or assertion_pseudocode
            ),
        ],
        smoke_tests=[
            SmokeTest(
                id="SM-03-SPARSE",
                execution_path_id="EP-03-1",
                description="Sparse smoke test",
                # No expected_effects
            ),
        ],
        coverage=Coverage(requirements_covered=["03-REQ-1"]),
    )

    result = _render_test_spec_slim(ts)

    # Should render without error
    assert isinstance(result, str)

    # IDs and descriptions should be present
    assert "TS-03-SPARSE" in result
    assert "PT-03-SPARSE" in result
    assert "EC-03-SPARSE" in result
    assert "SM-03-SPARSE" in result
    assert "Sparse test case with no verbose fields" in result
    assert "Sparse property test" in result
    assert "Sparse edge case test" in result
    assert "Sparse smoke test" in result


# ---------------------------------------------------------------------------
# 03-PROP-5: Level 2 slim render omits verbose assertion fields
# Validates: 03-REQ-6.1, 03-REQ-6.3
# ---------------------------------------------------------------------------


def test_property_slim_never_contains_verbose_fields() -> None:
    """For a TestSpec with verbose entries, slim render never contains them."""
    ts = _make_test_spec_with_verbose_entries()
    result = _render_test_spec_slim(ts)

    # All verbose fields from test cases
    assert "assert result == expected_output_one" not in result
    assert "assert response.status == 200" not in result
    assert "detailed input data for test one" not in result
    assert "detailed input data for test two" not in result
    assert "detailed expected output for test one" not in result
    assert "detailed expected output for test two" not in result

    # Property test verbose fields
    assert "any integer x in range(0, 1000)" not in result
    assert "assert f(x) >= 0" not in result

    # Smoke test verbose fields
    assert "renders output" not in result
    assert "logs event" not in result
    assert "updates cache" not in result

    # Edge case verbose fields
    assert "assert handle_empty() == default_value" not in result

    # But IDs and descriptions are preserved
    assert "TS-03-1" in result
    assert "TS-03-2" in result
    assert "PT-03-1" in result
    assert "EC-03-1" in result
    assert "SM-03-1" in result
