package lessons

import "github.com/BRO3886/vimtutor/internal/curriculum"

func init() {
	curriculum.Register(curriculum.Lesson{
		ID:            "02_word_motions",
		Title:         "Word Motions",
		Description:   "Jump between words at the speed of thought.",
		Difficulty:    curriculum.Beginner,
		Tags:          []string{"motion", "word"},
		Prerequisites: []curriculum.LessonID{"01_basic_motions"},
		Steps: []curriculum.Step{
			{
				ID:    "02a_intro",
				Title: "Word motions",
				Explanation: `h/j/k/l are fine for short distances.
For longer moves, use word motions:

  w  → jump to start of NEXT word
  b  → jump BACK to start of current/prev word
  e  → jump to END of current/next word

Capital versions (W, B, E) treat sequences of
non-whitespace as one big "WORD" (ignoring punctuation).

Press SPACE to continue.`,
				InitText:   "",
				CursorLine: 0, CursorCol: 0,
			},
			{
				ID:    "02b_w_forward",
				Title: "Jump forward with w",
				Explanation: `Move forward by words to land on "fox".
Use: w`,
				InitText:   "The quick brown fox jumps",
				CursorLine: 0, CursorCol: 0,
				Challenge: &curriculum.Challenge{
					Goal:         `Cursor on "fox" (col 16)`,
					ExpectedLine: 0,
					ExpectedCol:  16,
					Hint:         "Press w three times, or 3w",
				},
			},
			{
				ID:    "02c_b_backward",
				Title: "Jump back with b",
				Explanation: `Move backward to land on "quick".
Use: b`,
				InitText:   "The quick brown fox jumps",
				CursorLine: 0, CursorCol: 16,
				Challenge: &curriculum.Challenge{
					Goal:         `Cursor on "quick" (col 4)`,
					ExpectedLine: 0,
					ExpectedCol:  4,
					Hint:         "Press b twice, or 2b",
				},
			},
			{
				ID:    "02d_e_end",
				Title: "Jump to word end with e",
				Explanation: `Jump to the end of the next word.
Use: e`,
				InitText:   "The quick brown fox",
				CursorLine: 0, CursorCol: 0,
				Challenge: &curriculum.Challenge{
					Goal:         `Cursor on "e" of "The" (col 2)`,
					ExpectedLine: 0,
					ExpectedCol:  2,
					Hint:         "Press e once",
				},
			},
			{
				ID:    "02e_count_prefix",
				Title: "Count prefix",
				Explanation: `Any motion can be preceded by a count.
3w = jump 3 words forward.

Move to "jumps" using a count.`,
				InitText:   "The quick brown fox jumps over",
				CursorLine: 0, CursorCol: 0,
				Challenge: &curriculum.Challenge{
					Goal:         `Cursor on "jumps" (col 20)`,
					ExpectedLine: 0,
					ExpectedCol:  20,
					Hint:         "Try 4w to jump 4 words forward",
				},
			},
			{
				ID:    "02f_bigword",
				Title: "WORD vs word",
				Explanation: `Lowercase w treats punctuation as word boundaries.
Uppercase W skips over everything to the next whitespace.

  "foo.bar.baz"  → w stops at each dot
                 → W jumps over the whole thing

Move past "foo.bar.baz" in one W jump.`,
				InitText:   "start foo.bar.baz end",
				CursorLine: 0, CursorCol: 6,
				Challenge: &curriculum.Challenge{
					Goal:         `Cursor on "end" (col 18)`,
					ExpectedLine: 0,
					ExpectedCol:  18,
					Hint:         "Press W once to jump over the whole foo.bar.baz as one WORD",
				},
			},
		},
	})
}
