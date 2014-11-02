package buffer

import (
	"bufio"
	"io"
	"os"
)

import "github.com/ehedgehog/guineapig/examples/termboxed/screen"
import "github.com/ehedgehog/guineapig/examples/termboxed/grid"

type Type interface {
	Insert(ch rune)
	DeleteBack()
	DeleteForward()
	BackOne()
	UpOne()
	DownOne()
	ForwardOne()
	Return()
	Execute()
	PutLines(c screen.Canvas, first, n int)
	SetWhere(where grid.LineCol)
	Where() grid.LineCol
	Expose() (line int, content []string) // attempt to eliminate?
	ReadFromFile(fileName string, r io.Reader)
	WriteToFile(fileName []string)
}

// SimpleBuffer is a simplistic implementation of
// Buffer. It burns store like it was November 5th.
type SimpleBuffer struct {
	content  []string           // existing lines of text
	where    grid.LineCol       // current location in buffer (line, column)
	execute  func(Type, string) // execute command on buffer at line
	fileName string             // file name used for most recent read
}

func (b *SimpleBuffer) Expose() (line int, content []string) {
	return b.where.Line, b.content
}

func (b *SimpleBuffer) WriteToFile(fileNameOption []string) {
	fileName := b.fileName
	if len(fileNameOption) > 0 {
		fileName = fileNameOption[0]
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
	}
}

func (b *SimpleBuffer) ReadFromFile(fileName string, r io.Reader) {
	/*
		var x bytes.Buffer
		x.ReadFrom(r)
		all := x.Bytes()
		for len(all) > 0 {
			ch, size := utf8.DecodeRune(all)
			if ch == '\n' {
				b.Return()
			} else {
				b.Insert(ch)
			}
			all = all[size:]
		}
	*/
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		b.content = append(b.content, line)
	}
	b.where.Line = 0
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

func (b *SimpleBuffer) Execute() {
	b.makeRoom()
	b.execute(b, b.content[b.where.Line])
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

func New(execute func(Type, string)) Type {
	return &SimpleBuffer{
		content: []string{},
		execute: execute,
	}
}
