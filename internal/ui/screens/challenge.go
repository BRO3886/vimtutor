package screens

import (
	"fmt"
	"strings"
	"time"

	"github.com/BRO3886/vimtutor/internal/curriculum"
	"github.com/BRO3886/vimtutor/internal/engine"
	"github.com/BRO3886/vimtutor/internal/metrics"
	"github.com/BRO3886/vimtutor/internal/ui/theme"
	"github.com/BRO3886/vimtutor/internal/ui/components"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type challengeStatus int

const (
	statusPending challengeStatus = iota
	statusPassed
	statusFailed
)

// ChallengeMsg is sent when a challenge result is known.
type ChallengeMsg struct {
	Passed  bool
	Summary metrics.SessionSummary
}

// NextStepMsg signals moving to the next step.
type NextStepMsg struct{}

// LessonDoneMsg signals the lesson is complete.
type LessonDoneMsg struct {
	Summary metrics.SessionSummary
}

// ChallengeScreen is the main interactive lesson/challenge screen.
type ChallengeScreen struct {
	lesson   *curriculum.Lesson
	stepIdx  int
	buf      *engine.Buffer
	parser   *engine.Parser
	recorder *metrics.Recorder

	width  int
	height int

	status     challengeStatus
	attempts   int
	hintShown  bool
	feedback   string
	feedbackAt time.Time

	stepKeystrokes int
	stepMistakes   []string
	stepStart      time.Time

	// explanatory step (no challenge) - waiting for space
	waitingSpace bool

	// post-lesson result
	done    bool
	summary metrics.SessionSummary
}

// NewChallengeScreen creates a fresh challenge screen for a lesson.
func NewChallengeScreen(lesson *curriculum.Lesson, recorder *metrics.Recorder, width, height int) *ChallengeScreen {
	s := &ChallengeScreen{
		lesson:   lesson,
		stepIdx:  0,
		parser:   engine.NewParser(),
		recorder: recorder,
		width:    width,
		height:   height,
	}
	s.loadStep()
	return s
}

func (s *ChallengeScreen) loadStep() {
	if s.stepIdx >= len(s.lesson.Steps) {
		s.done = true
		return
	}
	step := s.lesson.Steps[s.stepIdx]
	s.buf = engine.NewBuffer(step.InitText)
	s.buf.SetCursor(step.CursorLine, step.CursorCol)
	s.parser.Reset()
	s.status = statusPending
	s.attempts = 0
	s.hintShown = false
	s.feedback = ""
	s.stepKeystrokes = 0
	s.stepMistakes = nil
	s.stepStart = time.Now()
	s.waitingSpace = step.Challenge == nil
	s.recorder.StartStep()
}

func (s *ChallengeScreen) currentStep() curriculum.Step {
	return s.lesson.Steps[s.stepIdx]
}

// Init implements tea.Model.
func (s *ChallengeScreen) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (s *ChallengeScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = m.Width
		s.height = m.Height

	case tea.KeyMsg:
		if s.done {
			return s, nil
		}

		key := m.String()

		// Global escape in normal mode resets parser
		if key == "esc" || key == "ctrl+[" {
			if s.buf.Mode() == engine.ModeInsert {
				engine.InsertKey(s.buf, key)
				s.stepKeystrokes++
				s.checkChallenge()
				return s, nil
			}
			s.parser.Reset()
			return s, nil
		}

		// Explanatory step: space to continue
		if s.waitingSpace {
			if key == " " || key == "enter" {
				s.advance()
			}
			return s, nil
		}

		// Status: passed - space/enter to continue
		if s.status == statusPassed {
			if key == " " || key == "enter" {
				s.advance()
			}
			return s, nil
		}

		// Insert mode
		if s.buf.Mode() == engine.ModeInsert {
			s.stepKeystrokes++
			engine.InsertKey(s.buf, key)
			s.checkChallenge()
			return s, nil
		}

		// Normal mode: feed key to parser
		cmd, needMore := s.parser.Feed(key)
		s.stepKeystrokes++

		if needMore {
			return s, nil
		}

		if cmd != nil {
			enteredInsert := engine.Execute(s.buf, cmd)
			_ = enteredInsert
			s.checkChallenge()
		}

	case clearFeedbackMsg:
		s.feedback = ""
	}

	return s, nil
}

type clearFeedbackMsg struct{}

func (s *ChallengeScreen) checkChallenge() {
	step := s.currentStep()
	if step.Challenge == nil {
		return
	}

	ch := step.Challenge
	text := s.buf.Text()
	cur := s.buf.GetCursor()

	passed, _ := ch.Check(text, cur.Line, cur.Col)

	if passed {
		s.status = statusPassed
		xp := metrics.XPChallengeFirst
		if s.attempts > 0 {
			xp = metrics.XPChallengeHinted
		}
		s.recorder.AddXP(xp)
		s.feedback = fmt.Sprintf("✓  Passed! +%dXP", xp)
		s.recorder.RecordResult(
			string(s.lesson.ID), step.ID,
			s.attempts+1, s.stepKeystrokes,
			true, s.stepMistakes,
		)
	}
}

func (s *ChallengeScreen) advance() {
	s.stepIdx++
	if s.stepIdx >= len(s.lesson.Steps) {
		s.done = true
		s.recorder.AddXP(metrics.XPLessonComplete)
		_, _, _, sum := s.recorder.Finalize()
		s.summary = sum
		return
	}
	s.loadStep()
}

// View implements tea.Model.
func (s *ChallengeScreen) View() string {
	if s.done {
		return s.renderComplete()
	}

	step := s.currentStep()

	// Layout areas
	headerH := 3
	explanationH := countLines(step.Explanation) + 2
	goalH := 3
	editorH := s.height - headerH - explanationH - goalH - 3 // 3 for status bar + padding
	if editorH < 5 {
		editorH = 5
	}

	var sections []string

	// Header: lesson title + progress
	sections = append(sections, s.renderHeader())

	// Divider
	sections = append(sections, theme.Styles.Separator.Render(strings.Repeat("─", s.width)))

	// Explanation
	if step.Explanation != "" {
		sections = append(sections, s.renderExplanation(step))
		sections = append(sections, theme.Styles.Separator.Render(strings.Repeat("─", s.width)))
	}

	// Goal
	if step.Challenge != nil {
		sections = append(sections, s.renderGoal(step))
	}

	// Editor
	if step.InitText != "" || step.Challenge != nil {
		editorContent := components.RenderEditor(s.buf, components.EditorOpts{
			Width:        s.width,
			Height:       editorH,
			ShowLineNums: true,
		})
		boxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#414868")).
			Padding(0, 1).
			Width(s.width - 2)
		sections = append(sections, boxStyle.Render(editorContent))
	}

	// Feedback
	if s.feedback != "" {
		switch s.status {
		case statusPassed:
			sections = append(sections, theme.Styles.Success.Render("  "+s.feedback))
		case statusFailed:
			sections = append(sections, theme.Styles.Failure.Render("  "+s.feedback))
		default:
			sections = append(sections, theme.Styles.Hint.Render("  "+s.feedback))
		}
	}

	// Hint
	if s.hintShown && step.Challenge != nil {
		sections = append(sections, theme.Styles.Hint.Render("  Hint: "+step.Challenge.Hint))
	}

	// Waiting for space indicator
	if s.waitingSpace || s.status == statusPassed {
		cont := theme.Styles.Subtle.Render("  Press SPACE to continue →")
		sections = append(sections, cont)
	}

	// Status bar
	pendingKeys := s.parser.PendingKeys()
	extra := fmt.Sprintf("step %d/%d", s.stepIdx+1, len(s.lesson.Steps))
	statusBar := components.RenderStatusBar(s.buf, pendingKeys, s.width, extra)
	sections = append(sections, statusBar)

	return strings.Join(sections, "\n")
}

func (s *ChallengeScreen) renderHeader() string {
	title := theme.Styles.Title.Render(fmt.Sprintf("  %s", s.lesson.Title))

	// Progress bar
	total := len(s.lesson.Steps)
	current := s.stepIdx
	if s.status == statusPassed {
		current++
	}
	pbar := theme.ProgressBar(current, total, 20)
	pct := fmt.Sprintf("%d%%", current*100/total)
	progress := pbar + " " + theme.Styles.Subtle.Render(pct)

	gap := s.width - lipgloss.Width(title) - lipgloss.Width(progress) - 4
	if gap < 0 {
		gap = 0
	}
	return title + strings.Repeat(" ", gap) + progress + "  "
}

func (s *ChallengeScreen) renderExplanation(step curriculum.Step) string {
	stepTitle := theme.Styles.Subtitle.Render(fmt.Sprintf("  %s", step.Title))
	text := theme.Styles.Explanation.Render(indentLines(step.Explanation, "  "))
	return stepTitle + "\n" + text
}

func (s *ChallengeScreen) renderGoal(step curriculum.Step) string {
	goalLabel := theme.Styles.Subtle.Render("  Goal  ")
	goalText := theme.Styles.Goal.Render(step.Challenge.Goal)
	return goalLabel + goalText
}

func (s *ChallengeScreen) renderComplete() string {
	sum := s.summary
	acc := fmt.Sprintf("%.0f%%", sum.Accuracy*100)

	lines := []string{
		"",
		theme.Styles.Success.Render("  ✓  Lesson Complete!"),
		"",
		theme.Styles.StatLabel.Render("  Steps completed: ") + theme.Styles.StatValue.Render(fmt.Sprintf("%d", sum.StepsCompleted)),
		theme.Styles.StatLabel.Render("  Accuracy:        ") + theme.Styles.StatValue.Render(acc),
		theme.Styles.StatLabel.Render("  XP earned:       ") + theme.Styles.XPFill.Render(fmt.Sprintf("+%d XP", sum.XPEarned)),
		"",
		theme.Styles.Subtle.Render("  Press q to return to menu"),
	}

	return strings.Join(lines, "\n")
}

// IsDone returns whether the lesson is complete.
func (s *ChallengeScreen) IsDone() bool {
	return s.done
}

// Summary returns the session summary (only valid when done).
func (s *ChallengeScreen) Summary() metrics.SessionSummary {
	return s.summary
}

func countLines(s string) int {
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}

func indentLines(text, prefix string) string {
	lines := strings.Split(text, "\n")
	for i, l := range lines {
		if l != "" {
			lines[i] = prefix + l
		}
	}
	return strings.Join(lines, "\n")
}
