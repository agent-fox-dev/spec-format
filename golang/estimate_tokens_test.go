package afspec

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// TS-03-5: EstimateTokens is exported and returns len(text) / 4
// Requirement: 03-REQ-1.5
// ---------------------------------------------------------------------------

func TestEstimateTokens_EightChars(t *testing.T) {
	defer requireImplemented(t)

	result := EstimateTokens("abcdefgh")
	if result != 2 {
		t.Errorf("EstimateTokens(\"abcdefgh\") = %d, want 2", result)
	}
}

func TestEstimateTokens_FourChars(t *testing.T) {
	defer requireImplemented(t)

	result := EstimateTokens("abcd")
	if result != 1 {
		t.Errorf("EstimateTokens(\"abcd\") = %d, want 1", result)
	}
}

func TestEstimateTokens_EmptyString(t *testing.T) {
	defer requireImplemented(t)

	result := EstimateTokens("")
	if result != 0 {
		t.Errorf("EstimateTokens(\"\") = %d, want 0", result)
	}
}

// ---------------------------------------------------------------------------
// 03-REQ-1.E1: Short strings (1-3 chars) return 0
// ---------------------------------------------------------------------------

func TestEstimateTokens_OneChar(t *testing.T) {
	defer requireImplemented(t)

	result := EstimateTokens("a")
	if result != 0 {
		t.Errorf("EstimateTokens(\"a\") = %d, want 0", result)
	}
}

func TestEstimateTokens_TwoChars(t *testing.T) {
	defer requireImplemented(t)

	result := EstimateTokens("ab")
	if result != 0 {
		t.Errorf("EstimateTokens(\"ab\") = %d, want 0", result)
	}
}

func TestEstimateTokens_ThreeChars(t *testing.T) {
	defer requireImplemented(t)

	result := EstimateTokens("abc")
	if result != 0 {
		t.Errorf("EstimateTokens(\"abc\") = %d, want 0", result)
	}
}

// ---------------------------------------------------------------------------
// 03-REQ-1.E2: Exactly 4 chars returns 1
// ---------------------------------------------------------------------------

func TestEstimateTokens_ExactlyFour(t *testing.T) {
	defer requireImplemented(t)

	result := EstimateTokens("wxyz")
	if result != 1 {
		t.Errorf("EstimateTokens(\"wxyz\") = %d, want 1", result)
	}
}

// ---------------------------------------------------------------------------
// 03-REQ-1.E3: Large string (1 million chars) returns correct result
// ---------------------------------------------------------------------------

func TestEstimateTokens_LargeString(t *testing.T) {
	defer requireImplemented(t)

	text := strings.Repeat("x", 1_000_000)
	result := EstimateTokens(text)
	if result != 250_000 {
		t.Errorf("EstimateTokens(1M chars) = %d, want 250000", result)
	}
}

// ---------------------------------------------------------------------------
// 03-PROP-1: Floor division identity for arbitrary string
// Validates: 03-REQ-1.1, 03-REQ-1.2, 03-REQ-1.3
// ---------------------------------------------------------------------------

func TestEstimateTokens_ArbitraryString(t *testing.T) {
	defer requireImplemented(t)

	text := "abcdefghijklmnopq" // 17 chars
	result := EstimateTokens(text)
	expected := len(text) / 4
	if result != expected {
		t.Errorf("EstimateTokens(%q) = %d, want %d", text, result, expected)
	}
}
