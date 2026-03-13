package ui

import (
	"fmt"
	"strings"

	"github.com/BRO3886/vimtutor/internal/curriculum"
	"github.com/BRO3886/vimtutor/internal/metrics"
	"github.com/BRO3886/vimtutor/internal/storage"
	"github.com/BRO3886/vimtutor/internal/ui/screens"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	// import all lessons
	_ "github.com/BRO3886/vimtutor/internal/curriculum/lessons"
)

type screenKind int

const (
	screenMenu screenKind = iota
	screenChallenge
	screenDashboard
)

// AppModel is the root Bubble Tea model.
type AppModel struct {
	db     *storage.DB
	screen screenKind
	width  int
	height int

	// Sub-models
	menu      *screens.MenuScreen
	challenge *screens.ChallengeScreen
	dashboard *screens.DashboardScreen

	// Transient message to show on menu
	flash string
}

// NewApp creates the root application model.
func NewApp(db *storage.DB) *AppModel {
	return &AppModel{
		db:     db,
		screen: screenMenu,
	}
}

// Init implements tea.Model.
func (a *AppModel) Init() tea.Cmd {
	return tea.EnterAltScreen
}

// Update implements tea.Model.
func (a *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = m.Width
		a.height = m.Height
		a.ensureMenu()
		a.menu.Update(m)

	case tea.KeyMsg:
		// Global quit
		if m.String() == "ctrl+c" {
			return a, tea.Quit
		}
		if m.String() == "q" && a.screen == screenMenu {
			return a, tea.Quit
		}
		if m.String() == "q" && a.screen == screenDashboard {
			a.screen = screenMenu
			a.ensureMenu()
			return a, nil
		}
		if m.String() == "q" && a.screen == screenChallenge {
			if a.challenge != nil && a.challenge.IsDone() {
				a.screen = screenMenu
				a.ensureMenu()
				return a, nil
			}
		}
	}

	switch a.screen {
	case screenMenu:
		a.ensureMenu()
		_, cmd := a.menu.Update(msg)
		// Check for screen transition messages
		switch msg.(type) {
		case screens.StartLessonMsg:
			slMsg := msg.(screens.StartLessonMsg)
			return a, tea.Batch(cmd, a.startLesson(slMsg.LessonID))
		case screens.ShowDashboardMsg:
			a.screen = screenDashboard
			a.dashboard = screens.NewDashboardScreen(a.db, a.width, a.height)
			return a, nil
		}
		return a, cmd

	case screenChallenge:
		if a.challenge == nil {
			a.screen = screenMenu
			a.ensureMenu()
			return a, nil
		}
		_, cmd := a.challenge.Update(msg)

		if a.challenge.IsDone() {
			// Save results
			sum := a.challenge.Summary()
			a.flash = fmt.Sprintf("Lesson complete! +%d XP earned", sum.XPEarned)
			a.screen = screenMenu
			a.ensureMenu()
		}
		return a, cmd

	case screenDashboard:
		if a.dashboard == nil {
			a.screen = screenMenu
			a.ensureMenu()
			return a, nil
		}
		_, cmd := a.dashboard.Update(msg)
		switch msg.(type) {
		case screens.BackToMenuMsg:
			a.screen = screenMenu
			a.ensureMenu()
			return a, nil
		}
		return a, cmd
	}

	return a, nil
}

// View implements tea.Model.
func (a *AppModel) View() string {
	if a.width == 0 {
		return "Loading..."
	}

	switch a.screen {
	case screenMenu:
		a.ensureMenu()
		view := a.menu.View()
		if a.flash != "" {
			flashLine := lipgloss.NewStyle().Foreground(lipgloss.Color("#9ece6a")).Render("  " + a.flash)
			view = flashLine + "\n" + view
		}
		return view

	case screenChallenge:
		if a.challenge == nil {
			return "Loading lesson..."
		}
		return a.challenge.View()

	case screenDashboard:
		if a.dashboard == nil {
			return "Loading stats..."
		}
		return a.dashboard.View()
	}

	return ""
}

func (a *AppModel) ensureMenu() {
	if a.menu == nil {
		a.menu = screens.NewMenuScreen(a.db, a.width, a.height)
	}
}

func (a *AppModel) startLesson(id curriculum.LessonID) tea.Cmd {
	return func() tea.Msg {
		lesson, err := curriculum.Get(id)
		if err != nil {
			return nil
		}

		sessionID, err := a.db.CreateSession(string(id), "lesson")
		if err != nil {
			sessionID = 0
		}

		recorder := metrics.NewRecorder(sessionID, string(id), "lesson")
		a.challenge = screens.NewChallengeScreen(lesson, recorder, a.width, a.height)
		a.screen = screenChallenge
		a.flash = ""
		return nil
	}
}

// Run starts the TUI application.
func Run(db *storage.DB) error {
	app := NewApp(db)
	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// RunLesson starts directly into a specific lesson.
func RunLesson(db *storage.DB, lessonID string) error {
	app := NewApp(db)

	// Find the lesson
	lesson, err := curriculum.Get(curriculum.LessonID(lessonID))
	if err != nil {
		return fmt.Errorf("lesson %q not found: %w", lessonID, err)
	}

	// We need to start with a window size — we'll let the WindowSizeMsg handle it
	_ = lesson

	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

// Exported for tests
func indentStr(s string, prefix string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = prefix + l
	}
	return strings.Join(lines, "\n")
}
