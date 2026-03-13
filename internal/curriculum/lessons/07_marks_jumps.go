package lessons

import "github.com/BRO3886/vimtutor/internal/curriculum"

func init() {
	curriculum.Register(curriculum.Lesson{
		ID:            "07_marks_jumps",
		Title:         "Marks & Jumps",
		Description:   "Teleport anywhere — marks, gg/G, and line jumps.",
		Difficulty:    curriculum.Intermediate,
		Tags:          []string{"motion", "marks"},
		Prerequisites: []curriculum.LessonID{"02_word_motions"},
		Steps: []curriculum.Step{
			{
				ID:    "07a_intro",
				Title: "Marks and jumps",
				Explanation: `Big file? Teleport instead of scrolling:

  gg      → go to first line
  G       → go to last line
  {n}G    → go to line n  (e.g. 5G)
  m{a-z}  → set a mark
  '{mark} → jump to marked line
  ` + "`" + `{mark} → jump to exact mark position

  Ctrl+o  → jump back (previous position)
  Ctrl+i  → jump forward

Press SPACE to continue.`,
				InitText:   "",
				CursorLine: 0, CursorCol: 0,
			},
			{
				ID:    "07b_gg_G",
				Title: "Top and bottom",
				Explanation: `Jump to the last line with G.`,
				InitText:   "line 1\nline 2\nline 3\nline 4\nline 5",
				CursorLine: 0, CursorCol: 0,
				Challenge: &curriculum.Challenge{
					Goal:         "Jump to line 5 (index 4)",
					ExpectedLine: 4,
					ExpectedCol:  -1,
					Hint:         "Press G to jump to the last line",
				},
			},
			{
				ID:    "07c_back_to_top",
				Title: "Back to top",
				Explanation: `Now jump back to the very first line.
Use: gg`,
				InitText:   "line 1\nline 2\nline 3\nline 4\nline 5",
				CursorLine: 4, CursorCol: 0,
				Challenge: &curriculum.Challenge{
					Goal:         "Jump to line 1 (index 0)",
					ExpectedLine: 0,
					ExpectedCol:  -1,
					Hint:         "Press gg to jump to the first line",
				},
			},
			{
				ID:    "07d_line_jump",
				Title: "Jump to specific line",
				Explanation: `Jump directly to line 3 (index 2).
Use: 3G`,
				InitText:   "alpha\nbeta\ngamma\ndelta\nepsilon",
				CursorLine: 0, CursorCol: 0,
				Challenge: &curriculum.Challenge{
					Goal:         "Jump to line 3",
					ExpectedLine: 2,
					ExpectedCol:  -1,
					Hint:         "3G jumps to line number 3",
				},
			},
		},
	})
}
