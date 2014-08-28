package screen

import "github.com/nsf/termbox-go"

type Writeable interface {
	PutString(x, y int, content string)
	SetCell(x, y int, ch rune, a, b termbox.Attribute)
	SetCursor(x, y int)
	Resize(x, y, w, h int)
	Size() (w, h int)
}

type Panel interface {
	Writeable
}

func (s *ScreenWritable) PutString(x, y int, content string) {
	for i, ch := range content {
		s.SetCell(x+i, y, ch, termbox.ColorDefault, termbox.ColorDefault)
	}
}

func (s *ScreenWritable) Resize(x, y, w, h int) {
	s.x, s.y = x, y
	s.w, s.h = w, h
}

func (s *ScreenWritable) Size() (w, h int) {
	return s.w, s.h
}

func (s *ScreenWritable) SetCursor(x, y int) {
	termbox.SetCursor(x, y)
}

func (s *ScreenWritable) SetCell(x, y int, ch rune, a, b termbox.Attribute) {
	termbox.SetCell(x, y, ch, a, b)
}

type ScreenWritable struct {
	x, y, w, h int
}

func NewScreenWriteable(x, y, w, h int) Writeable {
	return &ScreenWritable{x, y, w, h}
}
