"""Tests for folder-name validation — spec_id/spec_name vs. folder prefix/suffix.

Covers: TS-NS-1, TS-NS-2, TS-NS-3, TS-NS-4, TS-NS-5
Requirements: NS-REQ-1, NS-REQ-2, NS-REQ-3, NS-REQ-4, NS-REQ-5
"""

from __future__ import annotations

from pathlib import Path

import afspec
from afspec import validate_cross_file

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

_MINIMAL_PRD_TEMPLATE = """\
---
spec_id: "{spec_id}"
spec_name: "{spec_name}"
title: "Test"
status: "draft"
created_at: "2026-01-01T00:00:00Z"
updated_at: "2026-01-01T00:00:00Z"
owner: "test"
source: "test"
supersedes: []
tags: []
intent_hash: null
schema_version: 1
---
# Test
"""

_MINIMAL_REQUIREMENTS_TEMPLATE = """\
{{
  "$schema": "https://agent-fox.dev/schemas/requirements.v1.json",
  "spec_id": "{spec_id}",
  "spec_name": "{spec_name}",
  "schema_version": 1,
  "introduction": "Test.",
  "glossary": {{}},
  "requirements": [],
  "correctness_properties": [],
  "execution_paths": [],
  "error_handling": []
}}
"""

_MINIMAL_TEST_SPEC_TEMPLATE = """\
{{
  "$schema": "https://agent-fox.dev/schemas/test_spec.v1.json",
  "spec_id": "{spec_id}",
  "spec_name": "{spec_name}",
  "schema_version": 1,
  "test_cases": [],
  "property_tests": [],
  "edge_case_tests": [],
  "smoke_tests": [],
  "coverage": {{}}
}}
"""

_MINIMAL_TASKS_TEMPLATE = """\
{{
  "$schema": "https://agent-fox.dev/schemas/tasks.v1.json",
  "spec_id": "{spec_id}",
  "spec_name": "{spec_name}",
  "schema_version": 1,
  "test_commands": {{
    "spec_tests": "pytest",
    "all_tests": "pytest",
    "linter": "ruff"
  }},
  "dependencies": [],
  "task_groups": [],
  "traceability": []
}}
"""


def _write_spec_dir(dir_path: Path, spec_id: str, spec_name: str) -> None:
    """Write minimal spec files to dir_path with given spec_id and spec_name.

    All four artifacts (prd.md, requirements.json, test_spec.json, tasks.json)
    use the same spec_id and spec_name so that cross-file-7 does not fire.
    """
    dir_path.mkdir(parents=True, exist_ok=True)
    (dir_path / "prd.md").write_text(_MINIMAL_PRD_TEMPLATE.format(spec_id=spec_id, spec_name=spec_name))
    (dir_path / "requirements.json").write_text(
        _MINIMAL_REQUIREMENTS_TEMPLATE.format(spec_id=spec_id, spec_name=spec_name)
    )
    (dir_path / "test_spec.json").write_text(_MINIMAL_TEST_SPEC_TEMPLATE.format(spec_id=spec_id, spec_name=spec_name))
    (dir_path / "tasks.json").write_text(_MINIMAL_TASKS_TEMPLATE.format(spec_id=spec_id, spec_name=spec_name))


def _folder_name_errors(errors):
    """Return only errors whose rule is 'folder_name'."""
    return [e for e in errors if e.rule == "folder_name"]


# ---------------------------------------------------------------------------
# TS-NS-1: spec_id mismatch with folder prefix
# ---------------------------------------------------------------------------


class TestFolderNameSpecIDMismatch:
    """TS-NS-1: spec_id '99' does not match folder prefix '05' in '05_my_feature'."""

    def test_spec_id_mismatch_produces_folder_name_error(self, tmp_path: Path) -> None:
        # Spec lives in 05_my_feature but prd.md declares spec_id "99".
        spec_dir = tmp_path / "05_my_feature"
        _write_spec_dir(spec_dir, spec_id="99", spec_name="my_feature")

        spec = afspec.load_spec(spec_dir)
        errors = validate_cross_file(spec)

        fn_errors = _folder_name_errors(errors)
        assert len(fn_errors) >= 1, f"expected at least one folder_name error for spec_id mismatch, got {errors}"

        # Error must reference both the folder prefix and the frontmatter spec_id.
        messages = [e.message for e in fn_errors]
        assert any("99" in m and "05" in m for m in messages), (
            f"folder_name error should reference '99' (spec_id) and '05' (folder prefix); got: {messages}"
        )


# ---------------------------------------------------------------------------
# TS-NS-2: spec_name mismatch with folder suffix
# ---------------------------------------------------------------------------


class TestFolderNameSpecNameMismatch:
    """TS-NS-2: spec_name 'other_thing' does not match folder suffix 'my_feature'."""

    def test_spec_name_mismatch_produces_folder_name_error(self, tmp_path: Path) -> None:
        # Spec lives in 05_my_feature but prd.md declares spec_name "other_thing".
        spec_dir = tmp_path / "05_my_feature"
        _write_spec_dir(spec_dir, spec_id="05", spec_name="other_thing")

        spec = afspec.load_spec(spec_dir)
        errors = validate_cross_file(spec)

        fn_errors = _folder_name_errors(errors)
        assert len(fn_errors) >= 1, f"expected at least one folder_name error for spec_name mismatch, got {errors}"

        # Error must reference both the folder suffix and the frontmatter spec_name.
        messages = [e.message for e in fn_errors]
        assert any("other_thing" in m and "my_feature" in m for m in messages), (
            f"folder_name error should reference 'other_thing' (spec_name) and 'my_feature' "
            f"(folder suffix); got: {messages}"
        )


# ---------------------------------------------------------------------------
# TS-NS-3: matching spec_id and spec_name produce no folder-name errors
# ---------------------------------------------------------------------------


class TestFolderNameMatchingSpec:
    """TS-NS-3: spec_id '05' and spec_name 'my_feature' match '05_my_feature' folder."""

    def test_matching_spec_has_no_folder_name_errors(self, tmp_path: Path) -> None:
        spec_dir = tmp_path / "05_my_feature"
        _write_spec_dir(spec_dir, spec_id="05", spec_name="my_feature")

        spec = afspec.load_spec(spec_dir)
        errors = validate_cross_file(spec)

        fn_errors = _folder_name_errors(errors)
        assert fn_errors == [], f"expected zero folder_name errors for matching spec, got: {fn_errors}"


# ---------------------------------------------------------------------------
# TS-NS-4: non-spec directory name skips folder-name validation
# ---------------------------------------------------------------------------


class TestFolderNameNonSpecDirectory:
    """TS-NS-4: 'scratch' directory does not trigger folder-name validation."""

    def test_non_spec_directory_skips_folder_name_check(self, tmp_path: Path) -> None:
        # "scratch" does not match the NN_snake_case pattern — validation is skipped.
        spec_dir = tmp_path / "scratch"
        # Use deliberately mismatched values that would fire if the check ran.
        _write_spec_dir(spec_dir, spec_id="99", spec_name="wrong_name")

        spec = afspec.load_spec(spec_dir)
        errors = validate_cross_file(spec)

        fn_errors = _folder_name_errors(errors)
        assert fn_errors == [], f"expected zero folder_name errors for non-spec directory, got: {fn_errors}"


# ---------------------------------------------------------------------------
# TS-NS-5: both spec_id and spec_name mismatches reported independently
# ---------------------------------------------------------------------------


class TestFolderNameBothMismatches:
    """TS-NS-5: both spec_id '99' and spec_name 'other_thing' mismatches reported."""

    def test_both_mismatches_produce_two_folder_name_errors(self, tmp_path: Path) -> None:
        spec_dir = tmp_path / "05_my_feature"
        _write_spec_dir(spec_dir, spec_id="99", spec_name="other_thing")

        spec = afspec.load_spec(spec_dir)
        errors = validate_cross_file(spec)

        fn_errors = _folder_name_errors(errors)
        assert len(fn_errors) == 2, (
            f"expected exactly 2 folder_name errors (one for spec_id, one for spec_name), "
            f"got {len(fn_errors)}: {fn_errors}"
        )

        # One error must mention spec_id, another spec_name.
        messages = [e.message for e in fn_errors]
        assert any("spec_id" in m for m in messages), f"no folder_name error mentions 'spec_id': {messages}"
        assert any("spec_name" in m for m in messages), f"no folder_name error mentions 'spec_name': {messages}"
