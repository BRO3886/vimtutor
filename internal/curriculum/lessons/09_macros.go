package lessons

import "github.com/BRO3886/vimtutor/internal/curriculum"

func init() {
	curriculum.Register(curriculum.Lesson{
		ID:            "09_macros",
		Title:         "Macros",
		Description:   "Record and replay any sequence of commands.",
		Difficulty:    curriculum.Advanced,
		Tags:          []string{"macro", "advanced"},
		Prerequisites: []curriculum.LessonID{"08_registers"},
		Steps: []curriculum.Step{
			{
				ID:    "09a_intro",
				Title: "What are macros?",
				Explanation: `Macros record a sequence of keystrokes and replay them.

  q{a-z}   → start recording into register
  q        → stop recording
  @{a-z}   → replay the macro
  @@       → replay the last macro again
  5@a      → replay macro a five times

Macros are just text stored in registers.
A macro is perfect for repetitive structural edits.

Example workflow:
  1. Do the edit manually once, recording: qa ... q
  2. Move to the next thing to fix: j
  3. Replay: @a
  4. Keep going: 10@a

Press SPACE to continue.`,
				InitText:   "",
				CursorLine: 0, CursorCol: 0,
			},
			{
				ID:    "09b_basic_macro",
				Title: "Record and replay",
				Explanation: `Each line needs "- " prepended.
Record a macro to do it, then replay it.

Suggested approach:
  qa         → start recording
  I- <Esc>  → insert "- " at start
  j          → move to next line
  q          → stop recording
  @a         → replay on remaining lines`,
				InitText:   "apple\nbanana\ncherry",
				CursorLine: 0, CursorCol: 0,
				Challenge: &curriculum.Challenge{
					Goal:         `Prepend "- " to all three lines`,
					ExpectedText: "- apple\n- banana\n- cherry",
					Hint:         `Record with qa, use I- <Esc>j, stop with q, replay with 2@a`,
				},
			},
		},
	})
}
