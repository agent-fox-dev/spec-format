"""Tests for traceability-based ref inference in scoped rendering.

TS-02-1: When all subtasks in the target group have empty refs,
         render_individual_scoped invokes traceability inference and
         returns a scoped result filtered to the inferred refs.

TS-02-2: _infer_refs_from_traceability returns only requirement_id and
         test_spec_id values from entries whose task_id starts with the
         target group prefix.

TS-02-3: When traceability inference yields at least one ref, scoped
         rendering activates immediately without attempting text inference,
         and Python emits an INFO log.

These tests are in RED PHASE — they will fail because the inference
helpers do not exist yet and render_individual_scoped currently falls
back to the unscoped dump when subtask refs are empty.
"""

from __future__ import annotations

import logging

import pytest

import afspec.render as render_mod
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
from afspec.render import (
    render_individual,
    render_individual_scoped,
    render_tasks_scoped,
)

# ---------------------------------------------------------------------------
# Helpers — build structurally valid specs with configurable traceability
# ---------------------------------------------------------------------------

_REQ_ID_A = "02-REQ-1"
_REQ_ID_B = "02-REQ-2"
_CRITERION_A1 = "02-REQ-1.1"
_CRITERION_B1 = "02-REQ-2.1"
_TS_ID_A = "TS-02-1"
_TS_ID_B = "TS-02-2"


def _make_requirement(req_id: str, criterion_id: str) -> Requirement:
    """Build a single Requirement with one acceptance criterion."""
    return Requirement(
        id=req_id,
        title=f"Requirement {req_id}",
        user_story=UserStory(role="developer", goal="test inference", benefit="coverage"),
        acceptance_criteria=[
            Criterion(
                id=criterion_id,
                ears_pattern=EARSPattern.EVENT_DRIVEN,
                when="called",
                system="the system",
                action=f"performs action for {criterion_id}",
            ),
        ],
    )


def _build_spec_with_traceability(
    target_group: int,
    traceability: list[TraceabilityEntry],
    *,
    subtask_requirement_refs: list[str] | None = None,
    subtask_test_spec_refs: list[str] | None = None,
    extra_traceability_entries: list[TraceabilityEntry] | None = None,
) -> Spec:
    """Build a structurally valid Spec with configurable traceability.

    Creates a spec with two requirements (``02-REQ-1``, ``02-REQ-2``)
    and two test cases (``TS-02-1``, ``TS-02-2``).  The target group's
    subtasks get explicit refs only when explicitly provided via the
    optional keyword arguments; otherwise they are empty (the default).

    A ``wiring_verification`` group is always appended as the final
    group to satisfy the structural validation rule.

    Parameters
    ----------
    target_group:
        The id for the main TaskGroup under test.
    traceability:
        TraceabilityEntry records to wire into ``spec.tasks.traceability``.
    subtask_requirement_refs:
        If provided, set these as explicit requirement_refs on subtasks.
    subtask_test_spec_refs:
        If provided, set these as explicit test_spec_refs on subtasks.
    extra_traceability_entries:
        Additional traceability entries (e.g. for other groups) to
        include alongside the provided ones.
    """
    all_traceability = list(traceability)
    if extra_traceability_entries:
        all_traceability.extend(extra_traceability_entries)

    subtask = Subtask(
        id=f"{target_group}.1",
        title="Implement feature",
        details=["Some implementation details"],
        requirement_refs=subtask_requirement_refs or [],
        test_spec_refs=subtask_test_spec_refs or [],
    )

    # Main group under test
    main_group = TaskGroup(
        id=target_group,
        kind=TaskGroupKind.STANDARD,
        title="Main group",
        subtasks=[subtask],
        verification=VerificationSubtask(id=f"{target_group}.V", checks=["pass"]),
    )

    # A tests group is required as the first group
    tests_group = TaskGroup(
        id=1,
        kind=TaskGroupKind.TESTS,
        title="Tests group",
        subtasks=[
            Subtask(
                id="1.1",
                title="Write tests",
                requirement_refs=[_REQ_ID_A],
                test_spec_refs=[_TS_ID_A],
            ),
        ],
        verification=VerificationSubtask(id="1.V", checks=["pass"]),
    )

    groups: list[TaskGroup] = []
    if target_group == 1:
        # When testing group 1, make it a tests group directly
        main_group = TaskGroup(
            id=1,
            kind=TaskGroupKind.TESTS,
            title="Tests group",
            subtasks=[subtask],
            verification=VerificationSubtask(id="1.V", checks=["pass"]),
        )
        groups.append(main_group)
    else:
        groups.append(tests_group)
        groups.append(main_group)

    # Append wiring_verification as the final group
    wv_id = max(g.id for g in groups) + 1
    smoke_tsid = "TS-02-SMOKE-1"
    groups.append(
        TaskGroup(
            id=wv_id,
            kind=TaskGroupKind.WIRING_VERIFICATION,
            title="Wiring verification",
            subtasks=[
                Subtask(
                    id=f"{wv_id}.1",
                    title="Trace execution paths and stub/dead-code audit",
                    test_spec_refs=[smoke_tsid],
                    requirement_refs=[_REQ_ID_A],
                ),
            ],
            verification=VerificationSubtask(id=f"{wv_id}.V", checks=["done"]),
        ),
    )

    return Spec(
        prd=PRDDocument(
            frontmatter=PRDFrontmatter(
                spec_id="02",
                spec_name="scoped_rendering_ref_inference",
                title="Scoped Rendering Ref Inference",
                created_at="2024-01-01",
                updated_at="2024-01-01",
                owner="test",
                source="internal",
            ),
            body="Test spec for traceability inference.",
        ),
        requirements=Requirements(
            spec_id="02",
            spec_name="scoped_rendering_ref_inference",
            introduction="Test spec for traceability inference.",
            requirements=[
                _make_requirement(_REQ_ID_A, _CRITERION_A1),
                _make_requirement(_REQ_ID_B, _CRITERION_B1),
            ],
            execution_paths=[
                ExecutionPath(
                    id="02-PATH-1",
                    title="Main path",
                    steps=[
                        PathStep(actor="Caller", action="Invoke"),
                        PathStep(actor="System", action="Respond"),
                    ],
                ),
            ],
        ),
        test_spec=TestSpec(
            spec_id="02",
            spec_name="scoped_rendering_ref_inference",
            test_cases=[
                TestCase(
                    id=_TS_ID_A,
                    requirement_id=_CRITERION_A1,
                    kind="integration",
                    description="Test case A",
                ),
                TestCase(
                    id=_TS_ID_B,
                    requirement_id=_CRITERION_B1,
                    kind="unit",
                    description="Test case B",
                ),
            ],
            smoke_tests=[
                SmokeTest(
                    id="TS-02-SMOKE-1",
                    execution_path_id="02-PATH-1",
                    description="Wiring smoke test",
                ),
            ],
        ),
        tasks=Tasks(
            spec_id="02",
            spec_name="scoped_rendering_ref_inference",
            task_groups=groups,
            traceability=all_traceability,
        ),
    )


# ---------------------------------------------------------------------------
# TS-02-1: Traceability inference activates scoped rendering
# (02-REQ-1.1)
# ---------------------------------------------------------------------------


class TestTraceabilityInferenceActivatesScopedRendering:
    """When all subtasks have empty refs and traceability entries match the
    target group, render_individual_scoped should infer refs from traceability
    and return a scoped result filtered to those inferred refs.
    """

    def test_returns_scoped_result_with_inferred_requirement(self) -> None:
        """Result requirements section contains the inferred requirement."""
        spec = _build_spec_with_traceability(
            target_group=3,
            traceability=[
                TraceabilityEntry(
                    task_id="3.1",
                    requirement_id=_REQ_ID_A,
                    test_spec_id=_TS_ID_A,
                ),
            ],
        )
        result = render_individual_scoped(spec, target_group=3)
        assert _REQ_ID_A in result["requirements"]

    def test_returns_scoped_result_with_inferred_test_spec(self) -> None:
        """Result test_spec section contains the inferred test spec."""
        spec = _build_spec_with_traceability(
            target_group=3,
            traceability=[
                TraceabilityEntry(
                    task_id="3.1",
                    requirement_id=_REQ_ID_A,
                    test_spec_id=_TS_ID_A,
                ),
            ],
        )
        result = render_individual_scoped(spec, target_group=3)
        assert _TS_ID_A in result["test_spec"]

    def test_excludes_unreferenced_requirement(self) -> None:
        """Requirements not inferred via traceability are excluded."""
        spec = _build_spec_with_traceability(
            target_group=3,
            traceability=[
                TraceabilityEntry(
                    task_id="3.1",
                    requirement_id=_REQ_ID_A,
                    test_spec_id=_TS_ID_A,
                ),
            ],
        )
        result = render_individual_scoped(spec, target_group=3)

        # The unscoped render includes REQ-2; scoped should not
        unscoped = render_individual(spec)
        assert _REQ_ID_B in unscoped["requirements"], "Sanity: REQ-B is in unscoped"

        # Scoped result should NOT include the unreferenced requirement
        assert _REQ_ID_B not in result["requirements"]

    def test_excludes_unreferenced_test_spec(self) -> None:
        """Test specs not inferred via traceability are excluded."""
        spec = _build_spec_with_traceability(
            target_group=3,
            traceability=[
                TraceabilityEntry(
                    task_id="3.1",
                    requirement_id=_REQ_ID_A,
                    test_spec_id=_TS_ID_A,
                ),
            ],
        )
        result = render_individual_scoped(spec, target_group=3)

        # The unscoped render includes TS-02-2; scoped should not
        unscoped = render_individual(spec)
        assert _TS_ID_B in unscoped["test_spec"], "Sanity: TS-B is in unscoped"

        # Scoped result should NOT include the unreferenced test spec
        assert _TS_ID_B not in result["test_spec"]

    def test_result_is_not_full_unscoped_dump(self) -> None:
        """The result must differ from the full unscoped render_individual."""
        spec = _build_spec_with_traceability(
            target_group=3,
            traceability=[
                TraceabilityEntry(
                    task_id="3.1",
                    requirement_id=_REQ_ID_A,
                    test_spec_id=_TS_ID_A,
                ),
            ],
        )
        result = render_individual_scoped(spec, target_group=3)
        unscoped = render_individual(spec)

        # If inference worked, requirements and test_spec should be
        # scoped (smaller), not identical to unscoped
        assert result["requirements"] != unscoped["requirements"]
        assert result["test_spec"] != unscoped["test_spec"]


# ---------------------------------------------------------------------------
# TS-02-2: _infer_refs_from_traceability collects only matching entries
# (02-REQ-1.2)
# ---------------------------------------------------------------------------


class TestTraceabilityInferenceFiltersMatchingEntries:
    """_infer_refs_from_traceability / inferRefsFromTraceability returns only
    requirement_id and test_spec_id values from entries whose task_id starts
    with the target group prefix.
    """

    def test_collects_matching_requirement_id(self) -> None:
        """Requirement IDs from matching entries are collected."""
        spec = _build_spec_with_traceability(
            target_group=3,
            traceability=[
                TraceabilityEntry(
                    task_id="3.1",
                    requirement_id=_REQ_ID_A,
                    test_spec_id=_TS_ID_A,
                ),
            ],
            extra_traceability_entries=[
                TraceabilityEntry(
                    task_id="5.1",
                    requirement_id=_REQ_ID_B,
                    test_spec_id=_TS_ID_B,
                ),
            ],
        )
        result = render_individual_scoped(spec, target_group=3)

        # Only REQ-A (from group 3 traceability) should appear, not REQ-B
        assert _REQ_ID_A in result["requirements"]
        assert _REQ_ID_B not in result["requirements"]

    def test_collects_matching_test_spec_id(self) -> None:
        """Test spec IDs from matching entries are collected."""
        spec = _build_spec_with_traceability(
            target_group=3,
            traceability=[
                TraceabilityEntry(
                    task_id="3.1",
                    requirement_id=_REQ_ID_A,
                    test_spec_id=_TS_ID_A,
                ),
            ],
            extra_traceability_entries=[
                TraceabilityEntry(
                    task_id="5.1",
                    requirement_id=_REQ_ID_B,
                    test_spec_id=_TS_ID_B,
                ),
            ],
        )
        result = render_individual_scoped(spec, target_group=3)

        # Only TS-A (from group 3 traceability) should appear, not TS-B
        assert _TS_ID_A in result["test_spec"]
        assert _TS_ID_B not in result["test_spec"]

    def test_excludes_non_matching_group_entries(self) -> None:
        """Entries from other groups are fully excluded from inference."""
        spec = _build_spec_with_traceability(
            target_group=3,
            traceability=[
                TraceabilityEntry(
                    task_id="3.1",
                    requirement_id=_REQ_ID_A,
                    test_spec_id=_TS_ID_A,
                ),
            ],
            extra_traceability_entries=[
                TraceabilityEntry(
                    task_id="5.1",
                    requirement_id=_REQ_ID_B,
                    test_spec_id=_TS_ID_B,
                ),
            ],
        )
        result = render_individual_scoped(spec, target_group=3)
        unscoped = render_individual(spec)

        # Scoped should be smaller than unscoped — REQ-B/TS-B excluded
        assert _REQ_ID_B in unscoped["requirements"], "Sanity: REQ-B in unscoped"
        assert _REQ_ID_B not in result["requirements"]

    def test_collects_from_multiple_matching_entries(self) -> None:
        """Multiple matching traceability entries contribute their IDs."""
        spec = _build_spec_with_traceability(
            target_group=3,
            traceability=[
                TraceabilityEntry(
                    task_id="3.1",
                    requirement_id=_REQ_ID_A,
                    test_spec_id=_TS_ID_A,
                ),
                TraceabilityEntry(
                    task_id="3.2",
                    requirement_id=_REQ_ID_B,
                    test_spec_id=_TS_ID_B,
                ),
            ],
        )
        result = render_individual_scoped(spec, target_group=3)

        # Both requirements should be included
        assert _REQ_ID_A in result["requirements"]
        assert _REQ_ID_B in result["requirements"]
        # Both test specs should be included
        assert _TS_ID_A in result["test_spec"]
        assert _TS_ID_B in result["test_spec"]


# ---------------------------------------------------------------------------
# TS-02-3: Traceability inference short-circuits the chain
# (02-REQ-1.3)
# ---------------------------------------------------------------------------


class TestTraceabilityInferenceShortCircuits:
    """When traceability inference yields at least one ref, scoped rendering
    activates immediately without attempting text-based inference.
    Python additionally emits an INFO log.
    """

    def test_scoped_result_returned_when_traceability_has_refs(self) -> None:
        """Scoped rendering activates when traceability yields refs."""
        spec = _build_spec_with_traceability(
            target_group=3,
            traceability=[
                TraceabilityEntry(
                    task_id="3.1",
                    requirement_id=_REQ_ID_A,
                    test_spec_id=_TS_ID_A,
                ),
            ],
        )
        result = render_individual_scoped(spec, target_group=3)
        unscoped = render_individual(spec)

        # The result should be scoped (not the full dump)
        assert result["requirements"] != unscoped["requirements"]

    def test_info_log_emitted_on_traceability_inference(
        self,
        caplog: pytest.LogCaptureFixture,
    ) -> None:
        """An INFO log is emitted when traceability inference activates."""
        spec = _build_spec_with_traceability(
            target_group=3,
            traceability=[
                TraceabilityEntry(
                    task_id="3.1",
                    requirement_id=_REQ_ID_A,
                    test_spec_id=_TS_ID_A,
                ),
            ],
        )
        with caplog.at_level(logging.INFO, logger="afspec.render"):
            result = render_individual_scoped(spec, target_group=3)

        # Check the result is scoped (not full dump)
        assert _REQ_ID_A in result.get("requirements", "")

        # Verify that an INFO log mentioning inference was emitted.
        infer_logs = [
            r
            for r in caplog.records
            if r.levelno == logging.INFO
            and r.name == "afspec.render"
            and "infer" in r.message.lower()
        ]
        assert len(infer_logs) >= 1, (
            "Expected an INFO log from afspec.render mentioning inference"
        )

    def test_empty_traceability_proceeds_to_next_strategy(self) -> None:
        """When traceability is empty, the chain proceeds (no short-circuit).

        With no traceability and no text-based ID patterns in subtask text,
        the function should fall back to the full unscoped dump (for
        requirements and test_spec), but tasks should still be scoped.
        """
        spec = _build_spec_with_traceability(
            target_group=3,
            traceability=[],  # No traceability entries at all
        )
        result = render_individual_scoped(spec, target_group=3)
        unscoped = render_individual(spec)

        # With no inference possible, requirements and test_spec should
        # match the unscoped output (full dump fallback).
        assert result["requirements"] == unscoped["requirements"]
        assert result["test_spec"] == unscoped["test_spec"]

    def test_no_traceability_for_target_group_proceeds(self) -> None:
        """Traceability entries exist but none match the target group prefix."""
        spec = _build_spec_with_traceability(
            target_group=3,
            traceability=[],  # No entries for group 3
            extra_traceability_entries=[
                TraceabilityEntry(
                    task_id="5.1",
                    requirement_id=_REQ_ID_B,
                    test_spec_id=_TS_ID_B,
                ),
            ],
        )
        result = render_individual_scoped(spec, target_group=3)
        unscoped = render_individual(spec)

        # No matching traceability → should fall back to unscoped
        # (requirements and test_spec match unscoped)
        assert result["requirements"] == unscoped["requirements"]
        assert result["test_spec"] == unscoped["test_spec"]


# ---------------------------------------------------------------------------
# Edge cases: 02-REQ-1.E1, 02-REQ-1.E2, 02-REQ-1.E3
# ---------------------------------------------------------------------------


class TestTraceabilityInferenceEdgeCases:
    """Edge cases for traceability-based inference."""

    def test_empty_requirement_id_is_skipped(self) -> None:
        """TraceabilityEntry with empty requirement_id is skipped (02-REQ-1.E2)."""
        spec = _build_spec_with_traceability(
            target_group=3,
            traceability=[
                TraceabilityEntry(
                    task_id="3.1",
                    requirement_id="",  # empty — should be skipped
                    test_spec_id=_TS_ID_A,
                ),
            ],
        )
        result = render_individual_scoped(spec, target_group=3)

        # test_spec should be scoped (TS-A found via traceability)
        assert _TS_ID_A in result["test_spec"]

    def test_empty_test_spec_id_is_skipped(self) -> None:
        """TraceabilityEntry with empty test_spec_id is skipped (02-REQ-1.E2)."""
        spec = _build_spec_with_traceability(
            target_group=3,
            traceability=[
                TraceabilityEntry(
                    task_id="3.1",
                    requirement_id=_REQ_ID_A,
                    test_spec_id="",  # empty — should be skipped
                ),
            ],
        )
        result = render_individual_scoped(spec, target_group=3)

        # requirements should be scoped (REQ-A found via traceability)
        assert _REQ_ID_A in result["requirements"]

    def test_explicit_refs_skip_inference_entirely(self) -> None:
        """When subtasks have explicit refs, inference is skipped (02-REQ-1.E3).

        This tests that the existing scoped rendering logic is preserved
        when subtasks already carry refs.
        """
        spec = _build_spec_with_traceability(
            target_group=3,
            traceability=[
                # Traceability points to REQ-B/TS-B, but explicit refs
                # say REQ-A/TS-A — explicit refs should win.
                TraceabilityEntry(
                    task_id="3.1",
                    requirement_id=_REQ_ID_B,
                    test_spec_id=_TS_ID_B,
                ),
            ],
            subtask_requirement_refs=[_REQ_ID_A],
            subtask_test_spec_refs=[_TS_ID_A],
        )
        result = render_individual_scoped(spec, target_group=3)

        # Explicit refs (REQ-A, TS-A) should be used, not traceability
        assert _REQ_ID_A in result["requirements"]
        assert _TS_ID_A in result["test_spec"]


# ---------------------------------------------------------------------------
# Helper for text-based inference tests
# ---------------------------------------------------------------------------


def _build_spec_for_text_inference(
    target_group: int,
    subtask_title: str = "Implement feature",
    subtask_details: list[str] | None = None,
    *,
    traceability: list[TraceabilityEntry] | None = None,
) -> Spec:
    """Build a Spec for text-based inference tests.

    Like ``_build_spec_with_traceability`` but with configurable subtask
    ``title`` and ``details``, enabling tests where ID patterns are
    embedded in the subtask text for text-based inference to discover.

    Subtask ``requirement_refs`` and ``test_spec_refs`` are always empty.
    Traceability defaults to an empty list (no entries).
    """
    subtask = Subtask(
        id=f"{target_group}.1",
        title=subtask_title,
        details=subtask_details if subtask_details is not None else [],
        requirement_refs=[],
        test_spec_refs=[],
    )

    main_group = TaskGroup(
        id=target_group,
        kind=TaskGroupKind.STANDARD,
        title="Main group",
        subtasks=[subtask],
        verification=VerificationSubtask(id=f"{target_group}.V", checks=["pass"]),
    )

    tests_group = TaskGroup(
        id=1,
        kind=TaskGroupKind.TESTS,
        title="Tests group",
        subtasks=[
            Subtask(
                id="1.1",
                title="Write tests",
                requirement_refs=[_REQ_ID_A],
                test_spec_refs=[_TS_ID_A],
            ),
        ],
        verification=VerificationSubtask(id="1.V", checks=["pass"]),
    )

    groups: list[TaskGroup] = []
    if target_group == 1:
        main_group = TaskGroup(
            id=1,
            kind=TaskGroupKind.TESTS,
            title="Tests group",
            subtasks=[subtask],
            verification=VerificationSubtask(id="1.V", checks=["pass"]),
        )
        groups.append(main_group)
    else:
        groups.append(tests_group)
        groups.append(main_group)

    wv_id = max(g.id for g in groups) + 1
    groups.append(
        TaskGroup(
            id=wv_id,
            kind=TaskGroupKind.WIRING_VERIFICATION,
            title="Wiring verification",
            subtasks=[
                Subtask(
                    id=f"{wv_id}.1",
                    title="Trace execution paths and stub/dead-code audit",
                    test_spec_refs=["TS-02-SMOKE-1"],
                    requirement_refs=[_REQ_ID_A],
                ),
            ],
            verification=VerificationSubtask(id=f"{wv_id}.V", checks=["done"]),
        ),
    )

    return Spec(
        prd=PRDDocument(
            frontmatter=PRDFrontmatter(
                spec_id="02",
                spec_name="scoped_rendering_ref_inference",
                title="Scoped Rendering Ref Inference",
                created_at="2024-01-01",
                updated_at="2024-01-01",
                owner="test",
                source="internal",
            ),
            body="Test spec for text-based inference.",
        ),
        requirements=Requirements(
            spec_id="02",
            spec_name="scoped_rendering_ref_inference",
            introduction="Test spec for text-based inference.",
            requirements=[
                _make_requirement(_REQ_ID_A, _CRITERION_A1),
                _make_requirement(_REQ_ID_B, _CRITERION_B1),
            ],
            execution_paths=[
                ExecutionPath(
                    id="02-PATH-1",
                    title="Main path",
                    steps=[
                        PathStep(actor="Caller", action="Invoke"),
                        PathStep(actor="System", action="Respond"),
                    ],
                ),
            ],
        ),
        test_spec=TestSpec(
            spec_id="02",
            spec_name="scoped_rendering_ref_inference",
            test_cases=[
                TestCase(
                    id=_TS_ID_A,
                    requirement_id=_CRITERION_A1,
                    kind="integration",
                    description="Test case A",
                ),
                TestCase(
                    id=_TS_ID_B,
                    requirement_id=_CRITERION_B1,
                    kind="unit",
                    description="Test case B",
                ),
            ],
            smoke_tests=[
                SmokeTest(
                    id="TS-02-SMOKE-1",
                    execution_path_id="02-PATH-1",
                    description="Wiring smoke test",
                ),
            ],
        ),
        tasks=Tasks(
            spec_id="02",
            spec_name="scoped_rendering_ref_inference",
            task_groups=groups,
            traceability=traceability or [],
        ),
    )


# ---------------------------------------------------------------------------
# TS-02-4: Text-based inference activates scoped rendering
# (02-REQ-2.1)
# ---------------------------------------------------------------------------


class TestTextInferenceActivatesScopedRendering:
    """When traceability inference returns empty, render_individual_scoped
    invokes text-based inference and returns a scoped result from
    validated regex matches found in subtask title and details.
    """

    def test_req_id_in_title_activates_scoped_rendering(self) -> None:
        """Subtask title containing a known req ID triggers text inference."""
        spec = _build_spec_for_text_inference(
            target_group=2,
            subtask_title=f"Implement {_REQ_ID_A} logic",
        )
        result = render_individual_scoped(spec, target_group=2)
        assert _REQ_ID_A in result["requirements"]
        assert _REQ_ID_B not in result["requirements"]

    def test_req_id_in_details_activates_scoped_rendering(self) -> None:
        """Subtask details containing a known req ID triggers text inference."""
        spec = _build_spec_for_text_inference(
            target_group=2,
            subtask_title="Implement feature",
            subtask_details=[f"Must satisfy {_REQ_ID_A}"],
        )
        result = render_individual_scoped(spec, target_group=2)
        assert _REQ_ID_A in result["requirements"]
        assert _REQ_ID_B not in result["requirements"]

    def test_ts_id_in_title_activates_scoped_rendering(self) -> None:
        """Subtask title containing a known test spec ID triggers text inference."""
        spec = _build_spec_for_text_inference(
            target_group=2,
            subtask_title=f"Cover {_TS_ID_A} tests",
        )
        result = render_individual_scoped(spec, target_group=2)
        assert _TS_ID_A in result["test_spec"]
        assert _TS_ID_B not in result["test_spec"]

    def test_ts_id_in_details_activates_scoped_rendering(self) -> None:
        """Subtask details containing a known test spec ID triggers text inference."""
        spec = _build_spec_for_text_inference(
            target_group=2,
            subtask_title="Implement feature",
            subtask_details=[f"See {_TS_ID_A} for coverage"],
        )
        result = render_individual_scoped(spec, target_group=2)
        assert _TS_ID_A in result["test_spec"]
        assert _TS_ID_B not in result["test_spec"]

    def test_scoped_result_differs_from_unscoped(self) -> None:
        """Text-inferred scoped result differs from the full unscoped dump."""
        spec = _build_spec_for_text_inference(
            target_group=2,
            subtask_title=f"Implement {_REQ_ID_A} logic",
        )
        result = render_individual_scoped(spec, target_group=2)
        unscoped = render_individual(spec)
        assert result["requirements"] != unscoped["requirements"]


# ---------------------------------------------------------------------------
# TS-02-5: Regex patterns are module-level compiled constants
# (02-REQ-2.2)
# ---------------------------------------------------------------------------


class TestRegexPatternsAreModuleLevelConstants:
    """_REQ_ID_RE and _TS_ID_RE (Python) are instances of re.Pattern
    defined at module level in afspec.render, not compiled per call.
    """

    def test_req_id_re_is_compiled_pattern(self) -> None:
        """_REQ_ID_RE is a re.Pattern at module level."""
        import re

        assert hasattr(render_mod, "_REQ_ID_RE"), (
            "_REQ_ID_RE not defined at module level in afspec.render"
        )
        assert isinstance(render_mod._REQ_ID_RE, re.Pattern)

    def test_ts_id_re_is_compiled_pattern(self) -> None:
        """_TS_ID_RE is a re.Pattern at module level."""
        import re

        assert hasattr(render_mod, "_TS_ID_RE"), (
            "_TS_ID_RE not defined at module level in afspec.render"
        )
        assert isinstance(render_mod._TS_ID_RE, re.Pattern)

    def test_req_id_re_matches_requirement_id_pattern(self) -> None:
        """_REQ_ID_RE matches requirement ID strings like '02-REQ-1'."""
        assert hasattr(render_mod, "_REQ_ID_RE"), (
            "_REQ_ID_RE not defined at module level"
        )
        match = render_mod._REQ_ID_RE.search("Implement 02-REQ-1 logic")
        assert match is not None

    def test_ts_id_re_matches_test_spec_id_pattern(self) -> None:
        """_TS_ID_RE matches test spec ID strings like 'TS-02-1'."""
        assert hasattr(render_mod, "_TS_ID_RE"), (
            "_TS_ID_RE not defined at module level"
        )
        match = render_mod._TS_ID_RE.search("See TS-02-1 for tests")
        assert match is not None


# ---------------------------------------------------------------------------
# TS-02-6: _infer_refs_from_subtask_text filters to known IDs
# (02-REQ-2.3, 02-REQ-2.E1, 02-REQ-2.E2, 02-REQ-2.E3)
# ---------------------------------------------------------------------------


class TestTextInferenceFiltersToKnownIDs:
    """_infer_refs_from_subtask_text scans title and all details strings,
    collects regex matches, and filters to only IDs present in the spec.
    Invalid matches and unmentioned IDs are excluded.
    """

    def test_valid_req_and_ts_ids_are_returned(self) -> None:
        """Valid IDs from subtask text appear in inferred refs."""
        spec = _build_spec_for_text_inference(
            target_group=2,
            subtask_title=f"Work on {_REQ_ID_A}",
            subtask_details=[
                f"See {_TS_ID_A} for tests",
                "Also 99-REQ-999 is mentioned",
            ],
        )
        assert hasattr(render_mod, "_infer_refs_from_subtask_text"), (
            "_infer_refs_from_subtask_text not defined in afspec.render"
        )
        req_refs, ts_refs = render_mod._infer_refs_from_subtask_text(
            spec, target_group=2
        )
        assert _REQ_ID_A in req_refs
        assert _TS_ID_A in ts_refs

    def test_unknown_req_id_is_discarded(self) -> None:
        """Requirement IDs not present in the spec are filtered out."""
        spec = _build_spec_for_text_inference(
            target_group=2,
            subtask_title="Work on 99-REQ-999",
            subtask_details=["Also mentions 99-REQ-888"],
        )
        assert hasattr(render_mod, "_infer_refs_from_subtask_text"), (
            "_infer_refs_from_subtask_text not defined"
        )
        req_refs, _ts_refs = render_mod._infer_refs_from_subtask_text(
            spec, target_group=2
        )
        assert "99-REQ-999" not in req_refs
        assert "99-REQ-888" not in req_refs

    def test_unknown_ts_id_is_discarded(self) -> None:
        """Test spec IDs not present in the spec are filtered out."""
        spec = _build_spec_for_text_inference(
            target_group=2,
            subtask_title="See TS-99-1 for tests",
        )
        assert hasattr(render_mod, "_infer_refs_from_subtask_text"), (
            "_infer_refs_from_subtask_text not defined"
        )
        _req_refs, ts_refs = render_mod._infer_refs_from_subtask_text(
            spec, target_group=2
        )
        assert "TS-99-1" not in ts_refs

    def test_unmentioned_spec_ids_are_not_inferred(self) -> None:
        """IDs that exist in the spec but aren't in subtask text are excluded."""
        spec = _build_spec_for_text_inference(
            target_group=2,
            subtask_title=f"Work on {_REQ_ID_A}",
            subtask_details=[f"See {_TS_ID_A} for tests"],
        )
        assert hasattr(render_mod, "_infer_refs_from_subtask_text"), (
            "_infer_refs_from_subtask_text not defined"
        )
        req_refs, ts_refs = render_mod._infer_refs_from_subtask_text(
            spec, target_group=2
        )
        # REQ-B and TS-B exist in spec but weren't mentioned in text
        assert _REQ_ID_B not in req_refs
        assert _TS_ID_B not in ts_refs

    def test_all_invalid_matches_fall_through_to_unscoped(self) -> None:
        """When all regex matches are invalid, falls back to unscoped (02-REQ-2.E1)."""
        spec = _build_spec_for_text_inference(
            target_group=2,
            subtask_title="Work on 99-REQ-999",
            subtask_details=["Also TS-99-1 is mentioned"],
        )
        result = render_individual_scoped(spec, target_group=2)
        unscoped = render_individual(spec)
        assert result["requirements"] == unscoped["requirements"]
        assert result["test_spec"] == unscoped["test_spec"]

    def test_no_regex_matches_returns_empty_collections(self) -> None:
        """When text has no ID patterns at all, empty collections returned (02-REQ-2.E2)."""
        spec = _build_spec_for_text_inference(
            target_group=2,
            subtask_title="Plain title with no IDs",
            subtask_details=["Plain detail with no IDs"],
        )
        assert hasattr(render_mod, "_infer_refs_from_subtask_text"), (
            "_infer_refs_from_subtask_text not defined"
        )
        req_refs, ts_refs = render_mod._infer_refs_from_subtask_text(
            spec, target_group=2
        )
        assert len(req_refs) == 0
        assert len(ts_refs) == 0

    def test_empty_details_scans_title_only(self) -> None:
        """When details is empty, only title is scanned (02-REQ-2.E3)."""
        spec = _build_spec_for_text_inference(
            target_group=2,
            subtask_title=f"Implement {_REQ_ID_A} logic",
            subtask_details=[],
        )
        assert hasattr(render_mod, "_infer_refs_from_subtask_text"), (
            "_infer_refs_from_subtask_text not defined"
        )
        req_refs, _ts_refs = render_mod._infer_refs_from_subtask_text(
            spec, target_group=2
        )
        assert _REQ_ID_A in req_refs


# ---------------------------------------------------------------------------
# TS-02-7: Text inference activates with INFO log; traceability-first order
# (02-REQ-2.4, 02-PROP-6)
# ---------------------------------------------------------------------------


class TestTextInferenceInfoLogAndTraceabilityPriority:
    """When text-based inference yields at least one validated ref, scoped
    rendering activates and Python emits an INFO log.  Also validates that
    text inference is skipped when traceability inference succeeds
    (traceability-first order per 02-PROP-6).
    """

    def test_info_log_emitted_on_text_inference(
        self,
        caplog: pytest.LogCaptureFixture,
    ) -> None:
        """An INFO log is emitted when text inference activates scoped rendering."""
        spec = _build_spec_for_text_inference(
            target_group=2,
            subtask_title=f"Handle {_REQ_ID_A}",
        )
        with caplog.at_level(logging.INFO, logger="afspec.render"):
            result = render_individual_scoped(spec, target_group=2)

        # Result should be scoped to REQ-A (not full dump)
        assert _REQ_ID_A in result["requirements"]
        assert _REQ_ID_B not in result["requirements"]

        # Verify INFO log mentioning inference was emitted
        infer_logs = [
            r
            for r in caplog.records
            if r.levelno == logging.INFO
            and r.name == "afspec.render"
            and "infer" in r.message.lower()
        ]
        assert len(infer_logs) >= 1, (
            "Expected an INFO log from afspec.render mentioning inference"
        )

    def test_traceability_takes_priority_over_text_inference(self) -> None:
        """When traceability yields refs, text inference is not applied.

        Traceability points to REQ-A/TS-A, while subtask text mentions
        REQ-B. The result should use traceability-inferred refs only.
        """
        spec = _build_spec_for_text_inference(
            target_group=2,
            subtask_title=f"Handle {_REQ_ID_B}",
            traceability=[
                TraceabilityEntry(
                    task_id="2.1",
                    requirement_id=_REQ_ID_A,
                    test_spec_id=_TS_ID_A,
                ),
            ],
        )
        result = render_individual_scoped(spec, target_group=2)

        # Traceability refs (REQ-A) should be used
        assert _REQ_ID_A in result["requirements"]
        # Text-inferred refs (REQ-B from title) should NOT be applied
        assert _REQ_ID_B not in result["requirements"]

    def test_partial_traceability_short_circuits_text_inference(self) -> None:
        """Even partial traceability refs (only req, no ts) short-circuit
        text inference.  Requirements are scoped to traceability-inferred
        refs; test spec falls back to full rendering.
        """
        spec = _build_spec_for_text_inference(
            target_group=2,
            subtask_title=f"Cover {_TS_ID_B} tests",
            traceability=[
                TraceabilityEntry(
                    task_id="2.1",
                    requirement_id=_REQ_ID_A,
                    test_spec_id="",  # empty — no test spec from traceability
                ),
            ],
        )
        result = render_individual_scoped(spec, target_group=2)

        # Traceability found REQ-A → requirements should be scoped
        assert _REQ_ID_A in result["requirements"]
        assert _REQ_ID_B not in result["requirements"]

        # test_spec from traceability is empty, and text inference should
        # NOT be applied — test_spec should be full (no scoping)
        unscoped = render_individual(spec)
        assert result["test_spec"] == unscoped["test_spec"]


# ---------------------------------------------------------------------------
# TS-02-8: Partial inference — only requirement refs inferred
# (02-REQ-3.1)
# ---------------------------------------------------------------------------


class TestPartialInferenceReqRefsOnly:
    """When inference yields requirement_refs but no test_spec_refs,
    scoped rendering activates for requirements and falls back to full
    rendering for the test spec section (TS-02-8, 02-REQ-3.1).

    The OR condition ``if inferred_req or inferred_ts`` enables scoped
    rendering even when only one ref type is available.
    """

    def test_requirements_scoped_to_inferred_req(self) -> None:
        """Requirements section includes only the inferred requirement."""
        spec = _build_spec_with_traceability(
            target_group=4,
            traceability=[
                TraceabilityEntry(
                    task_id="4.1",
                    requirement_id=_REQ_ID_A,
                    test_spec_id="",  # empty — no TS ref from traceability
                ),
            ],
        )
        result = render_individual_scoped(spec, target_group=4)
        assert _REQ_ID_A in result["requirements"]

    def test_unreferenced_requirement_excluded(self) -> None:
        """Requirement not matched by inferred refs is excluded."""
        spec = _build_spec_with_traceability(
            target_group=4,
            traceability=[
                TraceabilityEntry(
                    task_id="4.1",
                    requirement_id=_REQ_ID_A,
                    test_spec_id="",
                ),
            ],
        )
        result = render_individual_scoped(spec, target_group=4)
        assert _REQ_ID_B not in result["requirements"]

    def test_test_spec_fully_rendered(self) -> None:
        """Test spec is fully rendered when no test spec refs are inferred."""
        spec = _build_spec_with_traceability(
            target_group=4,
            traceability=[
                TraceabilityEntry(
                    task_id="4.1",
                    requirement_id=_REQ_ID_A,
                    test_spec_id="",
                ),
            ],
        )
        result = render_individual_scoped(spec, target_group=4)
        unscoped = render_individual(spec)
        # All test specs should be present (full render, not scoped)
        assert _TS_ID_A in result["test_spec"]
        assert _TS_ID_B in result["test_spec"]
        assert result["test_spec"] == unscoped["test_spec"]

    def test_requirements_differ_from_unscoped(self) -> None:
        """Scoped requirements differ from the full unscoped render."""
        spec = _build_spec_with_traceability(
            target_group=4,
            traceability=[
                TraceabilityEntry(
                    task_id="4.1",
                    requirement_id=_REQ_ID_A,
                    test_spec_id="",
                ),
            ],
        )
        result = render_individual_scoped(spec, target_group=4)
        unscoped = render_individual(spec)
        assert result["requirements"] != unscoped["requirements"]

    def test_tasks_section_scoped_to_target_group(self) -> None:
        """Tasks section is scoped to the target group."""
        spec = _build_spec_with_traceability(
            target_group=4,
            traceability=[
                TraceabilityEntry(
                    task_id="4.1",
                    requirement_id=_REQ_ID_A,
                    test_spec_id="",
                ),
            ],
        )
        result = render_individual_scoped(spec, target_group=4)
        expected_tasks = render_tasks_scoped(spec.tasks, target_group=4)
        assert result["tasks"] == expected_tasks


# ---------------------------------------------------------------------------
# Partial inference — only test spec refs inferred
# (02-REQ-3.1 mirror case)
# ---------------------------------------------------------------------------


class TestPartialInferenceTSRefsOnly:
    """When inference yields test_spec_refs but no requirement_refs,
    scoped rendering activates for test spec and falls back to full
    rendering for requirements (02-REQ-3.1).
    """

    def test_test_spec_scoped_to_inferred_ts(self) -> None:
        """Test spec section includes only the inferred test spec."""
        spec = _build_spec_with_traceability(
            target_group=4,
            traceability=[
                TraceabilityEntry(
                    task_id="4.1",
                    requirement_id="",  # empty — no req ref from traceability
                    test_spec_id=_TS_ID_A,
                ),
            ],
        )
        result = render_individual_scoped(spec, target_group=4)
        assert _TS_ID_A in result["test_spec"]

    def test_unreferenced_test_spec_excluded(self) -> None:
        """Test spec not matched by inferred refs is excluded."""
        spec = _build_spec_with_traceability(
            target_group=4,
            traceability=[
                TraceabilityEntry(
                    task_id="4.1",
                    requirement_id="",
                    test_spec_id=_TS_ID_A,
                ),
            ],
        )
        result = render_individual_scoped(spec, target_group=4)
        assert _TS_ID_B not in result["test_spec"]

    def test_requirements_fully_rendered(self) -> None:
        """Requirements are fully rendered when no requirement refs inferred."""
        spec = _build_spec_with_traceability(
            target_group=4,
            traceability=[
                TraceabilityEntry(
                    task_id="4.1",
                    requirement_id="",
                    test_spec_id=_TS_ID_A,
                ),
            ],
        )
        result = render_individual_scoped(spec, target_group=4)
        unscoped = render_individual(spec)
        # All requirements should be present (full render, not scoped)
        assert _REQ_ID_A in result["requirements"]
        assert _REQ_ID_B in result["requirements"]
        assert result["requirements"] == unscoped["requirements"]

    def test_test_spec_differs_from_unscoped(self) -> None:
        """Scoped test spec differs from the full unscoped render."""
        spec = _build_spec_with_traceability(
            target_group=4,
            traceability=[
                TraceabilityEntry(
                    task_id="4.1",
                    requirement_id="",
                    test_spec_id=_TS_ID_A,
                ),
            ],
        )
        result = render_individual_scoped(spec, target_group=4)
        unscoped = render_individual(spec)
        assert result["test_spec"] != unscoped["test_spec"]


# ---------------------------------------------------------------------------
# TS-02-9 / TS-02-10: Unscoped fallback with scoped tasks
# (02-REQ-3.2, 02-REQ-3.3)
# ---------------------------------------------------------------------------


class TestUnscopedFallbackWithScopedTasks:
    """When both inference strategies return empty collections for both ref
    types, render_individual_scoped falls back to full unscoped rendering
    for requirements and test spec, but still scopes tasks to the target
    group (TS-02-9, TS-02-10, 02-REQ-3.2, 02-REQ-3.3).
    """

    def test_full_requirements_in_fallback(self) -> None:
        """All requirements are present in the fallback output."""
        spec = _build_spec_for_text_inference(
            target_group=2,
            subtask_title="Do some work",
            subtask_details=["Plain detail with no IDs"],
        )
        result = render_individual_scoped(spec, target_group=2)
        assert _REQ_ID_A in result["requirements"]
        assert _REQ_ID_B in result["requirements"]

    def test_full_test_spec_in_fallback(self) -> None:
        """All test specs are present in the fallback output."""
        spec = _build_spec_for_text_inference(
            target_group=2,
            subtask_title="Do some work",
            subtask_details=["Plain detail with no IDs"],
        )
        result = render_individual_scoped(spec, target_group=2)
        assert _TS_ID_A in result["test_spec"]
        assert _TS_ID_B in result["test_spec"]

    def test_requirements_match_unscoped(self) -> None:
        """Requirements match the full unscoped output exactly."""
        spec = _build_spec_for_text_inference(
            target_group=2,
            subtask_title="Do some work",
        )
        result = render_individual_scoped(spec, target_group=2)
        unscoped = render_individual(spec)
        assert result["requirements"] == unscoped["requirements"]

    def test_test_spec_matches_unscoped(self) -> None:
        """Test spec matches the full unscoped output exactly."""
        spec = _build_spec_for_text_inference(
            target_group=2,
            subtask_title="Do some work",
        )
        result = render_individual_scoped(spec, target_group=2)
        unscoped = render_individual(spec)
        assert result["test_spec"] == unscoped["test_spec"]

    def test_tasks_scoped_to_target_group(self) -> None:
        """Tasks section contains target group subtasks in detail, not others."""
        spec = _build_spec_for_text_inference(
            target_group=2,
            subtask_title="Do some work",
            subtask_details=["Plain detail with no IDs"],
        )
        result = render_individual_scoped(spec, target_group=2)
        # Target group subtask should be present in full detail
        assert "2.1" in result["tasks"]
        assert "Do some work" in result["tasks"]
        # Group 1's subtask should NOT appear in full detail
        # (group 1 is summarised as a one-liner, subtask IDs not listed)
        assert "1.1 Write tests" not in result["tasks"]

    def test_tasks_not_fully_unscoped(self) -> None:
        """Tasks section differs from the fully unscoped tasks render."""
        spec = _build_spec_for_text_inference(
            target_group=2,
            subtask_title="Do some work",
        )
        result = render_individual_scoped(spec, target_group=2)
        unscoped = render_individual(spec)
        # Scoped tasks should differ from fully unscoped tasks
        assert result["tasks"] != unscoped["tasks"]

    def test_tasks_match_render_tasks_scoped(self) -> None:
        """Tasks section matches render_tasks_scoped output exactly."""
        spec = _build_spec_for_text_inference(
            target_group=2,
            subtask_title="Do some work",
        )
        result = render_individual_scoped(spec, target_group=2)
        expected_tasks = render_tasks_scoped(spec.tasks, target_group=2)
        assert result["tasks"] == expected_tasks
