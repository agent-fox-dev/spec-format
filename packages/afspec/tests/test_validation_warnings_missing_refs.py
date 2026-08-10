"""Tests for missing subtask refs validation warning.

TS-02-11: validate / Validate invokes _check_missing_subtask_refs and
          appends its warnings to the ValidationResult when subtasks
          have empty requirement_refs or test_spec_refs.

TS-02-12: _check_missing_subtask_refs emits exactly one warning per
          subtask with the correct message format including the joined
          field names.

TS-02-13: _check_missing_subtask_refs skips the entire
          WIRING_VERIFICATION TaskGroup and emits no warnings for its
          subtasks even if they have empty refs.

TS-02-14: The field_names portion of the warning message uses
          ' and '.join(missing), producing 'requirement_refs and
          test_spec_refs' when both are empty.

These tests are in RED PHASE — they will fail because
_check_missing_subtask_refs does not exist yet in this branch and
validate() does not yet invoke it.
"""

from __future__ import annotations

import afspec.validation as validation_mod
from afspec.models import (
    Criterion,
    EARSPattern,
    ExecutionPath,
    PathStep,
    PRDDocument,
    PRDFrontmatter,
    Requirement,
    Requirements,
    SmokeTest,
    Spec,
    Subtask,
    TaskGroup,
    TaskGroupKind,
    Tasks,
    TestCase,
    TestSpec,
    TraceabilityEntry,
    UserStory,
    VerificationSubtask,
)
from afspec.validation import validate

# ---------------------------------------------------------------------------
# Helpers — build structurally valid specs with configurable subtask refs
# ---------------------------------------------------------------------------

_SPEC_ID = "MR"
_REQ_ID = f"{_SPEC_ID}-REQ-1"
_CRITERION_ID = f"{_SPEC_ID}-REQ-1.1"
_TS_ID = f"TS-{_SPEC_ID}-1"
_SMOKE_TS_ID = f"TS-{_SPEC_ID}-SMOKE-1"
_PATH_ID = f"{_SPEC_ID}-PATH-1"


def _build_valid_spec(
    middle_groups: list[TaskGroup],
    *,
    traceability: list[TraceabilityEntry] | None = None,
) -> Spec:
    """Build a structurally valid Spec with configurable middle groups.

    Wraps ``middle_groups`` between a ``kind=tests`` first group and a
    ``kind=wiring_verification`` final group to satisfy structural
    validation rules.  The tests group (id=1) has a subtask with
    populated refs so it does NOT trigger missing-refs warnings itself.
    """
    tests_group = TaskGroup(
        id=1,
        kind=TaskGroupKind.TESTS,
        title="Tests group",
        subtasks=[
            Subtask(
                id="1.1",
                title="Write tests",
                requirement_refs=[_REQ_ID],
                test_spec_refs=[_TS_ID],
            ),
        ],
        verification=VerificationSubtask(id="1.V", checks=["pass"]),
    )

    all_groups = [tests_group, *middle_groups]

    wv_id = max(g.id for g in all_groups) + 1
    wv_group = TaskGroup(
        id=wv_id,
        kind=TaskGroupKind.WIRING_VERIFICATION,
        title="Wiring verification",
        subtasks=[
            Subtask(
                id=f"{wv_id}.1",
                title="Trace execution paths and stub/dead-code audit",
                test_spec_refs=[_SMOKE_TS_ID],
                requirement_refs=[_REQ_ID],
            ),
        ],
        verification=VerificationSubtask(id=f"{wv_id}.V", checks=["done"]),
    )
    all_groups.append(wv_group)

    return Spec(
        prd=PRDDocument(
            frontmatter=PRDFrontmatter(
                spec_id=_SPEC_ID,
                spec_name="missing_refs_test",
                title="Missing Refs Test",
                created_at="2024-01-01",
                updated_at="2024-01-01",
                owner="test",
                source="internal",
            ),
            body="Test spec for missing refs validation warning.",
        ),
        requirements=Requirements(
            spec_id=_SPEC_ID,
            spec_name="missing_refs_test",
            introduction="Test spec for missing refs validation warning.",
            requirements=[
                Requirement(
                    id=_REQ_ID,
                    title="Missing refs test requirement",
                    user_story=UserStory(
                        role="developer",
                        goal="test validation",
                        benefit="coverage",
                    ),
                    acceptance_criteria=[
                        Criterion(
                            id=_CRITERION_ID,
                            ears_pattern=EARSPattern.UBIQUITOUS,
                            system="the system",
                            action="performs the action",
                        ),
                    ],
                ),
            ],
            execution_paths=[
                ExecutionPath(
                    id=_PATH_ID,
                    title="Main path",
                    steps=[
                        PathStep(actor="User", action="Invoke"),
                        PathStep(actor="System", action="Respond"),
                    ],
                ),
            ],
        ),
        test_spec=TestSpec(
            spec_id=_SPEC_ID,
            spec_name="missing_refs_test",
            test_cases=[
                TestCase(
                    id=_TS_ID,
                    requirement_id=_CRITERION_ID,
                    kind="unit",
                    description="Test case for missing refs",
                ),
            ],
            smoke_tests=[
                SmokeTest(
                    id=_SMOKE_TS_ID,
                    execution_path_id=_PATH_ID,
                    description="Wiring smoke test",
                ),
            ],
        ),
        tasks=Tasks(
            spec_id=_SPEC_ID,
            spec_name="missing_refs_test",
            task_groups=all_groups,
            traceability=traceability
            or [
                TraceabilityEntry(
                    requirement_id=_CRITERION_ID,
                    test_spec_id=_TS_ID,
                    task_id="1.1",
                ),
            ],
        ),
    )


def _collect_missing_refs_warnings(spec: Spec) -> list:
    """Call _check_missing_subtask_refs for each group, collecting all warnings.

    Mirrors how ``validate()`` invokes the check per group.
    """
    warnings = []
    for group in spec.tasks.task_groups:
        warnings.extend(validation_mod._check_missing_subtask_refs(group))
    return warnings


# ---------------------------------------------------------------------------
# TS-02-11: validate invokes _check_missing_subtask_refs (02-REQ-4.1)
# ---------------------------------------------------------------------------


class TestValidateInvokesCheckMissingSubtaskRefs:
    """TS-02-11: validate() should invoke _check_missing_subtask_refs and
    append its warnings to the ValidationResult when subtasks have empty
    requirement_refs or test_spec_refs.
    """

    def test_warning_present_for_subtask_with_empty_requirement_refs(self) -> None:
        """validate() emits a warning for a subtask with empty requirement_refs."""
        spec = _build_valid_spec(
            middle_groups=[
                TaskGroup(
                    id=2,
                    kind=TaskGroupKind.STANDARD,
                    title="Standard group",
                    subtasks=[
                        Subtask(
                            id="2.1",
                            title="Implement feature",
                            requirement_refs=[],
                            test_spec_refs=[_TS_ID],
                        ),
                    ],
                    verification=VerificationSubtask(id="2.V", checks=["pass"]),
                ),
            ],
        )
        result = validate(spec)
        warning_ids = [w.entity_id for w in result.warnings]
        assert "2.1" in warning_ids, f"Expected warning for subtask '2.1', got entity_ids: {warning_ids}"

    def test_warning_message_contains_requirement_refs(self) -> None:
        """Warning message references 'requirement_refs' when only req refs empty."""
        spec = _build_valid_spec(
            middle_groups=[
                TaskGroup(
                    id=2,
                    kind=TaskGroupKind.STANDARD,
                    title="Standard group",
                    subtasks=[
                        Subtask(
                            id="2.1",
                            title="Implement feature",
                            requirement_refs=[],
                            test_spec_refs=[_TS_ID],
                        ),
                    ],
                    verification=VerificationSubtask(id="2.V", checks=["pass"]),
                ),
            ],
        )
        result = validate(spec)
        subtask_warnings = [w for w in result.warnings if w.entity_id == "2.1"]
        assert len(subtask_warnings) >= 1, f"Expected at least one warning for subtask 2.1, got: {result.warnings}"
        assert "requirement_refs" in subtask_warnings[0].message

    def test_warning_message_matches_expected_format(self) -> None:
        """Warning follows format: 'Subtask {id} has empty requirement_refs — ...'."""
        spec = _build_valid_spec(
            middle_groups=[
                TaskGroup(
                    id=2,
                    kind=TaskGroupKind.STANDARD,
                    title="Standard group",
                    subtasks=[
                        Subtask(
                            id="2.1",
                            title="Implement feature",
                            requirement_refs=[],
                            test_spec_refs=[_TS_ID],
                        ),
                    ],
                    verification=VerificationSubtask(id="2.V", checks=["pass"]),
                ),
            ],
        )
        result = validate(spec)
        subtask_warnings = [w for w in result.warnings if w.entity_id == "2.1"]
        expected_msg = "Subtask 2.1 has empty requirement_refs — scoped rendering will fall back to full spec dump"
        assert any(w.message == expected_msg for w in subtask_warnings), (
            f"Expected message '{expected_msg}', got: {[w.message for w in subtask_warnings]}"
        )

    def test_warnings_collection_has_at_least_one_entry(self) -> None:
        """ValidationResult.warnings has at least one entry for the affected subtask."""
        spec = _build_valid_spec(
            middle_groups=[
                TaskGroup(
                    id=2,
                    kind=TaskGroupKind.STANDARD,
                    title="Standard group",
                    subtasks=[
                        Subtask(
                            id="2.1",
                            title="Implement feature",
                            requirement_refs=[],
                            test_spec_refs=[],
                        ),
                    ],
                    verification=VerificationSubtask(id="2.V", checks=["pass"]),
                ),
            ],
        )
        result = validate(spec)
        assert len(result.warnings) >= 1


# ---------------------------------------------------------------------------
# TS-02-12: Exactly one warning per subtask with correct message format
# (02-REQ-4.2)
# ---------------------------------------------------------------------------


class TestCheckMissingSubtaskRefsMessageFormat:
    """TS-02-12: _check_missing_subtask_refs emits exactly one warning per
    subtask with the correct message format including the joined field
    names.  Three subtasks with different empty-ref combinations produce
    exactly three warnings.
    """

    def test_function_exists(self) -> None:
        """_check_missing_subtask_refs is defined in afspec.validation."""
        assert hasattr(validation_mod, "_check_missing_subtask_refs"), (
            "_check_missing_subtask_refs not defined in afspec.validation"
        )

    def test_exactly_three_warnings_for_three_affected_subtasks(self) -> None:
        """Three subtasks with different empty-ref combos produce exactly 3 warnings."""
        assert hasattr(validation_mod, "_check_missing_subtask_refs"), "_check_missing_subtask_refs not defined"
        group = TaskGroup(
            id=2,
            kind=TaskGroupKind.STANDARD,
            title="Standard group",
            subtasks=[
                Subtask(
                    id="2.1",
                    title="Sub 1",
                    requirement_refs=[],
                    test_spec_refs=[_TS_ID],
                ),
                Subtask(
                    id="2.2",
                    title="Sub 2",
                    requirement_refs=[_REQ_ID],
                    test_spec_refs=[],
                ),
                Subtask(
                    id="2.3",
                    title="Sub 3",
                    requirement_refs=[],
                    test_spec_refs=[],
                ),
            ],
            verification=VerificationSubtask(id="2.V", checks=["pass"]),
        )
        warnings = validation_mod._check_missing_subtask_refs(group)
        assert len(warnings) == 3

    def test_empty_requirement_refs_only_message(self) -> None:
        """Subtask with empty requirement_refs only produces correct message."""
        assert hasattr(validation_mod, "_check_missing_subtask_refs"), "_check_missing_subtask_refs not defined"
        group = TaskGroup(
            id=2,
            kind=TaskGroupKind.STANDARD,
            title="Standard group",
            subtasks=[
                Subtask(
                    id="2.1",
                    title="Sub 1",
                    requirement_refs=[],
                    test_spec_refs=[_TS_ID],
                ),
                Subtask(
                    id="2.2",
                    title="Sub 2",
                    requirement_refs=[_REQ_ID],
                    test_spec_refs=[],
                ),
                Subtask(
                    id="2.3",
                    title="Sub 3",
                    requirement_refs=[],
                    test_spec_refs=[],
                ),
            ],
            verification=VerificationSubtask(id="2.V", checks=["pass"]),
        )
        warnings = validation_mod._check_missing_subtask_refs(group)
        messages = {w.message for w in warnings}
        expected = "Subtask 2.1 has empty requirement_refs — scoped rendering will fall back to full spec dump"
        assert expected in messages

    def test_empty_test_spec_refs_only_message(self) -> None:
        """Subtask with empty test_spec_refs only produces correct message."""
        assert hasattr(validation_mod, "_check_missing_subtask_refs"), "_check_missing_subtask_refs not defined"
        group = TaskGroup(
            id=2,
            kind=TaskGroupKind.STANDARD,
            title="Standard group",
            subtasks=[
                Subtask(
                    id="2.1",
                    title="Sub 1",
                    requirement_refs=[],
                    test_spec_refs=[_TS_ID],
                ),
                Subtask(
                    id="2.2",
                    title="Sub 2",
                    requirement_refs=[_REQ_ID],
                    test_spec_refs=[],
                ),
                Subtask(
                    id="2.3",
                    title="Sub 3",
                    requirement_refs=[],
                    test_spec_refs=[],
                ),
            ],
            verification=VerificationSubtask(id="2.V", checks=["pass"]),
        )
        warnings = validation_mod._check_missing_subtask_refs(group)
        messages = {w.message for w in warnings}
        expected = "Subtask 2.2 has empty test_spec_refs — scoped rendering will fall back to full spec dump"
        assert expected in messages

    def test_both_empty_message(self) -> None:
        """Subtask with both empty produces message with joined field names."""
        assert hasattr(validation_mod, "_check_missing_subtask_refs"), "_check_missing_subtask_refs not defined"
        group = TaskGroup(
            id=2,
            kind=TaskGroupKind.STANDARD,
            title="Standard group",
            subtasks=[
                Subtask(
                    id="2.1",
                    title="Sub 1",
                    requirement_refs=[],
                    test_spec_refs=[_TS_ID],
                ),
                Subtask(
                    id="2.2",
                    title="Sub 2",
                    requirement_refs=[_REQ_ID],
                    test_spec_refs=[],
                ),
                Subtask(
                    id="2.3",
                    title="Sub 3",
                    requirement_refs=[],
                    test_spec_refs=[],
                ),
            ],
            verification=VerificationSubtask(id="2.V", checks=["pass"]),
        )
        warnings = validation_mod._check_missing_subtask_refs(group)
        messages = {w.message for w in warnings}
        expected = (
            "Subtask 2.3 has empty requirement_refs and test_spec_refs "
            "— scoped rendering will fall back to full spec dump"
        )
        assert expected in messages

    def test_no_warning_for_subtask_with_populated_refs(self) -> None:
        """Subtask with both refs populated does not trigger a warning."""
        assert hasattr(validation_mod, "_check_missing_subtask_refs"), "_check_missing_subtask_refs not defined"
        group = TaskGroup(
            id=2,
            kind=TaskGroupKind.STANDARD,
            title="Standard group",
            subtasks=[
                Subtask(
                    id="2.1",
                    title="Fully populated",
                    requirement_refs=[_REQ_ID],
                    test_spec_refs=[_TS_ID],
                ),
            ],
            verification=VerificationSubtask(id="2.V", checks=["pass"]),
        )
        warnings = validation_mod._check_missing_subtask_refs(group)
        warning_ids = [w.entity_id for w in warnings]
        assert "2.1" not in warning_ids

    def test_placeholder_refs_do_not_trigger_warning(self) -> None:
        """Non-empty refs with placeholder strings like 'TBD' do not trigger
        a warning (02-REQ-4.E2).
        """
        assert hasattr(validation_mod, "_check_missing_subtask_refs"), "_check_missing_subtask_refs not defined"
        group = TaskGroup(
            id=2,
            kind=TaskGroupKind.STANDARD,
            title="Standard group",
            subtasks=[
                Subtask(
                    id="2.1",
                    title="Placeholder refs",
                    requirement_refs=["TBD"],
                    test_spec_refs=["TBD"],
                ),
            ],
        )
        warnings = validation_mod._check_missing_subtask_refs(group)
        assert len(warnings) == 0


# ---------------------------------------------------------------------------
# TS-02-13: WIRING_VERIFICATION groups are entirely skipped (02-REQ-4.3)
# ---------------------------------------------------------------------------


class TestCheckMissingSubtaskRefsSkipsWiringVerification:
    """TS-02-13: _check_missing_subtask_refs skips the entire
    WIRING_VERIFICATION TaskGroup and emits no warnings for its subtasks
    even if they have empty refs.
    """

    def test_no_warning_for_wv_group(self) -> None:
        """No warning emitted for subtasks in a WIRING_VERIFICATION group."""
        assert hasattr(validation_mod, "_check_missing_subtask_refs"), "_check_missing_subtask_refs not defined"
        wv_group = TaskGroup(
            id=1,
            kind=TaskGroupKind.WIRING_VERIFICATION,
            title="Wiring verification",
            subtasks=[
                Subtask(
                    id="W.1",
                    title="Wiring check",
                    requirement_refs=[],
                    test_spec_refs=[],
                ),
            ],
        )
        warnings = validation_mod._check_missing_subtask_refs(wv_group)
        warning_ids = [w.entity_id for w in warnings]
        assert "W.1" not in warning_ids
        assert len(warnings) == 0

    def test_warning_emitted_for_standard_group(self) -> None:
        """Warning IS emitted for a STANDARD group with empty refs."""
        assert hasattr(validation_mod, "_check_missing_subtask_refs"), "_check_missing_subtask_refs not defined"
        std_group = TaskGroup(
            id=2,
            kind=TaskGroupKind.STANDARD,
            title="Standard group",
            subtasks=[
                Subtask(
                    id="S.1",
                    title="Standard check",
                    requirement_refs=[],
                    test_spec_refs=[],
                ),
            ],
        )
        warnings = validation_mod._check_missing_subtask_refs(std_group)
        warning_ids = [w.entity_id for w in warnings]
        assert "S.1" in warning_ids

    def test_mixed_groups_only_warns_for_non_wv(self) -> None:
        """Iterating over WV and STANDARD groups, only STANDARD emits warnings.

        This mirrors how validate() calls the check per group, collecting
        results across all groups.
        """
        assert hasattr(validation_mod, "_check_missing_subtask_refs"), "_check_missing_subtask_refs not defined"
        wv_group = TaskGroup(
            id=1,
            kind=TaskGroupKind.WIRING_VERIFICATION,
            title="Wiring verification",
            subtasks=[
                Subtask(
                    id="W.1",
                    title="Wiring check",
                    requirement_refs=[],
                    test_spec_refs=[],
                ),
            ],
        )
        std_group = TaskGroup(
            id=2,
            kind=TaskGroupKind.STANDARD,
            title="Standard group",
            subtasks=[
                Subtask(
                    id="S.1",
                    title="Standard check",
                    requirement_refs=[],
                    test_spec_refs=[],
                ),
            ],
        )
        all_warnings = []
        for group in [wv_group, std_group]:
            all_warnings.extend(validation_mod._check_missing_subtask_refs(group))
        warning_ids = [w.entity_id for w in all_warnings]
        assert "W.1" not in warning_ids
        assert "S.1" in warning_ids
        assert len(all_warnings) == 1

    def test_no_warnings_when_all_groups_are_wv(self) -> None:
        """No warnings when all TaskGroups are WIRING_VERIFICATION (02-REQ-4.E3)."""
        assert hasattr(validation_mod, "_check_missing_subtask_refs"), "_check_missing_subtask_refs not defined"
        wv_group = TaskGroup(
            id=1,
            kind=TaskGroupKind.WIRING_VERIFICATION,
            title="WV 1",
            subtasks=[
                Subtask(
                    id="1.1",
                    title="Check",
                    requirement_refs=[],
                    test_spec_refs=[],
                ),
            ],
        )
        warnings = validation_mod._check_missing_subtask_refs(wv_group)
        assert len(warnings) == 0

    def test_wv_group_with_multiple_empty_ref_subtasks_skipped(self) -> None:
        """All subtasks in a WV group are skipped, even if multiple have empty
        refs (02-REQ-4.E4).
        """
        assert hasattr(validation_mod, "_check_missing_subtask_refs"), "_check_missing_subtask_refs not defined"
        wv_group = TaskGroup(
            id=1,
            kind=TaskGroupKind.WIRING_VERIFICATION,
            title="Wiring verification",
            subtasks=[
                Subtask(
                    id="1.1",
                    title="Check 1",
                    requirement_refs=[],
                    test_spec_refs=[],
                ),
                Subtask(
                    id="1.2",
                    title="Check 2",
                    requirement_refs=[],
                    test_spec_refs=[],
                ),
            ],
        )
        warnings = validation_mod._check_missing_subtask_refs(wv_group)
        assert len(warnings) == 0


# ---------------------------------------------------------------------------
# TS-02-14: field_names joined with ' and ' (02-REQ-4.4)
# ---------------------------------------------------------------------------


class TestFieldNamesJoinedWithAnd:
    """TS-02-14: The field_names portion of the warning message uses
    ' and '.join(missing) serialisation, producing 'requirement_refs and
    test_spec_refs' when both are empty.
    """

    def test_both_empty_joined_with_and(self) -> None:
        """Message contains 'requirement_refs and test_spec_refs' when both empty."""
        assert hasattr(validation_mod, "_check_missing_subtask_refs"), "_check_missing_subtask_refs not defined"
        group = TaskGroup(
            id=2,
            kind=TaskGroupKind.STANDARD,
            title="Standard group",
            subtasks=[
                Subtask(
                    id="2.1",
                    title="Sub 1",
                    requirement_refs=[],
                    test_spec_refs=[],
                ),
            ],
            verification=VerificationSubtask(id="2.V", checks=["pass"]),
        )
        warnings = validation_mod._check_missing_subtask_refs(group)
        assert len(warnings) == 1
        assert "requirement_refs and test_spec_refs" in warnings[0].message

    def test_no_comma_in_joined_field_names(self) -> None:
        """Joined field names use ' and ', not a comma separator."""
        assert hasattr(validation_mod, "_check_missing_subtask_refs"), "_check_missing_subtask_refs not defined"
        group = TaskGroup(
            id=2,
            kind=TaskGroupKind.STANDARD,
            title="Standard group",
            subtasks=[
                Subtask(
                    id="2.1",
                    title="Sub 1",
                    requirement_refs=[],
                    test_spec_refs=[],
                ),
            ],
            verification=VerificationSubtask(id="2.V", checks=["pass"]),
        )
        warnings = validation_mod._check_missing_subtask_refs(group)
        assert len(warnings) == 1
        assert "," not in warnings[0].message

    def test_single_empty_field_has_no_and_joiner(self) -> None:
        """When only one ref type is empty, ' and ' joiner is not present."""
        assert hasattr(validation_mod, "_check_missing_subtask_refs"), "_check_missing_subtask_refs not defined"
        group = TaskGroup(
            id=2,
            kind=TaskGroupKind.STANDARD,
            title="Standard group",
            subtasks=[
                Subtask(
                    id="2.1",
                    title="Sub 1",
                    requirement_refs=[],
                    test_spec_refs=[_TS_ID],
                ),
            ],
            verification=VerificationSubtask(id="2.V", checks=["pass"]),
        )
        warnings = validation_mod._check_missing_subtask_refs(group)
        assert len(warnings) == 1
        assert "requirement_refs" in warnings[0].message
        assert " and " not in warnings[0].message
