package draw

import (
	"fmt"

	"github.com/nsf/termbox-go"
)

const (
	glyph_hbar      = '─'
	glyph_vbar      = '│'
	glyph_corner_tl = '┌'
	glyph_corner_tr = '┐'
	glyph_corner_bl = '└'
	glyph_corner_br = '┘'
	glyph_plus      = '┼'
	glyph_T         = '┬'
	glyph_pin       = '┴'
	glyph_lstile    = '├'
	glyph_rstile    = '┤'
)

type XY struct {
	X, Y int
}

type WH struct {
	W, H int
}

type Scrolling struct {
	Lines  int
	OnLine int
}

type BoxInfo struct {
	Loc  XY
	Size WH
	Off  Scrolling
}

func Box(content string, b BoxInfo) {

	xbase, ybase := b.Loc.X, b.Loc.Y
	w, h := b.Size.W, b.Size.H

	for x := 1; x < w-1; x += 1 {
		for _, y := range []int{0, h - 1} {
			termbox.SetCell(xbase+x, ybase+y, glyph_hbar, termbox.ColorDefault, termbox.ColorDefault)
		}
	}

	for y := 1; y < h-1; y += 1 {
		for _, x := range []int{0, w - 1} {
			termbox.SetCell(xbase+x, ybase+y, glyph_vbar, termbox.ColorDefault, termbox.ColorDefault)
		}
	}

	county := []rune(fmt.Sprintf("─┤ xy, wh, off: %v ", b))
	for i, r := range county {
		termbox.SetCell(xbase+i+1, ybase, r, termbox.ColorDefault, termbox.ColorDefault)
	}

	termbox.SetCell(xbase, ybase, glyph_corner_tl, termbox.ColorDefault, termbox.ColorDefault)
	termbox.SetCell(xbase+w-1, ybase, glyph_corner_tr, termbox.ColorDefault, termbox.ColorDefault)
	termbox.SetCell(xbase+w-1, ybase+h-1, glyph_corner_br, termbox.ColorDefault, termbox.ColorDefault)
	termbox.SetCell(xbase, ybase+h-1, glyph_corner_bl, termbox.ColorDefault, termbox.ColorDefault)

	scrollbar(b)
}

const (
	topOffset = 4
	botOffset = 4
)

func Say(x, y int, message string) {
	for _, ch := range message {
		termbox.SetCell(x, y, ch, termbox.ColorDefault, termbox.ColorDefault)
		x += 1
	}
}

func scrollbar(b BoxInfo) {
	//

	xbase, ybase := b.Loc.X, b.Loc.Y
	w, h := b.Size.W, b.Size.H

	if b.Off.Lines < h {
		return
	}

	//
	x := xbase + w - 1
	y := ybase + topOffset
	bigy := ybase + h - 1 - botOffset

	termbox.SetCell(x, y, glyph_pin, termbox.ColorDefault, termbox.ColorDefault)
	//
	contentSize := b.Off.Lines
	currentLineIndex := b.Off.OnLine

	y += 1
	//
	zoneSize := bigy - y
	barSize := zoneSize * h / contentSize
	if barSize == 0 {
		barSize = 1
	}
	downset := currentLineIndex * (zoneSize - barSize) / contentSize

	// Say(10, 8, fmt.Sprintf("zoneSize %v, h %v, barSize %v", zoneSize, h, barSize))
	// Say(10, 9, fmt.Sprintf("line %v, gapSize %v, contentsSize %v", currentLineIndex, zoneSize-barSize, contentSize))
	// Say(10, 10, fmt.Sprintf("zoneSize %v, barSize %v, downset %v", zoneSize, barSize, downset))

	//
	for yy := y; yy < y+downset; yy += 1 {

		termbox.SetCell(x, yy, ' ', termbox.ColorDefault, termbox.ColorDefault)

	}

	for yy := y + downset; yy < y+downset+barSize; yy += 1 {
		termbox.SetCell(x, yy, ' ', termbox.ColorDefault, termbox.ColorCyan)
	}

	for yy := y + downset + barSize; yy < bigy; yy += 1 {

		termbox.SetCell(x, yy, ' ', termbox.ColorDefault, termbox.ColorDefault)

	}
	//
	termbox.SetCell(x, bigy, glyph_T, termbox.ColorDefault, termbox.ColorDefault)
}
