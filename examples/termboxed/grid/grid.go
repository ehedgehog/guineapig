package grid

// LineCol is a location on a canvas or like surface.
type LineCol struct {
	Line int
	Col  int
}

func (where LineCol) ColPlus(dCol int) LineCol {
	return LineCol{Col: where.Col + dCol, Line: where.Line}
}

func (where LineCol) LineMinus(dRow int) LineCol {
	return LineCol{Col: where.Col, Line: where.Line - dRow}
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
