package panel

import "github.com/limetext/termbox-go"

type Pen interface {
	FG() termbox.Attribute
	BG() termbox.Attribute
}

type Panel interface {
	Size() (w, h int)
	Loc() (x, y int)
	ShowString(x, y int, content string, p Pen) error
	ShowRune(x, y int, content rune, p Pen) error
}

func ShowString(p Panel, x, y int, content string, with Pen) error {
	i := 0
	for _, ch := range content {
		p.ShowRune(x+i, y, ch, with)
		i += 1
	}
	return nil
}

type Shelf struct {
	Items []Panel
}
