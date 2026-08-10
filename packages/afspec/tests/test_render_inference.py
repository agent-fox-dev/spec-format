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
from afspec.render import render_individual, render_individual_scoped

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
