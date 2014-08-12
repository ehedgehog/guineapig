package main

// import "github.com/limetext/termbox-go"

import "github.com/nsf/termbox-go"
import "fmt"

// import "log"

// import _ "github.com/ehedgehog/guineapig/examples/termboxed/panel"

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

type Writeable interface {
	PutString(x, y int, content string)
	SetCursor(x, y int)
}

type Panel interface {
	Writeable
}

type SimplePanel struct {
	x, y int
	w, h int
}

func (s *SimplePanel) PutString(x, y int, content string) {
	for i, ch := range content {
		termbox.SetCell(x+i, y, ch, termbox.ColorDefault, termbox.ColorDefault)
	}
}

func (s *SimplePanel) SetCursor(x, y int) {
	termbox.SetCursor(x, y)
}

func NewPanel(x, y int, w, h int) Panel {
	return &SimplePanel{x, y, w, h}
}

type Buffer interface {
	Insert(ch rune)
	DeleteBack()
	DeleteForward()
	BackOne()
	UpOne()
	DownOne()
	ForwardOne()
	Return()
	PutAll(w Writeable)
	ScrollUp()
	ScrollDown()
	ScrollTop()
}

type SimpleBuffer struct {
	content        []string
	line           int
	col            int
	verticalOffset int
	width          int
	height         int
}

func (b *SimpleBuffer) makeRoom() {
	if b.line >= len(b.content) {
		content := make([]string, b.line+1)
		copy(content, b.content)
		b.content = content
	}
	for b.col > len(b.content[b.line]) {
		b.content[b.line] += "        "
	}
}

func (b *SimpleBuffer) ScrollUp() {
	if b.verticalOffset > 0 {
		b.verticalOffset -= 1
		b.line -= 1
	}
}

func (b *SimpleBuffer) ScrollDown() {
	b.verticalOffset += 1
}

func (b *SimpleBuffer) ScrollTop() {
	b.line = 0
	b.verticalOffset = 0
}

func (b *SimpleBuffer) Insert(ch rune) {

	b.makeRoom()

	loc := b.col
	runes := []rune(b.content[b.line])

	A := []rune{}
	B := append(A, runes[0:loc]...)
	C := append(B, ch)
	D := append(C, runes[loc:]...)

	b.col += 1
	b.content[b.line] = string(D)
}

func (b *SimpleBuffer) Return() {

	b.makeRoom()

	lines := append(b.content, "")

	right := lines[b.line][b.col:]
	left := lines[b.line][0:b.col]

	copy(lines[b.line+1:], lines[b.line:])
	lines[b.line] = left
	lines[b.line+1] = right
	b.line += 1
	b.col = 0
	b.content = lines
}

func (b *SimpleBuffer) UpOne() {
	if b.line > 0 {
		if b.line == b.verticalOffset && b.verticalOffset > 0 {
			b.verticalOffset -= 1
		}
		b.line -= 1
	}
}

func (b *SimpleBuffer) DownOne() {
	b.line += 1
	if b.line-b.verticalOffset > b.height-3 {
		b.verticalOffset += 1
	}
}

func (b *SimpleBuffer) BackOne() {
	if b.col > 0 {
		b.col -= 1
	}
}

func (b *SimpleBuffer) ForwardOne() {
	b.col += 1
}

func (b *SimpleBuffer) DeleteBack() {
	b.makeRoom()
	if b.col > 0 {
		content := b.content[b.line]
		before := content[0 : b.col-1]
		after := content[b.col:]
		newContent := before + after
		b.content[b.line] = newContent
		b.BackOne()
	}
}

func (b *SimpleBuffer) DeleteForward() {
	b.ForwardOne()
	b.DeleteBack()
}

func (b *SimpleBuffer) PutAll(w Writeable) {
	box(0, fmt.Sprintf("offset: %v, line: %v, cursor(col %v, line %v), height: %v", b.verticalOffset, b.line, b.col+1, b.line+1-b.verticalOffset, b.height), 0, 0, b.width, b.height)
	limit := b.height - 2
	if len(b.content) < limit {
		limit = len(b.content)
	}
	line := 1
	for i := b.verticalOffset; i < limit; i += 1 {
		w.PutString(1, line, b.content[i])
		line += 1
	}
	w.SetCursor(b.col+1, b.line+1-b.verticalOffset)
}

func NewBuffer(w, h int) Buffer {
	return &SimpleBuffer{content: []string{""}, width: w, height: h}
}

type EventHandler interface {
	Handle(e *termbox.Event) error
}

type Loc struct {
	X, Y int
}

type Editor struct {
	p Panel
	b Buffer
	l Loc
}

type SimpleEventHandler struct {
	e       *Editor
	count   int
	content string
}

func NewSimpleEventHandler(x, y int, w, h int) EventHandler {
	p := NewPanel(x, y, w, h)
	b := NewBuffer(w, h)
	l := Loc{0, 0}
	e := &Editor{p, b, l}
	return &SimpleEventHandler{e, 0, ""}
}

func (eh *SimpleEventHandler) Handle(e *termbox.Event) error {

	if e.Type == termbox.EventKey {
		if e.Ch == 0 {
			switch e.Key {
			case termbox.KeySpace:
				eh.content = string(append([]rune(eh.content), ' '))
				eh.e.b.Insert(' ')
			case termbox.KeyBackspace2:
				eh.e.b.DeleteBack()
			case termbox.KeyDelete:
				eh.e.b.DeleteForward()
			case termbox.KeyArrowLeft:
				eh.e.b.BackOne()
			case termbox.KeyEnter:
				eh.e.b.Return()
			case termbox.KeyArrowRight:
				eh.e.b.ForwardOne()
			case termbox.KeyArrowUp:
				eh.e.b.UpOne()
			case termbox.KeyArrowDown:
				eh.e.b.DownOne()
			case termbox.KeyF1:
				eh.e.b.ScrollUp()
			case termbox.KeyF2:
				eh.e.b.ScrollDown()
			case termbox.KeyF3:
				eh.e.b.ScrollTop()
			default:
				b := eh.e.b
				report := fmt.Sprintf("<key: %#d>", uint(e.Key))
				for _, ch := range report {
					b.Insert(rune(ch))
				}
			}
		} else {
			eh.content = string(append([]rune(eh.content), e.Ch))
			eh.e.b.Insert(e.Ch)
		}
	}

	w, h := termbox.Size()
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	eh.count += 1
	_, _ = w, h

	eh.e.b.PutAll(eh.e.p)
	termbox.Flush()
	return nil
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

	county := []rune(fmt.Sprintf("─┤ %v ", content))
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

	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)
	w, h := termbox.Size()
	eh := NewSimpleEventHandler(10, 10, w, h)
	eh.Handle(&termbox.Event{})

	for {
		ev := termbox.PollEvent()
		if ev.Type == termbox.EventKey && ev.Key == termbox.KeyCtrlX {
			return
		}
		eh.Handle(&ev)
	}
}
