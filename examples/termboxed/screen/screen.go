package screen

import (
	"github.com/ehedgehog/guineapig/examples/termboxed/grid"
	"github.com/gdamore/tcell"
)

////////////////////////////////////////////////////////////////

type Canvas interface {
	Size() grid.Size
	SetCell(where grid.LineCol, ch rune, s tcell.Style)
	SetCursor(where grid.LineCol)
}

var DefaultStyle tcell.Style

var StyleBackCyan = DefaultStyle.Background(tcell.ColorDarkCyan)

var StyleBackYellow = DefaultStyle.Background(tcell.ColorDarkCyan)

func PutString(c Canvas, x, y int, content string, s tcell.Style) {
	i := 0
	size := c.Size()
	w := size.Width
	limit := w - x
	// sprime := *s.(*StyleStruct)
	for _, ch := range content {
		if i > limit {
			break
		}
		// sprime.fg += 1
		scurrent := s
		//		if i&1 == 0 {
		//			scurrent = &sprime
		//		}
		if ch == '\t' {
			for counter := 0; counter < 4; counter += 1 {
				c.SetCell(grid.LineCol{Col: x + i, Line: y}, ch, scurrent)
				i += 1
			}
		} else {
			c.SetCell(grid.LineCol{Col: x + i, Line: y}, ch, scurrent)
			i += 1
		}

	}
}

///////////////////////////////////////////////////////////////

type TermboxCanvas struct {
	size grid.Size
}

var TheScreen tcell.Screen

func init() {
	TheScreen, _ = tcell.NewScreen()
}

func NewTermboxCanvas() *TermboxCanvas {
	w, h := TheScreen.Size()
	return &TermboxCanvas{grid.Size{Width: w, Height: h}}
}

func (t *TermboxCanvas) Size() grid.Size {
	return t.size
}

func (t *TermboxCanvas) SetCursor(where grid.LineCol) {
	TheScreen.ShowCursor(where.Col, where.Line)
}

func (t *TermboxCanvas) SetCell(where grid.LineCol, glyph rune, s tcell.Style) {
	TheScreen.SetContent(where.Col, where.Line, glyph, []rune{}, s)
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

func (s *SubCanvas) SetCell(where grid.LineCol, glyph rune, st tcell.Style) {
	s.outer.SetCell(where.Plus(s.offset), glyph, st)
}

func NewSubCanvas(outer Canvas, dx, dy, w, h int) Canvas {
	return &SubCanvas{outer, grid.LineCol{Col: dx, Line: dy}, grid.Size{Width: w, Height: h}}
}

///////////////////////////////////////////////////////////////
