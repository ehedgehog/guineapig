package layouts

import "github.com/gdamore/tcell"
import "github.com/ehedgehog/guineapig/examples/termboxed/events"
import "github.com/ehedgehog/guineapig/examples/termboxed/screen"
import "github.com/ehedgehog/guineapig/examples/termboxed/grid"
import "github.com/ehedgehog/guineapig/examples/termboxed/bounds"

type Shelf struct {
	Block
}

func NewShelf(generator func() events.Handler, elements ...events.Handler) events.Handler {
	return &Shelf{
		Block: Block{
			focus:     0,
			elements:  elements,
			generator: generator,
			bounds:    make([]int, len(elements)),
		},
	}
}

func (s *Shelf) New() events.Handler {
	return NewShelf(s.generator)
}

func (s *Shelf) Geometry() grid.Geometry {
	minw, maxw, minh, maxh := 0, 0, 0, 0
	for _, eh := range s.elements {
		g := eh.Geometry()
		minh = bounds.Max(minh, g.MinHeight)
		maxh = bounds.Max(maxh, g.MaxHeight)
		minw = minw + g.MinWidth
		maxw = maxw + g.MaxWidth
	}
	return grid.Geometry{MinWidth: minw, MaxWidth: maxw, MinHeight: minh, MaxHeight: maxh}
}

func (b *Shelf) Key(e *tcell.EventKey) error {
	if e.Key() == tcell.KeyCtrlT {
		b.elements = append(b.elements, b.generator())
		b.bounds = append(b.bounds, 0)
		b.ResizeTo(b.recentSize)
		return nil
	}
	return b.elements[b.focus].Key(e)
}

func (s *Shelf) Mouse(e *tcell.EventMouse) error {
	x := 0
	for i, w := range s.bounds {
		nextX := x + w
		mx, my := e.Position()
		if mx < nextX {
			mx -= x
			s.focus = i
			return s.elements[i].Mouse(tcell.NewEventMouse(mx, my, e.Buttons(), e.Modifiers()))
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
		if g.MinWidth != g.MaxWidth {
			count += 1
		}
	}
	totalSpare := w - g.MinWidth
	spare := totalSpare / count
	x := 0
	for i, eh := range s.elements {
		g := eh.Geometry()
		if g.MinWidth == g.MaxWidth {
			w := g.MinWidth
			s.bounds[i] = w
			c := screen.NewSubCanvas(outer, x, 0, w, h)
			eh.ResizeTo(c)
			x += w
		} else {
			w := g.MinWidth + spare
			s.bounds[i] = w
			c := screen.NewSubCanvas(outer, x, 0, w, h)
			eh.ResizeTo(c)
			x += w
		}
	}
	s.recentSize = outer
	return nil
}
