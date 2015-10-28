package edit

import "github.com/ehedgehog/guineapig/examples/termboxed/text"
import "github.com/ehedgehog/guineapig/examples/termboxed/grid"

type State struct {
	where  grid.LineCol
	buffer text.Buffer
	marked grid.MarkedRange
	offset grid.Offset
}

func (s *State) Buffer() text.Buffer {
	return s.buffer
}

func (s *State) Where() grid.LineCol {
	return s.where
}

func (s *State) SetWhere(where grid.LineCol) {
	s.where = where
}
