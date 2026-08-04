"""Spec linting: discover and validate specification packs.

Consolidates spec discovery, validation finding types, and the
lint runner into a single module.
"""

from __future__ import annotations

import logging
from dataclasses import dataclass, field
from pathlib import Path

import afspec
from afspec.discovery import parse_spec_dir_name

logger = logging.getLogger(__name__)

# ---------------------------------------------------------------------------
# Finding types
# ---------------------------------------------------------------------------

SEVERITY_ERROR = "error"
SEVERITY_WARNING = "warning"
SEVERITY_HINT = "hint"

SEVERITY_ORDER = {SEVERITY_ERROR: 0, SEVERITY_WARNING: 1, SEVERITY_HINT: 2}


@dataclass(frozen=True)
class Finding:
    """A single validation finding."""

    spec_name: str
    file: str
    rule: str
    severity: str
    message: str
    line: int | None


def sort_findings(findings: list[Finding]) -> list[Finding]:
    return sorted(
        findings,
        key=lambda f: (f.spec_name, f.file, SEVERITY_ORDER.get(f.severity, 99)),
    )


def compute_exit_code(findings: list[Finding]) -> int:
    return 1 if any(f.severity == SEVERITY_ERROR for f in findings) else 0


# ---------------------------------------------------------------------------
# Spec discovery
# ---------------------------------------------------------------------------


@dataclass(frozen=True)
class SpecInfo:
    """Metadata about a discovered specification folder."""

    name: str
    prefix: int
    path: Path
    has_tasks: bool
    has_prd: bool


class LintError(Exception):
    """Raised when lint cannot proceed (e.g. missing specs directory)."""


def discover_specs(
    specs_dir: Path,
    filter_spec: str | None = None,
) -> list[SpecInfo]:
    """Discover spec folders with a requirements.json file present."""
    if not specs_dir.is_dir():
        raise LintError(f"No specifications found: '{specs_dir}' does not exist")

    specs: list[SpecInfo] = []
    found_candidates = False
    for entry in sorted(specs_dir.iterdir()):
        if not entry.is_dir():
            continue
        parsed = parse_spec_dir_name(entry.name)
        if parsed is None:
            continue

        found_candidates = True
        prefix, _ = parsed

        if not (entry / "requirements.json").is_file():
            logger.debug("Spec folder '%s' has no requirements.json, skipping", entry.name)
            continue

        has_tasks = (entry / "tasks.json").is_file()
        has_prd = (entry / "prd.md").is_file()

        if not has_tasks:
            logger.warning("Spec folder '%s' has no tasks.json", entry.name)

        specs.append(
            SpecInfo(name=entry.name, prefix=prefix, path=entry, has_tasks=has_tasks, has_prd=has_prd)
        )

    if not specs:
        if not found_candidates:
            raise LintError(f"No specifications found in '{specs_dir}'")
        return []

    specs.sort(key=lambda s: s.prefix)

    if filter_spec is not None:
        filtered = [s for s in specs if s.name == filter_spec]
        if not filtered:
            available = ", ".join(s.name for s in specs)
            raise LintError(f"Spec '{filter_spec}' not found. Available specs: {available}")
        return filtered

    return specs


# ---------------------------------------------------------------------------
# Lint runner
# ---------------------------------------------------------------------------

_KNOWN_SEVERITIES = {"error", "warning", "hint"}


def _map_afspec_findings(
    spec_name: str,
    errors: list[afspec.ValidationError],
) -> list[Finding]:
    findings: list[Finding] = []
    for ve in errors:
        raw_severity = getattr(ve, "severity", None)
        severity = raw_severity if raw_severity in _KNOWN_SEVERITIES else "error"
        findings.append(
            Finding(
                spec_name=spec_name,
                file=ve.file,
                rule=ve.rule,
                severity=severity,
                message=ve.message,
                line=getattr(ve, "line", None),
            )
        )
    return findings


def _validate_spec(spec: SpecInfo) -> list[Finding]:
    try:
        loaded = afspec.load_spec(spec.path)
        result = afspec.validate(loaded)
        return _map_afspec_findings(spec.name, result.errors)
    except Exception as exc:
        return [
            Finding(
                spec_name=spec.name,
                file=str(spec.path),
                rule="afspec-error",
                severity="error",
                message=str(exc),
                line=None,
            )
        ]


def _is_spec_implemented(spec: SpecInfo) -> bool:
    try:
        loaded = afspec.load_spec(spec.path)
        if not loaded.tasks or not loaded.tasks.task_groups:
            return False
        return all(all(st.state == afspec.SubtaskState.DONE for st in g.subtasks) for g in loaded.tasks.task_groups)
    except Exception:
        return False


@dataclass(frozen=True)
class LintResult:
    findings: list[Finding] = field(default_factory=list)
    exit_code: int = 0


def run_lint_specs(
    specs_dir: Path,
    *,
    lint_all: bool = False,
    progress_callback: None = None,
) -> LintResult:
    """Run spec linting and return structured results."""
    if not specs_dir.exists():
        raise LintError(f"Specs directory not found: {specs_dir}")

    if progress_callback is not None:
        progress_callback("Discovering specs...")
    try:
        discovered: list[SpecInfo] = discover_specs(specs_dir)
    except LintError:
        no_spec_finding = Finding(
            spec_name="(none)",
            file=str(specs_dir),
            rule="no-specs",
            severity="error",
            message=f"No specifications found in {specs_dir} directory",
            line=None,
        )
        return LintResult(findings=[no_spec_finding], exit_code=1)

    if not lint_all:
        filtered = [s for s in discovered if not _is_spec_implemented(s)]
        skipped = len(discovered) - len(filtered)
        if skipped > 0:
            logger.info("Skipping %d fully-implemented spec(s) (use --all to include)", skipped)
        if not filtered:
            return LintResult(findings=[], exit_code=0)
        discovered = filtered

    if progress_callback is not None:
        progress_callback(f"Validating {len(discovered)} spec(s)...")

    findings: list[Finding] = []
    for spec in discovered:
        findings.extend(_validate_spec(spec))

    findings = sort_findings(findings)
    exit_code = compute_exit_code(findings)
    return LintResult(findings=findings, exit_code=exit_code)
