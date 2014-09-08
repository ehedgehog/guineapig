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

type WH struct {
	W, H int
}

type Scrolling struct {
	Lines  int
	OnLine int
}

type BoxInfo struct {
	Size WH
	Off  Scrolling
}

//func Box(sw screen.Canvas, content string, b BoxInfo) {
//
//	xbase, ybase := b.Loc.X, b.Loc.Y
//	w, h := b.Size.W, b.Size.H
//
//	for x := 1; x < w-1; x += 1 {
//		for _, y := range []int{0, h - 1} {
//			sw.SetCell(xbase+x, ybase+y, Glyph_hbar, screen.DefaultStyle)
//		}
//	}
//
//	for y := 1; y < h-1; y += 1 {
//		for _, x := range []int{0, w - 1} {
//			sw.SetCell(xbase+x, ybase+y, Glyph_vbar, screen.DefaultStyle)
//		}
//	}
//
//	county := []rune(fmt.Sprintf("─┤ xy, wh, off: %v ", b))
//	for i, r := range county {
//		sw.SetCell(xbase+i+1, ybase, r, screen.DefaultStyle)
//	}
//
//	sw.SetCell(xbase, ybase, Glyph_corner_tl, screen.DefaultStyle)
//	sw.SetCell(xbase+w-1, ybase, Glyph_corner_tr, screen.DefaultStyle)
//	sw.SetCell(xbase+w-1, ybase+h-1, Glyph_corner_br, screen.DefaultStyle)
//	sw.SetCell(xbase, ybase+h-1, Glyph_corner_bl, screen.DefaultStyle)
//
//	Scrollbar(sw, b)
//}

const (
	topOffset = 2
	botOffset = 2
)

//func Say(x, y int, message string) {
//	for _, ch := range message {
//		termbox.SetCell(x, y, ch, termbox.ColorDefault, termbox.ColorDefault)
//		x += 1
//	}
//}

//func max(a, b int) int {
//	if a < b {
//		return b
//	} else {
//		return a
//	}
//}

func Scrollbar(sw screen.Canvas, b BoxInfo) {
	//
	h := b.Size.H

	for yy := 0; yy < h; yy += 1 {
		sw.SetCell(0, yy, Glyph_vbar, screen.DefaultStyle)
	}

	if b.Off.Lines < h {
		return
	}

	//
	y := topOffset
	bigy := h - 1 - botOffset

	sw.SetCell(0, y, Glyph_pin, screen.DefaultStyle)
	//
	contentSize := b.Off.Lines
	currentLineIndex := b.Off.OnLine

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
