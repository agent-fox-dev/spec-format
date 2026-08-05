package afspec

import "testing"

// requireImplemented recovers from panics caused by "not implemented" stubs
// and converts them to test failures. Place as a deferred call at the top of
// any test that calls unimplemented functions.
//
//	func TestFoo(t *testing.T) {
//	    defer requireImplemented(t)
//	    // ... test body that may call unimplemented stubs
//	}
func requireImplemented(t *testing.T) {
	t.Helper()
	if r := recover(); r != nil {
		t.Fatalf("not yet implemented (panic: %v)", r)
	}
}
