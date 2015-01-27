package buffer

import (
	"bufio"
	//	"errors"
	"io"
	"os"
)

import "github.com/ehedgehog/guineapig/examples/termboxed/screen"
import "github.com/ehedgehog/guineapig/examples/termboxed/grid"

type Type interface {
	// Insert inserts the rune at the current position.
	Insert(where grid.LineCol, ch rune)

	// DeleteLine(n) deletes line n from the buffer
	DeleteLine(where grid.LineCol) grid.LineCol

	MoveLines(where grid.LineCol, firstLine, lastLine int)

	DeleteLines(where grid.LineCol, lowLine, highLine int) grid.LineCol

	// DeleteBack delete the previous rune if not at line start. Otherwise
	// it does nothing.
	DeleteBack(grid.LineCol) grid.LineCol

	// DeleteForward delete the current rune if there are any runes
	// remaining on the current line. Otherwise it does nothing.
	DeleteForward(grid.LineCol) grid.LineCol

	// Return inserts a newline (and hence a new line) at the current position.
	Return(grid.LineCol) grid.LineCol

	Execute(grid.LineCol) (grid.LineCol, error)

	PutLines(c screen.Canvas, first, n int)

	// attempt to eliminate?
	Expose() []string

	// ReadFromFile reads from r inserting the content at the current position.
	ReadFromFile(where grid.LineCol, fileName string, r io.Reader) (grid.LineCol, error)

	WriteToFile(fileName []string) error
}

// SimpleBuffer is a simplistic implementation of
// Buffer. It burns store like it was November 5th.
type SimpleBuffer struct {
	content  []string                 // existing lines of text
	execute  func(Type, string) error // execute command on buffer at line
	fileName string                   // file name used for most recent read
}

func (b *SimpleBuffer) Expose() (content []string) {
	return b.content
}

func (b *SimpleBuffer) MoveLines(where grid.LineCol, firstLine, lastLine int) {
	lines := b.content
	target := where.Line
	newContent := make([]string, 0, len(lines))

	if target < firstLine {
		newContent = append(newContent, lines[0:target+1]...)
		newContent = append(newContent, lines[firstLine:lastLine+1]...)
		newContent = append(newContent, lines[target+1:firstLine]...)
		newContent = append(newContent, lines[lastLine+1:]...)

	} else if target > lastLine {
		panic("target > lastLine, not implemented")
	} else {
		panic("target within range")
	}

	b.content = newContent
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

func (b *SimpleBuffer) Insert(where grid.LineCol, ch rune) {

	b.makeRoom(where)

	loc := where.Col
	runes := []rune(b.content[where.Line])

	A := []rune{}
	B := append(A, runes[0:loc]...)
	C := append(B, ch)
	D := append(C, runes[loc:]...)

	b.content[where.Line] = string(D)
}

func (b *SimpleBuffer) Execute(where grid.LineCol) (grid.LineCol, error) {
	b.makeRoom(where)
	return where, b.execute(b, b.content[where.Line])
}

func New(execute func(Type, string) error) Type {
	return &SimpleBuffer{
		content: []string{},
		execute: execute,
	}
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
	where.DownOne()
	where.Col = 0
	b.content = lines
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
		where.LeftOne()
	}
	return where
}

func (b *SimpleBuffer) DeleteForward(where grid.LineCol) grid.LineCol {
	where.RightOne()
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
