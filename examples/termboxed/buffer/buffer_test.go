package buffer

import "testing"

func TestCanCreateBuffer(t *testing.T) {
	const (
		width  = 80
		height = 50
	)
	b := New(width, height).(*SimpleBuffer)
	eq(t, "should be at line 0", b.line, 0)
	eq(t, "should be at column 0", b.col, 0)
	eq(t, "should have no scroll offset", b.verticalOffset, 0)
	eq(t, "should have given width", b.width, width)
	eq(t, "should have given height", b.height, height)
}

func eq(t *testing.T, oops string, a, b interface{}) {
	if a != b {
		t.Errorf("%s: got %v, expected %v.", oops, a, b)
	}
}
