package events

import "github.com/gdamore/tcell"
import "github.com/ehedgehog/guineapig/examples/termboxed/screen"
import "github.com/ehedgehog/guineapig/examples/termboxed/grid"

type Handler interface {
	Key(e *tcell.EventKey) error
	Mouse(e *tcell.EventMouse) error
	ResizeTo(outer screen.Canvas) error
	Paint() error
	SetCursor() error
	Geometry() grid.Geometry
	New() Handler
}
