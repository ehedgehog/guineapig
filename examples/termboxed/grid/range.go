package grid

type MarkedRange struct {
	firstMarkedLine int
	lastMarkedLine  int
}

func (mr *MarkedRange) Range() (first, last int) {
	if mr.firstMarkedLine == 0 {
		return -1, -1
	}
	return mr.firstMarkedLine - 1, mr.lastMarkedLine - 1
}

func (mr *MarkedRange) Clear() {
	mr.firstMarkedLine, mr.lastMarkedLine = 0, 0
}

func (mr *MarkedRange) IsActive() bool {
	return mr.firstMarkedLine > 0
}

func (mr *MarkedRange) SetLow(lineNumber int) {
	mr.firstMarkedLine = lineNumber + 1
	if mr.lastMarkedLine < mr.firstMarkedLine {
		mr.lastMarkedLine = mr.firstMarkedLine
	}
}

func (mr *MarkedRange) SetHigh(lineNumber int) {
	mr.lastMarkedLine = lineNumber + 1
	if mr.firstMarkedLine > mr.lastMarkedLine {
		mr.firstMarkedLine = mr.lastMarkedLine
	}
}

func (mr *MarkedRange) MoveAfter(target int) int {
	first, last := mr.Range()
	diff := last - first
	if target > last {
		target = target - (diff + 1)
	}
	mr.SetLow(target + 1)
	mr.SetHigh(target + 1 + diff)
	return target
}

func (mr *MarkedRange) Return(lineNumber int) {
	if mr.IsActive() {
		if lineNumber <= mr.lastMarkedLine {
			mr.lastMarkedLine += 1
		}
		if lineNumber < mr.firstMarkedLine {
			mr.firstMarkedLine += 1
		}
	}
}

func (mr *MarkedRange) RemoveLine(lineNumber int) {
	if mr.IsActive() {
		first, last := mr.Range()
		if lineNumber <= last {
			mr.lastMarkedLine -= 1
		}
		if lineNumber < first {
			mr.firstMarkedLine -= 1
		}
	}
}
