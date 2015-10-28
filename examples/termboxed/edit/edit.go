package edit

import (
	"log"

	"github.com/ehedgehog/guineapig/examples/termboxed/screen"
	"github.com/ehedgehog/guineapig/examples/termboxed/text"
)
import "github.com/ehedgehog/guineapig/examples/termboxed/grid"

type State struct {
	Where  grid.LineCol
	Buffer text.Buffer
	Marked grid.MarkedRange
	Offset grid.Offset
}

type Panel struct {
	Canvas    screen.Canvas
	PaintFunc func(*Panel)
}

func (p *Panel) SetCursor(where grid.LineCol) {
	p.Canvas.SetCursor(where)
}

func (p *Panel) Size() grid.Size {
	return p.Canvas.Size()
}

func (p *Panel) Paint() {
	if p.PaintFunc == nil {
		log.Println("Paint -- no paint function provided.")
	} else {
		p.PaintFunc(p)
	}
}
