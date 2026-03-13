package screens

import (
	"fmt"
	"strings"

	"github.com/BRO3886/vimtutor/internal/curriculum"
	"github.com/BRO3886/vimtutor/internal/metrics"
	"github.com/BRO3886/vimtutor/internal/storage"
	"github.com/BRO3886/vimtutor/internal/ui/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	// import all lessons to trigger their init() calls
	_ "github.com/BRO3886/vimtutor/internal/curriculum/lessons"
)

// StartLessonMsg signals that a lesson should start.
type StartLessonMsg struct {
	LessonID curriculum.LessonID
}

// ShowDashboardMsg signals to show the dashboard.
type ShowDashboardMsg struct{}

// MenuScreen is the main lesson selection menu.
type MenuScreen struct {
	lessons  []curriculum.Lesson
	progress map[string]metrics.UserProgress
	stats    *metrics.OverallStats
	cursor   int
	width    int
	height   int
	db       *storage.DB
}

// NewMenuScreen creates a new menu screen.
func NewMenuScreen(db *storage.DB, width, height int) *MenuScreen {
	lessons := curriculum.All()
	progress, _ := db.GetAllProgress()
	stats, _ := db.GetOverallStats()

	return &MenuScreen{
		lessons:  lessons,
		progress: progress,
		stats:    stats,
		width:    width,
		height:   height,
		db:       db,
	}
}

// Init implements tea.Model.
func (m *MenuScreen) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m *MenuScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.lessons)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "enter", " ", "l":
			if m.cursor < len(m.lessons) {
				lesson := m.lessons[m.cursor]
				if m.isUnlocked(lesson.ID) {
					return m, func() tea.Msg {
						return StartLessonMsg{LessonID: lesson.ID}
					}
				}
			}
		case "s":
			return m, func() tea.Msg {
				return ShowDashboardMsg{}
			}
		case "g":
			m.cursor = 0
		case "G":
			m.cursor = len(m.lessons) - 1
		}
	}
	return m, nil
}

// View implements tea.Model.
func (m *MenuScreen) View() string {
	var sections []string

	// Title banner
	sections = append(sections, m.renderBanner())
	sections = append(sections, theme.Styles.Separator.Render(strings.Repeat("─", m.width)))

	// Stats row
	if m.stats != nil {
		sections = append(sections, m.renderStatsRow())
		sections = append(sections, theme.Styles.Separator.Render(strings.Repeat("─", m.width)))
	}

	// Lesson list
	sections = append(sections, m.renderLessonList())

	// Footer
	sections = append(sections, theme.Styles.Separator.Render(strings.Repeat("─", m.width)))
	sections = append(sections, m.renderFooter())

	return strings.Join(sections, "\n")
}

func (m *MenuScreen) renderBanner() string {
	banner := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7aa2f7")).
		Bold(true).
		Padding(1, 2).
		Render("vimtutor  —  learn vim by doing")

	return banner
}

func (m *MenuScreen) renderStatsRow() string {
	if m.stats == nil {
		return ""
	}

	xpCur, xpNeeded := metrics.XPToNextLevel(m.stats.TotalXP)

	xpBar := theme.ProgressBar(xpCur, xpNeeded, 12)
	levelStr := theme.Styles.Level.Render(fmt.Sprintf("Lv.%d %s", m.stats.Level+1, m.stats.LevelName))
	xpStr := theme.Styles.XPFill.Render(fmt.Sprintf("%dXP", m.stats.TotalXP))
	streakStr := theme.Styles.Streak.Render(fmt.Sprintf("🔥 %dd streak", m.stats.StreakDays))

	left := "  " + levelStr + "  " + xpBar + " " + xpStr
	right := streakStr + "  "
	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}

	return left + strings.Repeat(" ", gap) + right
}

func (m *MenuScreen) renderLessonList() string {
	var rows []string
	rows = append(rows, "")

	for i, lesson := range m.lessons {
		unlocked := m.isUnlocked(lesson.ID)
		prog := m.progress[string(lesson.ID)]

		var row string
		if i == m.cursor {
			if unlocked {
				arrow := theme.Styles.MenuItemActive.Render("▶  ")
				title := theme.Styles.MenuItemActive.Render(fmt.Sprintf("%-30s", lesson.Title))
				diff := m.difficultyBadge(lesson.Difficulty)
				completions := ""
				if prog.CompletedCount > 0 {
					completions = theme.Styles.Success.Render(fmt.Sprintf("  ✓ %dx", prog.CompletedCount))
				}
				row = arrow + title + diff + completions
			} else {
				row = theme.Styles.MenuLocked.Render(fmt.Sprintf("   🔒 %-30s  (complete previous lesson)", lesson.Title))
			}
		} else {
			if unlocked {
				num := theme.Styles.Subtle.Render(fmt.Sprintf("   %2d. ", i+1))
				title := theme.Styles.MenuItem.Render(fmt.Sprintf("%-30s", lesson.Title))
				diff := m.difficultyBadge(lesson.Difficulty)
				completions := ""
				if prog.CompletedCount > 0 {
					completions = theme.Styles.Success.Render(fmt.Sprintf("  ✓ %dx", prog.CompletedCount))
				}
				row = num + title + diff + completions
			} else {
				row = theme.Styles.MenuLocked.Render(fmt.Sprintf("   %2d. 🔒 %s", i+1, lesson.Title))
			}
		}

		rows = append(rows, row)
	}

	rows = append(rows, "")
	return strings.Join(rows, "\n")
}

func (m *MenuScreen) difficultyBadge(d curriculum.Difficulty) string {
	switch d {
	case curriculum.Beginner:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#9ece6a")).Render("  beginner")
	case curriculum.Intermediate:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#e0af68")).Render("  intermediate")
	case curriculum.Advanced:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#f7768e")).Render("  advanced")
	}
	return ""
}

func (m *MenuScreen) isUnlocked(id curriculum.LessonID) bool {
	// First lesson is always unlocked
	if len(m.lessons) > 0 && m.lessons[0].ID == id {
		return true
	}
	return m.db.IsLessonUnlocked(string(id))
}

func (m *MenuScreen) renderFooter() string {
	keys := []string{
		"j/k navigate",
		"enter start lesson",
		"s stats",
		"q quit",
	}
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = theme.Styles.Subtle.Render("  " + k)
	}
	return strings.Join(parts, theme.Styles.Separator.Render("  ·"))
}
