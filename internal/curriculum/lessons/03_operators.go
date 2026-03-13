package lessons

import "github.com/BRO3886/vimtutor/internal/curriculum"

func init() {
	curriculum.Register(curriculum.Lesson{
		ID:            "03_operators",
		Title:         "Operators: d, c, y",
		Description:   "Delete, change, and yank text. The operator-motion combo is vim's superpower.",
		Difficulty:    curriculum.Beginner,
		Tags:          []string{"operator", "edit"},
		Prerequisites: []curriculum.LessonID{"02_word_motions"},
		Steps: []curriculum.Step{
			{
				ID:    "03a_intro",
				Title: "The operator grammar",
				Explanation: `Vim edits follow a grammar:  [count] operator motion

  d  → delete
  c  → change (delete + enter insert mode)
  y  → yank (copy)

Examples:
  dw   → delete from cursor to next word
  d$   → delete to end of line
  3dw  → delete 3 words
  dd   → delete the whole line (doubled = linewise)
  cc   → change the whole line
  yy   → yank the whole line

Press SPACE to continue.`,
				InitText:   "",
				CursorLine: 0, CursorCol: 0,
			},
			{
				ID:    "03b_dw",
				Title: "Delete a word",
				Explanation: `Delete "quick " from the sentence.
Use: dw (cursor is on 'q' in "quick")`,
				InitText:   "The quick brown fox",
				CursorLine: 0, CursorCol: 4,
				Challenge: &curriculum.Challenge{
					Goal:         `Delete "quick " to get "The brown fox"`,
					ExpectedText: "The brown fox",
					ExpectedLine: 0,
					Hint:         "Press dw to delete from cursor to next word start",
				},
			},
			{
				ID:    "03c_dd",
				Title: "Delete the whole line",
				Explanation: `Delete line 2 ("delete this line").
Use: dd`,
				InitText:   "keep this\ndelete this line\nkeep this too",
				CursorLine: 1, CursorCol: 0,
				Challenge: &curriculum.Challenge{
					Goal:         "Delete the middle line",
					ExpectedText: "keep this\nkeep this too",
					ExpectedLine: 1,
					Hint:         "dd deletes the entire current line",
				},
			},
			{
				ID:    "03d_cw",
				Title: "Change a word",
				Explanation: `Change "quick" to "fast".
Use: cw (then type the replacement, then Esc)`,
				InitText:   "The quick brown fox",
				CursorLine: 0, CursorCol: 4,
				Challenge: &curriculum.Challenge{
					Goal:         `Change "quick" to "fast"`,
					ExpectedText: "The fast brown fox",
					Hint:         "cw deletes the word and enters insert mode. Type 'fast' then Esc.",
				},
			},
			{
				ID:    "03e_d_dollar",
				Title: "Delete to end of line",
				Explanation: `Delete everything from "brown" to the end.
Use: d$`,
				InitText:   "The quick brown fox",
				CursorLine: 0, CursorCol: 10,
				Challenge: &curriculum.Challenge{
					Goal:         "Delete from cursor to end of line",
					ExpectedText: "The quick ",
					ExpectedLine: 0,
					Hint:         "d$ deletes from cursor to end of line",
				},
			},
			{
				ID:    "03f_yank_paste",
				Title: "Yank and paste",
				Explanation: `Yank line 1, then paste it below line 3.
Step 1: yy on line 1 (cursor already there)
Step 2: move to line 3 with 2j
Step 3: p to paste below`,
				InitText:   "first\nsecond\nthird",
				CursorLine: 0, CursorCol: 0,
				Challenge: &curriculum.Challenge{
					Goal:         "Duplicate line 1 below line 3",
					ExpectedText: "first\nsecond\nthird\nfirst",
					ExpectedLine: 3,
					Hint:         "yy yanks the current line. Move with 2j. Then p pastes below.",
				},
			},
			{
				ID:    "03g_dot",
				Title: "The dot command",
				Explanation: `The . command repeats the last change.

Delete the first "bad", then use . to delete the next two.`,
				InitText:   "bad good bad good bad",
				CursorLine: 0, CursorCol: 0,
				Challenge: &curriculum.Challenge{
					Goal:         "Remove all three 'bad' occurrences",
					ExpectedText: "good good good",
					Hint:         "dw deletes 'bad ', then use w to move to next 'bad', then .",
				},
			},
		},
	})
}
