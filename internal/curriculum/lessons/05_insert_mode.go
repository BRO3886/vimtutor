package lessons

import "github.com/BRO3886/vimtutor/internal/curriculum"

func init() {
	curriculum.Register(curriculum.Lesson{
		ID:            "05_insert_mode",
		Title:         "Insert Mode Mastery",
		Description:   "Enter insert mode with precision — there are 6 ways in.",
		Difficulty:    curriculum.Beginner,
		Tags:          []string{"insert", "basics"},
		Prerequisites: []curriculum.LessonID{"03_operators"},
		Steps: []curriculum.Step{
			{
				ID:    "05a_intro",
				Title: "Ways to enter insert mode",
				Explanation: `Never just smash i. Pick the right entry point:

  i  → insert before cursor
  I  → insert at start of line (first non-blank)
  a  → append after cursor
  A  → append at end of line
  o  → open new line BELOW and insert
  O  → open new line ABOVE and insert

Mastering these means fewer keystrokes to get where you need.

Press SPACE to continue.`,
				InitText:   "",
				CursorLine: 0, CursorCol: 0,
			},
			{
				ID:    "05b_append",
				Title: "Append at end of line",
				Explanation: `Add " world" to the end of the line.
Cursor is at col 0. Use A (append at end of line).`,
				InitText:   "hello",
				CursorLine: 0, CursorCol: 0,
				Challenge: &curriculum.Challenge{
					Goal:         `Append " world" to make "hello world"`,
					ExpectedText: "hello world",
					Hint:         "A moves cursor to end of line and enters insert mode",
				},
			},
			{
				ID:    "05c_insert_at_start",
				Title: "Insert at line start",
				Explanation: `Add "// " at the start of the line to comment it out.
Use: I (insert at first non-blank)`,
				InitText:   "    fmt.Println()",
				CursorLine: 0, CursorCol: 10,
				Challenge: &curriculum.Challenge{
					Goal:         "Comment out the line",
					ExpectedText: "    // fmt.Println()",
					Hint:         "I positions cursor at first non-blank char and enters insert mode",
				},
			},
			{
				ID:    "05d_open_below",
				Title: "Open line below",
				Explanation: `Add a new line "    return nil" below line 1.
Use: o (open new line below)`,
				InitText:   "func foo() error {",
				CursorLine: 0, CursorCol: 0,
				Challenge: &curriculum.Challenge{
					Goal:         "Add return nil on a new line below",
					ExpectedText: "func foo() error {\n\treturn nil",
					Hint:         "o opens a new line below and enters insert mode",
				},
			},
			{
				ID:    "05e_open_above",
				Title: "Open line above",
				Explanation: `Insert "// start of block" ABOVE the current line.
Use: O (open new line above)`,
				InitText:   "x := 42",
				CursorLine: 0, CursorCol: 0,
				Challenge: &curriculum.Challenge{
					Goal:         "Add a comment line above",
					ExpectedText: "// start of block\nx := 42",
					Hint:         "O opens a new line above cursor and enters insert mode",
				},
			},
			{
				ID:    "05f_ctrl_w",
				Title: "Delete word in insert mode",
				Explanation: `While in insert mode, Ctrl+w deletes the previous word.
This saves you from hammering backspace.

Change "wrng" to "right" using cw, then notice:
after typing "ri", press Ctrl+w to delete "ri" and retype.`,
				InitText:   "this is wrng",
				CursorLine: 0, CursorCol: 8,
				Challenge: &curriculum.Challenge{
					Goal:         `Change "wrng" to "right"`,
					ExpectedText: "this is right",
					Hint:         "cw then type 'right' then Esc",
				},
			},
		},
	})
}
