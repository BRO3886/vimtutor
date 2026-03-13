package lessons

import "github.com/BRO3886/vimtutor/internal/curriculum"

func init() {
	curriculum.Register(curriculum.Lesson{
		ID:            "04_text_objects",
		Title:         "Text Objects",
		Description:   "This is where vim clicks. Operate on semantic chunks of text.",
		Difficulty:    curriculum.Intermediate,
		Tags:          []string{"text-object", "operator"},
		Prerequisites: []curriculum.LessonID{"03_operators"},
		Steps: []curriculum.Step{
			{
				ID:    "04a_intro",
				Title: "Text objects: the game changer",
				Explanation: `Text objects let you operate on structured text:

  iw   → inner word
  aw   → around word (includes trailing space)
  i"   → inside double quotes
  a"   → around double quotes (includes the quotes)
  i(   → inside parentheses
  a(   → around parentheses (includes parens)
  i{   → inside curly braces
  a{   → around curly braces

These combine with any operator:
  ci"   → change inside quotes
  da(   → delete around parens
  yi{   → yank inside braces

Press SPACE to continue.`,
				InitText:   "",
				CursorLine: 0, CursorCol: 0,
			},
			{
				ID:    "04b_ciw",
				Title: "Change inner word",
				Explanation: `Change the word "quick" to "slow".
Cursor is anywhere on "quick".
Use: ciw, then type "slow", then Esc`,
				InitText:   `The quick brown fox`,
				CursorLine: 0, CursorCol: 5,
				Challenge: &curriculum.Challenge{
					Goal:         `Change "quick" to "slow" using ciw`,
					ExpectedText: "The slow brown fox",
					Hint:         "ciw deletes the word under cursor and enters insert mode",
				},
			},
			{
				ID:    "04c_ci_quotes",
				Title: "Change inside quotes",
				Explanation: `Change what's inside the double quotes.
Replace "hello world" with "goodbye".
Use: ci"`,
				InitText:   `print("hello world")`,
				CursorLine: 0, CursorCol: 8,
				Challenge: &curriculum.Challenge{
					Goal:         `Change the string to "goodbye"`,
					ExpectedText: `print("goodbye")`,
					Hint:         `ci" deletes everything inside the quotes and enters insert mode`,
				},
			},
			{
				ID:    "04d_di_parens",
				Title: "Delete inside parens",
				Explanation: `Delete the arguments inside the function call.
Use: di(`,
				InitText:   `foo(x, y, z)`,
				CursorLine: 0, CursorCol: 5,
				Challenge: &curriculum.Challenge{
					Goal:         "Delete everything inside the parens",
					ExpectedText: "foo()",
					Hint:         "di( deletes everything between the parentheses",
				},
			},
			{
				ID:    "04e_da_parens",
				Title: "Delete around parens",
				Explanation: `Delete the parens AND their contents.
Use: da(`,
				InitText:   `foo(x, y, z) bar`,
				CursorLine: 0, CursorCol: 5,
				Challenge: &curriculum.Challenge{
					Goal:         "Delete parens and their contents",
					ExpectedText: "foo bar",
					Hint:         "da( includes the parentheses themselves",
				},
			},
			{
				ID:    "04f_ci_braces",
				Title: "Change inside braces",
				Explanation: `Replace the map contents with "key: val".
Use: ci{`,
				InitText:   `m := map[string]int{"foo": 1, "bar": 2}`,
				CursorLine: 0, CursorCol: 22,
				Challenge: &curriculum.Challenge{
					Goal:         `Replace map contents with "key: val"`,
					ExpectedText: `m := map[string]int{"key: val"}`,
					Hint:         `ci{ deletes inside the curly braces and enters insert mode`,
				},
			},
			{
				ID:    "04g_combined",
				Title: "Put it together",
				Explanation: `Change the function argument from "bad_name" to "good_name".
This requires: ci" or ciw depending on where your cursor is.`,
				InitText:   `rename("bad_name", true)`,
				CursorLine: 0, CursorCol: 9,
				Challenge: &curriculum.Challenge{
					Goal:         `Change "bad_name" to "good_name"`,
					ExpectedText: `rename("good_name", true)`,
					Hint:         `ci" will delete inside the quotes around bad_name`,
				},
			},
		},
	})
}
