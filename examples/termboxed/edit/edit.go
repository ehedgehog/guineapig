package edit

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ehedgehog/guineapig/examples/termboxed/bounds"
	"github.com/ehedgehog/guineapig/examples/termboxed/draw"
	"github.com/ehedgehog/guineapig/examples/termboxed/events"
	"github.com/ehedgehog/guineapig/examples/termboxed/screen"
	"github.com/ehedgehog/guineapig/examples/termboxed/text"
	"github.com/gdamore/tcell"
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

type EditorPanel struct {
	topBar    *Panel
	bottomBar *Panel
	leftBar   *Panel
	rightBar  *Panel
	textBox   *Panel

	current *State
	main    State
	command State
}

func (ep *EditorPanel) New() events.Handler {
	return NewEditorPanel()
}

func NewEditorPanel() events.Handler {
	mb := text.NewBuffer(func(b text.Buffer, s string) error { return nil })
	var ep *EditorPanel
	ep = &EditorPanel{
		main: State{Buffer: mb},

		command: State{Buffer: text.NewBuffer(func(b text.Buffer, s string) error {
			content := b.Expose()
			line := ep.command.Where.Line
			blobs := strings.Split(content[line], " ")
			command := commands[blobs[0]]
			if command == nil {
				return errors.New("not a command: " + blobs[0])
			} else {
				return command(ep, blobs)
			}
		}),
		},
	}
	ep.current = &ep.main
	return ep
}

func (ep *EditorPanel) Geometry() grid.Geometry {
	minw := 2
	maxw := 1000
	minh := 2
	maxh := 1000
	return grid.Geometry{MinWidth: minw, MaxWidth: maxw, MinHeight: minh, MaxHeight: maxh}
}

func readIntoBuffer(ep *EditorPanel, b text.Buffer, fileName string) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	w, err := b.ReadFromFile(ep.main.Where, fileName, f)
	ep.main.Where = w
	return err
}

var commands = map[string]func(*EditorPanel, []string) error{
	"r": func(ep *EditorPanel, blobs []string) error {
		b := ep.main.Buffer
		return readIntoBuffer(ep, b, blobs[1])
	},
	"mr": func(ep *EditorPanel, blobs []string) error {
		b := ep.main.Buffer
		if ep.main.Marked.IsActive() {
			target := ep.main.Where.Line
			first, last := ep.main.Marked.Range()
			if first <= target && target <= last {
				return errors.New("range overlaps target")
			}
			b.MoveLines(ep.main.Where, first, last)
			ep.main.Where.Line = ep.main.Marked.MoveAfter(target)
			return nil
		} else {
			return errors.New("no marked range")
		}
	},
	"w": func(ep *EditorPanel, blobs []string) error {
		b := ep.main.Buffer
		return b.WriteToFile(blobs[1:])
	},
	"d": func(ep *EditorPanel, blobs []string) error {
		b := ep.main.Buffer
		lineNumber := ep.main.Where.Line
		b.DeleteLine(ep.main.Where)
		ep.main.Marked.RemoveLine(lineNumber)
		return nil
	},
	"dr": func(ep *EditorPanel, blobs []string) error {
		if ep.main.Marked.IsActive() {
			b := ep.main.Buffer
			first, last := ep.main.Marked.Range()
			ep.current.Where = b.DeleteLines(ep.current.Where, first, last)
			ep.main.Marked.Clear()
			return nil
		} else {
			return errors.New("no marked range")
		}
	},
}

func (ep *EditorPanel) Key(e *tcell.EventKey) error {
	b := ep.current.Buffer
	if e.Key() != tcell.KeyRune {
		switch e.Key() {

		case 0:
			// nothing

		case tcell.KeyF1:
			ep.current = &ep.command
			ep.command.Buffer.Return(ep.command.Where)
			ep.current.Where = grid.LineCol{Line: ep.command.Where.Line + 1, Col: 0}

		case tcell.KeyF2:
			ep.command.Where, _ = ep.command.Buffer.Execute(ep.command.Where)

		case tcell.KeyCtrlB:
			if ep.current == &ep.main {
				ep.current = &ep.command
			} else {
				ep.current = &ep.main
			}

		case tcell.KeySpace:
			b.Insert(ep.current.Where, ' ')
			ep.current.Where.RightOne()

		case tcell.KeyBackspace2:
			ep.current.Where = b.DeleteBack(ep.current.Where)

		case tcell.KeyDelete:
			ep.current.Where = b.DeleteForward(ep.current.Where)

		case tcell.KeyF3:
			ep.main.Marked.SetLow(ep.main.Where.Line)

		case tcell.KeyF4:
			ep.main.Marked.SetHigh(ep.main.Where.Line)

		case tcell.KeyPgUp:
			where := ep.current.Where
			vo := ep.current.Offset.Vertical
			if where.Line-vo == 0 {
				top := bounds.Max(0, where.Line-ep.textBox.Size().Height)
				ep.current.Where = grid.LineCol{top, where.Col}
			} else {
				ep.current.Where = grid.LineCol{vo, where.Col}
			}

		case tcell.KeyPgDn:
			where := ep.current.Where
			vo := ep.current.Offset.Vertical
			height := ep.textBox.Size().Height
			if where.Line-vo == height-1 {
				// forward one page
				bot := where.Line + height
				ep.current.Where = grid.LineCol{bot, where.Col}
			} else {
				// bottom of this page
				ep.current.Where = grid.LineCol{vo + height - 1, where.Col}
			}

		case tcell.KeyEnd:
			where := ep.current.Where
			if where.Col == 0 {
				contents := b.Expose()
				line := contents[ep.current.Where.Line]
				where.Col = len(line)
			} else {
				where.Col = 0
			}
			ep.current.Where = where

		case tcell.KeyEnter:
			if ep.current == &ep.main {
				ep.current.Where = b.Return(ep.current.Where)
				ep.main.Marked.Return(ep.current.Where.Line)
			} else {
				_, err := b.Execute(ep.current.Where)
				if err == nil {
					report(ep, b, "OK")
				} else {
					report(ep, b, err.Error())
				}
				ep.current = &ep.main
			}

		case tcell.KeyRight:
			ep.current.Where.RightOne()

		case tcell.KeyUp:
			ep.current.Where.UpOne()

		case tcell.KeyDown:
			ep.current.Where.DownOne()

		case tcell.KeyLeft:
			ep.current.Where.LeftOne()

		default:
			report := fmt.Sprintf("<key: %#d>\n", uint(e.Key()))
			for _, ch := range report {
				b.Insert(ep.current.Where, rune(ch))
				ep.current.Where.RightOne()
			}
		}
	} else {
		b.Insert(ep.current.Where, e.Rune())
		ep.current.Where.RightOne()
	}
	return nil
}

func report(ep *EditorPanel, b text.Buffer, message string) {
	b.Insert(ep.current.Where, ' ')
	ep.current.Where.RightOne()
	b.Insert(ep.current.Where, '(')
	ep.current.Where.RightOne()
	for _, rune := range message {
		b.Insert(ep.current.Where, rune)
		ep.current.Where.RightOne()
	}
	b.Insert(ep.current.Where, ')')
	ep.current.Where.RightOne()
	b.Insert(ep.current.Where, ' ')
	ep.current.Where.RightOne()
}

func (ep *EditorPanel) Mouse(e *tcell.EventMouse) error {
	x, y := e.Position()
	size := ep.textBox.Size()
	w, h := size.Width, size.Height
	if 0 < x && x < w+1 && 0 < y && y < h+1 {
		ep.current = &ep.main
		ep.current.Where = grid.LineCol{y - 1, x - 1}

		// hack to adjust beteen buffer & cancas coordinates.
		ep.current.Where.Line -= 1
		ep.current.Where.Line += ep.current.Offset.Vertical
		ep.current.Where.Col -= 6

	} else if x >= delta && y == 0 {
		ep.command.Where = grid.LineCol{0, x - delta}
		ep.current = &ep.command
	}
	return nil
}

func (ep *EditorPanel) AdjustScrolling() {
	size := ep.textBox.Size()
	line := ep.current.Where.Line
	h := size.Height
	if line < ep.current.Offset.Vertical {
		ep.current.Offset.Vertical = line
	}
	if line > ep.current.Offset.Vertical+h-1 {
		ep.current.Offset.Vertical = line - h + 1
	}
}

func (ep *EditorPanel) Paint() error {
	ep.AdjustScrolling()
	ep.topBar.Paint()
	ep.bottomBar.Paint()
	ep.leftBar.Paint()
	ep.rightBar.Paint()
	ep.textBox.Paint()
	return nil
}

const delta = 5

func textPainterFor(tb *TextBox, s *State) func(*Panel) {
	return func(p *Panel) {
		h := tb.lineInfo.Size().Height
		v := s.Offset.Vertical
		s.Buffer.PutLines(tb.lineContent, v, h)

		if s.Marked.IsActive() {
			first, last := s.Marked.Range()
			for line := first - v; line < last-v+1; line += 1 {
				tb.lineInfo.SetCell(grid.LineCol{line, tryTagSize - 1}, '║', markStyle)
			}
		}

		ln := 0
		numberStyle := screen.DefaultStyle
		//
		for i := v; i < v+h; i += 1 {
			s := fmt.Sprintf("%4v", i)
			for j, ch := range s {
				tb.lineInfo.SetCell(grid.LineCol{ln, j}, rune(ch), numberStyle)
			}
			ln += 1
		}
	}
}

func rightPainterFor(s *State) func(*Panel) {
	return func(p *Panel) {
		b := s.Buffer
		content := b.Expose()
		line := s.Where.Line
		length := bounds.Max(line, len(content))
		draw.Scrollbar(p.Canvas, draw.ScrollInfo{length, line})
	}
}

func bottomPainter(p *Panel) {
	c := p.Canvas
	w := c.Size().Width
	c.SetCell(grid.LineCol{Col: 0, Line: 0}, draw.Glyph_corner_bl, screen.DefaultStyle)
	for i := 1; i < w; i += 1 {
		c.SetCell(grid.LineCol{Col: i, Line: 0}, draw.Glyph_hbar, screen.DefaultStyle)
	}
	c.SetCell(grid.LineCol{Col: w - 1, Line: 0}, draw.Glyph_corner_br, screen.DefaultStyle)
}

func topPainterFor(s *State) func(*Panel) {
	return func(p *Panel) {
		c := p.Canvas
		w := c.Size().Width
		c.SetCell(grid.LineCol{Col: 0, Line: 0}, draw.Glyph_corner_tl, screen.DefaultStyle)
		for i := 1; i < w; i += 1 {
			c.SetCell(grid.LineCol{Col: i, Line: 0}, draw.Glyph_hbar, screen.DefaultStyle)
		}
		screen.PutString(c, 2, 0, "─┤ ", screen.DefaultStyle)
		c.SetCell(grid.LineCol{Col: w - 1, Line: 0}, draw.Glyph_corner_tr, screen.DefaultStyle)
		tline := s.Where.Line
		s.Buffer.PutLines(screen.NewSubCanvas(c, delta, 0, w-delta-2, 1), tline, 1)
	}
}

func leftPainter(p *Panel) {
	c := p.Canvas
	h := c.Size().Height
	for j := 0; j < h; j += 1 {
		c.SetCell(grid.LineCol{Col: 0, Line: j}, draw.Glyph_vbar, screen.DefaultStyle)
	}
}

func (ep *EditorPanel) ResizeTo(outer screen.Canvas) error {
	size := outer.Size()
	w, h := size.Width, size.Height

	ep.leftBar = &Panel{Canvas: screen.NewSubCanvas(outer, 0, 1, 1, h-2), PaintFunc: leftPainter}
	ep.rightBar = &Panel{Canvas: screen.NewSubCanvas(outer, w-1, 1, 1, h-2), PaintFunc: rightPainterFor(&ep.main)}
	ep.topBar = &Panel{Canvas: screen.NewSubCanvas(outer, 0, 0, w, 1), PaintFunc: topPainterFor(&ep.command)}
	ep.bottomBar = &Panel{Canvas: screen.NewSubCanvas(outer, 0, h-1, w, 1), PaintFunc: bottomPainter}

	textBox := NewTextBox(ep, outer, 1, 1, w-2, h-2)
	ep.textBox = &Panel{Canvas: textBox, PaintFunc: textPainterFor(textBox, &ep.main)}
	return nil
}

const tryTagSize = 6

func NewTextBox(ep *EditorPanel, outer screen.Canvas, dx, dy, w, h int) *TextBox {
	sub := screen.NewSubCanvas(outer, dx, dy, w, h)
	lineInfo := screen.NewSubCanvas(sub, 0, 0, tryTagSize, h)
	page := screen.NewSubCanvas(sub, tryTagSize, 0, w-tryTagSize, h)
	return &TextBox{lineInfo: lineInfo, lineContent: page, embedded: sub}
}

type TextBox struct {
	embedded    screen.Canvas
	lineInfo    screen.Canvas
	lineContent screen.Canvas
}

func (tb *TextBox) Size() grid.Size {
	return tb.embedded.Size()
}

func (s *TextBox) SetCell(where grid.LineCol, glyph rune, st tcell.Style) {
	if where.Col < tryTagSize {
		s.lineInfo.SetCell(where, glyph, st)
	} else {
		s.lineContent.SetCell(where.ColPlus(-tryTagSize), glyph, st)
	}
}

var markStyle = screen.DefaultStyle.Foreground(tcell.ColorBrightRed)

var hereStyle = screen.DefaultStyle.Foreground(tcell.ColorRed)

func (t *TextBox) SetCursor(where grid.LineCol) {
	// col := bounds.Max(0, where.Col-tryTagSize)
	t.lineContent.SetCursor(grid.LineCol{where.Line, where.Col})
}

func (ep *EditorPanel) SetCursor() error {
	if ep.current == &ep.main {
		ep.textBox.SetCursor(ep.current.Where.LineMinus(ep.current.Offset.Vertical))
	} else {
		where := ep.command.Where
		ep.topBar.SetCursor(grid.LineCol{0, where.Col + delta})
	}
	return nil
}
