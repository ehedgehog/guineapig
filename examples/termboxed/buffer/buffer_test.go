package buffer

import "testing"

func TestCanCreateBuffer(t *testing.T) {
	const (
		width  = 80
		height = 50
	)
	b := New(execFunction, width, height).(*SimpleBuffer)
	eq(t, "should be at line 0", b.line, 0)
	eq(t, "should be at column 0", b.col, 0)
}

func execFunction(b Type, args string) {
	// nothing (yet)
}

func eq(t *testing.T, oops string, a, b interface{}) {
	if a != b {
		t.Errorf("%s: got %v, expected %v.", oops, a, b)
	}
}
