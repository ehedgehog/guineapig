package screen

import (
	"github.com/ehedgehog/guineapig/examples/termboxed/grid"
	"github.com/nsf/termbox-go"
)

////////////////////////////////////////////////////////////////

type Canvas interface {
	Size() grid.Size
	SetCell(where grid.LineCol, ch rune, s Style)
	SetCursor(where grid.LineCol)
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

var StyleBackYellow = &StyleStruct{termbox.ColorDefault, termbox.ColorYellow}

func PutString(c Canvas, x, y int, content string, s Style) {
	i := 0
	size := c.Size()
	w := size.Width
	limit := w - x
	sprime := *s.(*StyleStruct)
	for _, ch := range content {
		if i > limit {
			break
		}
		// sprime.fg += 1
		scurrent := s
		if i&1 == 0 {
			scurrent = &sprime
		}
		c.SetCell(grid.LineCol{Col: x + i, Line: y}, ch, scurrent)
		i += 1
	}
}

///////////////////////////////////////////////////////////////

type TermboxCanvas struct {
	size grid.Size
}

func NewTermboxCanvas() *TermboxCanvas {
	w, h := termbox.Size()
	return &TermboxCanvas{grid.Size{Width: w, Height: h}}
}

func (t *TermboxCanvas) Size() grid.Size {
	return t.size
}

func (t *TermboxCanvas) SetCursor(where grid.LineCol) {
	termbox.SetCursor(where.Col, where.Line)
}

func (t *TermboxCanvas) SetCell(where grid.LineCol, glyph rune, s Style) {
	termbox.SetCell(where.Col, where.Line, glyph, s.Foreground(), s.Background())
}

///////////////////////////////////////////////////////////////

type SubCanvas struct {
	outer  Canvas
	offset grid.LineCol
	size   grid.Size
}

func (s *SubCanvas) Size() grid.Size {
	return s.size
}

func (s *SubCanvas) SetCursor(where grid.LineCol) {
	s.outer.SetCursor(where.Plus(s.offset))
}

func (s *SubCanvas) SetCell(where grid.LineCol, glyph rune, st Style) {
	s.outer.SetCell(where.Plus(s.offset), glyph, st)
}

func NewSubCanvas(outer Canvas, dx, dy, w, h int) Canvas {
	return &SubCanvas{outer, grid.LineCol{Col: dx, Line: dy}, grid.Size{Width: w, Height: h}}
}

///////////////////////////////////////////////////////////////
