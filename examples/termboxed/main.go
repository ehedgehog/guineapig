package main

// import "github.com/limetext/termbox-go"

import "github.com/nsf/termbox-go"
import (
	"github.com/ehedgehog/guineapig/examples/termboxed/buffer"
	"github.com/ehedgehog/guineapig/examples/termboxed/screen"
)

import "fmt"

// import "log"

// import _ "github.com/ehedgehog/guineapig/examples/termboxed/panel"

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

func NewPanel(x, y int, w, h int) screen.Panel {
	return &SimplePanel{x, y, w, h}
}

type EventHandler interface {
	Handle(e *termbox.Event) error
}

type Loc struct {
	X, Y int
}

type Editor struct {
	p screen.Panel
	b buffer.Type
	l Loc
}

type SimpleEventHandler struct {
	e *Editor
}

func NewSimpleEventHandler(x, y int, w, h int) EventHandler {
	p := NewPanel(x, y, w, h)
	b := buffer.New(w, h)
	l := Loc{0, 0}
	e := &Editor{p, b, l}
	return &SimpleEventHandler{e}
}

func (eh *SimpleEventHandler) Handle(e *termbox.Event) error {

	if e.Type == termbox.EventKey {
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
	} else if e.Type == termbox.EventMouse {
		b := eh.e.b
		report := fmt.Sprintf("<mouse: %v, %v>\n", e.MouseX, e.MouseY)
		for _, ch := range report {
			b.Insert(rune(ch))
		}

	}

	w, h := termbox.Size()
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	_, _ = w, h

	eh.e.b.PutAll(eh.e.p)
	termbox.Flush()
	return nil
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
