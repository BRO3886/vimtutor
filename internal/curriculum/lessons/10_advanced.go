package lessons

import "github.com/BRO3886/vimtutor/internal/curriculum"

func init() {
	curriculum.Register(curriculum.Lesson{
		ID:            "10_advanced",
		Title:         "Advanced Combinations",
		Description:   "Compose everything you know into powerful combos.",
		Difficulty:    curriculum.Advanced,
		Tags:          []string{"advanced", "combo"},
		Prerequisites: []curriculum.LessonID{"09_macros"},
		Steps: []curriculum.Step{
			{
				ID:    "10a_intro",
				Title: "The vim mindset",
				Explanation: `At this point you know the building blocks.
The real skill is composing them efficiently.

Some power combos:
  ci"   → change string literal
  da(   → delete a function call's argument list
  >i{   → indent the body of a function
  d/foo → delete up to the next "foo"
  gUiw  → uppercase the current word

The goal is always: fewest keystrokes, maximum precision.
Think before you type. Plan your edit.

Press SPACE to continue.`,
				InitText:   "",
				CursorLine: 0, CursorCol: 0,
			},
			{
				ID:    "10b_combo1",
				Title: "Swap function arguments",
				Explanation: `Swap the two arguments: change foo(b, a) to foo(a, b).
This requires deleting "b, " and repositioning "a".`,
				InitText:   "foo(b, a)",
				CursorLine: 0, CursorCol: 4,
				Challenge: &curriculum.Challenge{
					Goal:         "Reorder to foo(a, b)",
					ExpectedText: "foo(a, b)",
					Hint:         "Try: dw then move past ', ' and p, or use a creative combination",
				},
			},
			{
				ID:    "10c_combo2",
				Title: "Indent a block",
				Explanation: `Indent the function body (lines 2-4) with >>.
First select with Vj... or use >> line by line.

Alternatively: position cursor inside the block, use >i{ to indent.`,
				InitText:   "func main() {\nfmt.Println(1)\nfmt.Println(2)\nfmt.Println(3)\n}",
				CursorLine: 1, CursorCol: 0,
				Challenge: &curriculum.Challenge{
					Goal:         "Add one level of indentation to lines 2-4",
					ExpectedText: "func main() {\n\tfmt.Println(1)\n\tfmt.Println(2)\n\tfmt.Println(3)\n}",
					Hint:         "Use >> on each line or visual mode V + select + >",
				},
			},
			{
				ID:    "10d_combo3",
				Title: "Cleanup trailing content",
				Explanation: `Delete everything after the semicolon on each line.
Line 1 should become: "x := 1"`,
				InitText:   "x := 1; // unused\ny := 2; // also unused",
				CursorLine: 0, CursorCol: 0,
				Challenge: &curriculum.Challenge{
					Goal:         "Remove trailing comments from both lines",
					ExpectedText: "x := 1\ny := 2",
					Hint:         "dt; or d$ after positioning, then j and repeat or use . command",
				},
			},
		},
	})
}
