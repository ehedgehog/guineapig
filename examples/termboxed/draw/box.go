package draw

import (
	"github.com/ehedgehog/guineapig/examples/termboxed/bounds"
	"github.com/ehedgehog/guineapig/examples/termboxed/screen"
)

const (
	Glyph_hbar      = '─'
	Glyph_vbar      = '│'
	Glyph_corner_tl = '┌'
	Glyph_corner_tr = '┐'
	Glyph_corner_bl = '└'
	Glyph_corner_br = '┘'
	Glyph_plus      = '┼'
	Glyph_T         = '┬'
	Glyph_pin       = '┴'
	Glyph_lstile    = '├'
	Glyph_rstile    = '┤'
)

type ScrollInfo struct {
	Lines  int
	OnLine int
}

const (
	topOffset = 2
	botOffset = 2
)

func Scrollbar(sw screen.Canvas, s ScrollInfo) {
	//

	_, h := sw.Size()

	contentSize := s.Lines
	currentLineIndex := s.OnLine

	for yy := 0; yy < h; yy += 1 {
		sw.SetCell(0, yy, Glyph_vbar, screen.DefaultStyle)
	}

	if contentSize < h {
		return
	}

	//
	y := topOffset
	bigy := h - 1 - botOffset

	sw.SetCell(0, y, Glyph_pin, screen.DefaultStyle)
	//

	y += 1
	//
	zoneSize := bigy - y
	barSize := bounds.Max(1, zoneSize*h/contentSize)
	downset := currentLineIndex * (zoneSize - barSize) / contentSize

	//
	for yy := y; yy < y+downset; yy += 1 {
		sw.SetCell(0, yy, ' ', screen.DefaultStyle)
	}

	for yy := y + downset; yy < y+downset+barSize; yy += 1 {
		sw.SetCell(0, yy, ' ', screen.StyleBackCyan)
	}

	for yy := y + downset + barSize; yy < bigy; yy += 1 {
		sw.SetCell(0, yy, ' ', screen.DefaultStyle)
	}
	//
	sw.SetCell(0, bigy, Glyph_T, screen.DefaultStyle)
}
