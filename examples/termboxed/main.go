package main

import "github.com/nsf/termbox-go"
import "github.com/ehedgehog/guineapig/examples/termboxed/buffer"
import "github.com/ehedgehog/guineapig/examples/termboxed/screen"
import "fmt"

// import "github.com/limetext/termbox-go"
// import "log"
// import _ "github.com/ehedgehog/guineapig/examples/termboxed/panel"

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

type Editor struct {
	p screen.Panel
	b buffer.Type
	l Loc
}

type EditorPanel struct {
	panel  screen.Canvas
	buffer buffer.Type
	where  Loc
}

func NewEditorPanel() EventHandler {
	ws := screen.Canvas(nil)
	return &EditorPanel{ws, buffer.New(0, 0), Loc{0, 0}}
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
	panic("mouse")
	return nil
}

func (ep *EditorPanel) Paint() error {
	ep.buffer.PutAll(ep.panel)
	return nil
}

func (eh *EditorPanel) ResizeTo(outer screen.Canvas) error {
	eh.panel = outer
	return nil
}

func (ep *EditorPanel) SetCursor() error {
	x, y := ep.buffer.Where()
	ep.panel.SetCursor(x, y)
	return nil
}

type SideBySide struct {
	Focus EventHandler
	A, B  EventHandler
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
	return nil
}

func (s *SideBySide) ResizeTo(outer screen.Canvas) error {
	w, h := outer.Size()
	aw := w / 2
	bw := w - aw
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
	return &SideBySide{A, A, B}
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
