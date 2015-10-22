package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	//	"strconv"
	"strings"
)

import "github.com/gdamore/tcell"

import "github.com/ehedgehog/guineapig/examples/termboxed/bounds"
import "github.com/ehedgehog/guineapig/examples/termboxed/text"
import "github.com/ehedgehog/guineapig/examples/termboxed/draw"
import "github.com/ehedgehog/guineapig/examples/termboxed/screen"
import "github.com/ehedgehog/guineapig/examples/termboxed/grid"
import "github.com/ehedgehog/guineapig/examples/termboxed/layouts"
import "github.com/ehedgehog/guineapig/examples/termboxed/events"

type Offset struct {
	vertical   int
	horizontal int
}

type State struct {
	where  grid.LineCol
	buffer text.Buffer
	marked grid.MarkedRange
	offset Offset
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

type Panel struct {
	canvas screen.Canvas
	paint  func(*Panel)
}

func (p *Panel) SetCursor(where grid.LineCol) {
	p.canvas.SetCursor(where)
}

func (p *Panel) Size() grid.Size {
	return p.canvas.Size()
}

func (p *Panel) Paint() {
	if p.paint == nil {
		log.Println("Paint -- no paint function provided.")
	} else {
		p.paint(p)
	}
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
	w, err := b.ReadFromFile(ep.main.where, fileName, f)
	ep.main.where = w
	return err
}

var commands = map[string]func(*EditorPanel, []string) error{
	"r": func(ep *EditorPanel, blobs []string) error {
		b := ep.main.buffer
		return readIntoBuffer(ep, b, blobs[1])
	},
	"mr": func(ep *EditorPanel, blobs []string) error {
		b := ep.main.buffer
		if ep.main.marked.IsActive() {
			target := ep.main.where.Line
			first, last := ep.main.marked.Range()
			if first <= target && target <= last {
				return errors.New("range overlaps target")
			}
			b.MoveLines(ep.main.where, first, last)
			ep.main.where.Line = ep.main.marked.MoveAfter(target)
			return nil
		} else {
			return errors.New("no marked range")
		}
	},
	"w": func(ep *EditorPanel, blobs []string) error {
		b := ep.main.buffer
		return b.WriteToFile(blobs[1:])
	},
	"d": func(ep *EditorPanel, blobs []string) error {
		b := ep.main.buffer
		lineNumber := ep.main.where.Line
		b.DeleteLine(ep.main.where)
		ep.main.marked.RemoveLine(lineNumber)
		return nil
	},
	"dr": func(ep *EditorPanel, blobs []string) error {
		if ep.main.marked.IsActive() {
			b := ep.main.buffer
			first, last := ep.main.marked.Range()
			ep.current.where = b.DeleteLines(ep.current.where, first, last)
			ep.main.marked.Clear()
			return nil
		} else {
			return errors.New("no marked range")
		}
	},
}

func NewEditorPanel() events.Handler {
	mb := text.NewBuffer(func(b text.Buffer, s string) error { return nil })
	var ep *EditorPanel
	ep = &EditorPanel{
		main: State{buffer: mb},

		command: State{buffer: text.NewBuffer(func(b text.Buffer, s string) error {
			content := b.Expose()
			line := ep.command.where.Line
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

func (ep *EditorPanel) New() events.Handler {
	return NewEditorPanel()
}

func (ep *EditorPanel) Key(e *tcell.EventKey) error {
	b := ep.current.buffer
	if e.Key() != tcell.KeyRune  {
		switch e.Key() {

		case 0:
			// nothing

		case tcell.KeyF1:
			ep.current = &ep.command
			ep.command.buffer.Return(ep.command.where)
			ep.current.where = grid.LineCol{Line: ep.command.where.Line + 1, Col: 0}

		case tcell.KeyF2:
			ep.command.where, _ = ep.command.buffer.Execute(ep.command.where)

		case tcell.KeyCtrlB:
			if ep.current == &ep.main {
				ep.current = &ep.command
			} else {
				ep.current = &ep.main
			}

		case tcell.KeySpace:
			b.Insert(ep.current.where, ' ')
			ep.current.where.RightOne()

		case tcell.KeyBackspace2:
			ep.current.where = b.DeleteBack(ep.current.where)

		case tcell.KeyDelete:
			ep.current.where = b.DeleteForward(ep.current.where)

		case tcell.KeyF3:
			ep.main.marked.SetLow(ep.main.where.Line)

		case tcell.KeyF4:
			ep.main.marked.SetHigh(ep.main.where.Line)

		case tcell.KeyPgUp:
			where := ep.current.where
			vo := ep.current.offset.vertical
			if where.Line-vo == 0 {
				top := bounds.Max(0, where.Line-ep.textBox.Size().Height)
				ep.current.where = grid.LineCol{top, where.Col}
			} else {
				ep.current.where = grid.LineCol{vo, where.Col}
			}

		case tcell.KeyPgDn:
			where := ep.current.where
			vo := ep.current.offset.vertical
			height := ep.textBox.Size().Height
			if where.Line-vo == height-1 {
				// forward one page
				bot := where.Line + height
				ep.current.where = grid.LineCol{bot, where.Col}
			} else {
				// bottom of this page
				ep.current.where = grid.LineCol{vo + height - 1, where.Col}
			}

		case tcell.KeyEnd:
			where := ep.current.where
			if where.Col == 0 {
				contents := b.Expose()
				line := contents[ep.current.where.Line]
				where.Col = len(line)
			} else {
				where.Col = 0
			}
			ep.current.where = where

		case tcell.KeyEnter:
			if ep.current == &ep.main {
				ep.current.where = b.Return(ep.current.where)
				ep.main.marked.Return(ep.current.where.Line)
			} else {
				_, err := b.Execute(ep.current.where)
				if err == nil {
					report(ep, b, "OK")
				} else {
					report(ep, b, err.Error())
				}
				ep.current = &ep.main
			}

		case tcell.KeyRight:
			ep.current.where.RightOne()

		case tcell.KeyUp:
			ep.current.where.UpOne()

		case tcell.KeyDown:
			ep.current.where.DownOne()

		case tcell.KeyLeft:
			ep.current.where.LeftOne()

		default:
			report := fmt.Sprintf("<key: %#d>\n", uint(e.Key()))
			for _, ch := range report {
				b.Insert(ep.current.where, rune(ch))
				ep.current.where.RightOne()
			}
		}
	} else {
		b.Insert(ep.current.where, e.Rune())
		ep.current.where.RightOne()
	}
	return nil
}

func report(ep *EditorPanel, b text.Buffer, message string) {
	b.Insert(ep.current.where, ' ')
	ep.current.where.RightOne()
	b.Insert(ep.current.where, '(')
	ep.current.where.RightOne()
	for _, rune := range message {
		b.Insert(ep.current.where, rune)
		ep.current.where.RightOne()
	}
	b.Insert(ep.current.where, ')')
	ep.current.where.RightOne()
	b.Insert(ep.current.where, ' ')
	ep.current.where.RightOne()
}

func (ep *EditorPanel) Mouse(e *tcell.EventMouse) error {
	x, y := e.Position()
	size := ep.textBox.Size()
	w, h := size.Width, size.Height
	if 0 < x && x < w+1 && 0 < y && y < h+1 {
		ep.current.where = grid.LineCol{y - 1, x - 1}
		ep.current = &ep.main
	} else if x >= delta && y == 0 {
		ep.command.where = grid.LineCol{0, x - delta}
		ep.current = &ep.command
	}
	return nil
}

func (ep *EditorPanel) AdjustScrolling() {
	size := ep.textBox.Size()
	line := ep.current.where.Line
	h := size.Height
	if line < ep.current.offset.vertical {
		ep.current.offset.vertical = line
	}
	if line > ep.current.offset.vertical+h-1 {
		ep.current.offset.vertical = line - h + 1
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
		v := s.offset.vertical
		s.buffer.PutLines(tb.lineContent, v, h)

		if s.marked.IsActive() {
			first, last := s.marked.Range()
			for line := first - v; line < last-v+1; line += 1 {
				tb.lineInfo.SetCell(grid.LineCol{line, tryTagSize - 1}, ' ', markStyle)
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
		b := s.buffer
		content := b.Expose()
		line := s.where.Line
		length := bounds.Max(line, len(content))
		draw.Scrollbar(p.canvas, draw.ScrollInfo{length, line})
	}
}

func bottomPainter(p *Panel) {
	c := p.canvas
	w := c.Size().Width
	c.SetCell(grid.LineCol{Col: 0, Line: 0}, draw.Glyph_corner_bl, screen.DefaultStyle)
	for i := 1; i < w; i += 1 {
		c.SetCell(grid.LineCol{Col: i, Line: 0}, draw.Glyph_hbar, screen.DefaultStyle)
	}
	c.SetCell(grid.LineCol{Col: w - 1, Line: 0}, draw.Glyph_corner_br, screen.DefaultStyle)
}

func topPainterFor(s *State) func(*Panel) {
	return func(p *Panel) {
		c := p.canvas
		w := c.Size().Width
		c.SetCell(grid.LineCol{Col: 0, Line: 0}, draw.Glyph_corner_tl, screen.DefaultStyle)
		for i := 1; i < w; i += 1 {
			c.SetCell(grid.LineCol{Col: i, Line: 0}, draw.Glyph_hbar, screen.DefaultStyle)
		}
		screen.PutString(c, 2, 0, "─┤ ", screen.DefaultStyle)
		c.SetCell(grid.LineCol{Col: w - 1, Line: 0}, draw.Glyph_corner_tr, screen.DefaultStyle)
		tline := s.where.Line
		s.buffer.PutLines(screen.NewSubCanvas(c, delta, 0, w-delta-2, 1), tline, 1)
	}
}

func leftPainter(p *Panel) {
	c := p.canvas
	h := c.Size().Height
	for j := 0; j < h; j += 1 {
		c.SetCell(grid.LineCol{Col: 0, Line: j}, draw.Glyph_vbar, screen.DefaultStyle)
	}
}

func (ep *EditorPanel) ResizeTo(outer screen.Canvas) error {
	size := outer.Size()
	w, h := size.Width, size.Height

	ep.leftBar = &Panel{canvas: screen.NewSubCanvas(outer, 0, 1, 1, h-2), paint: leftPainter}
	ep.rightBar = &Panel{canvas: screen.NewSubCanvas(outer, w-1, 1, 1, h-2), paint: rightPainterFor(&ep.main)}
	ep.topBar = &Panel{canvas: screen.NewSubCanvas(outer, 0, 0, w, 1), paint: topPainterFor(&ep.command)}
	ep.bottomBar = &Panel{canvas: screen.NewSubCanvas(outer, 0, h-1, w, 1), paint: bottomPainter}

	textBox := NewTextBox(ep, outer, 1, 1, w-2, h-2)
	ep.textBox = &Panel{canvas: textBox, paint: textPainterFor(textBox, &ep.main)}
	return nil
}

const tryTagSize = 6

func NewTextBox(ep *EditorPanel, outer screen.Canvas, dx, dy, w, h int) *TextBox {
	sub := screen.NewSubCanvas(outer, dx, dy, w, h)
	lineInfo := screen.NewSubCanvas(sub, 0, 0, tryTagSize, h)
	page := screen.NewSubCanvas(sub, tryTagSize, 0, w-tryTagSize, h)
	return &TextBox{lineInfo: lineInfo, lineContent: page, Canvas: sub}
}

type TextBox struct {
	screen.Canvas
	lineInfo    screen.Canvas
	lineContent screen.Canvas
}

var markStyle = screen.DefaultStyle.Background(tcell.ColorYellow)

var hereStyle = screen.DefaultStyle.Foreground(tcell.ColorRed)

func (t *TextBox) SetCursor(where grid.LineCol) {
	// col := bounds.Max(0, where.Col-tryTagSize)
	t.lineContent.SetCursor(grid.LineCol{where.Line, where.Col})
}

func (ep *EditorPanel) SetCursor() error {
	if ep.current == &ep.main {
		where := ep.current.where.LineMinus(ep.current.offset.vertical)
		ep.textBox.SetCursor(where)
	} else {
		where := ep.command.where
		ep.topBar.SetCursor(grid.LineCol{0, where.Col + delta})
	}
	return nil
}

func main() {
	err := screen.TheScreen.Init()
	if err != nil {
		panic(err)
	}
	defer screen.TheScreen.Fini()

	// tcell.SetColorMode(tcell.ColorMode256)
	// tcell.SetColorPalette(makeColours())
	// tcell.SetInputMode(tcell.InputEsc | tcell.InputMouse)

	page := screen.NewTermboxCanvas()

	edA := layouts.NewStack(NewEditorPanel, NewEditorPanel())
	// edB := NewStack(NewEditorPanel, NewEditorPanel())
	//	eh := NewSideBySide(edA, edB)

	eh := layouts.NewShelf(func() events.Handler { return layouts.NewStack(NewEditorPanel, NewEditorPanel()) }, edA)

	eh.ResizeTo(page)
	screen.TheScreen.EnableMouse()

	for {
		screen.TheScreen.Clear()
		eh.Paint()
		eh.SetCursor()
		screen.TheScreen.Show()
		ev := screen.TheScreen.PollEvent()

		switch ev := ev.(type) {
		case *tcell.EventMouse:	
			eh.Mouse(ev)
		case *tcell.EventKey: eh.Key(ev)
			if ev.Key() ==  tcell.KeyCtrlX { return }
		case *tcell.EventResize: 
			page = screen.NewTermboxCanvas()
			eh.ResizeTo(page)
		}
	}
}
