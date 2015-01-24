package buffer

import (
	"bufio"
	"errors"
	"io"
	"os"
)

import "github.com/ehedgehog/guineapig/examples/termboxed/screen"
import "github.com/ehedgehog/guineapig/examples/termboxed/grid"

type Type interface {
	// Insert inserts the rune at the current position and moves right.
	Insert(where grid.LineCol, ch rune) grid.LineCol

	// DeleteLine(n) deletes line n from the buffer
	DeleteLine(where grid.LineCol) grid.LineCol

	DeleteLines(where grid.LineCol, lowLine, highLine int) grid.LineCol

	// DeleteBack delete the previous rune if not at line start. Otherwise
	// it does nothing.
	DeleteBack(grid.LineCol) grid.LineCol

	// DeleteForward delete the current rune if there are any runes
	// remaining on the current line. Otherwise it does nothing.
	DeleteForward(grid.LineCol) grid.LineCol

	// BackOne moves left one rune if not at line start. Otherwise
	// it does nothing.
	BackOne(grid.LineCol) grid.LineCol

	// UpOne returns a LineCol one line up from where, or where
	// unmodified if it i at the first line.
	UpOne(where grid.LineCol) grid.LineCol

	// DownOne moves down one line, preserving the column.
	DownOne(where grid.LineCol) grid.LineCol

	// ForwardOne moves right one rune.
	ForwardOne(where grid.LineCol) grid.LineCol

	// Return inserts a newline (and hence a new line) at the current position.
	Return(grid.LineCol) grid.LineCol

	Execute(grid.LineCol) (grid.LineCol, error)

	PutLines(c screen.Canvas, first, n int)

	// SetWhere sets the current position to be where.
	// SetWhere(where grid.LineCol)

	// Where returns the current position
	// Where() grid.LineCol

	// attempt to eliminate?
	Expose() []string

	// ReadFromFile reads from r inserting the content at the current position.
	ReadFromFile(where grid.LineCol, fileName string, r io.Reader) (grid.LineCol, error)

	WriteToFile(fileName []string) error
}

// SimpleBuffer is a simplistic implementation of
// Buffer. It burns store like it was November 5th.
type SimpleBuffer struct {
	content []string // existing lines of text
	//	where    grid.LineCol             // current location in buffer (line, column)
	execute  func(Type, string) error // execute command on buffer at line
	fileName string                   // file name used for most recent read
}

func (b *SimpleBuffer) Expose() (content []string) {
	return b.content
}

func (b *SimpleBuffer) DeleteLines(where grid.LineCol, lowLine, highLine int) grid.LineCol {
	if 0 <= lowLine && lowLine <= highLine && highLine <= len(b.content) {
		b.content = append(b.content[0:lowLine], b.content[highLine+1:]...)
		if where.Line >= lowLine {
			if where.Line <= highLine {
				where.Line = lowLine
			} else {
				where.Line = where.Line - (highLine - lowLine + 1)
			}
		}
	}
	return where
}

func (b *SimpleBuffer) DeleteLine(where grid.LineCol) grid.LineCol {
	line := where.Line
	if line == 0 {
		b.content = b.content[1:]
	} else if line < len(b.content) {
		b.content = append(b.content[0:line], b.content[line+1:]...)
	} else {
		// nothing to do -- deleting virtual line
	}
	return where
}

func (b *SimpleBuffer) WriteToFile(fileNameOption []string) error {
	fileName := ""
	if len(fileNameOption) > 0 {
		fileName = fileNameOption[0]
	}
	if len(fileName) == 0 {
		fileName = b.fileName
	}
	f, err := os.Create(fileName)
	if err == nil {
		// horrid. couldn't we at least contrive to call WriteString?
		// or use bufio and buffer up the bytes?
		defer f.Close()
		for _, line := range b.content {
			f.Write([]byte(line))
			f.Write([]byte{'\n'})
		}
	} else {
		return err
	}
	return nil
}

func (b *SimpleBuffer) ReadFromFile(where grid.LineCol, fileName string, r io.Reader) (grid.LineCol, error) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		b.content = append(b.content, line)
	}
	where.Line = 0
	b.fileName = fileName
	return where, nil
}

func (b *SimpleBuffer) makeRoom(where grid.LineCol) {
	line, col := where.Line, where.Col
	if line >= len(b.content) {
		content := make([]string, line+1)
		copy(content, b.content)
		b.content = content
	}
	for col > len(b.content[line]) {
		b.content[line] += "        "
	}
}

func (b *SimpleBuffer) Insert(where grid.LineCol, ch rune) grid.LineCol {

	b.makeRoom(where)

	loc := where.Col
	runes := []rune(b.content[where.Line])

	A := []rune{}
	B := append(A, runes[0:loc]...)
	C := append(B, ch)
	D := append(C, runes[loc:]...)

	where.Col += 1
	b.content[where.Line] = string(D)

	return where
}

func (b *SimpleBuffer) Execute(where grid.LineCol) (grid.LineCol, error) {
	b.makeRoom(where)
	// return b.execute(b, b.content[where.Line])
	return where, errors.New("execute not implemented yet.")
}

func (b *SimpleBuffer) Return(where grid.LineCol) grid.LineCol {

	b.makeRoom(where)

	lines := append(b.content, "")

	line, col := where.Line, where.Col
	right := lines[line][col:]
	left := lines[line][0:col]

	copy(lines[line+1:], lines[line:])
	lines[line] = left
	lines[line+1] = right
	where = b.DownOne(where)
	where.Col = 0
	b.content = lines
	return where
}

func (b *SimpleBuffer) UpOne(where grid.LineCol) grid.LineCol {
	if where.Line > 0 {
		where.Line -= 1
	}
	return where
}

func (b *SimpleBuffer) DownOne(where grid.LineCol) grid.LineCol {
	where.Line += 1
	return where
}

func (b *SimpleBuffer) BackOne(where grid.LineCol) grid.LineCol {
	if where.Col > 0 {
		where.Col -= 1
	}
	return where
}

func (b *SimpleBuffer) ForwardOne(where grid.LineCol) grid.LineCol {
	where.Col += 1
	return where
}

func (b *SimpleBuffer) DeleteBack(where grid.LineCol) grid.LineCol {
	b.makeRoom(where)
	line, col := where.Line, where.Col
	if col > 0 {
		content := b.content[line]
		before := content[0 : col-1]
		after := content[col:]
		newContent := before + after
		b.content[line] = newContent
		return b.BackOne(where)
	}
	return where
}

func (b *SimpleBuffer) DeleteForward(where grid.LineCol) grid.LineCol {
	where = b.ForwardOne(where)
	return b.DeleteBack(where)
}

func (b *SimpleBuffer) PutLines(w screen.Canvas, first, n int) {
	content := b.content
	row := 0
	for line := first; line < len(content) && row < n; line += 1 {
		screen.PutString(w, 0, row, content[line], screen.DefaultStyle)
		row += 1
	}
}

// func (s *SimpleBuffer) Where() grid.LineCol {
// 	return s.where
//}

// func (s *SimpleBuffer) SetWhere(where grid.LineCol) {
// 	s.where = where
// }

func New(execute func(Type, string) error) Type {
	return &SimpleBuffer{
		content: []string{},
		execute: execute,
	}
}
