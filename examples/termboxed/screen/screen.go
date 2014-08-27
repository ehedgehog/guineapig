package screen

import "github.com/nsf/termbox-go"

type Writeable interface {
	PutString(x, y int, content string)
	SetCell(x, y int, ch rune, a, b termbox.Attribute)
	SetCursor(x, y int)
	Resize(x, y, w, h int)
	Size() (w, h int)
}

type Panel interface {
	Writeable
}
