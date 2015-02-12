package main

import (
	"errors"
	"fmt"
	// 	"log"
	"os"
	//	"strconv"
	"strings"
)

import termbox "github.com/limetext/termbox-go"

import "github.com/ehedgehog/guineapig/examples/termboxed/bounds"
import "github.com/ehedgehog/guineapig/examples/termboxed/buffer"
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
	buffer buffer.Type
	marked grid.MarkedRange
	offset Offset
}

type EditorPanel struct {
	topBar    screen.Canvas
	bottomBar screen.Canvas
	leftBar   screen.Canvas
	rightBar  screen.Canvas
	textBox   screen.Canvas

	current *State
	main    State
	command State
}

func (ep *EditorPanel) Geometry() grid.Geometry {
	minw := 2
	maxw := 1000
	minh := 2
	maxh := 1000
	return grid.Geometry{MinWidth: minw, MaxWidth: maxw, MinHeight: minh, MaxHeight: maxh}
}

func readIntoBuffer(ep *EditorPanel, b buffer.Type, fileName string) error {
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
	mb := buffer.New(func(b buffer.Type, s string) error { return nil })
	var ep *EditorPanel
	ep = &EditorPanel{
		main: State{buffer: mb},

		command: State{buffer: buffer.New(func(b buffer.Type, s string) error {
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

func (ep *EditorPanel) Key(e *termbox.Event) error {
	b := ep.current.buffer
	if e.Ch == 0 {
		switch e.Key {

		case 0:
			// nothing

		case termbox.KeyF1:
			ep.current = &ep.command
			ep.command.buffer.Return(ep.command.where)
			ep.current.where = grid.LineCol{Line: ep.command.where.Line + 1, Col: 0}

		case termbox.KeyF2:
			ep.command.where, _ = ep.command.buffer.Execute(ep.command.where)

		case termbox.KeyCtrlB:
			if ep.current == &ep.main {
				ep.current = &ep.command
			} else {
				ep.current = &ep.main
			}

		case termbox.KeySpace:
			b.Insert(ep.current.where, ' ')
			ep.current.where.RightOne()

		case termbox.KeyBackspace2:
			ep.current.where = b.DeleteBack(ep.current.where)

		case termbox.KeyDelete:
			ep.current.where = b.DeleteForward(ep.current.where)

		case termbox.KeyF3:
			ep.main.marked.SetLow(ep.main.where.Line)

		case termbox.KeyF4:
			ep.main.marked.SetHigh(ep.main.where.Line)

		case termbox.KeyPgup:
			where := ep.current.where
			vo := ep.current.offset.vertical
			if where.Line-vo == 0 {
				top := bounds.Max(0, where.Line-ep.textBox.Size().Height)
				ep.current.where = grid.LineCol{top, where.Col}
			} else {
				ep.current.where = grid.LineCol{vo, where.Col}
			}

		case termbox.KeyPgdn:
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

		case termbox.KeyEnd:
			where := ep.current.where
			if where.Col == 0 {
				contents := b.Expose()
				line := contents[ep.current.where.Line]
				where.Col = len(line)
			} else {
				where.Col = 0
			}
			ep.current.where = where

		case termbox.KeyEnter:
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

		case termbox.KeyArrowRight:
			ep.current.where.RightOne()

		case termbox.KeyArrowUp:
			ep.current.where.UpOne()

		case termbox.KeyArrowDown:
			ep.current.where.DownOne()

		case termbox.KeyArrowLeft:
			ep.current.where.LeftOne()

		default:
			report := fmt.Sprintf("<key: %#d>\n", uint(e.Key))
			for _, ch := range report {
				b.Insert(ep.current.where, rune(ch))
				ep.current.where.RightOne()
			}
		}
	} else {
		b.Insert(ep.current.where, e.Ch)
		ep.current.where.RightOne()
	}
	return nil
}

func report(ep *EditorPanel, b buffer.Type, message string) {
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

func (ep *EditorPanel) Mouse(e *termbox.Event) error {
	x, y := e.MouseX, e.MouseY
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
	bottomSize := ep.bottomBar.Size()
	w := bottomSize.Width
	content := ep.main.buffer.Expose()
	line := ep.current.where.Line
	textBoxSize := ep.textBox.Size()
	textHeight := textBoxSize.Height
	ep.main.buffer.PutLines(ep.textBox, ep.main.offset.vertical, textHeight)
	//
	ep.bottomBar.SetCell(grid.LineCol{Col: 0, Line: 0}, draw.Glyph_corner_bl, screen.DefaultStyle)
	for i := 1; i < w; i += 1 {
		ep.bottomBar.SetCell(grid.LineCol{Col: i, Line: 0}, draw.Glyph_hbar, screen.DefaultStyle)
	}
	ep.bottomBar.SetCell(grid.LineCol{Col: w - 1, Line: 0}, draw.Glyph_corner_br, screen.DefaultStyle)
	//
	leftBarSize := ep.leftBar.Size()
	lh := leftBarSize.Height
	for j := 0; j < lh; j += 1 {
		ep.leftBar.SetCell(grid.LineCol{Col: 0, Line: j}, draw.Glyph_vbar, screen.DefaultStyle)
	}
	//
	ep.topBar.SetCell(grid.LineCol{Col: 0, Line: 0}, draw.Glyph_corner_tl, screen.DefaultStyle)
	for i := 1; i < w; i += 1 {
		ep.topBar.SetCell(grid.LineCol{Col: i, Line: 0}, draw.Glyph_hbar, screen.DefaultStyle)
	}
	screen.PutString(ep.topBar, 2, 0, "─┤ ", screen.DefaultStyle)
	ep.topBar.SetCell(grid.LineCol{Col: w - 1, Line: 0}, draw.Glyph_corner_tr, screen.DefaultStyle)
	//
	// HACK -- shouldn't need to remake each time
	tline := ep.command.where.Line
	ep.command.buffer.PutLines(screen.NewSubCanvas(ep.topBar, delta, 0, w-delta-2, 1), tline, 1)
	//
	length := bounds.Max(line, len(content))
	draw.Scrollbar(ep.rightBar, draw.ScrollInfo{length, line})
	//
	return nil
}

const delta = 5

func (eh *EditorPanel) ResizeTo(outer screen.Canvas) error {
	size := outer.Size()
	w, h := size.Width, size.Height
	eh.leftBar = screen.NewSubCanvas(outer, 0, 1, 1, h-2)
	eh.rightBar = screen.NewSubCanvas(outer, w-1, 1, 1, h-2)
	eh.topBar = screen.NewSubCanvas(outer, 0, 0, w, 1)
	eh.bottomBar = screen.NewSubCanvas(outer, 0, h-1, w, 1)
	// eh.textBox = screen.NewSubCanvas(outer, 1, 1, w-2, h-2)
	eh.textBox = NewTextBox(eh, outer, 1, 1, w-2, h-2)
	return nil
}

const tryTagSize = 6

func NewTextBox(ep *EditorPanel, outer screen.Canvas, dx, dy, w, h int) screen.Canvas {
	sub := screen.NewSubCanvas(outer, dx, dy, w, h)
	return &TextBox{tagSize: tryTagSize, ep: ep, SubCanvas: *sub.(*screen.SubCanvas)}
}

type TextBox struct {
	tagSize int
	ep      *EditorPanel
	screen.SubCanvas
}

var markStyle = screen.MakeStyle(termbox.ColorDefault, termbox.ColorYellow)

var hereStyle = screen.MakeStyle(termbox.ColorRed, termbox.ColorDefault)

func (t *TextBox) SetCell(where grid.LineCol, ch rune, s screen.Style) {
	if where.Col == 0 {
		ep := t.ep

		verticalOffset := ep.main.offset.vertical

		numberStyle := screen.DefaultStyle
		if where.Line+verticalOffset == ep.main.where.Line {
			numberStyle = hereStyle
		}

		first, last := ep.main.marked.Range()
		if first-verticalOffset <= where.Line && where.Line <= last-verticalOffset {
			t.SubCanvas.SetCell(grid.LineCol{where.Line, tryTagSize - 1}, ' ', markStyle)
		}
		s := fmt.Sprintf("%4v", where.Line+verticalOffset)
		for i, ch := range s {
			t.SubCanvas.SetCell(grid.LineCol{where.Line, i}, ch, numberStyle)
		}
		//	for i := 0; i < t.tagSize; i += 1 {
		//		t.SubCanvas.SetCell(grid.LineCol{where.Line, i}, '_', s)
		//	}
	}
	t.SubCanvas.SetCell(where.ColPlus(t.tagSize), ch, s)
}

func (t *TextBox) SetCursor(where grid.LineCol) {
	t.SubCanvas.SetCursor(where.ColPlus(t.tagSize))
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

func makeColours() []termbox.RGB {
	result := termbox.Palette256 // make([]termbox.RGB, 256)
	result[termbox.ColorBlack] = termbox.RGB{0, 0, 0}
	result[termbox.ColorRed] = termbox.RGB{255, 0, 0}
	result[termbox.ColorGreen] = termbox.RGB{0, 255, 0}
	result[termbox.ColorYellow] = termbox.RGB{255, 255, 0}
	result[termbox.ColorBlue] = termbox.RGB{0, 0, 255}
	result[termbox.ColorMagenta] = termbox.RGB{255, 0, 255}
	result[termbox.ColorCyan] = termbox.RGB{0, 255, 255}
	result[termbox.ColorWhite] = termbox.RGB{255, 255, 255}
	return result
}

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	termbox.SetColorMode(termbox.ColorMode256)
	termbox.SetColorPalette(makeColours())
	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)

	page := screen.NewTermboxCanvas()

	edA := layouts.NewStack(NewEditorPanel, NewEditorPanel())
	// edB := NewStack(NewEditorPanel, NewEditorPanel())
	//	eh := NewSideBySide(edA, edB)

	eh := layouts.NewShelf(func() events.Handler { return layouts.NewStack(NewEditorPanel, NewEditorPanel()) }, edA)

	eh.ResizeTo(page)

	for {
		eh.Paint()
		eh.SetCursor()
		termbox.Flush()
		ev := termbox.PollEvent()
		if ev.Type == termbox.EventKey && ev.Key == termbox.KeyCtrlX {
			return
		}
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		if ev.Type == termbox.EventMouse {
			eh.Mouse(&ev)
		}
		if ev.Type == termbox.EventKey {
			eh.Key(&ev)
		}
		if ev.Type == termbox.EventResize {
			page = screen.NewTermboxCanvas()
			eh.ResizeTo(page)
		}
	}
}
