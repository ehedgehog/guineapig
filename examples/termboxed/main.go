package main

import "github.com/gdamore/tcell"

import "github.com/ehedgehog/guineapig/examples/termboxed/screen"

import "github.com/ehedgehog/guineapig/examples/termboxed/layouts"
import "github.com/ehedgehog/guineapig/examples/termboxed/events"
import "github.com/ehedgehog/guineapig/examples/termboxed/edit"

func main() {
	err := screen.TheScreen.Init()
	if err != nil {
		panic(err)
	}
	defer screen.TheScreen.Fini()

	page := screen.NewTermboxCanvas()

	edA := layouts.NewStack(edit.NewEditorPanel, edit.NewEditorPanel())

	eh := layouts.NewShelf(func() events.Handler { return layouts.NewStack(edit.NewEditorPanel, edit.NewEditorPanel()) }, edA)

	eh.ResizeTo(page)
	screen.TheScreen.EnableMouse()

	for {
		screen.TheScreen.Clear()
		eh.Paint()
		eh.SetCursor()
		screen.TheScreen.Show()
		ev := screen.TheScreen.PollEvent()

		switch ev := ev.(type) {
		case *tcell.EventMouse:
			if ev.Buttons() > 0 {
				eh.Mouse(ev)
			}
		case *tcell.EventKey:
			eh.Key(ev)
			if ev.Key() == tcell.KeyCtrlX {
				return
			}
		case *tcell.EventResize:
			page = screen.NewTermboxCanvas()
			eh.ResizeTo(page)
		}
	}
}
