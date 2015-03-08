package layouts

import "github.com/ehedgehog/guineapig/examples/termboxed/screen"
import "github.com/ehedgehog/guineapig/examples/termboxed/events"

type Block struct {
	generator  func() events.Handler
	elements   []events.Handler
	bounds     []int
	focus      int
	recentSize screen.Canvas
}

func (b *Block) SetCursor() error {
	return b.elements[b.focus].SetCursor()
}

func (b *Block) Paint() error {
	for _, e := range b.elements {
		e.Paint()
	}
	return nil
}
