package edit

import "github.com/ehedgehog/guineapig/examples/termboxed/text"
import "github.com/ehedgehog/guineapig/examples/termboxed/grid"

type State struct {
	Where  grid.LineCol
	Buffer text.Buffer
	Marked grid.MarkedRange
	Offset grid.Offset
}
