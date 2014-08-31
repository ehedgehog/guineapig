package main

import "github.com/nsf/termbox-go"
import "github.com/ehedgehog/guineapig/examples/termboxed/buffer"
import "github.com/ehedgehog/guineapig/examples/termboxed/screen"
import "fmt"

// import "github.com/limetext/termbox-go"
// import "log"
// import _ "github.com/ehedgehog/guineapig/examples/termboxed/panel"

type SimplePanel struct {
	outer screen.Panel
	x, y  int
	w, h  int
}

func (s *SimplePanel) PutString(x, y int, content string) {
	s.outer.PutString(s.x+x, s.y+y, content)
}

func (s *SimplePanel) Resize(x, y, w, h int) {
	s.x, s.y = x, y
	s.w, s.h = w, h
}

func (s *SimplePanel) SetCursor(x, y int) {
	termbox.SetCursor(x+s.x, y+s.y)
}

func (s *SimplePanel) Size() (w, h int) {
	return s.w, s.h
}

func (s *SimplePanel) SetCell(x, y int, ch rune, a, b termbox.Attribute) {
	s.outer.SetCell(x+s.x, y+s.y, ch, a, b)
}

func NewPanel(outer screen.Panel, x, y int, w, h int) screen.Panel {
	return &SimplePanel{outer, x, y, w, h}
}

type EventHandler interface {
	HandleKey(e *termbox.Event) error
	HandleMouse(e *termbox.Event) error
	// HandleResize(x, y, w, h int) error
	HandleResizeTo(outer screen.Canvas) error
	HandlePaint() error
	HandleSetCursor() error
}

type Loc struct {
	X, Y int
}

type Editor struct {
	p screen.Panel
	b buffer.Type
	l Loc
}

type EditorEventHandler struct {
	e *Editor
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

func (ep *EditorPanel) HandleKey(e *termbox.Event) error {
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

func (ep *EditorPanel) HandleMouse(e *termbox.Event) error {
	panic("mouse")
	return nil
}

func (ep *EditorPanel) HandlePaint() error {
	ep.buffer.PutAll(ep.panel)
	return nil
}

func (eh *EditorPanel) HandleResizeTo(outer screen.Canvas) error {
	eh.panel = outer
	return nil
}

func (ep *EditorPanel) HandleSetCursor() error {
	x, y := ep.buffer.Where()
	ep.panel.SetCursor(x, y)
	return nil
}

type SideBySide2 struct {
	Focus EventHandler
	A, B  EventHandler
}

func (s *SideBySide2) HandleKey(e *termbox.Event) error {
	if e.Key == termbox.KeyCtrlA {
		if s.Focus == s.A {
			s.Focus = s.B
		} else {
			s.Focus = s.A
		}
	} else {
		s.Focus.HandleKey(e)
	}
	return nil
}

func (s *SideBySide2) HandleMouse(e *termbox.Event) error {
	return nil
}

func (s *SideBySide2) HandleResizeTo(outer screen.Canvas) error {
	w, h := outer.Size()
	aw := w / 2
	bw := w - aw
	s.A.HandleResizeTo(screen.NewSubCanvas(outer, 0, 0, aw, h))
	s.B.HandleResizeTo(screen.NewSubCanvas(outer, aw, 0, bw, h))
	return nil
}

func (s *SideBySide2) HandlePaint() error {
	s.A.HandlePaint()
	s.B.HandlePaint()
	return nil
}

func (s *SideBySide2) HandleSetCursor() error {
	return s.Focus.HandleSetCursor()
}

func NewSideBySide2(A, B EventHandler) EventHandler {
	return &SideBySide2{A, A, B}
}

func NewEditorEventHandler(x, y int, w, h int) EventHandler {
	ws := screen.NewScreenWriteable(x, y, w, h)
	p := NewPanel(ws, x, y, w, h)
	b := buffer.New(w, h)
	l := Loc{0, 0}
	e := &Editor{p, b, l}
	return &EditorEventHandler{e}
}

func (eh *EditorEventHandler) HandleResizeTo(outer screen.Canvas) error {
	panic("old resize")
	// eh.e.p.Resize(x, y, w, h)
	return nil
}

func (eh *EditorEventHandler) HandlePaint() error {
	panic("old paint") // eh.e.b.PutAll(eh.e.p)
	return nil
}

func (eh *EditorEventHandler) HandleKey(e *termbox.Event) error {
	if e.Ch == 0 {
		switch e.Key {
		case 0:
			// nothing
		case termbox.KeySpace:
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
			report := fmt.Sprintf("<key: %#d>\n", uint(e.Key))
			for _, ch := range report {
				b.Insert(rune(ch))
			}
		}
	} else {
		eh.e.b.Insert(e.Ch)
	}
	return nil
}

func (eh *EditorEventHandler) HandleSetCursor() error {
	b := eh.e.b
	w := eh.e.p
	x, y := b.Where()
	w.SetCursor(x+1, y+1)
	return nil
}

func (eh *EditorEventHandler) HandleMouse(e *termbox.Event) error {
	b := eh.e.b
	report := fmt.Sprintf("<mouse: %v, %v>\n", e.MouseX, e.MouseY)
	for _, ch := range report {
		b.Insert(rune(ch))
	}
	return nil
}

//func Handle(eh EventHandler, e *termbox.Event) {
//	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
//	if e.Type == termbox.EventMouse {
//		eh.HandleMouse(e)
//	}
//	if e.Type == termbox.EventKey {
//		eh.HandleKey(e)
//	}
//	if e.Type == termbox.EventResize {
//		eh.HandleResize(0, 0, e.Width, e.Height)
//	}
//}

//type SideBySide struct {
//	Focus EventHandler
//	A, B  EventHandler
//}
//
//func (s *SideBySide) HandleKey(e *termbox.Event) error {
//	if e.Key == termbox.KeyCtrlA {
//		if s.Focus == s.A {
//			s.Focus = s.B
//		} else {
//			s.Focus = s.A
//		}
//	} else {
//		s.Focus.HandleKey(e)
//	}
//	return nil
//}
//
//func (s *SideBySide) HandleMouse(e *termbox.Event) error {
//	return nil
//}
//
//func (s *SideBySide) HandleResize(x, y, w, h int) error {
//	aw := w / 2
//	bw := w - aw
//	s.A.HandleResize(x, y, aw, h)
//	s.B.HandleResize(x+aw, y, bw, h)
//	return nil
//}
//
//func (s *SideBySide) HandlePaint() error {
//	s.A.HandlePaint()
//	s.B.HandlePaint()
//	return nil
//}
//
//func (s *SideBySide) HandleSetCursor() error {
//	return s.Focus.HandleSetCursor()
//}
//
//func NewSideBySide(A, B EventHandler) EventHandler {
//	return &SideBySide{A, A, B}
//}

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)

	// w, h := termbox.Size()
	page := screen.NewTermboxCanvas()

	edA := NewEditorPanel()
	edB := NewEditorPanel()
	ed := NewSideBySide2(edA, edB)

	// ehA := NewEditorEventHandler(0, 0, w/2, h)
	// ehB := NewEditorEventHandler(w-w/2, 0, w/2, h)
	// eh := NewSideBySide(ehA, ehB)
	eh := ed

	eh.HandleResizeTo(page)

	for {
		eh.HandlePaint()
		eh.HandleSetCursor()
		termbox.Flush()
		ev := termbox.PollEvent()
		if ev.Type == termbox.EventKey && ev.Key == termbox.KeyCtrlX {
			return
		}
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		if ev.Type == termbox.EventMouse {
			eh.HandleMouse(&ev)
		}
		if ev.Type == termbox.EventKey {
			eh.HandleKey(&ev)
		}
		if ev.Type == termbox.EventResize {
			// 	eh.HandleResize(0, 0, ev.Width, ev.Height)
			page = screen.NewTermboxCanvas()
			eh.HandleResizeTo(page)
		}
		_ = page
	}
}

//func OLDmain() {
//
//	err := termbox.Init()
//	if err != nil {
//		panic(err)
//	}
//	defer termbox.Close()
//
//	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)
//
//	w, h := termbox.Size()
//
//	ehA := NewEditorEventHandler(0, 0, w/2, h)
//	ehB := NewEditorEventHandler(w-w/2, 0, w/2, h)
//	eh := NewSideBySide(ehA, ehB)
//
//	eh.HandleResize(0, 0, w, h)
//
//	for {
//		eh.HandlePaint()
//		eh.HandleSetCursor()
//		termbox.Flush()
//		ev := termbox.PollEvent()
//		if ev.Type == termbox.EventKey && ev.Key == termbox.KeyCtrlX {
//			return
//		}
//		Handle(eh, &ev)
//	}
//}
