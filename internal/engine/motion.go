package engine

import (
	"strings"
	"unicode"
)

// Motion represents a cursor movement.
type Motion struct {
	Key   string
	Count int
	Char  rune // for f/t/F/T
}

// Range represents a text range with start and end positions.
type Range struct {
	L1, C1, L2, C2 int
	Linewise        bool
}

// ApplyMotion moves the cursor according to the motion.
// Returns the range swept if this is used with an operator.
func (b *Buffer) ApplyMotion(m Motion) Range {
	startLine, startCol := b.cursor.Line, b.cursor.Col
	count := m.Count
	if count <= 0 {
		count = 1
	}

	for i := 0; i < count; i++ {
		b.applyMotionOnce(m)
	}

	endLine, endCol := b.cursor.Line, b.cursor.Col

	// Determine range direction
	// Range end is exclusive: [C1, C2)
	if endLine < startLine || (endLine == startLine && endCol < startCol) {
		return Range{endLine, endCol, startLine, startCol, false}
	}

	r := Range{startLine, startCol, endLine, endCol, false}

	// Some motions are linewise
	switch m.Key {
	case "gg", "G", "j", "k", "+", "-":
		r = Range{startLine, 0, endLine, len(b.lines[endLine]), true}
		if startLine > endLine {
			r = Range{endLine, 0, startLine, len(b.lines[startLine]), true}
		}
	}

	return r
}

func (b *Buffer) applyMotionOnce(m Motion) {
	line := b.cursor.Line
	col := b.cursor.Col
	lineLen := len(b.lines[line])

	switch m.Key {
	case "h":
		if col > 0 {
			b.cursor.Col--
		}

	case "l":
		max := maxCol(b.lines[line], b.mode)
		if col < max {
			b.cursor.Col++
		}

	case "j":
		if line < len(b.lines)-1 {
			b.cursor.Line++
			b.clampCol()
		}

	case "k":
		if line > 0 {
			b.cursor.Line--
			b.clampCol()
		}

	case "0", "^":
		if m.Key == "0" {
			b.cursor.Col = 0
		} else {
			// ^ goes to first non-blank
			b.cursor.Col = firstNonBlank(b.lines[line])
		}

	case "$":
		if lineLen > 0 {
			b.cursor.Col = lineLen - 1
		} else {
			b.cursor.Col = 0
		}

	case "w":
		b.cursor.Line, b.cursor.Col = nextWordStart(b.lines, line, col, false)

	case "W":
		b.cursor.Line, b.cursor.Col = nextWordStart(b.lines, line, col, true)

	case "b":
		b.cursor.Line, b.cursor.Col = prevWordStart(b.lines, line, col, false)

	case "B":
		b.cursor.Line, b.cursor.Col = prevWordStart(b.lines, line, col, true)

	case "e":
		b.cursor.Line, b.cursor.Col = nextWordEnd(b.lines, line, col, false)

	case "E":
		b.cursor.Line, b.cursor.Col = nextWordEnd(b.lines, line, col, true)

	case "gg":
		b.cursor.Line = 0
		b.cursor.Col = firstNonBlank(b.lines[0])

	case "G":
		last := len(b.lines) - 1
		b.cursor.Line = last
		b.cursor.Col = firstNonBlank(b.lines[last])

	case "f":
		if m.Char != 0 {
			newCol := findChar(b.lines[line], col+1, m.Char, true)
			if newCol >= 0 {
				b.cursor.Col = newCol
				b.lastFind = m.Char
				b.lastFindDir = 1
				b.lastFindTill = false
			}
		}

	case "F":
		if m.Char != 0 {
			newCol := findChar(b.lines[line], col-1, m.Char, false)
			if newCol >= 0 {
				b.cursor.Col = newCol
				b.lastFind = m.Char
				b.lastFindDir = -1
				b.lastFindTill = false
			}
		}

	case "t":
		if m.Char != 0 {
			newCol := findChar(b.lines[line], col+1, m.Char, true)
			if newCol >= 0 {
				b.cursor.Col = newCol - 1
				b.lastFind = m.Char
				b.lastFindDir = 1
				b.lastFindTill = true
			}
		}

	case "T":
		if m.Char != 0 {
			newCol := findChar(b.lines[line], col-1, m.Char, false)
			if newCol >= 0 {
				b.cursor.Col = newCol + 1
				b.lastFind = m.Char
				b.lastFindDir = -1
				b.lastFindTill = true
			}
		}

	case ";":
		// repeat last find
		if b.lastFind != 0 {
			fm := Motion{Char: b.lastFind}
			if b.lastFindDir == 1 {
				if b.lastFindTill {
					fm.Key = "t"
				} else {
					fm.Key = "f"
				}
			} else {
				if b.lastFindTill {
					fm.Key = "T"
				} else {
					fm.Key = "F"
				}
			}
			b.applyMotionOnce(fm)
		}

	case ",":
		// reverse last find
		if b.lastFind != 0 {
			fm := Motion{Char: b.lastFind}
			if b.lastFindDir == -1 {
				if b.lastFindTill {
					fm.Key = "t"
				} else {
					fm.Key = "f"
				}
			} else {
				if b.lastFindTill {
					fm.Key = "T"
				} else {
					fm.Key = "F"
				}
			}
			b.applyMotionOnce(fm)
		}

	case "%":
		b.matchBracket()

	case "g_":
		// end of line non-blank
		b.cursor.Col = lastNonBlank(b.lines[line])
	}
}

// --- Word motion helpers ---

func isWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

func nextWordStart(lines [][]rune, line, col int, bigword bool) (int, int) {
	l := lines[line]
	col++
	for col >= len(l) {
		line++
		if line >= len(lines) {
			return len(lines) - 1, len(lines[len(lines)-1]) - 1
		}
		l = lines[line]
		col = 0
		if len(l) > 0 {
			break
		}
	}
	// skip rest of current word class
	if col < len(l) {
		startClass := charClass(l[col-1], bigword)
		for col < len(l) && charClass(l[col], bigword) == startClass {
			col++
		}
	}
	// skip whitespace
	for col < len(l) && unicode.IsSpace(l[col]) {
		col++
	}
	if col >= len(l) {
		// move to next line
		for {
			line++
			if line >= len(lines) {
				line = len(lines) - 1
				col = max(0, len(lines[line])-1)
				return line, col
			}
			l = lines[line]
			col = 0
			if len(l) == 0 {
				return line, 0
			}
			for col < len(l) && unicode.IsSpace(l[col]) {
				col++
			}
			if col < len(l) {
				return line, col
			}
		}
	}
	return line, col
}

func prevWordStart(lines [][]rune, line, col int, bigword bool) (int, int) {
	col--
	if col < 0 {
		line--
		if line < 0 {
			return 0, 0
		}
		col = len(lines[line]) - 1
	}
	l := lines[line]
	// skip whitespace backwards
	for col >= 0 && unicode.IsSpace(l[col]) {
		col--
		if col < 0 {
			line--
			if line < 0 {
				return 0, 0
			}
			l = lines[line]
			col = len(l) - 1
		}
	}
	// find start of this word
	if col >= 0 {
		wClass := charClass(l[col], bigword)
		for col > 0 && charClass(l[col-1], bigword) == wClass {
			col--
		}
	}
	return line, max(0, col)
}

func nextWordEnd(lines [][]rune, line, col int, bigword bool) (int, int) {
	col++
	l := lines[line]
	// skip whitespace
	for col < len(l) && unicode.IsSpace(l[col]) {
		col++
	}
	if col >= len(l) {
		line++
		if line >= len(lines) {
			return len(lines) - 1, len(lines[len(lines)-1]) - 1
		}
		l = lines[line]
		col = 0
		for col < len(l) && unicode.IsSpace(l[col]) {
			col++
		}
	}
	// move to end of word
	if col < len(l) {
		wClass := charClass(l[col], bigword)
		for col+1 < len(l) && charClass(l[col+1], bigword) == wClass {
			col++
		}
	}
	return line, col
}

type charType int

const (
	charSpace charType = iota
	charWord
	charPunct
)

func charClass(r rune, bigword bool) charType {
	if unicode.IsSpace(r) {
		return charSpace
	}
	if bigword {
		return charWord
	}
	if isWordChar(r) {
		return charWord
	}
	return charPunct
}

func firstNonBlank(line []rune) int {
	for i, r := range line {
		if !unicode.IsSpace(r) {
			return i
		}
	}
	return 0
}

func lastNonBlank(line []rune) int {
	for i := len(line) - 1; i >= 0; i-- {
		if !unicode.IsSpace(line[i]) {
			return i
		}
	}
	return 0
}

func findChar(line []rune, start int, ch rune, forward bool) int {
	if forward {
		for i := start; i < len(line); i++ {
			if line[i] == ch {
				return i
			}
		}
	} else {
		for i := start; i >= 0; i-- {
			if line[i] == ch {
				return i
			}
		}
	}
	return -1
}

// matchBracket jumps to matching bracket.
func (b *Buffer) matchBracket() {
	line := b.lines[b.cursor.Line]
	col := b.cursor.Col
	if col >= len(line) {
		return
	}
	ch := line[col]
	var open, close rune
	var forward bool
	switch ch {
	case '(':
		open, close, forward = '(', ')', true
	case ')':
		open, close, forward = '(', ')', false
	case '[':
		open, close, forward = '[', ']', true
	case ']':
		open, close, forward = '[', ']', false
	case '{':
		open, close, forward = '{', '}', true
	case '}':
		open, close, forward = '{', '}', false
	default:
		return
	}

	depth := 0
	if forward {
		for l := b.cursor.Line; l < len(b.lines); l++ {
			startC := 0
			if l == b.cursor.Line {
				startC = col
			}
			for c := startC; c < len(b.lines[l]); c++ {
				r := b.lines[l][c]
				if r == open {
					depth++
				} else if r == close {
					depth--
					if depth == 0 {
						b.cursor.Line = l
						b.cursor.Col = c
						return
					}
				}
			}
		}
	} else {
		for l := b.cursor.Line; l >= 0; l-- {
			endC := len(b.lines[l]) - 1
			if l == b.cursor.Line {
				endC = col
			}
			for c := endC; c >= 0; c-- {
				r := b.lines[l][c]
				if r == close {
					depth++
				} else if r == open {
					depth--
					if depth == 0 {
						b.cursor.Line = l
						b.cursor.Col = c
						return
					}
				}
			}
		}
	}
}

// TextObjectRange computes the range for a text object like iw, aw, i", a(, etc.
func (b *Buffer) TextObjectRange(obj string) Range {
	line := b.cursor.Line
	col := b.cursor.Col
	l := b.lines[line]

	inner := strings.HasPrefix(obj, "i")
	var delim rune
	switch {
	case obj == "iw" || obj == "aw":
		return b.wordObjectRange(line, col, obj == "aw")
	case obj == "iW" || obj == "aW":
		return b.bigWordObjectRange(line, col, obj == "aW")
	case obj == "il" || obj == "al":
		// line object
		return Range{line, 0, line, len(l), true}
	default:
		// delimiter objects
		switch obj[1] {
		case '"':
			delim = '"'
		case '\'':
			delim = '\''
		case '`':
			delim = '`'
		case '(':
			delim = '('
		case ')':
			delim = ')'
		case '[':
			delim = '['
		case ']':
			delim = ']'
		case '{':
			delim = '{'
		case '}':
			delim = '}'
		case '<':
			delim = '<'
		case '>':
			delim = '>'
		default:
			return Range{line, col, line, col + 1, false}
		}
	}

	return b.delimObjectRange(line, col, delim, inner)
}

func (b *Buffer) wordObjectRange(line, col int, around bool) Range {
	l := b.lines[line]
	if col >= len(l) {
		return Range{line, col, line, col, false}
	}

	wClass := charClass(l[col], false)
	start := col
	end := col

	for start > 0 && charClass(l[start-1], false) == wClass {
		start--
	}
	for end < len(l)-1 && charClass(l[end+1], false) == wClass {
		end++
	}

	if around {
		// include trailing whitespace
		if end+1 < len(l) && unicode.IsSpace(l[end+1]) {
			for end+1 < len(l) && unicode.IsSpace(l[end+1]) {
				end++
			}
		} else if start > 0 && unicode.IsSpace(l[start-1]) {
			for start > 0 && unicode.IsSpace(l[start-1]) {
				start--
			}
		}
	}

	return Range{line, start, line, end + 1, false}
}

func (b *Buffer) bigWordObjectRange(line, col int, around bool) Range {
	l := b.lines[line]
	if col >= len(l) {
		return Range{line, col, line, col, false}
	}

	start := col
	end := col

	if !unicode.IsSpace(l[col]) {
		for start > 0 && !unicode.IsSpace(l[start-1]) {
			start--
		}
		for end < len(l)-1 && !unicode.IsSpace(l[end+1]) {
			end++
		}
	}

	if around {
		if end+1 < len(l) && unicode.IsSpace(l[end+1]) {
			for end+1 < len(l) && unicode.IsSpace(l[end+1]) {
				end++
			}
		}
	}

	return Range{line, start, line, end + 1, false}
}

func (b *Buffer) delimObjectRange(line, col int, delim rune, inner bool) Range {
	l := b.lines[line]

	var open, close rune
	switch delim {
	case '(':
		open, close = '(', ')'
	case '[':
		open, close = '[', ']'
	case '{':
		open, close = '{', '}'
	case '<':
		open, close = '<', '>'
	default:
		// symmetric delimiter: find nearest pair on current line
		first := -1
		second := -1
		for i, r := range l {
			if r == delim {
				if first == -1 {
					first = i
				} else {
					second = i
					if first <= col && col <= second {
						if inner {
							return Range{line, first + 1, line, second, false}
						}
						return Range{line, first, line, second + 1, false}
					}
					first = second
					second = -1
				}
			}
		}
		return Range{line, col, line, col + 1, false}
	}

	// balanced bracket search
	depth := 0
	startC := -1
	// search backward from cursor for opening bracket
	for i := col; i >= 0; i-- {
		r := l[i]
		if r == close {
			depth++
		} else if r == open {
			if depth == 0 {
				startC = i
				break
			}
			depth--
		}
	}
	if startC == -1 {
		return Range{line, col, line, col + 1, false}
	}
	// search forward for matching close
	depth = 0
	endC := -1
	for i := startC; i < len(l); i++ {
		r := l[i]
		if r == open {
			depth++
		} else if r == close {
			depth--
			if depth == 0 {
				endC = i
				break
			}
		}
	}
	if endC == -1 {
		return Range{line, col, line, col + 1, false}
	}

	if inner {
		return Range{line, startC + 1, line, endC, false}
	}
	return Range{line, startC, line, endC + 1, false}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
