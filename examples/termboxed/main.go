package main

import (
	"fmt"
	"os"
	"strings"
)

import "github.com/nsf/termbox-go"
import "github.com/ehedgehog/guineapig/examples/termboxed/bounds"
import "github.com/ehedgehog/guineapig/examples/termboxed/buffer"
import "github.com/ehedgehog/guineapig/examples/termboxed/draw"
import "github.com/ehedgehog/guineapig/examples/termboxed/screen"

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

type Loc struct {
	X, Y int
}

type EditorPanel struct {
	topBar         screen.Canvas
	bottomBar      screen.Canvas
	leftBar        screen.Canvas
	rightBar       screen.Canvas
	textBox        screen.Canvas
	mainBuffer     buffer.Type
	lineBuffer     buffer.Type
	focusBuffer    *buffer.Type
	verticalOffset int
	where          Loc
}

func (ep *EditorPanel) Geometry() Geometry {
	minw := 2
	maxw := 1000
	minh := 2
	maxh := 1000
	return Geometry{minWidth: minw, maxWidth: maxw, minHeight: minh, maxHeight: maxh}
}

func readIntoBuffer(b buffer.Type, fileName string) {
	f, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	b.ReadFrom(f)
}

var commands = map[string]func(buffer.Type, []string){
	"r": func(b buffer.Type, blobs []string) {
		readIntoBuffer(b, blobs[1])
	},
}

func NewEditorPanel() EventHandler {
	mb := buffer.New(func(b buffer.Type, s string) {}, 0, 0)
	ep := &EditorPanel{
		mainBuffer: mb, // buffer.New(func(b buffer.Type, s string) {}, 0, 0),
		lineBuffer: buffer.New(func(b buffer.Type, s string) {
			line, content := b.Expose()
			blobs := strings.Split(content[line], " ")
			// c := screen.NewSubCanvas(screen.NewTermboxCanvas(), 40, 40, 40, 40)
			command := commands[blobs[0]]
			if command == nil {
			} else {
				command(mb, blobs)
			}
			// screen.PutString(c, 0, 0, "-- "+blobs[0]+" --", screen.DefaultStyle)
		}, 0, 0),
		where: Loc{0, 0},
	}
	ep.focusBuffer = &ep.mainBuffer
	return ep
}

func (ep *EditorPanel) New() EventHandler {
	return NewEditorPanel()
}

func (ep *EditorPanel) Key(e *termbox.Event) error {
	b := *ep.focusBuffer
	if e.Ch == 0 {
		switch e.Key {
		case 0:
			// nothing
		case termbox.KeyCtrlB:
			if ep.focusBuffer == &ep.mainBuffer {
				ep.focusBuffer = &ep.lineBuffer
			} else {
				ep.focusBuffer = &ep.mainBuffer
			}
		case termbox.KeySpace:
			b.Insert(' ')
		case termbox.KeyBackspace2:
			b.DeleteBack()
		case termbox.KeyDelete:
			b.DeleteForward()
		case termbox.KeyArrowLeft:
			b.BackOne()
		case termbox.KeyEnter:
			if ep.focusBuffer == &ep.mainBuffer {
				b.Return()
			} else {
				b.Execute()
			}
		case termbox.KeyArrowRight:
			b.ForwardOne()
		case termbox.KeyArrowUp:
			b.UpOne()
		case termbox.KeyArrowDown:
			b.DownOne()
		default:
			report := fmt.Sprintf("<key: %#d>\n", uint(e.Key))
			for _, ch := range report {
				b.Insert(rune(ch))
			}
		}
	} else {
		b.Insert(e.Ch)
	}
	return nil
}

func (ep *EditorPanel) Mouse(e *termbox.Event) error {
	x, y := e.MouseX, e.MouseY
	w, h := ep.textBox.Size()
	if 0 < x && x < w+1 && 0 < y && y < h+1 {
		ep.mainBuffer.SetWhere(x-1, y-1)
		ep.focusBuffer = &ep.mainBuffer
	} else if x >= delta && y == 0 {
		ep.lineBuffer.SetWhere(x-delta, 0)
		ep.focusBuffer = &ep.lineBuffer
	}
	return nil
}

func (ep *EditorPanel) AdjustScrolling() {
	line, _ := ep.mainBuffer.Expose()
	_, h := ep.textBox.Size()
	if line < ep.verticalOffset {
		ep.verticalOffset = line
	}
	if line > ep.verticalOffset+h-1 {
		ep.verticalOffset = line - h + 1
	}
}

func (ep *EditorPanel) Paint() error {
	ep.AdjustScrolling()
	w, _ := ep.bottomBar.Size()
	line, content := ep.mainBuffer.Expose()
	_, textHeight := ep.textBox.Size()
	ep.mainBuffer.PutLines(ep.textBox, ep.verticalOffset, textHeight)
	//
	ep.bottomBar.SetCell(0, 0, draw.Glyph_corner_bl, screen.DefaultStyle)
	for i := 1; i < w; i += 1 {
		ep.bottomBar.SetCell(i, 0, draw.Glyph_hbar, screen.DefaultStyle)
	}
	ep.bottomBar.SetCell(w-1, 0, draw.Glyph_corner_br, screen.DefaultStyle)
	//
	_, lh := ep.leftBar.Size()
	for j := 0; j < lh; j += 1 {
		ep.leftBar.SetCell(0, j, draw.Glyph_vbar, screen.DefaultStyle)
	}
	//
	ep.topBar.SetCell(0, 0, draw.Glyph_corner_tl, screen.DefaultStyle)
	for i := 1; i < w; i += 1 {
		ep.topBar.SetCell(i, 0, draw.Glyph_hbar, screen.DefaultStyle)
	}
	screen.PutString(ep.topBar, 2, 0, "─┤ ", screen.DefaultStyle)
	ep.topBar.SetCell(w-1, 0, draw.Glyph_corner_tr, screen.DefaultStyle)
	//
	// HACK -- shouldn't need to remake each time
	tline, _ := ep.lineBuffer.Expose()
	ep.lineBuffer.PutLines(screen.NewSubCanvas(ep.topBar, delta, 0, w-delta-2, 1), tline, 1)
	//
	length := bounds.Max(line, len(content))
	draw.Scrollbar(ep.rightBar, draw.ScrollInfo{length, line})
	//
	return nil
}

const delta = 5

func (eh *EditorPanel) ResizeTo(outer screen.Canvas) error {
	w, h := outer.Size()
	eh.leftBar = screen.NewSubCanvas(outer, 0, 1, 1, h-2)
	eh.rightBar = screen.NewSubCanvas(outer, w-1, 1, 1, h-2)
	eh.topBar = screen.NewSubCanvas(outer, 0, 0, w, 1)
	eh.bottomBar = screen.NewSubCanvas(outer, 0, h-1, w, 1)
	eh.textBox = screen.NewSubCanvas(outer, 1, 1, w-2, h-2)
	return nil
}

func (ep *EditorPanel) SetCursor() error {
	if ep.focusBuffer == &ep.mainBuffer {
		x, y := ep.mainBuffer.Where()
		ep.textBox.SetCursor(x, y)
	} else {
		x, _ := ep.lineBuffer.Where()
		ep.topBar.SetCursor(x+delta, 0)
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
	w, h := outer.Size()
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
	w, h := outer.Size()
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
	w, h := outer.Size()
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

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)

	page := screen.NewTermboxCanvas()

	edA := NewStack(NewEditorPanel, NewEditorPanel())
	// edB := NewStack(NewEditorPanel, NewEditorPanel())
	//	eh := NewSideBySide(edA, edB)

	eh := NewShelf(func() EventHandler {
		return NewStack(NewEditorPanel, NewEditorPanel())
	},
		edA)

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
