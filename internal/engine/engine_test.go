package engine_test

import (
	"testing"

	"github.com/BRO3886/vimtutor/internal/engine"
)

func feed(p *engine.Parser, keys ...string) *engine.ParsedCmd {
	var last *engine.ParsedCmd
	for _, k := range keys {
		cmd, _ := p.Feed(k)
		if cmd != nil {
			last = cmd
		}
	}
	return last
}

func execKeys(buf *engine.Buffer, keys ...string) {
	p := engine.NewParser()
	for _, k := range keys {
		if buf.Mode() == engine.ModeInsert {
			engine.InsertKey(buf, k)
			continue
		}
		cmd, needMore := p.Feed(k)
		if needMore {
			continue
		}
		engine.Execute(buf, cmd)
	}
}

func TestBasicMotions(t *testing.T) {
	buf := engine.NewBuffer("hello world\nfoo bar")
	buf.SetCursor(0, 0)

	// l moves right
	execKeys(buf, "l", "l", "l")
	c := buf.GetCursor()
	if c.Col != 3 {
		t.Errorf("expected col 3, got %d", c.Col)
	}

	// j moves down
	execKeys(buf, "j")
	c = buf.GetCursor()
	if c.Line != 1 {
		t.Errorf("expected line 1, got %d", c.Line)
	}

	// 0 goes to start
	execKeys(buf, "0")
	c = buf.GetCursor()
	if c.Col != 0 {
		t.Errorf("expected col 0, got %d", c.Col)
	}

	// $ goes to end
	execKeys(buf, "$")
	c = buf.GetCursor()
	if c.Col != len("foo bar")-1 {
		t.Errorf("expected col %d, got %d", len("foo bar")-1, c.Col)
	}
}

func TestWordMotions(t *testing.T) {
	buf := engine.NewBuffer("The quick brown fox")
	buf.SetCursor(0, 0)

	// w moves to next word
	execKeys(buf, "w")
	c := buf.GetCursor()
	if c.Col != 4 {
		t.Errorf("w: expected col 4, got %d", c.Col)
	}

	// 2w moves 2 words (feed as separate keystrokes)
	execKeys(buf, "2", "w")
	c = buf.GetCursor()
	if c.Col != 16 {
		t.Errorf("2w: expected col 16, got %d", c.Col)
	}

	// b moves back
	execKeys(buf, "b")
	c = buf.GetCursor()
	if c.Col != 10 {
		t.Errorf("b: expected col 10, got %d", c.Col)
	}
}

func TestDeleteWord(t *testing.T) {
	buf := engine.NewBuffer("The quick brown fox")
	buf.SetCursor(0, 4)
	execKeys(buf, "d", "w")

	if buf.Text() != "The brown fox" {
		t.Errorf("dw: expected 'The brown fox', got %q", buf.Text())
	}
}

func TestDeleteLine(t *testing.T) {
	buf := engine.NewBuffer("keep this\ndelete this\nkeep too")
	buf.SetCursor(1, 0)
	execKeys(buf, "d", "d")

	if buf.Text() != "keep this\nkeep too" {
		t.Errorf("dd: expected 'keep this\\nkeep too', got %q", buf.Text())
	}
}

func TestChangeWord(t *testing.T) {
	buf := engine.NewBuffer("The quick brown fox")
	buf.SetCursor(0, 4)
	execKeys(buf, "c", "w") // enters insert mode
	// now type replacement
	execKeys(buf, "f", "a", "s", "t", "esc")

	if buf.Text() != "The fast brown fox" {
		t.Errorf("cw: expected 'The fast brown fox', got %q", buf.Text())
	}
}

func TestInsertAndAppend(t *testing.T) {
	buf := engine.NewBuffer("hello")
	buf.SetCursor(0, 4)
	execKeys(buf, "a") // append after cursor
	execKeys(buf, " ", "w", "o", "r", "l", "d", "esc")

	if buf.Text() != "hello world" {
		t.Errorf("a: expected 'hello world', got %q", buf.Text())
	}
}

func TestTextObjectInnerWord(t *testing.T) {
	buf := engine.NewBuffer("The quick brown")
	buf.SetCursor(0, 5) // cursor on 'u' in "quick"
	execKeys(buf, "d", "i", "w")

	if buf.Text() != "The  brown" {
		t.Errorf("diw: expected 'The  brown', got %q", buf.Text())
	}
}

func TestTextObjectInnerQuotes(t *testing.T) {
	buf := engine.NewBuffer(`print("hello world")`)
	buf.SetCursor(0, 8) // inside the string
	execKeys(buf, "d", "i", `"`)

	if buf.Text() != `print("")` {
		t.Errorf(`di": expected 'print("")', got %q`, buf.Text())
	}
}

func TestFindChar(t *testing.T) {
	buf := engine.NewBuffer("hello world fox")
	buf.SetCursor(0, 0)
	execKeys(buf, "f", "o")

	c := buf.GetCursor()
	if c.Col != 4 {
		t.Errorf("fo: expected col 4, got %d", c.Col)
	}
}

func TestGGandG(t *testing.T) {
	buf := engine.NewBuffer("line1\nline2\nline3\nline4")
	buf.SetCursor(0, 0)

	execKeys(buf, "G")
	c := buf.GetCursor()
	if c.Line != 3 {
		t.Errorf("G: expected line 3, got %d", c.Line)
	}

	execKeys(buf, "g", "g")
	c = buf.GetCursor()
	if c.Line != 0 {
		t.Errorf("gg: expected line 0, got %d", c.Line)
	}
}

func TestYankPaste(t *testing.T) {
	buf := engine.NewBuffer("first\nsecond\nthird")
	buf.SetCursor(0, 0)

	execKeys(buf, "y", "y") // yank line
	execKeys(buf, "j", "j") // move to line 3
	execKeys(buf, "p")      // paste below

	expected := "first\nsecond\nthird\nfirst"
	if buf.Text() != expected {
		t.Errorf("yy/p: expected %q, got %q", expected, buf.Text())
	}
}

func TestOpenLineBelow(t *testing.T) {
	buf := engine.NewBuffer("hello")
	buf.SetCursor(0, 0)
	execKeys(buf, "o")
	execKeys(buf, "w", "o", "r", "l", "d", "esc")

	expected := "hello\nworld"
	if buf.Text() != expected {
		t.Errorf("o: expected %q, got %q", expected, buf.Text())
	}
}

func TestParserCountPrefix(t *testing.T) {
	p := engine.NewParser()
	cmd := feed(p, "3", "w")
	if cmd == nil {
		t.Fatal("expected parsed cmd, got nil")
	}
	if cmd.Motion.Count != 3 {
		t.Errorf("3w: expected count 3, got %d", cmd.Motion.Count)
	}
}

func TestParserTextObject(t *testing.T) {
	p := engine.NewParser()
	cmd := feed(p, "d", "i", "w")
	if cmd == nil {
		t.Fatal("expected parsed cmd, got nil")
	}
	if cmd.Operator != "d" {
		t.Errorf("diw: expected operator 'd', got %q", cmd.Operator)
	}
	if cmd.Motion.Key != "iw" {
		t.Errorf("diw: expected motion 'iw', got %q", cmd.Motion.Key)
	}
}
