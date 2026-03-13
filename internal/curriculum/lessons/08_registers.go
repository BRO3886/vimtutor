package lessons

import "github.com/BRO3886/vimtutor/internal/curriculum"

func init() {
	curriculum.Register(curriculum.Lesson{
		ID:            "08_registers",
		Title:         "Registers",
		Description:   "Vim's clipboard system — multiple named slots.",
		Difficulty:    curriculum.Intermediate,
		Tags:          []string{"registers", "yank"},
		Prerequisites: []curriculum.LessonID{"03_operators"},
		Steps: []curriculum.Step{
			{
				ID:    "08a_intro",
				Title: "Named registers",
				Explanation: `Vim has many registers. The most useful:

  "    → unnamed (default) register — every d/c/y goes here
  0    → yank register — last y always goes here (not deleted by d)
  a-z  → named registers — you control these

Usage:
  "ayy    → yank line into register a
  "ap     → paste from register a
  "0p     → paste last yanked text (ignoring deletes)

Tip: When you yank then delete something, "0p retrieves
     your original yank even though "" was overwritten.

Press SPACE to continue.`,
				InitText:   "",
				CursorLine: 0, CursorCol: 0,
			},
			{
				ID:    "08b_default_register",
				Title: "Yank and paste",
				Explanation: `Yank line 1 and paste it after line 3.
Use: yy then 2j then p`,
				InitText:   "important\nfoo\nbar",
				CursorLine: 0, CursorCol: 0,
				Challenge: &curriculum.Challenge{
					Goal:         `Duplicate line 1 after line 3`,
					ExpectedText: "important\nfoo\nbar\nimportant",
					Hint:         "yy yanks the line, 2j moves down, p pastes below",
				},
			},
			{
				ID:    "08c_named_register",
				Title: "Use a named register",
				Explanation: `Yank line 1 into register 'a'.
Then move to line 3 and paste from register 'a'.
Use: "ayy, then 2j, then "ap`,
				InitText:   "SAVE THIS\nignore\nPASTE HERE",
				CursorLine: 0, CursorCol: 0,
				Challenge: &curriculum.Challenge{
					Goal:         `Paste "SAVE THIS" after "PASTE HERE"`,
					ExpectedText: "SAVE THIS\nignore\nPASTE HERE\nSAVE THIS",
					Hint:         `"ayy yanks to register a; "ap pastes from register a`,
				},
			},
		},
	})
}
