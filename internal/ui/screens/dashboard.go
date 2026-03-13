package screens

import (
	"fmt"
	"strings"
	"time"

	"github.com/BRO3886/vimtutor/internal/metrics"
	"github.com/BRO3886/vimtutor/internal/storage"
	"github.com/BRO3886/vimtutor/internal/ui/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BackToMenuMsg signals returning to the menu.
type BackToMenuMsg struct{}

// DashboardScreen shows progress, stats, and charts.
type DashboardScreen struct {
	stats  *metrics.OverallStats
	width  int
	height int
}

// NewDashboardScreen creates a dashboard with current stats.
func NewDashboardScreen(db *storage.DB, width, height int) *DashboardScreen {
	stats, _ := db.GetOverallStats()
	return &DashboardScreen{
		stats:  stats,
		width:  width,
		height: height,
	}
}

// Init implements tea.Model.
func (d *DashboardScreen) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (d *DashboardScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "backspace":
			return d, func() tea.Msg { return BackToMenuMsg{} }
		}
	}
	return d, nil
}

// View implements tea.Model.
func (d *DashboardScreen) View() string {
	if d.stats == nil {
		return theme.Styles.Subtle.Render("  No stats yet. Complete a lesson first!")
	}

	var sections []string

	// Title
	sections = append(sections, theme.Styles.Title.Render("  Stats & Progress"))
	sections = append(sections, theme.Styles.Separator.Render(strings.Repeat("─", d.width)))

	// Level + XP
	sections = append(sections, d.renderLevel())
	sections = append(sections, "")

	// Key stats row
	sections = append(sections, d.renderKeyStats())
	sections = append(sections, "")

	// Accuracy sparkline
	sections = append(sections, d.renderSparkline())
	sections = append(sections, "")

	// Top mistakes
	sections = append(sections, d.renderTopMistakes())
	sections = append(sections, "")

	// Footer
	sections = append(sections, theme.Styles.Separator.Render(strings.Repeat("─", d.width)))
	sections = append(sections, theme.Styles.Subtle.Render("  q/esc → back to menu"))

	return strings.Join(sections, "\n")
}

func (d *DashboardScreen) renderLevel() string {
	st := d.stats
	xpCur, xpNeeded := metrics.XPToNextLevel(st.TotalXP)

	levelBig := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff9e64")).Bold(true).
		Render(fmt.Sprintf("  Level %d", st.Level+1))
	levelName := theme.Styles.Level.Render(fmt.Sprintf(" — %s", st.LevelName))
	xpInfo := theme.Styles.Subtle.Render(fmt.Sprintf("  (%d/%d XP to next level)", xpCur, xpNeeded))

	bar := theme.ProgressBar(xpCur, xpNeeded, 30)

	return levelBig + levelName + "\n  " + bar + xpInfo
}

func (d *DashboardScreen) renderKeyStats() string {
	st := d.stats

	acc := fmt.Sprintf("%.1f%%", st.AvgAccuracy*100)
	totalTime := formatDuration(st.TotalTimeSecs)

	col1 := theme.Styles.StatLabel.Render("  Sessions:  ") + theme.Styles.StatValue.Render(fmt.Sprintf("%d", st.TotalSessions))
	col2 := theme.Styles.StatLabel.Render("  Streak:    ") + theme.Styles.Streak.Render(fmt.Sprintf("%d days 🔥", st.StreakDays))
	col3 := theme.Styles.StatLabel.Render("  Accuracy:  ") + theme.Styles.StatValue.Render(acc)
	col4 := theme.Styles.StatLabel.Render("  Time:      ") + theme.Styles.StatValue.Render(totalTime)
	col5 := theme.Styles.StatLabel.Render("  Lessons:   ") + theme.Styles.StatValue.Render(fmt.Sprintf("%d", st.LessonsCompleted))

	return col1 + "\n" + col2 + "\n" + col3 + "\n" + col4 + "\n" + col5
}

func (d *DashboardScreen) renderSparkline() string {
	st := d.stats
	if len(st.DailyXP) == 0 {
		return theme.Styles.Subtle.Render("  XP history: not enough data yet")
	}

	values := make([]float64, len(st.DailyXP))
	for i, day := range st.DailyXP {
		values[i] = float64(day.XPEarned)
	}

	spark := theme.Sparkline(values, 30)
	label := theme.Styles.StatLabel.Render("  XP (14 days):  ")

	// Date range
	first := ""
	last := ""
	if len(st.DailyXP) > 0 {
		first = formatDate(st.DailyXP[0].Date)
		last = formatDate(st.DailyXP[len(st.DailyXP)-1].Date)
	}
	dateRange := theme.Styles.Subtle.Render(fmt.Sprintf("  %s → %s", first, last))

	return label + spark + "\n" + dateRange
}

func (d *DashboardScreen) renderTopMistakes() string {
	st := d.stats
	if len(st.TopMistakes) == 0 {
		return theme.Styles.Subtle.Render("  Top mistakes: none yet")
	}

	var rows []string
	rows = append(rows, theme.Styles.StatLabel.Render("  Most missed:"))

	maxCount := st.TopMistakes[0].Count
	for i, m := range st.TopMistakes {
		if i >= 5 {
			break
		}
		bar := theme.ProgressBar(m.Count, maxCount, 10)
		rows = append(rows, fmt.Sprintf("    %-8s %s  %d",
			theme.Styles.StatValue.Render(m.Key),
			bar,
			m.Count,
		))
	}

	return strings.Join(rows, "\n")
}

func formatDuration(secs float64) string {
	d := time.Duration(secs) * time.Second
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

func formatDate(s string) string {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return s
	}
	return t.Format("Jan 2")
}
