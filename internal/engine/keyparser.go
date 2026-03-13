package engine

import (
	"strconv"
	"strings"
)

// ParsedCmd represents a fully parsed vim command.
type ParsedCmd struct {
	Count    int    // 0 means no explicit count (treat as 1)
	Operator string // "d", "c", "y", ">", "<", "r", "" for pure motions
	Motion   Motion // the motion or text object
	Register rune   // register name, 0 for default
	Linewise bool   // dd, cc, yy style
	Action   string // special actions: "i","I","a","A","o","O","p","P","u","x","J","esc","enter","backspace"
	Raw      string // original keystroke sequence
}

type parseState int

const (
	stateStart parseState = iota
	stateCount
	stateRegister
	stateOperator
	stateMotion
	stateAwaitChar  // after f/t/F/T/r
	stateGPrefix    // after 'g'
	stateZPrefix    // after 'z'
)

// Parser is a streaming key sequence parser.
type Parser struct {
	state    parseState
	pending  []string
	count    int
	register rune
	operator string
}

// NewParser creates a fresh parser.
func NewParser() *Parser {
	return &Parser{state: stateStart}
}

// Reset clears parser state.
func (p *Parser) Reset() {
	p.state = stateStart
	p.pending = p.pending[:0]
	p.count = 0
	p.register = 0
	p.operator = ""
}

// Feed processes one key string (e.g. "d", "ctrl+w", "esc", "G").
// Returns a ParsedCmd when a complete command is assembled, and needMore=false.
// Returns needMore=true when more keys are needed.
func (p *Parser) Feed(key string) (cmd *ParsedCmd, needMore bool) {
	p.pending = append(p.pending, key)
	raw := strings.Join(p.pending, "")

	switch p.state {
	case stateStart:
		return p.handleStart(key, raw)
	case stateCount:
		return p.handleCount(key, raw)
	case stateRegister:
		return p.handleRegister(key, raw)
	case stateOperator:
		return p.handleOperator(key, raw)
	case stateMotion:
		return p.handleMotion(key, raw)
	case stateAwaitChar:
		return p.handleAwaitChar(key, raw)
	case stateGPrefix:
		return p.handleGPrefix(key, raw)
	case stateZPrefix:
		return p.handleZPrefix(key, raw)
	}
	return p.emitReset(nil)
}

// PendingKeys returns the keys buffered so far (for display in status bar).
func (p *Parser) PendingKeys() string {
	return strings.Join(p.pending, "")
}

func (p *Parser) handleStart(key, raw string) (*ParsedCmd, bool) {
	// Escape always resets
	if key == "esc" || key == "ctrl+[" {
		return p.emitAction("esc", raw)
	}

	// Register prefix
	if key == `"` {
		p.state = stateRegister
		return nil, true
	}

	// Count prefix (1-9, not 0 at start)
	if d := digit(key); d > 0 {
		p.count = d
		p.state = stateCount
		return nil, true
	}

	// Operators
	if isOperator(key) {
		p.operator = key
		p.state = stateOperator
		return nil, true
	}

	// Pure motions
	if m, ok := pureMot(key); ok {
		if needsChar(m) {
			p.state = stateAwaitChar
			return nil, true
		}
		return p.emitMotion(m, raw)
	}

	// Insert mode entry actions
	switch key {
	case "i", "I", "a", "A", "o", "O":
		return p.emitAction(key, raw)
	case "p":
		return p.emitAction("p", raw)
	case "P":
		return p.emitAction("P", raw)
	case "x":
		return p.emitAction("x", raw)
	case "X":
		return p.emitAction("X", raw)
	case "J":
		return p.emitAction("J", raw)
	case "u":
		return p.emitAction("u", raw)
	case "r":
		p.operator = "r"
		p.state = stateAwaitChar
		return nil, true
	case "~":
		return p.emitAction("~", raw)
	case "g":
		p.state = stateGPrefix
		return nil, true
	case "z":
		p.state = stateZPrefix
		return nil, true
	case ":":
		return p.emitAction(":", raw)
	case "enter", "ctrl+m":
		return p.emitAction("enter", raw)
	case "backspace", "ctrl+h":
		return p.emitAction("backspace", raw)
	}

	// Unknown key - just reset
	return p.emitReset(nil)
}

func (p *Parser) handleCount(key, raw string) (*ParsedCmd, bool) {
	if d := digit(key); d >= 0 {
		p.count = p.count*10 + d
		return nil, true
	}
	// After count, expect operator or motion
	if isOperator(key) {
		p.operator = key
		p.state = stateOperator
		return nil, true
	}
	if m, ok := pureMot(key); ok {
		m.Count = p.count
		if needsChar(m) {
			p.state = stateAwaitChar
			return nil, true
		}
		return p.emitMotion(m, raw)
	}
	switch key {
	case "g":
		p.state = stateGPrefix
		return nil, true
	case "esc", "ctrl+[":
		return p.emitAction("esc", raw)
	}
	return p.emitReset(nil)
}

func (p *Parser) handleRegister(key, raw string) (*ParsedCmd, bool) {
	// expect a register char a-z, A-Z, ", +, 0
	if len(key) == 1 {
		r := rune(key[0])
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '"' || r == '+' || r == '0' {
			p.register = r
			p.state = stateStart
			// next key will be the operator
			p.pending = p.pending[:len(p.pending)-1] // remove register char from pending display
			return nil, true
		}
	}
	return p.emitReset(nil)
}

func (p *Parser) handleOperator(key, raw string) (*ParsedCmd, bool) {
	op := p.operator
	// Double operator = linewise (dd, cc, yy, >>)
	if key == op || (op == ">" && key == ">") || (op == "<" && key == "<") {
		return p.emitCmd(&ParsedCmd{
			Count:    p.count,
			Operator: op,
			Linewise: true,
			Register: p.register,
			Raw:      raw,
		})
	}

	// Count after operator
	if d := digit(key); d > 0 && op != "r" {
		if p.count == 0 {
			p.count = d
		} else {
			p.count = p.count*10 + d
		}
		return nil, true
	}

	// Expect motion or text object
	if m, ok := pureMot(key); ok {
		m.Count = p.count
		if needsChar(m) {
			p.state = stateAwaitChar
			return nil, true
		}
		return p.emitCmd(&ParsedCmd{
			Count:    1,
			Operator: op,
			Motion:   m,
			Register: p.register,
			Raw:      raw,
		})
	}

	// Text object prefix 'i' or 'a'
	if key == "i" || key == "a" {
		p.pending = append(p.pending, key) // handled in motionFromTO
		p.state = stateMotion
		return nil, true
	}

	// g prefix
	if key == "g" {
		p.state = stateGPrefix
		return nil, true
	}

	if key == "esc" {
		return p.emitAction("esc", raw)
	}

	return p.emitReset(nil)
}

func (p *Parser) handleMotion(key, raw string) (*ParsedCmd, bool) {
	// Complete text object: the last two keys should be like "iw", "a("
	n := len(p.pending)
	if n >= 2 {
		toKey := p.pending[n-2] + key
		if m, ok := textObjectMot(toKey); ok {
			return p.emitCmd(&ParsedCmd{
				Count:    p.count,
				Operator: p.operator,
				Motion:   m,
				Register: p.register,
				Raw:      raw,
			})
		}
	}
	return p.emitReset(nil)
}

func (p *Parser) handleAwaitChar(key, raw string) (*ParsedCmd, bool) {
	if len(key) != 1 {
		return p.emitReset(nil)
	}
	ch := rune(key[0])

	// replace 'r' operator
	if p.operator == "r" {
		return p.emitCmd(&ParsedCmd{
			Count:    p.count,
			Operator: "r",
			Motion:   Motion{Key: "r", Char: ch},
			Raw:      raw,
		})
	}

	// get the key that triggered stateAwaitChar (second-to-last pending key)
	n := len(p.pending)
	motKey := ""
	if n >= 2 {
		motKey = p.pending[n-2]
	}

	m, ok := pureMot(motKey)
	if !ok {
		return p.emitReset(nil)
	}
	m.Char = ch
	m.Count = p.count

	if p.operator != "" {
		return p.emitCmd(&ParsedCmd{
			Count:    1,
			Operator: p.operator,
			Motion:   m,
			Register: p.register,
			Raw:      raw,
		})
	}
	return p.emitMotion(m, raw)
}

func (p *Parser) handleGPrefix(key, raw string) (*ParsedCmd, bool) {
	switch key {
	case "g":
		m := Motion{Key: "gg", Count: p.count}
		if p.operator != "" {
			return p.emitCmd(&ParsedCmd{Count: 1, Operator: p.operator, Motion: m, Raw: raw})
		}
		return p.emitMotion(m, raw)
	case "_":
		m := Motion{Key: "g_", Count: p.count}
		if p.operator != "" {
			return p.emitCmd(&ParsedCmd{Count: 1, Operator: p.operator, Motion: m, Raw: raw})
		}
		return p.emitMotion(m, raw)
	case "j":
		m := Motion{Key: "j", Count: p.count}
		return p.emitMotion(m, raw)
	case "k":
		m := Motion{Key: "k", Count: p.count}
		return p.emitMotion(m, raw)
	}
	return p.emitReset(nil)
}

func (p *Parser) handleZPrefix(key, raw string) (*ParsedCmd, bool) {
	switch key {
	case "z":
		return p.emitAction("zz", raw) // center screen (we'll just accept it)
	case "t":
		return p.emitAction("zt", raw)
	case "b":
		return p.emitAction("zb", raw)
	}
	return p.emitReset(nil)
}

// --- Emit helpers ---

func (p *Parser) emitMotion(m Motion, raw string) (*ParsedCmd, bool) {
	if p.count > 0 && m.Count == 0 {
		m.Count = p.count
	}
	return p.emitCmd(&ParsedCmd{
		Count:  1,
		Motion: m,
		Raw:    raw,
	})
}

func (p *Parser) emitAction(action, raw string) (*ParsedCmd, bool) {
	cmd := &ParsedCmd{
		Count:  max(1, p.count),
		Action: action,
		Raw:    raw,
	}
	return p.emitCmd(cmd)
}

func (p *Parser) emitCmd(cmd *ParsedCmd) (*ParsedCmd, bool) {
	p.Reset()
	return cmd, false
}

func (p *Parser) emitReset(cmd *ParsedCmd) (*ParsedCmd, bool) {
	p.Reset()
	return cmd, false
}

// --- Classification helpers ---

func isOperator(key string) bool {
	switch key {
	case "d", "c", "y", ">", "<":
		return true
	}
	return false
}

func pureMot(key string) (Motion, bool) {
	switch key {
	case "h", "j", "k", "l",
		"w", "W", "b", "B", "e", "E",
		"0", "^", "$",
		"G", "gg",
		"f", "F", "t", "T",
		"%", ";", ",",
		"g_":
		return Motion{Key: key}, true
	}
	return Motion{}, false
}

func textObjectMot(to string) (Motion, bool) {
	switch to {
	case "iw", "aw", "iW", "aW",
		"i\"", "a\"", "i'", "a'", "i`", "a`",
		"i(", "a(", "i)", "a)",
		"i[", "a[", "i]", "a]",
		"i{", "a{", "i}", "a}",
		"i<", "a<", "i>", "a>",
		"il", "al":
		return Motion{Key: to}, true
	}
	return Motion{}, false
}

func needsChar(m Motion) bool {
	switch m.Key {
	case "f", "F", "t", "T":
		return true
	}
	return false
}

func digit(key string) int {
	if len(key) == 1 && key[0] >= '0' && key[0] <= '9' {
		d, _ := strconv.Atoi(key)
		return d
	}
	return -1
}
