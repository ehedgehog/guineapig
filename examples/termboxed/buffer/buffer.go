package buffer

import (
	"fmt"

	"github.com/ehedgehog/guineapig/examples/termboxed/draw"
	"github.com/ehedgehog/guineapig/examples/termboxed/screen"
)

type Type interface {
	Insert(ch rune)
	DeleteBack()
	DeleteForward()
	BackOne()
	UpOne()
	DownOne()
	ForwardOne()
	Return()
	PutAll(w screen.Writeable)
	ScrollUp()
	ScrollDown()
	ScrollTop()
}

// SimpleBuffer is a simplistic implementation of
// Buffer. It burns store like it was November 5th.
type SimpleBuffer struct {
	content        []string
	line           int
	col            int
	verticalOffset int
	width          int
	height         int
}

func (b *SimpleBuffer) makeRoom() {
	if b.line >= len(b.content) {
		content := make([]string, b.line+1)
		copy(content, b.content)
		b.content = content
	}
	for b.col > len(b.content[b.line]) {
		b.content[b.line] += "        "
	}
}

func (b *SimpleBuffer) ScrollUp() {
	if b.verticalOffset > 0 {
		b.verticalOffset -= 1
		b.line -= 1
	}
}

func (b *SimpleBuffer) ScrollDown() {
	b.verticalOffset += 1
}

func (b *SimpleBuffer) ScrollTop() {
	b.line = 0
	b.verticalOffset = 0
}

func (b *SimpleBuffer) Insert(ch rune) {

	b.makeRoom()

	loc := b.col
	runes := []rune(b.content[b.line])

	A := []rune{}
	B := append(A, runes[0:loc]...)
	C := append(B, ch)
	D := append(C, runes[loc:]...)

	b.col += 1
	b.content[b.line] = string(D)
}

func (b *SimpleBuffer) Return() {

	b.makeRoom()

	lines := append(b.content, "")

	right := lines[b.line][b.col:]
	left := lines[b.line][0:b.col]

	copy(lines[b.line+1:], lines[b.line:])
	lines[b.line] = left
	lines[b.line+1] = right
	b.DownOne() // b.line += 1
	b.col = 0
	b.content = lines
}

func (b *SimpleBuffer) UpOne() {
	if b.line > 0 {
		if b.line == b.verticalOffset && b.verticalOffset > 0 {
			b.verticalOffset -= 1
		}
		b.line -= 1
	}
}

func (b *SimpleBuffer) DownOne() {
	b.line += 1
	if b.line-b.verticalOffset > b.height-3 {
		b.verticalOffset += 1
	}
}

func (b *SimpleBuffer) BackOne() {
	if b.col > 0 {
		b.col -= 1
	}
}

func (b *SimpleBuffer) ForwardOne() {
	b.col += 1
}

func (b *SimpleBuffer) DeleteBack() {
	b.makeRoom()
	if b.col > 0 {
		content := b.content[b.line]
		before := content[0 : b.col-1]
		after := content[b.col:]
		newContent := before + after
		b.content[b.line] = newContent
		b.BackOne()
	}
}

func (b *SimpleBuffer) DeleteForward() {
	b.ForwardOne()
	b.DeleteBack()
}

func min(x, y int) int {
	if x < y {
		return x
	} else {
		return y
	}
}

func (b *SimpleBuffer) PutAll(w screen.Writeable) {

	loc := draw.XY{0, 0}
	size := draw.WH{b.width, b.height}

	length := len(b.content)
	if b.line > length {
		length = b.line
	}
	off := draw.Scrolling{length, b.line}

	info := draw.BoxInfo{loc, size, off}

	draw.Box(fmt.Sprintf("offset: %v, line: %v, cursor(col %v, line %v), height: %v", b.verticalOffset, b.line, b.col+1, b.line+1-b.verticalOffset, b.height), info)

	vertical := b.height - 2
	limit := min(vertical, len(b.content)-b.verticalOffset)

	w.PutString(80, 0, fmt.Sprintf(" range [%v, %v] ", b.verticalOffset, limit))

	wLine := 1
	bLine := b.verticalOffset
	content := b.content
	for {
		if bLine >= 0 && bLine < len(content) {
			w.PutString(1, wLine, content[bLine])
		} else {
			// w.PutString(1, wLine, "?")
		}
		wLine += 1
		bLine += 1
		if bLine == len(content) || wLine > vertical {
			break
		}
	}
	w.SetCursor(b.col+1, b.line+1-b.verticalOffset)
}

func New(w, h int) Type {
	return &SimpleBuffer{content: []string{""}, width: w, height: h}
}
