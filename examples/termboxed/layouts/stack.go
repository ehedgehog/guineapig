package layouts

import (
	"log"

	"github.com/gdamore/tcell"
)
import "github.com/ehedgehog/guineapig/examples/termboxed/events"
import "github.com/ehedgehog/guineapig/examples/termboxed/screen"
import "github.com/ehedgehog/guineapig/examples/termboxed/grid"
import "github.com/ehedgehog/guineapig/examples/termboxed/bounds"

type Stack struct {
	Block
}

func NewStack(generator func() events.Handler, elements ...events.Handler) events.Handler {
	return &Stack{
		Block: Block{
			focus:     0,
			elements:  elements,
			generator: generator,
			bounds:    make([]int, len(elements)),
		},
	}
}

func (s *Stack) New() events.Handler {
	return NewStack(s.generator)
}

func (s *Stack) Geometry() grid.Geometry {
	minw, maxw, minh, maxh := 0, 0, 0, 0
	for _, eh := range s.elements {
		g := eh.Geometry()
		minw = bounds.Max(minw, g.MinWidth)
		maxw = bounds.Max(maxw, g.MaxWidth)
		minh = minh + g.MinHeight
		maxh = maxh + g.MaxHeight
	}
	return grid.Geometry{MinWidth: minw, MaxWidth: maxw, MinHeight: minh, MaxHeight: maxh}
}

func (b *Stack) Key(e *tcell.EventKey) error {
	if e.Key() == tcell.KeyCtrlU {
		b.elements = append(b.elements, b.generator())
		b.bounds = append(b.bounds, 0)
		b.ResizeTo(b.recentSize)
		return nil
	}
	return b.elements[b.focus].Key(e)
}

func (s *Stack) Mouse(e *tcell.EventMouse) error {
	a, b := e.Position()
	log.Println("Stack.Shelf", a, b)
	y := 0
	for i, h := range s.bounds {
		nextY := y + h
		mx, my := e.Position()
		if my < nextY {
			my -= y
			s.focus = i
			return s.elements[i].Mouse(tcell.NewEventMouse(mx, my, e.Buttons(), e.Modifiers()))
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
		if g.MinHeight != g.MaxHeight {
			count += 1
		}
	}
	totalSpare := h - g.MinHeight
	spare := totalSpare / count
	y := 0
	for i, eh := range s.elements {
		g := eh.Geometry()
		if g.MinHeight == g.MaxHeight {
			h := g.MinHeight
			s.bounds[i] = h
			c := screen.NewSubCanvas(outer, 0, y, w, h)
			eh.ResizeTo(c)
			y += h
		} else {
			h := g.MinHeight + spare
			s.bounds[i] = h
			c := screen.NewSubCanvas(outer, 0, y, w, h)
			eh.ResizeTo(c)
			y += h
		}
	}
	s.recentSize = outer
	return nil
}
