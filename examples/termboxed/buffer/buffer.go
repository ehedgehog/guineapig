package buffer

import (
	"bufio"
	"io"
	"os"
)

import "github.com/ehedgehog/guineapig/examples/termboxed/screen"
import "github.com/ehedgehog/guineapig/examples/termboxed/grid"

type Type interface {
	// Insert inserts the rune at the current position and moves right.
	Insert(ch rune)

	// DeleteLine(n) deletes line n from the buffer
	DeleteLine(line int)

	DeleteLines(lowLine, highLine int)

	// DeleteBack delete the previous rune if not at line start. Otherwise
	// it does nothing.
	DeleteBack()

	// DeleteForward delete the current rune if there are any runes
	// remaining on the current line. Otherwise it does nothing.
	DeleteForward()

	// BackOne moves left one rune if not at line start. Otherwise
	// it does nothing.
	BackOne()

	// UpOne moves up one line if not at first line, preserving the column.
	UpOne()

	// DownOne moves down one line, preserving the column.
	DownOne()

	// ForwardOne moves right one rune.
	ForwardOne()

	// Return inserts a newline (and hence a new line) at the current position.
	Return()

	Execute() error

	PutLines(c screen.Canvas, first, n int)

	// SetWhere sets the current position to be where.
	SetWhere(where grid.LineCol)

	// Where returns the current position
	Where() grid.LineCol

	// attempt to eliminate?
	Expose() (line int, content []string)

	// ReadFromFile reads from r inserting the content at the current position.
	ReadFromFile(fileName string, r io.Reader) error

	WriteToFile(fileName []string) error
}

// SimpleBuffer is a simplistic implementation of
// Buffer. It burns store like it was November 5th.
type SimpleBuffer struct {
	content  []string                 // existing lines of text
	where    grid.LineCol             // current location in buffer (line, column)
	execute  func(Type, string) error // execute command on buffer at line
	fileName string                   // file name used for most recent read
}

func (b *SimpleBuffer) Expose() (line int, content []string) {
	return b.where.Line, b.content
}

func (b *SimpleBuffer) DeleteLines(lowLine, highLine int) {
	if 0 <= lowLine && lowLine <= highLine && highLine <= len(b.content) {
		b.content = append(b.content[0:lowLine], b.content[highLine+1:]...)
		if b.where.Line >= lowLine {
			if b.where.Line <= highLine {
				b.where.Line = lowLine
			} else {
				b.where.Line = b.where.Line - (highLine - lowLine + 1)
			}
		}
	}
}

func (b *SimpleBuffer) DeleteLine(line int) {
	if line == 0 {
		b.content = b.content[1:]
	} else if line < len(b.content) {
		b.content = append(b.content[0:line], b.content[line+1:]...)
	} else {
		// nothing to do -- deleting virtual line
	}
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

func (b *SimpleBuffer) ReadFromFile(fileName string, r io.Reader) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		b.content = append(b.content, line)
	}
	b.where.Line = 0
	b.fileName = fileName
	return nil
}

func (b *SimpleBuffer) makeRoom() {
	line, col := b.where.Line, b.where.Col
	if line >= len(b.content) {
		content := make([]string, line+1)
		copy(content, b.content)
		b.content = content
	}
	for col > len(b.content[line]) {
		b.content[line] += "        "
	}
}

func (b *SimpleBuffer) Insert(ch rune) {

	b.makeRoom()

	loc := b.where.Col
	runes := []rune(b.content[b.where.Line])

	A := []rune{}
	B := append(A, runes[0:loc]...)
	C := append(B, ch)
	D := append(C, runes[loc:]...)

	b.where.Col += 1
	b.content[b.where.Line] = string(D)
}

func (b *SimpleBuffer) Execute() error {
	b.makeRoom()
	return b.execute(b, b.content[b.where.Line])
}

func (b *SimpleBuffer) Return() {

	b.makeRoom()

	lines := append(b.content, "")

	line, col := b.where.Line, b.where.Col
	right := lines[line][col:]
	left := lines[line][0:col]

	copy(lines[line+1:], lines[line:])
	lines[line] = left
	lines[line+1] = right
	b.DownOne() // b.line += 1
	b.where.Col = 0
	b.content = lines
}

func (b *SimpleBuffer) UpOne() {
	if b.where.Line > 0 {
		b.where.Line -= 1
	}
}

func (b *SimpleBuffer) DownOne() {
	b.where.Line += 1
}

func (b *SimpleBuffer) BackOne() {
	if b.where.Col > 0 {
		b.where.Col -= 1
	}
}

func (b *SimpleBuffer) ForwardOne() {
	b.where.Col += 1
}

func (b *SimpleBuffer) DeleteBack() {
	b.makeRoom()
	line, col := b.where.Line, b.where.Col
	if col > 0 {
		content := b.content[line]
		before := content[0 : col-1]
		after := content[col:]
		newContent := before + after
		b.content[line] = newContent
		b.BackOne()
	}
}

func (b *SimpleBuffer) DeleteForward() {
	b.ForwardOne()
	b.DeleteBack()
}

func (b *SimpleBuffer) PutLines(w screen.Canvas, first, n int) {
	content := b.content
	row := 0
	for line := first; line < len(content) && row < n; line += 1 {
		screen.PutString(w, 0, row, content[line], screen.DefaultStyle)
		row += 1
	}
}

func (s *SimpleBuffer) Where() grid.LineCol {
	return s.where
}

func (s *SimpleBuffer) SetWhere(where grid.LineCol) {
	s.where = where
}

func New(execute func(Type, string) error) Type {
	return &SimpleBuffer{
		content: []string{},
		execute: execute,
	}
}
