package main

import (
	"errors"
	"fmt"
	// 	"log"
	"os"
	//	"strconv"
	"strings"
)

import termbox "github.com/limetext/termbox-go"

import "github.com/ehedgehog/guineapig/examples/termboxed/bounds"
import "github.com/ehedgehog/guineapig/examples/termboxed/buffer"
import "github.com/ehedgehog/guineapig/examples/termboxed/draw"
import "github.com/ehedgehog/guineapig/examples/termboxed/screen"
import "github.com/ehedgehog/guineapig/examples/termboxed/grid"

type Geometry struct {
	minWidth  int
	maxWidth  int
	minHeight int
	maxHeight int
}

type EventHandler interface {
	Key(e *termbox.Event) error
	Mouse(e *termbox.Event) error
	ResizeTo(outer screen.Canvas) error
	Paint() error
	SetCursor() error
	Geometry() Geometry
	New() EventHandler
}

type Focus struct {
	where  *grid.LineCol
	buffer *buffer.Type
}

type EditorPanel struct {
	firstMarkedLine int
	lastMarkedLine  int

	topBar    screen.Canvas
	bottomBar screen.Canvas
	leftBar   screen.Canvas
	rightBar  screen.Canvas
	textBox   screen.Canvas

	mainBuffer buffer.Type
	lineBuffer buffer.Type

	focus Focus
	// focusBuffer *buffer.Type

	verticalOffset int
	where          grid.LineCol
	otherWhere     grid.LineCol
}

func (ep *EditorPanel) Geometry() Geometry {
	minw := 2
	maxw := 1000
	minh := 2
	maxh := 1000
	return Geometry{minWidth: minw, maxWidth: maxw, minHeight: minh, maxHeight: maxh}
}

func readIntoBuffer(ep *EditorPanel, b buffer.Type, fileName string) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	w, err := b.ReadFromFile(ep.where, fileName, f)
	ep.where = w
	return err
}

var commands = map[string]func(*EditorPanel, []string) error{
	"r": func(ep *EditorPanel, blobs []string) error {
		b := ep.mainBuffer
		return readIntoBuffer(ep, b, blobs[1])
	},
	"w": func(ep *EditorPanel, blobs []string) error {
		b := ep.mainBuffer
		return b.WriteToFile(blobs[1:])
	},
	"d": func(ep *EditorPanel, blobs []string) error {
		b := ep.mainBuffer
		lineNumber := ep.where.Line
		b.DeleteLine(ep.where)
		if ep.firstMarkedLine > 0 {
			first, last := ep.firstMarkedLine-1, ep.lastMarkedLine-1
			if lineNumber <= last {
				ep.lastMarkedLine -= 1
				if lineNumber < first {
					ep.firstMarkedLine -= 1
				}
			}
		}
		return nil
	},
	"dr": func(ep *EditorPanel, blobs []string) error {
		if ep.firstMarkedLine > 0 {
			b := ep.mainBuffer
			ep.where = b.DeleteLines(ep.where, ep.firstMarkedLine-1, ep.lastMarkedLine-1)
			ep.firstMarkedLine, ep.lastMarkedLine = 0, 0
			return nil
		} else {
			return errors.New("no marked range")
		}
	},
}

func NewEditorPanel() EventHandler {
	mb := buffer.New(func(b buffer.Type, s string) error { return nil })
	var ep *EditorPanel
	ep = &EditorPanel{
		mainBuffer: mb, // buffer.New(func(b buffer.Type, s string) {}, 0, 0),
		lineBuffer: buffer.New(func(b buffer.Type, s string) error {
			content := b.Expose()
			line := ep.where.Line
			blobs := strings.Split(content[line], " ")
			// c := screen.NewSubCanvas(screen.NewTermboxCanvas(), 40, 40, 40, 40)
			command := commands[blobs[0]]
			if command == nil {
				return errors.New("not a command: " + blobs[0])
			} else {
				return command(ep, blobs)
			}
			// screen.PutString(c, 0, 0, "-- "+blobs[0]+" --", screen.DefaultStyle)
		}),
		// where: grid.LineCol{Line: 0, Col: 0},
	}
	ep.focus = Focus{buffer: &ep.mainBuffer, where: &ep.otherWhere}
	return ep
}

func (ep *EditorPanel) New() EventHandler {
	return NewEditorPanel()
}

func (ep *EditorPanel) Key(e *termbox.Event) error {
	b := *ep.focus.buffer
	if e.Ch == 0 {
		switch e.Key {

		case 0:
			// nothing

		case termbox.KeyF1:
			ep.focus = Focus{buffer: &ep.lineBuffer, where: &ep.otherWhere}
			// ep.otherWhere = ep.lineBuffer.Return(ep.otherWhere)

		case termbox.KeyF2:
			ep.otherWhere, _ = ep.lineBuffer.Execute(ep.otherWhere)

		case termbox.KeyCtrlB:
			if ep.focus.buffer == &ep.mainBuffer {
				ep.focus = Focus{buffer: &ep.lineBuffer, where: &ep.otherWhere}
			} else {
				ep.focus = Focus{buffer: &ep.mainBuffer, where: &ep.where}
			}

		case termbox.KeySpace:
			b.Insert(*ep.focus.where, ' ')
			ep.focus.where.RightOne()

		case termbox.KeyBackspace2:
			*ep.focus.where = b.DeleteBack(*ep.focus.where)

		case termbox.KeyDelete:
			*ep.focus.where = b.DeleteForward(*ep.focus.where)

		case termbox.KeyF3:
			where := ep.focus.where
			ep.firstMarkedLine = where.Line + 1
			if ep.lastMarkedLine < ep.firstMarkedLine {
				ep.lastMarkedLine = ep.firstMarkedLine
			}

		case termbox.KeyF4:
			where := ep.focus.where
			ep.lastMarkedLine = where.Line + 1
			if ep.firstMarkedLine > ep.lastMarkedLine {
				ep.firstMarkedLine = ep.lastMarkedLine
			}

		case termbox.KeyPgup:
			where := ep.where
			vo := ep.verticalOffset
			if where.Line-vo == 0 {
				top := bounds.Max(0, where.Line-ep.textBox.Size().Height)
				ep.where = grid.LineCol{top, where.Col}
			} else {
				ep.where = grid.LineCol{vo, where.Col}
			}

		case termbox.KeyPgdn:
			where := ep.where
			vo := ep.verticalOffset
			height := ep.textBox.Size().Height
			if where.Line-vo == height-1 {
				// forward one page
				bot := where.Line + height
				ep.where = grid.LineCol{bot, where.Col}
			} else {
				// bottom of this page
				ep.where = grid.LineCol{vo + height - 1, where.Col}
			}

		case termbox.KeyEnd:
			where := ep.focus.where
			if where.Col == 0 {
				contents := b.Expose()
				line := contents[ep.focus.where.Line]
				where.Col = len(line)
			} else {
				where.Col = 0
			}
			ep.focus.where = where

		case termbox.KeyEnter:
			if ep.focus.buffer == &ep.mainBuffer {
				ep.where = b.Return(ep.where)
				if ep.firstMarkedLine > 0 {
					lineNumber := ep.focus.where.Line
					first, last := ep.firstMarkedLine+1, ep.lastMarkedLine+1
					if lineNumber < last {
						ep.lastMarkedLine += 1
					}
					if lineNumber < first-1 {
						ep.firstMarkedLine += 1
					}
					report(ep, ep.lineBuffer, "zingo")
				}

			} else {
				_, err := b.Execute(*ep.focus.where)
				if err == nil {
					report(ep, b, "OK")
				} else {
					report(ep, b, err.Error())
				}
				ep.focus = Focus{buffer: &ep.mainBuffer, where: &ep.where}
			}

		case termbox.KeyArrowRight:
			ep.focus.where.RightOne()

		case termbox.KeyArrowUp:
			ep.focus.where.UpOne()

		case termbox.KeyArrowDown:
			ep.focus.where.DownOne()

		case termbox.KeyArrowLeft:
			ep.focus.where.LeftOne()

		default:
			report := fmt.Sprintf("<key: %#d>\n", uint(e.Key))
			for _, ch := range report {
				b.Insert(*ep.focus.where, rune(ch))
				ep.focus.where.RightOne()
			}
		}
	} else {
		b.Insert(*ep.focus.where, e.Ch)
		ep.focus.where.RightOne()
	}
	return nil
}

func report(ep *EditorPanel, b buffer.Type, message string) {
	b.Insert(*ep.focus.where, ' ')
	ep.focus.where.RightOne()
	b.Insert(*ep.focus.where, '(')
	ep.focus.where.RightOne()
	for _, rune := range message {
		b.Insert(*ep.focus.where, rune)
		ep.focus.where.RightOne()
	}
	b.Insert(*ep.focus.where, ')')
	ep.focus.where.RightOne()
	b.Insert(*ep.focus.where, ' ')
	ep.focus.where.RightOne()
}

func (ep *EditorPanel) Mouse(e *termbox.Event) error {
	x, y := e.MouseX, e.MouseY
	size := ep.textBox.Size()
	w, h := size.Width, size.Height
	if 0 < x && x < w+1 && 0 < y && y < h+1 {
		ep.where = grid.LineCol{y - 1, x - 1}
		ep.focus = Focus{buffer: &ep.mainBuffer, where: &ep.where}
	} else if x >= delta && y == 0 {
		ep.otherWhere = grid.LineCol{0, x - delta}
		ep.focus = Focus{where: &ep.otherWhere, buffer: &ep.lineBuffer}
	}
	return nil
}

func (ep *EditorPanel) AdjustScrolling() {
	size := ep.textBox.Size()
	line := ep.where.Line
	h := size.Height
	if line < ep.verticalOffset {
		ep.verticalOffset = line
	}
	if line > ep.verticalOffset+h-1 {
		ep.verticalOffset = line - h + 1
	}
}

func (ep *EditorPanel) Paint() error {
	ep.AdjustScrolling()
	bottomSize := ep.bottomBar.Size()
	w := bottomSize.Width
	content := ep.mainBuffer.Expose()
	line := ep.where.Line
	textBoxSize := ep.textBox.Size()
	textHeight := textBoxSize.Height
	ep.mainBuffer.PutLines(ep.textBox, ep.verticalOffset, textHeight)
	//
	ep.bottomBar.SetCell(grid.LineCol{Col: 0, Line: 0}, draw.Glyph_corner_bl, screen.DefaultStyle)
	for i := 1; i < w; i += 1 {
		ep.bottomBar.SetCell(grid.LineCol{Col: i, Line: 0}, draw.Glyph_hbar, screen.DefaultStyle)
	}
	ep.bottomBar.SetCell(grid.LineCol{Col: w - 1, Line: 0}, draw.Glyph_corner_br, screen.DefaultStyle)
	//
	leftBarSize := ep.leftBar.Size()
	lh := leftBarSize.Height
	for j := 0; j < lh; j += 1 {
		ep.leftBar.SetCell(grid.LineCol{Col: 0, Line: j}, draw.Glyph_vbar, screen.DefaultStyle)
	}
	//
	ep.topBar.SetCell(grid.LineCol{Col: 0, Line: 0}, draw.Glyph_corner_tl, screen.DefaultStyle)
	for i := 1; i < w; i += 1 {
		ep.topBar.SetCell(grid.LineCol{Col: i, Line: 0}, draw.Glyph_hbar, screen.DefaultStyle)
	}
	screen.PutString(ep.topBar, 2, 0, "─┤ ", screen.DefaultStyle)
	ep.topBar.SetCell(grid.LineCol{Col: w - 1, Line: 0}, draw.Glyph_corner_tr, screen.DefaultStyle)
	//
	// HACK -- shouldn't need to remake each time
	tline := ep.otherWhere.Line
	ep.lineBuffer.PutLines(screen.NewSubCanvas(ep.topBar, delta, 0, w-delta-2, 1), tline, 1)
	//
	length := bounds.Max(line, len(content))
	draw.Scrollbar(ep.rightBar, draw.ScrollInfo{length, line})
	//
	return nil
}

const delta = 5

func (eh *EditorPanel) ResizeTo(outer screen.Canvas) error {
	size := outer.Size()
	w, h := size.Width, size.Height
	eh.leftBar = screen.NewSubCanvas(outer, 0, 1, 1, h-2)
	eh.rightBar = screen.NewSubCanvas(outer, w-1, 1, 1, h-2)
	eh.topBar = screen.NewSubCanvas(outer, 0, 0, w, 1)
	eh.bottomBar = screen.NewSubCanvas(outer, 0, h-1, w, 1)
	// eh.textBox = screen.NewSubCanvas(outer, 1, 1, w-2, h-2)
	eh.textBox = NewTextBox(eh, outer, 1, 1, w-2, h-2)
	return nil
}

const tryTagSize = 6

func NewTextBox(ep *EditorPanel, outer screen.Canvas, dx, dy, w, h int) screen.Canvas {
	sub := screen.NewSubCanvas(outer, dx, dy, w, h)
	return &TextBox{tagSize: tryTagSize, ep: ep, SubCanvas: *sub.(*screen.SubCanvas)}
}

type TextBox struct {
	tagSize int
	ep      *EditorPanel
	screen.SubCanvas
}

var markStyle = screen.MakeStyle(termbox.ColorDefault, termbox.ColorYellow)

func (t *TextBox) SetCell(where grid.LineCol, ch rune, s screen.Style) {
	if where.Col == 0 {

		// log.Println("range:", t.ep.firstMarkedLine, "to", t.ep.lastMarkedLine)
		ep := t.ep
		if ep.firstMarkedLine-1-ep.verticalOffset <= where.Line && where.Line <= ep.lastMarkedLine-1-ep.verticalOffset {
			t.SubCanvas.SetCell(grid.LineCol{where.Line, tryTagSize - 1}, ' ', markStyle)
		}
		s := fmt.Sprintf("%4v", where.Line+ep.verticalOffset)
		for i, ch := range s {
			t.SubCanvas.SetCell(grid.LineCol{where.Line, i}, ch, screen.DefaultStyle)
		}
		//	for i := 0; i < t.tagSize; i += 1 {
		//		t.SubCanvas.SetCell(grid.LineCol{where.Line, i}, '_', s)
		//	}
	}
	t.SubCanvas.SetCell(where.ColPlus(t.tagSize), ch, s)
}

func (t *TextBox) SetCursor(where grid.LineCol) {
	t.SubCanvas.SetCursor(where.ColPlus(t.tagSize))
}

func (ep *EditorPanel) SetCursor() error {
	if ep.focus.buffer == &ep.mainBuffer {
		where := ep.where.LineMinus(ep.verticalOffset)
		ep.textBox.SetCursor(where) // (where.ColPlus(delta))
	} else {
		where := ep.otherWhere                                  // .LineMinus(ep.verticalOffset)
		ep.topBar.SetCursor(grid.LineCol{0, where.Col + delta}) // (where)
	}
	return nil
}

type Block struct {
	generator  func() EventHandler
	elements   []EventHandler
	bounds     []int
	focus      int
	recentSize screen.Canvas
}

type Stack struct {
	Block
}

func (b *Block) SetCursor() error {
	return b.elements[b.focus].SetCursor()
}

func (b *Block) Paint() error {
	for _, e := range b.elements {
		e.Paint()
	}
	return nil
}

func (b *Shelf) Key(e *termbox.Event) error {
	if e.Ch == 0 && e.Key == termbox.KeyCtrlT {
		b.elements = append(b.elements, b.generator())
		b.bounds = append(b.bounds, 0)
		b.ResizeTo(b.recentSize)
		return nil
	}
	return b.elements[b.focus].Key(e)
}

func (b *Stack) Key(e *termbox.Event) error {
	if e.Ch == 0 && e.Key == termbox.KeyCtrlU {
		b.elements = append(b.elements, b.generator())
		b.bounds = append(b.bounds, 0)
		b.ResizeTo(b.recentSize)
		return nil
	}
	return b.elements[b.focus].Key(e)
}

func NewStack(generator func() EventHandler, elements ...EventHandler) EventHandler {
	return &Stack{
		Block: Block{
			focus:     0,
			elements:  elements,
			generator: generator,
			bounds:    make([]int, len(elements)),
		},
	}
}

func (s *Stack) New() EventHandler {
	return NewStack(s.generator)
}

func (s *Stack) Geometry() Geometry {
	minw, maxw, minh, maxh := 0, 0, 0, 0
	for _, eh := range s.elements {
		g := eh.Geometry()
		minw = bounds.Max(minw, g.minWidth)
		maxw = bounds.Max(maxw, g.maxWidth)
		minh = minh + g.minHeight
		maxh = maxh + g.maxHeight
	}
	return Geometry{minWidth: minw, maxWidth: maxw, minHeight: minh, maxHeight: maxh}
}

func (s *Stack) Mouse(e *termbox.Event) error {
	y := 0
	for i, h := range s.bounds {
		nextY := y + h
		if e.MouseY < nextY {
			e.MouseY -= y
			s.focus = i
			return s.elements[i].Mouse(e)
		}
		y = nextY
	}
	panic("stack Mouse")
}

func (s *Stack) ResizeTo(outer screen.Canvas) error {
	g := s.Geometry()
	size := outer.Size()
	w, h := size.Width, size.Height
	count := 0
	for _, eh := range s.elements {
		g := eh.Geometry()
		if g.minHeight != g.maxHeight {
			count += 1
		}
	}
	totalSpare := h - g.minHeight
	spare := totalSpare / count
	y := 0
	for i, eh := range s.elements {
		g := eh.Geometry()
		if g.minHeight == g.maxHeight {
			h := g.minHeight
			s.bounds[i] = h
			c := screen.NewSubCanvas(outer, 0, y, w, h)
			eh.ResizeTo(c)
			y += h
		} else {
			h := g.minHeight + spare
			s.bounds[i] = h
			c := screen.NewSubCanvas(outer, 0, y, w, h)
			eh.ResizeTo(c)
			y += h
		}
	}
	s.recentSize = outer
	return nil
}

type Shelf struct {
	Block
}

func NewShelf(generator func() EventHandler, elements ...EventHandler) EventHandler {
	return &Shelf{
		Block: Block{
			focus:     0,
			elements:  elements,
			generator: generator,
			bounds:    make([]int, len(elements)),
		},
	}
}

func (s *Shelf) New() EventHandler {
	return NewShelf(s.generator)
}

func (s *Shelf) Geometry() Geometry {
	minw, maxw, minh, maxh := 0, 0, 0, 0
	for _, eh := range s.elements {
		g := eh.Geometry()
		minh = bounds.Max(minh, g.minHeight)
		maxh = bounds.Max(maxh, g.maxHeight)
		minw = minw + g.minWidth
		maxw = maxw + g.maxWidth
	}
	return Geometry{minWidth: minw, maxWidth: maxw, minHeight: minh, maxHeight: maxh}
}

func (s *Shelf) Mouse(e *termbox.Event) error {
	x := 0
	for i, w := range s.bounds {
		nextX := x + w
		if e.MouseX < nextX {
			e.MouseX -= x
			s.focus = i
			return s.elements[i].Mouse(e)
		}
		x = nextX
	}
	panic("shelf Mouse")
}

func (s *Shelf) ResizeTo(outer screen.Canvas) error {
	g := s.Geometry()
	size := outer.Size()
	w, h := size.Width, size.Height
	count := 0
	for _, eh := range s.elements {
		g := eh.Geometry()
		if g.minWidth != g.maxWidth {
			count += 1
		}
	}
	totalSpare := w - g.minWidth
	spare := totalSpare / count
	x := 0
	for i, eh := range s.elements {
		g := eh.Geometry()
		if g.minWidth == g.maxWidth {
			w := g.minWidth
			s.bounds[i] = w
			c := screen.NewSubCanvas(outer, x, 0, w, h)
			eh.ResizeTo(c)
			x += w
		} else {
			w := g.minWidth + spare
			s.bounds[i] = w
			c := screen.NewSubCanvas(outer, x, 0, w, h)
			eh.ResizeTo(c)
			x += w
		}
	}
	s.recentSize = outer
	return nil
}

type SideBySide struct {
	widthA int
	Focus  EventHandler
	A, B   EventHandler
}

func (s *SideBySide) Geometry() Geometry {
	ga, gb := s.A.Geometry(), s.B.Geometry()
	minw := ga.minWidth + gb.minWidth
	maxw := ga.maxWidth + gb.maxWidth
	minh := bounds.Max(ga.minHeight, gb.minHeight)
	maxh := bounds.Max(ga.maxHeight, gb.maxHeight)
	return Geometry{minWidth: minw, maxWidth: maxw, minHeight: minh, maxHeight: maxh}
}

func (s *SideBySide) Key(e *termbox.Event) error {
	if e.Key == termbox.KeyCtrlA {
		if s.Focus == s.A {
			s.Focus = s.B
		} else {
			s.Focus = s.A
		}
	} else {
		s.Focus.Key(e)
	}
	return nil
}

func (s *SideBySide) Mouse(e *termbox.Event) error {
	x := e.MouseX
	if x > s.widthA {
		s.Focus = s.B
		e.MouseX -= s.widthA
	} else {
		s.Focus = s.A
	}
	s.Focus.Mouse(e)
	return nil
}

func (s *SideBySide) ResizeTo(outer screen.Canvas) error {
	size := outer.Size()
	w, h := size.Width, size.Height
	aw := w / 2
	bw := w - aw
	s.widthA = aw
	s.A.ResizeTo(screen.NewSubCanvas(outer, 0, 0, aw, h))
	s.B.ResizeTo(screen.NewSubCanvas(outer, aw, 0, bw, h))
	return nil
}

func (s *SideBySide) Paint() error {
	s.A.Paint()
	s.B.Paint()
	return nil
}

func (s *SideBySide) SetCursor() error {
	return s.Focus.SetCursor()
}

func NewSideBySide(A, B EventHandler) EventHandler {
	return &SideBySide{0, A, A, B}
}

func (s *SideBySide) New() EventHandler {
	panic("SideBySide.New")
}

func makeColours() []termbox.RGB {
	result := termbox.Palette256 // make([]termbox.RGB, 256)
	result[termbox.ColorBlack] = termbox.RGB{0, 0, 0}
	result[termbox.ColorRed] = termbox.RGB{255, 0, 0}
	result[termbox.ColorGreen] = termbox.RGB{0, 255, 0}
	result[termbox.ColorYellow] = termbox.RGB{255, 255, 0}
	result[termbox.ColorBlue] = termbox.RGB{0, 0, 255}
	result[termbox.ColorMagenta] = termbox.RGB{255, 0, 255}
	result[termbox.ColorCyan] = termbox.RGB{0, 255, 255}
	result[termbox.ColorWhite] = termbox.RGB{255, 255, 255}
	return result
}

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	termbox.SetColorMode(termbox.ColorMode256)
	termbox.SetColorPalette(makeColours())
	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)

	page := screen.NewTermboxCanvas()

	edA := NewStack(NewEditorPanel, NewEditorPanel())
	// edB := NewStack(NewEditorPanel, NewEditorPanel())
	//	eh := NewSideBySide(edA, edB)

	eh := NewShelf(func() EventHandler { return NewStack(NewEditorPanel, NewEditorPanel()) }, edA)

	eh.ResizeTo(page)

	for {
		eh.Paint()
		eh.SetCursor()
		termbox.Flush()
		ev := termbox.PollEvent()
		if ev.Type == termbox.EventKey && ev.Key == termbox.KeyCtrlX {
			return
		}
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		if ev.Type == termbox.EventMouse {
			eh.Mouse(&ev)
		}
		if ev.Type == termbox.EventKey {
			eh.Key(&ev)
		}
		if ev.Type == termbox.EventResize {
			page = screen.NewTermboxCanvas()
			eh.ResizeTo(page)
		}
	}
}
