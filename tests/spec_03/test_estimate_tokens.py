"""Tests for the estimate_tokens utility function.

Test Spec: TS-03-1, TS-03-2, TS-03-3, TS-03-4
Requirements: 03-REQ-1
"""

from __future__ import annotations

import afspec

# ---------------------------------------------------------------------------
# TS-03-1: estimate_tokens returns len(text) // 4 for arbitrary strings
# Requirement: 03-REQ-1.1
# ---------------------------------------------------------------------------


def test_estimate_tokens_arbitrary_string() -> None:
    """Verify estimate_tokens returns len(text) // 4 for a 17-char string."""
    result = afspec.estimate_tokens("abcdefghijklmnopq")
    assert result == 4
    assert result == len("abcdefghijklmnopq") // 4


# ---------------------------------------------------------------------------
# TS-03-2: estimate_tokens('abcd') returns exactly 1
# Requirement: 03-REQ-1.2
# ---------------------------------------------------------------------------


def test_estimate_tokens_four_chars() -> None:
    """Verify estimate_tokens('abcd') returns integer 1."""
    result = afspec.estimate_tokens("abcd")
    assert result == 1
    assert isinstance(result, int)


# ---------------------------------------------------------------------------
# TS-03-3: estimate_tokens('') returns 0 for empty string
# Requirement: 03-REQ-1.3
# ---------------------------------------------------------------------------


def test_estimate_tokens_empty_string() -> None:
    """Verify estimate_tokens('') returns integer 0."""
    result = afspec.estimate_tokens("")
    assert result == 0


# ---------------------------------------------------------------------------
# TS-03-4: estimate_tokens is exported from afspec.__init__ and in __all__
# Requirement: 03-REQ-1.4
# ---------------------------------------------------------------------------


def test_estimate_tokens_exported() -> None:
    """Verify estimate_tokens is in afspec.__all__ and callable."""
    assert "estimate_tokens" in afspec.__all__
    assert callable(afspec.estimate_tokens)


# ---------------------------------------------------------------------------
# 03-REQ-1.E1: Strings of length 1, 2, or 3 return 0 (floor division)
# ---------------------------------------------------------------------------


def test_estimate_tokens_length_one() -> None:
    """Strings of length 1 return 0 due to floor division."""
    assert afspec.estimate_tokens("a") == 0


def test_estimate_tokens_length_two() -> None:
    """Strings of length 2 return 0 due to floor division."""
    assert afspec.estimate_tokens("ab") == 0


def test_estimate_tokens_length_three() -> None:
    """Strings of length 3 return 0 due to floor division."""
    assert afspec.estimate_tokens("abc") == 0


# ---------------------------------------------------------------------------
# 03-REQ-1.E2: String of exactly 4 chars returns 1
# ---------------------------------------------------------------------------


def test_estimate_tokens_exactly_four() -> None:
    """String of exactly 4 characters returns 1."""
    assert afspec.estimate_tokens("wxyz") == 1


# ---------------------------------------------------------------------------
# 03-REQ-1.E3: Large string (1 million chars) returns correct result
# ---------------------------------------------------------------------------


def test_estimate_tokens_large_string() -> None:
    """1,000,000-character string returns 250,000 without overflow."""
    text = "x" * 1_000_000
    result = afspec.estimate_tokens(text)
    assert result == 250_000


# ---------------------------------------------------------------------------
# 03-PROP-1: Floor division identity property
# Validates: 03-REQ-1.1, 03-REQ-1.2, 03-REQ-1.3
# ---------------------------------------------------------------------------


def test_estimate_tokens_eight_chars() -> None:
    """8-character string returns 2."""
    assert afspec.estimate_tokens("12345678") == 2


def test_estimate_tokens_five_chars() -> None:
    """5-character string returns 1 (floor division)."""
    assert afspec.estimate_tokens("hello") == 1


def test_estimate_tokens_seven_chars() -> None:
    """7-character string returns 1 (floor division)."""
    assert afspec.estimate_tokens("goodbye") == 1
