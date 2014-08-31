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

////////////////////////////////////////////////////////////////

type Canvas interface {
	Size() (w, h int)
	SetCell(x, y int, ch rune, s Style)
	SetCursor(x, y int)
}

type Style interface {
	Foreground() termbox.Attribute
	Background() termbox.Attribute
}

type StyleStruct struct {
	fg, bg termbox.Attribute
}

func (ss *StyleStruct) Foreground() termbox.Attribute {
	return ss.fg
}

func (ss *StyleStruct) Background() termbox.Attribute {
	return ss.bg
}

var DefaultStyle = &StyleStruct{termbox.ColorDefault, termbox.ColorDefault}

var StyleBackCyan = &StyleStruct{termbox.ColorDefault, termbox.ColorCyan}

func PutString(c Canvas, x, y int, content string, s Style) {
	i := 0
	for _, ch := range content {
		c.SetCell(x+i, y, ch, s)
		i += 1
	}
}

///////////////////////////////////////////////////////////////

type TermboxCanvas struct {
	width  int
	height int
}

func NewTermboxCanvas() *TermboxCanvas {
	w, h := termbox.Size()
	return &TermboxCanvas{w, h}
}

func (t *TermboxCanvas) Size() (w, h int) {
	return t.width, t.height
}

func (t *TermboxCanvas) SetCursor(x, y int) {
	termbox.SetCursor(x, y)
}

func (t *TermboxCanvas) SetCell(x, y int, glyph rune, s Style) {
	termbox.SetCell(x, y, glyph, s.Foreground(), s.Background())
}

///////////////////////////////////////////////////////////////

type SubCanvas struct {
	parent  Canvas
	xOffset int
	yOffset int
	width   int
	height  int
}

func (s *SubCanvas) Size() (w, h int) {
	return s.width, s.height
}

func (s *SubCanvas) SetCursor(x, y int) {
	s.parent.SetCursor(x+s.xOffset, y+s.yOffset)
}

func (s *SubCanvas) SetCell(x, y int, glyph rune, st Style) {
	s.parent.SetCell(x+s.xOffset, y+s.yOffset, glyph, st)
}

func NewSubCanvas(outer Canvas, dx, dy, w, h int) Canvas {
	return &SubCanvas{outer, dx, dy, w, h}
}

///////////////////////////////////////////////////////////////

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
