package engine

import (
	"strings"
	"unicode/utf8"
)

// Mode represents vim editor modes.
type Mode int

const (
	ModeNormal Mode = iota
	ModeInsert
	ModeVisual
	ModeVisualLine
)

func (m Mode) String() string {
	switch m {
	case ModeNormal:
		return "NORMAL"
	case ModeInsert:
		return "INSERT"
	case ModeVisual:
		return "VISUAL"
	case ModeVisualLine:
		return "V-LINE"
	default:
		return "NORMAL"
	}
}

// Cursor holds the line and column position (0-indexed).
type Cursor struct {
	Line int
	Col  int
}

// Buffer is the core text buffer with cursor state.
type Buffer struct {
	lines  [][]rune
	cursor Cursor
	mode   Mode
	// yank registers: '"' is default, 'a'-'z' are named
	registers map[rune]string
	// last find character for f/t/F/T repeating with ; and ,
	lastFind     rune
	lastFindDir  int // 1 = forward, -1 = backward
	lastFindTill bool
}

// NewBuffer creates a buffer with the given initial text.
func NewBuffer(text string) *Buffer {
	b := &Buffer{
		registers: make(map[rune]string),
	}
	b.SetText(text)
	return b
}

// SetText replaces the buffer content and resets cursor.
func (b *Buffer) SetText(text string) {
	lines := strings.Split(text, "\n")
	b.lines = make([][]rune, len(lines))
	for i, l := range lines {
		b.lines[i] = []rune(l)
	}
	b.cursor = Cursor{0, 0}
}

// SetCursor sets cursor position, clamping to valid range.
func (b *Buffer) SetCursor(line, col int) {
	b.cursor.Line = clamp(line, 0, len(b.lines)-1)
	b.clampCol()
	if col >= 0 {
		b.cursor.Col = clamp(col, 0, maxCol(b.lines[b.cursor.Line], b.mode))
	}
}

// Text returns the full buffer content as a string.
func (b *Buffer) Text() string {
	parts := make([]string, len(b.lines))
	for i, l := range b.lines {
		parts[i] = string(l)
	}
	return strings.Join(parts, "\n")
}

// Snapshot returns a canonical string for comparison (same as Text for now).
func (b *Buffer) Snapshot() string {
	return b.Text()
}

// Cursor returns the current cursor position.
func (b *Buffer) GetCursor() Cursor {
	return b.cursor
}

// Mode returns the current mode.
func (b *Buffer) Mode() Mode {
	return b.mode
}

// SetMode sets the buffer mode and adjusts cursor if needed.
func (b *Buffer) SetMode(m Mode) {
	b.mode = m
	b.clampCol()
}

// LineCount returns the number of lines.
func (b *Buffer) LineCount() int {
	return len(b.lines)
}

// Line returns the content of line i as a string.
func (b *Buffer) Line(i int) string {
	if i < 0 || i >= len(b.lines) {
		return ""
	}
	return string(b.lines[i])
}

// CurrentLine returns the current line content.
func (b *Buffer) CurrentLine() string {
	return b.Line(b.cursor.Line)
}

// RuneAt returns the rune at (line, col).
func (b *Buffer) RuneAt(line, col int) rune {
	if line < 0 || line >= len(b.lines) {
		return 0
	}
	l := b.lines[line]
	if col < 0 || col >= len(l) {
		return 0
	}
	return l[col]
}

// CurrentRune returns the rune under the cursor.
func (b *Buffer) CurrentRune() rune {
	return b.RuneAt(b.cursor.Line, b.cursor.Col)
}

// --- Low-level text mutations ---

// insertRune inserts a rune at cursor position and advances cursor.
func (b *Buffer) insertRune(r rune) {
	l := b.lines[b.cursor.Line]
	col := b.cursor.Col
	newLine := make([]rune, len(l)+1)
	copy(newLine, l[:col])
	newLine[col] = r
	copy(newLine[col+1:], l[col:])
	b.lines[b.cursor.Line] = newLine
	b.cursor.Col++
}

// insertNewlineBelow inserts a new empty line below cursor and moves down.
func (b *Buffer) insertNewlineBelow() {
	line := b.cursor.Line
	rest := b.lines[line][b.cursor.Col:]
	b.lines[line] = b.lines[line][:b.cursor.Col]
	newLine := make([]rune, len(rest))
	copy(newLine, rest)
	b.lines = append(b.lines[:line+1], append([][]rune{newLine}, b.lines[line+1:]...)...)
	b.cursor.Line++
	b.cursor.Col = 0
}

// deleteRune deletes the rune at cursor position.
func (b *Buffer) deleteRune() {
	l := b.lines[b.cursor.Line]
	if len(l) == 0 {
		return
	}
	col := b.cursor.Col
	if col >= len(l) {
		col = len(l) - 1
	}
	b.lines[b.cursor.Line] = append(l[:col], l[col+1:]...)
	b.clampCol()
}

// deleteLineRange deletes lines [start, end] inclusive.
func (b *Buffer) deleteLineRange(start, end int) string {
	if start < 0 {
		start = 0
	}
	if end >= len(b.lines) {
		end = len(b.lines) - 1
	}
	deleted := make([]string, end-start+1)
	for i := start; i <= end; i++ {
		deleted[i-start] = string(b.lines[i])
	}
	b.lines = append(b.lines[:start], b.lines[end+1:]...)
	if len(b.lines) == 0 {
		b.lines = [][]rune{{}}
	}
	b.cursor.Line = clamp(start, 0, len(b.lines)-1)
	b.clampCol()
	return strings.Join(deleted, "\n")
}

// deleteRange deletes text from (l1,c1) to (l2,c2) exclusive end.
// Returns the deleted text. Assumes l1 <= l2.
func (b *Buffer) deleteRange(l1, c1, l2, c2 int) string {
	if l1 == l2 {
		line := b.lines[l1]
		if c2 > len(line) {
			c2 = len(line)
		}
		deleted := string(line[c1:c2])
		b.lines[l1] = append(line[:c1], line[c2:]...)
		b.cursor = Cursor{l1, c1}
		b.clampCol()
		return deleted
	}

	// multi-line
	var sb strings.Builder
	sb.WriteString(string(b.lines[l1][c1:]))
	sb.WriteRune('\n')
	for i := l1 + 1; i < l2; i++ {
		sb.WriteString(string(b.lines[i]))
		sb.WriteRune('\n')
	}
	sb.WriteString(string(b.lines[l2][:c2]))

	newLine := append(b.lines[l1][:c1], b.lines[l2][c2:]...)
	b.lines = append(b.lines[:l1], b.lines[l2+1:]...)
	b.lines[l1] = newLine
	b.cursor = Cursor{l1, c1}
	b.clampCol()
	return sb.String()
}

// pasteAfter pastes text after the cursor.
func (b *Buffer) pasteAfter(text string, linewise bool) {
	if linewise {
		lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
		insertAt := b.cursor.Line + 1
		runeLines := make([][]rune, len(lines))
		for i, l := range lines {
			runeLines[i] = []rune(l)
		}
		b.lines = append(b.lines[:insertAt], append(runeLines, b.lines[insertAt:]...)...)
		b.cursor.Line = insertAt
		b.cursor.Col = 0
		return
	}
	// character paste - insert after cursor
	col := b.cursor.Col
	if len(b.lines[b.cursor.Line]) > 0 {
		col++
	}
	l := b.lines[b.cursor.Line]
	newLine := make([]rune, 0, len(l)+utf8.RuneCountInString(text))
	newLine = append(newLine, l[:col]...)
	newLine = append(newLine, []rune(text)...)
	newLine = append(newLine, l[col:]...)
	b.lines[b.cursor.Line] = newLine
	b.cursor.Col = col + utf8.RuneCountInString(text) - 1
}

// pasteBefore pastes text before the cursor.
func (b *Buffer) pasteBefore(text string, linewise bool) {
	if linewise {
		lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
		insertAt := b.cursor.Line
		runeLines := make([][]rune, len(lines))
		for i, l := range lines {
			runeLines[i] = []rune(l)
		}
		b.lines = append(b.lines[:insertAt], append(runeLines, b.lines[insertAt:]...)...)
		b.cursor.Line = insertAt
		b.cursor.Col = 0
		return
	}
	col := b.cursor.Col
	l := b.lines[b.cursor.Line]
	newLine := make([]rune, 0, len(l)+utf8.RuneCountInString(text))
	newLine = append(newLine, l[:col]...)
	newLine = append(newLine, []rune(text)...)
	newLine = append(newLine, l[col:]...)
	b.lines[b.cursor.Line] = newLine
	b.cursor.Col = col
}

// joinLine joins the current line with the next.
func (b *Buffer) joinLine() {
	if b.cursor.Line >= len(b.lines)-1 {
		return
	}
	cur := b.lines[b.cursor.Line]
	next := b.lines[b.cursor.Line+1]
	joinCol := len(cur)
	joined := make([]rune, 0, len(cur)+1+len(next))
	joined = append(joined, cur...)
	if len(cur) > 0 && len(next) > 0 {
		joined = append(joined, ' ')
	}
	joined = append(joined, next...)
	b.lines[b.cursor.Line] = joined
	b.lines = append(b.lines[:b.cursor.Line+1], b.lines[b.cursor.Line+2:]...)
	b.cursor.Col = joinCol
}

// --- Helpers ---

func (b *Buffer) clampCol() {
	if b.cursor.Line < 0 || b.cursor.Line >= len(b.lines) {
		b.cursor.Col = 0
		return
	}
	max := maxCol(b.lines[b.cursor.Line], b.mode)
	if b.cursor.Col > max {
		b.cursor.Col = max
	}
	if b.cursor.Col < 0 {
		b.cursor.Col = 0
	}
}

func maxCol(line []rune, mode Mode) int {
	if len(line) == 0 {
		return 0
	}
	if mode == ModeInsert {
		return len(line)
	}
	return len(line) - 1
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
