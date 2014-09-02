package main

import "fmt"
import "github.com/nsf/termbox-go"
import "github.com/ehedgehog/guineapig/examples/termboxed/bounds"
import "github.com/ehedgehog/guineapig/examples/termboxed/buffer"
import "github.com/ehedgehog/guineapig/examples/termboxed/draw"
import "github.com/ehedgehog/guineapig/examples/termboxed/screen"

type EventHandler interface {
	Key(e *termbox.Event) error
	Mouse(e *termbox.Event) error
	ResizeTo(outer screen.Canvas) error
	Paint() error
	SetCursor() error
}

type Loc struct {
	X, Y int
}

type EditorPanel struct {
	topBar    screen.Canvas
	bottomBar screen.Canvas
	leftBar   screen.Canvas
	rightBar  screen.Canvas
	textBox   screen.Canvas
	buffer    buffer.Type
	where     Loc
}

func NewEditorPanel() EventHandler {
	return &EditorPanel{
		buffer: buffer.New(0, 0),
		where:  Loc{0, 0},
	}
}

func (ep *EditorPanel) Key(e *termbox.Event) error {
	buffer := ep.buffer
	if e.Ch == 0 {
		switch e.Key {
		case 0:
			// nothing
		case termbox.KeySpace:
			buffer.Insert(' ')
		case termbox.KeyBackspace2:
			buffer.DeleteBack()
		case termbox.KeyDelete:
			buffer.DeleteForward()
		case termbox.KeyArrowLeft:
			buffer.BackOne()
		case termbox.KeyEnter:
			buffer.Return()
		case termbox.KeyArrowRight:
			buffer.ForwardOne()
		case termbox.KeyArrowUp:
			buffer.UpOne()
		case termbox.KeyArrowDown:
			buffer.DownOne()
		case termbox.KeyF1:
			buffer.ScrollUp()
		case termbox.KeyF2:
			buffer.ScrollDown()
		case termbox.KeyF3:
			buffer.ScrollTop()
		default:
			b := buffer
			report := fmt.Sprintf("<key: %#d>\n", uint(e.Key))
			for _, ch := range report {
				b.Insert(rune(ch))
			}
		}
	} else {
		buffer.Insert(e.Ch)
	}
	return nil
}

func (ep *EditorPanel) Mouse(e *termbox.Event) error {
	x, y := e.MouseX, e.MouseY
	ep.buffer.SetWhere(x-1, y-1)
	return nil
}

func (ep *EditorPanel) Paint() error {
	w, _ := ep.bottomBar.Size()
	ep.buffer.PutAll(ep.textBox)
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
	line, content := ep.buffer.Expose()
	_, sh := ep.rightBar.Size()
	size := draw.WH{1, sh}
	length := bounds.Max(line, len(content))
	off := draw.Scrolling{length, line}
	info := draw.BoxInfo{draw.XY{0, 0}, size, off}
	draw.Scrollbar(ep.rightBar, info)
	//
	return nil
}

func (eh *EditorPanel) ResizeTo(outer screen.Canvas) error {
	w, h := outer.Size()
	eh.leftBar = screen.NewSubCanvas(outer, 0, 1, 1, h-2)
	eh.rightBar = screen.NewSubCanvas(outer, w-2, 1, 1, h-2)
	eh.topBar = screen.NewSubCanvas(outer, 0, 0, 1, h)
	eh.bottomBar = screen.NewSubCanvas(outer, 0, h-1, w, 1)
	eh.textBox = screen.NewSubCanvas(outer, 1, 1, w-2, h-2)
	return nil
}

func (ep *EditorPanel) SetCursor() error {
	x, y := ep.buffer.Where()
	ep.textBox.SetCursor(x, y)
	return nil
}

type SideBySide struct {
	widthA int
	Focus  EventHandler
	A, B   EventHandler
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

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)

	page := screen.NewTermboxCanvas()

	edA := NewEditorPanel()
	edB := NewEditorPanel()
	eh := NewSideBySide(edA, edB)

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
