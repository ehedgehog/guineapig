package main

import termbox "github.com/limetext/termbox-go"

import "github.com/ehedgehog/guineapig/examples/termboxed/bounds"
import "github.com/ehedgehog/guineapig/examples/termboxed/screen"
import "github.com/ehedgehog/guineapig/examples/termboxed/events"
import "github.com/ehedgehog/guineapig/examples/termboxed/grid"

type SideBySide struct {
	widthA int
	Focus  events.EventHandler
	A, B   events.EventHandler
}

func (s *SideBySide) Geometry() grid.Geometry {
	ga, gb := s.A.Geometry(), s.B.Geometry()
	minw := ga.MinWidth + gb.MinWidth
	maxw := ga.MaxWidth + gb.MaxWidth
	minh := bounds.Max(ga.MinHeight, gb.MinHeight)
	maxh := bounds.Max(ga.MaxHeight, gb.MaxHeight)
	return grid.Geometry{MinWidth: minw, MaxWidth: maxw, MinHeight: minh, MaxHeight: maxh}
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

func NewSideBySide(A, B events.EventHandler) events.EventHandler {
	return &SideBySide{0, A, A, B}
}

func (s *SideBySide) New() events.EventHandler {
	panic("SideBySide.New")
}
