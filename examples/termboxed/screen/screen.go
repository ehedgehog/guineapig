package screen

type Writeable interface {
	PutString(x, y int, content string)
	SetCursor(x, y int)
	Resize(x, y, w, h int)
}

type Panel interface {
	Writeable
}
