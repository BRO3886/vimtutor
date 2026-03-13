package lessons

import "github.com/BRO3886/vimtutor/internal/curriculum"

func init() {
	curriculum.Register(curriculum.Lesson{
		ID:            "06_search",
		Title:         "Search & Find",
		Description:   "Navigate by character and pattern. f, t, /, *, and friends.",
		Difficulty:    curriculum.Intermediate,
		Tags:          []string{"search", "motion"},
		Prerequisites: []curriculum.LessonID{"02_word_motions"},
		Steps: []curriculum.Step{
			{
				ID:    "06a_intro",
				Title: "Find and search",
				Explanation: `The fastest way to move long distances:

  f{char}  → find next {char} on current line (inclusive)
  t{char}  → till {char} (stops one before)
  F{char}  → find backward
  T{char}  → till backward
  ;        → repeat last f/t/F/T
  ,        → repeat last f/t/F/T backward

  *        → search for word under cursor (forward)
  #        → search for word under cursor (backward)

Tip: dt, is a common pattern — "delete till comma"

Press SPACE to continue.`,
				InitText:   "",
				CursorLine: 0, CursorCol: 0,
			},
			{
				ID:    "06b_f_motion",
				Title: "Find character",
				Explanation: `Jump to the 'x' using f.
Use: fx`,
				InitText:   "hello world fox jumps",
				CursorLine: 0, CursorCol: 0,
				Challenge: &curriculum.Challenge{
					Goal:         "Jump to the 'x' character",
					ExpectedLine: 0,
					ExpectedCol:  14,
					Hint:         "fx finds the next 'x' on the current line",
				},
			},
			{
				ID:    "06c_t_motion",
				Title: "Till character",
				Explanation: `Jump to just before the comma using t.
Use: t,`,
				InitText:   "foo(a, b, c)",
				CursorLine: 0, CursorCol: 0,
				Challenge: &curriculum.Challenge{
					Goal:         "Land just before the first comma (col 4)",
					ExpectedLine: 0,
					ExpectedCol:  4,
					Hint:         "t, jumps to one before the comma",
				},
			},
			{
				ID:    "06d_dt",
				Title: "Delete till",
				Explanation: `Delete the first argument "a, " up to (not including) "b".
Use: dt, or similar`,
				InitText:   "func(a, b, c)",
				CursorLine: 0, CursorCol: 5,
				Challenge: &curriculum.Challenge{
					Goal:         `Delete "a, " leaving "func(b, c)"`,
					ExpectedText: "func(b, c)",
					Hint:         "dt, deletes from cursor till the comma (not including the comma)",
				},
			},
			{
				ID:    "06e_semicolon",
				Title: "Repeat find with ;",
				Explanation: `Jump to each 'o' using f then repeat with ;
Start on 'o' in "foo", then use ; to reach the 'o' in "fox".`,
				InitText:   "foo bar fox",
				CursorLine: 0, CursorCol: 0,
				Challenge: &curriculum.Challenge{
					Goal:         "Land on 'o' in 'fox' (col 9)",
					ExpectedLine: 0,
					ExpectedCol:  9,
					Hint:         "fo then ; to repeat the find",
				},
			},
			{
				ID:    "06f_percent",
				Title: "Match bracket with %",
				Explanation: `Jump from the opening paren to its matching close.
Use: %`,
				InitText:   "result := foo(bar(baz()))",
				CursorLine: 0, CursorCol: 13,
				Challenge: &curriculum.Challenge{
					Goal:         "Jump to the matching closing paren (col 24)",
					ExpectedLine: 0,
					ExpectedCol:  24,
					Hint:         "% jumps to the matching bracket/paren/brace",
				},
			},
		},
	})
}
