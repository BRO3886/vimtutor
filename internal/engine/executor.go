package engine

// Execute takes a parsed command and applies it to the buffer.
// Returns true if the buffer entered insert mode (caller needs to switch to insert handling).
func Execute(b *Buffer, cmd *ParsedCmd) (enteredInsert bool) {
	if cmd == nil {
		return false
	}

	count := cmd.Count
	if count <= 0 {
		count = 1
	}

	// Pure action commands
	if cmd.Action != "" {
		return b.execAction(cmd)
	}

	// Operator + motion/text-object commands
	if cmd.Operator != "" {
		b.execOperator(cmd)
		return cmd.Operator == "c"
	}

	// Pure motion
	if cmd.Motion.Key != "" {
		m := cmd.Motion
		if m.Count <= 0 {
			m.Count = count
		}
		b.ApplyMotion(m)
	}

	return false
}

func (b *Buffer) execAction(cmd *ParsedCmd) bool {
	switch cmd.Action {
	case "i":
		b.SetMode(ModeInsert)
		return true
	case "I":
		b.cursor.Col = firstNonBlank(b.lines[b.cursor.Line])
		b.SetMode(ModeInsert)
		return true
	case "a":
		l := b.lines[b.cursor.Line]
		if len(l) > 0 && b.cursor.Col < len(l) {
			b.cursor.Col++
		}
		b.SetMode(ModeInsert)
		return true
	case "A":
		b.cursor.Col = len(b.lines[b.cursor.Line])
		b.SetMode(ModeInsert)
		return true
	case "o":
		// open line below
		line := b.cursor.Line
		b.lines = append(b.lines[:line+1], append([][]rune{{}}, b.lines[line+1:]...)...)
		b.cursor.Line++
		b.cursor.Col = 0
		b.SetMode(ModeInsert)
		return true
	case "O":
		// open line above
		line := b.cursor.Line
		b.lines = append(b.lines[:line], append([][]rune{{}}, b.lines[line:]...)...)
		b.cursor.Col = 0
		b.SetMode(ModeInsert)
		return true
	case "p":
		reg := cmd.Register
		if reg == 0 {
			reg = '"'
		}
		text := b.registers[reg]
		if text != "" {
			linewise := b.registers['_'+reg] == "1"
			b.pasteAfter(text, linewise)
		}
	case "P":
		reg := cmd.Register
		if reg == 0 {
			reg = '"'
		}
		text := b.registers[reg]
		if text != "" {
			linewise := b.registers['_'+reg] == "1"
			b.pasteBefore(text, linewise)
		}
	case "x":
		if len(b.lines[b.cursor.Line]) > 0 {
			deleted := string([]rune{b.lines[b.cursor.Line][b.cursor.Col]})
			b.registers['"'] = deleted
			b.deleteRune()
		}
	case "X":
		if b.cursor.Col > 0 {
			b.cursor.Col--
			deleted := string([]rune{b.lines[b.cursor.Line][b.cursor.Col]})
			b.registers['"'] = deleted
			b.deleteRune()
		}
	case "J":
		b.joinLine()
	case "~":
		// toggle case
		l := b.lines[b.cursor.Line]
		if b.cursor.Col < len(l) {
			r := l[b.cursor.Col]
			if r >= 'a' && r <= 'z' {
				l[b.cursor.Col] = r - 32
			} else if r >= 'A' && r <= 'Z' {
				l[b.cursor.Col] = r + 32
			}
			b.cursor.Col = min(b.cursor.Col+1, len(l)-1)
		}
	case "u":
		// undo - not fully implemented, just a placeholder
	case "esc", "ctrl+[":
		// transition back to normal mode, adjust cursor
		if b.mode == ModeInsert {
			if b.cursor.Col > 0 && len(b.lines[b.cursor.Line]) > 0 {
				b.cursor.Col--
			}
		}
		b.SetMode(ModeNormal)
	case "enter":
		// in normal mode, move down
		if b.cursor.Line < len(b.lines)-1 {
			b.cursor.Line++
			b.cursor.Col = firstNonBlank(b.lines[b.cursor.Line])
		}
	case "backspace":
		if b.cursor.Col > 0 {
			b.cursor.Col--
		}
	case "zz", "zt", "zb":
		// scroll commands - no-op in challenge mode
	case ":":
		// command mode - no-op for now
	}
	return false
}

func (b *Buffer) execOperator(cmd *ParsedCmd) {
	op := cmd.Operator
	reg := cmd.Register
	if reg == 0 {
		reg = '"'
	}

	if cmd.Linewise {
		// dd, cc, yy, >>, <<
		count := cmd.Count
		if count <= 0 {
			count = 1
		}
		startLine := b.cursor.Line
		endLine := min(startLine+count-1, len(b.lines)-1)

		switch op {
		case "d":
			deleted := b.deleteLineRange(startLine, endLine)
			b.registers[reg] = deleted + "\n"
			b.registers['_'+reg] = "1"
		case "c":
			deleted := b.deleteLineRange(startLine, endLine)
			b.registers[reg] = deleted + "\n"
			b.registers['_'+reg] = "1"
			// ensure we have at least one empty line
			if len(b.lines) == 0 {
				b.lines = [][]rune{{}}
			}
			b.cursor.Col = 0
			b.SetMode(ModeInsert)
		case "y":
			lines := make([]string, endLine-startLine+1)
			for i := startLine; i <= endLine; i++ {
				lines[i-startLine] = string(b.lines[i])
			}
			b.registers[reg] = joinLines(lines) + "\n"
			b.registers['_'+reg] = "1"
		case ">":
			for i := startLine; i <= endLine; i++ {
				b.lines[i] = append([]rune("\t"), b.lines[i]...)
			}
		case "<":
			for i := startLine; i <= endLine; i++ {
				if len(b.lines[i]) > 0 && (b.lines[i][0] == '\t' || b.lines[i][0] == ' ') {
					b.lines[i] = b.lines[i][1:]
				}
			}
		}
		return
	}

	// Motion or text object based operator
	m := cmd.Motion
	if m.Count <= 0 {
		m.Count = cmd.Count
	}

	// vim quirk: cw = ce, cW = cE (change doesn't include trailing space)
	if op == "c" {
		if m.Key == "w" {
			m.Key = "e"
		} else if m.Key == "W" {
			m.Key = "E"
		}
	}

	var r Range
	if isTextObject(m.Key) {
		r = b.TextObjectRange(m.Key)
	} else {
		savedLine, savedCol := b.cursor.Line, b.cursor.Col
		r = b.ApplyMotion(m)
		b.cursor.Line, b.cursor.Col = savedLine, savedCol
		// For `e`/`E` motions with operators, include the char under cursor end
		if m.Key == "e" || m.Key == "E" {
			r.C2++
		}
	}

	switch op {
	case "d":
		if r.Linewise {
			deleted := b.deleteLineRange(r.L1, r.L2)
			b.registers[reg] = deleted + "\n"
			b.registers['_'+reg] = "1"
		} else {
			deleted := b.deleteRange(r.L1, r.C1, r.L2, r.C2)
			b.registers[reg] = deleted
			b.registers['_'+reg] = ""
		}
	case "c":
		if r.Linewise {
			deleted := b.deleteLineRange(r.L1, r.L2)
			b.registers[reg] = deleted + "\n"
			b.registers['_'+reg] = "1"
			if len(b.lines) == 0 {
				b.lines = [][]rune{{}}
			}
		} else {
			deleted := b.deleteRange(r.L1, r.C1, r.L2, r.C2)
			b.registers[reg] = deleted
			b.registers['_'+reg] = ""
		}
		b.SetMode(ModeInsert)
	case "y":
		if r.Linewise {
			lines := make([]string, r.L2-r.L1+1)
			for i := r.L1; i <= r.L2; i++ {
				lines[i-r.L1] = string(b.lines[i])
			}
			b.registers[reg] = joinLines(lines) + "\n"
			b.registers['_'+reg] = "1"
		} else {
			// extract text without modifying buffer
			var text string
			if r.L1 == r.L2 {
				l := b.lines[r.L1]
				c2 := r.C2
				if c2 > len(l) {
					c2 = len(l)
				}
				text = string(l[r.C1:c2])
			}
			b.registers[reg] = text
			b.registers['_'+reg] = ""
		}
		// yank doesn't move cursor
		b.cursor.Line, b.cursor.Col = r.L1, r.C1
	case "r":
		// replace single char
		if m.Key == "r" && m.Char != 0 {
			l := b.lines[b.cursor.Line]
			if b.cursor.Col < len(l) {
				l[b.cursor.Col] = m.Char
			}
		}
	case ">":
		startLine, endLine := r.L1, r.L2
		if !r.Linewise {
			endLine = r.L1
		}
		for i := startLine; i <= endLine; i++ {
			b.lines[i] = append([]rune("  "), b.lines[i]...)
		}
	case "<":
		startLine, endLine := r.L1, r.L2
		if !r.Linewise {
			endLine = r.L1
		}
		for i := startLine; i <= endLine; i++ {
			if len(b.lines[i]) >= 2 && b.lines[i][0] == ' ' && b.lines[i][1] == ' ' {
				b.lines[i] = b.lines[i][2:]
			} else if len(b.lines[i]) > 0 && b.lines[i][0] == '\t' {
				b.lines[i] = b.lines[i][1:]
			}
		}
	}
}

// InsertKey handles a single key in insert mode.
func InsertKey(b *Buffer, key string) bool {
	switch key {
	case "esc", "ctrl+[":
		if b.cursor.Col > 0 && len(b.lines[b.cursor.Line]) > 0 {
			b.cursor.Col--
		}
		b.SetMode(ModeNormal)
		return false
	case "enter", "ctrl+m":
		b.insertNewlineBelow()
	case "backspace", "ctrl+h":
		if b.cursor.Col > 0 {
			b.cursor.Col--
			b.deleteRune()
		} else if b.cursor.Line > 0 {
			// merge with previous line
			prevLine := b.cursor.Line - 1
			prevLen := len(b.lines[prevLine])
			merged := append(b.lines[prevLine], b.lines[b.cursor.Line]...)
			b.lines[prevLine] = merged
			b.lines = append(b.lines[:b.cursor.Line], b.lines[b.cursor.Line+1:]...)
			b.cursor.Line = prevLine
			b.cursor.Col = prevLen
		}
	case "ctrl+w":
		// delete word backwards in insert mode
		for b.cursor.Col > 0 {
			b.cursor.Col--
			r := b.lines[b.cursor.Line][b.cursor.Col]
			b.deleteRune()
			if !isWordChar(r) {
				break
			}
		}
	case "ctrl+u":
		// delete to start of line
		b.lines[b.cursor.Line] = b.lines[b.cursor.Line][b.cursor.Col:]
		b.cursor.Col = 0
	case "tab":
		b.insertRune('\t')
	default:
		if len(key) == 1 {
			b.insertRune(rune(key[0]))
		}
	}
	return true
}

func isTextObject(key string) bool {
	if len(key) < 2 {
		return false
	}
	return (key[0] == 'i' || key[0] == 'a') && key != "insert"
}

func joinLines(lines []string) string {
	result := ""
	for i, l := range lines {
		if i > 0 {
			result += "\n"
		}
		result += l
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
