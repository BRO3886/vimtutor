package lessons

import "github.com/BRO3886/vimtutor/internal/curriculum"

func init() {
	curriculum.Register(curriculum.Lesson{
		ID:          "01_basic_motions",
		Title:       "Basic Motions",
		Description: "Navigate with h, j, k, l — the vim way to move.",
		Difficulty:  curriculum.Beginner,
		Tags:        []string{"motion", "basics"},
		Steps: []curriculum.Step{
			{
				ID:    "01a_intro",
				Title: "The Home Row",
				Explanation: `In vim, your fingers never leave the home row.

The four navigation keys are:
  h ← left      l → right
  j ↓ down       k ↑ up

Forget the arrow keys. They work, but using them is a bad habit.
Your goal: reach for h/j/k/l without thinking.

Press SPACE to continue.`,
				InitText:   "",
				CursorLine: 0, CursorCol: 0,
			},
			{
				ID:    "01b_hjkl",
				Title: "Move right",
				Explanation: `Move the cursor right to land on the ★.
Use: l (right)`,
				InitText:   "The quick brown fox★",
				CursorLine: 0, CursorCol: 0,
				Challenge: &curriculum.Challenge{
					Goal:         "Move the cursor to the ★ character",
					ExpectedText: "The quick brown fox★",
					ExpectedLine: 0,
					ExpectedCol:  19,
					Hint:         "Press l repeatedly, or use a count: 19l",
				},
			},
			{
				ID:    "01c_down",
				Title: "Move down",
				Explanation: `Move the cursor down to the line with ★.
Use: j (down)`,
				InitText:   "line one\nline two\nline three ★\nline four",
				CursorLine: 0, CursorCol: 0,
				Challenge: &curriculum.Challenge{
					Goal:         "Move down to line three",
					ExpectedLine: 2,
					ExpectedCol:  -1,
					Hint:         "Press j twice to move down two lines",
				},
			},
			{
				ID:    "01d_up",
				Title: "Move up",
				Explanation: `Now move up to line one.
Use: k (up)`,
				InitText:   "★ line one\nline two\nline three",
				CursorLine: 2, CursorCol: 0,
				Challenge: &curriculum.Challenge{
					Goal:         "Move back up to line one",
					ExpectedLine: 0,
					ExpectedCol:  -1,
					Hint:         "Press k twice",
				},
			},
			{
				ID:    "01e_combined",
				Title: "Navigate to the X",
				Explanation: `Navigate to the X using h, j, k, l.
Try combining counts: 3j then 5l.`,
				InitText:   "........\n........\n........\n.....X..",
				CursorLine: 0, CursorCol: 0,
				Challenge: &curriculum.Challenge{
					Goal:         "Land on the X character",
					ExpectedLine: 3,
					ExpectedCol:  5,
					Hint:         "3j moves down 3 lines, 5l moves right 5 columns",
				},
			},
			{
				ID:    "01f_line_ends",
				Title: "Line start and end",
				Explanation: `Two essential motions:
  0 → jump to column 0 (start of line)
  $ → jump to end of line

The cursor is in the middle. Go to the end with $.`,
				InitText:   "The quick brown fox jumps",
				CursorLine: 0, CursorCol: 9,
				Challenge: &curriculum.Challenge{
					Goal:         "Jump to the end of the line",
					ExpectedLine: 0,
					ExpectedCol:  24,
					Hint:         "Press $ to jump to end of line",
				},
			},
			{
				ID:    "01g_line_start",
				Title: "Back to the start",
				Explanation: `Now jump back to the very beginning of the line.
Use: 0`,
				InitText:   "The quick brown fox jumps",
				CursorLine: 0, CursorCol: 24,
				Challenge: &curriculum.Challenge{
					Goal:         "Jump to column 0",
					ExpectedLine: 0,
					ExpectedCol:  0,
					Hint:         "Press 0 (zero) to go to column 0",
				},
			},
		},
	})
}
