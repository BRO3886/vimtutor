package curriculum

// LessonID uniquely identifies a lesson.
type LessonID string

// Difficulty levels for lessons.
type Difficulty int

const (
	Beginner     Difficulty = 1
	Intermediate Difficulty = 2
	Advanced     Difficulty = 3
)

func (d Difficulty) String() string {
	switch d {
	case Beginner:
		return "Beginner"
	case Intermediate:
		return "Intermediate"
	case Advanced:
		return "Advanced"
	}
	return "Unknown"
}

// Lesson is the top-level teaching unit.
type Lesson struct {
	ID            LessonID
	Title         string
	Description   string
	Difficulty    Difficulty
	Tags          []string
	Steps         []Step
	Prerequisites []LessonID
}

// Step is one teaching moment within a lesson.
// It may be explanatory (no challenge) or interactive (has a challenge).
type Step struct {
	ID          string
	Title       string
	Explanation string // shown above the editor
	InitText    string // buffer content when this step starts
	CursorLine  int
	CursorCol   int
	Challenge   *Challenge // nil = read-only explanatory step
}

// Challenge defines an interactive task the user must complete.
type Challenge struct {
	Goal          string // human-readable description
	ExpectedText  string // what the buffer should contain after completion
	ExpectedLine  int    // expected cursor line (-1 = don't check)
	ExpectedCol   int    // expected cursor col (-1 = don't check)
	MaxKeystrokes int    // 0 = unlimited
	Hint          string // shown after 2 failed attempts
}

// Check returns whether the given text and cursor match the expected output.
func (c *Challenge) Check(text string, line, col int) (passed bool, reason string) {
	if c.ExpectedText != "" && text != c.ExpectedText {
		return false, "buffer content doesn't match expected"
	}
	if c.ExpectedLine >= 0 && line != c.ExpectedLine {
		return false, "cursor on wrong line"
	}
	if c.ExpectedCol >= 0 && col != c.ExpectedCol {
		return false, "cursor in wrong column"
	}
	return true, ""
}
