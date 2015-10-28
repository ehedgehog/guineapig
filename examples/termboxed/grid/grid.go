package grid

type Geometry struct {
	MinWidth  int
	MaxWidth  int
	MinHeight int
	MaxHeight int
}

// LineCol is a location on a canvas or like surface.
type LineCol struct {
	Line int
	Col  int
}

func (where *LineCol) UpOne() {
	if where.Line > 0 {
		where.Line -= 1
	}
}

func (where *LineCol) DownOne() {
	where.Line += 1
}

func (where *LineCol) LeftOne() {
	if where.Col > 0 {
		where.Col -= 1
	}
}

func (where *LineCol) RightOne() {
	where.Col += 1
}

func (where LineCol) ColMinus(dCol int) LineCol {
	return LineCol{Col: where.Col - dCol, Line: where.Line}
}

func (where LineCol) ColPlus(dCol int) LineCol {
	return LineCol{Col: where.Col + dCol, Line: where.Line}
}

func (where LineCol) LineMinus(dRow int) LineCol {
	return LineCol{Col: where.Col, Line: where.Line - dRow}
}

func (where LineCol) LinePlus(dRow int) LineCol {
	return LineCol{Col: where.Col, Line: where.Line + dRow}
}

func (where LineCol) Plus(offset LineCol) LineCol {
	return LineCol{Col: where.Col + offset.Col, Line: where.Line + offset.Line}
}

// Size is a width x height representation of the size of
// a surface.
type Size struct {
	Width  int
	Height int
}

type Offset struct {
	Vertical   int
	Horizontal int
}
