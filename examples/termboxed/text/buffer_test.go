package text

import "testing"

func TestCanCreateBuffer(t *testing.T) {
//	b := New(execFunction).(*SimpleBuffer)
//	eq(t, "should be at line 0", b.line, 0)
//	eq(t, "should be at column 0", b.col, 0)
//	line, content := b.Expose()
//	eq(t, "should be at line 0", line, 0)
//	eq(t, "should have no content", len(content), 0)
}
//
//type Command func(Buffer)
//
//type Predicate func(Buffer)
//
//type Test struct {
//	commands []Command
//	test     Predicate
//}
//
//func TestSequence(t *testing.T) {
//	seq := &Test{
//		[]Command{
//			Command(func(b Buffer) { b.ForwardOne() }),
//			Command(func(b Buffer) { b.(*SimpleBuffer).makeRoom() }),
//		},
//		func(b Buffer) {
//			col, line := b.Where()
//			eq(t, "should be at first line", line, 0)
//			eq(t, "should be one char along", col, 1)
//		},
//	}
//	b := New(execFunction)
//	for _, c := range seq.commands {
//		c(b)
//	}
//	seq.test(b)
//}
//
//func TestSequence(t *testing.T) {
//	b := New(execFunction)
//	b.ForwardOne()
//	b.(*SimpleBuffer).makeRoom()
//	col, line := b.Where()
//	eq(t, "should be at first line", line, 0)
//	eq(t, "should be one char along", col, 1)
//}
//
//func TestInsertCharacterInEmptyBuffer(t *testing.T) {
//	b := New(execFunction)
//	b.Insert('1')
//	line, content := b.Expose()
//	eq(t, "should be at line 0", line, 0)
//	eq(t, "should have just one line", len(content), 1)
//	eq(t, "line should be '1'", content[0], "1")
//}
//
//func execFunction(b Buffer, args string) {
//	// nothing (yet)
//}
//
//func eq(t *testing.T, oops string, a, b interface{}) {
//	if a != b {
//		t.Errorf("%s: got %v, expected %v.", oops, a, b)
//	}
//}
