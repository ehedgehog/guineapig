package screen

type Writeable interface {
	PutString(x, y int, content string)
	SetCursor(x, y int)
}

type Panel interface {
	Writeable
}
