package main

// import "github.com/limetext/termbox-go"

import "fmt"

import (
	_ "github.com/ehedgehog/guineapig/examples/termboxed/panel"
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

func draw(count int, content string) {
	w, h := termbox.Size()
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	box(count, content, w/2, 0, w-w/2, h)
	box(count, content, 0, 0, w/2, h)

	termbox.Flush()
}

func box(count int, content string, xbase, ybase int, w, h int) {

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

	for i, ch := range []rune(content) {
		termbox.SetCell(xbase+1+i, ybase+1, ch, termbox.ColorDefault, termbox.ColorDefault)
	}

	termbox.SetCursor(xbase+1+len([]rune(content)), ybase+1)

	x := []rune(content)
	if len(x) > 0 {
		count = int(x[len(x)-1])
	}

	county := []rune(fmt.Sprintf("─┤ %d ", count))
	for i, r := range county {
		termbox.SetCell(xbase+i+1, ybase, r, termbox.ColorDefault, termbox.ColorDefault)
	}

	termbox.SetCell(xbase, ybase, glyph_corner_tl, termbox.ColorDefault, termbox.ColorDefault)
	termbox.SetCell(xbase+w-1, ybase, glyph_corner_tr, termbox.ColorDefault, termbox.ColorDefault)
	termbox.SetCell(xbase+w-1, ybase+h-1, glyph_corner_br, termbox.ColorDefault, termbox.ColorDefault)
	termbox.SetCell(xbase, ybase+h-1, glyph_corner_bl, termbox.ColorDefault, termbox.ColorDefault)

	termbox.Flush()
}

func main() {

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	termbox.SetInputMode(termbox.InputMouse)

	count := 0
	content := ""
	mx := 0

	for {
		draw(mx, content)
		count += 1
		ev := termbox.PollEvent()
		if ev.Type == termbox.EventKey && ev.Key == termbox.KeyEsc {
			return
		}
		if ev.Type == termbox.EventKey {
			ch := ev.Ch
			if ch == 0 {
				ch = ' '
			}
			content = string(append([]rune(content), ch))
		}
		mx = ev.MouseX
	}
}
