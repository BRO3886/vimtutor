package theme

import "github.com/charmbracelet/lipgloss"

// Color palette — a dark, rich vim-inspired theme.
var (
	// Core colors
	colorBg        = lipgloss.Color("#1a1b26") // tokyo night bg
	colorBgAlt     = lipgloss.Color("#24283b")
	colorBgHighlight = lipgloss.Color("#2f3241")
	colorFg        = lipgloss.Color("#c0caf5")
	colorFgMuted   = lipgloss.Color("#565f89")
	colorFgSubtle  = lipgloss.Color("#414868")

	// Accent colors
	colorBlue      = lipgloss.Color("#7aa2f7")
	colorCyan      = lipgloss.Color("#7dcfff")
	colorGreen     = lipgloss.Color("#9ece6a")
	colorYellow    = lipgloss.Color("#e0af68")
	colorOrange    = lipgloss.Color("#ff9e64")
	colorRed       = lipgloss.Color("#f7768e")
	colorPurple    = lipgloss.Color("#bb9af7")
	colorMagenta   = lipgloss.Color("#ff007c")

	// Mode colors
	colorNormal = lipgloss.Color("#7aa2f7") // blue
	colorInsert = lipgloss.Color("#9ece6a") // green
	colorVisual = lipgloss.Color("#bb9af7") // purple

	// XP/level
	colorXP    = lipgloss.Color("#e0af68")
	colorLevel = lipgloss.Color("#ff9e64")
)

// Styles are reusable lipgloss styles.
var Styles = struct {
	// Base
	App    lipgloss.Style
	Panel  lipgloss.Style
	Subtle lipgloss.Style

	// Status bar
	StatusBar     lipgloss.Style
	StatusMode    lipgloss.Style
	StatusNormal  lipgloss.Style
	StatusInsert  lipgloss.Style
	StatusVisual  lipgloss.Style
	StatusInfo    lipgloss.Style
	StatusPending lipgloss.Style

	// Editor
	EditorBox      lipgloss.Style
	LineNumber     lipgloss.Style
	LineNumberCur  lipgloss.Style
	CurrentLine    lipgloss.Style
	CursorBlock    lipgloss.Style

	// UI elements
	Title       lipgloss.Style
	Subtitle    lipgloss.Style
	Explanation lipgloss.Style
	Goal        lipgloss.Style
	Hint        lipgloss.Style
	Success     lipgloss.Style
	Failure     lipgloss.Style
	Separator   lipgloss.Style

	// Progress
	ProgressFill  lipgloss.Style
	ProgressEmpty lipgloss.Style
	XPBar         lipgloss.Style
	XPFill        lipgloss.Style

	// Menu
	MenuItem       lipgloss.Style
	MenuItemActive lipgloss.Style
	MenuTag        lipgloss.Style
	MenuLocked     lipgloss.Style

	// Stats
	StatLabel lipgloss.Style
	StatValue lipgloss.Style
	StatBig   lipgloss.Style
	Streak    lipgloss.Style
	Level     lipgloss.Style
}{
	App:    lipgloss.NewStyle().Background(colorBg).Foreground(colorFg),
	Panel:  lipgloss.NewStyle().Background(colorBgAlt).Foreground(colorFg).Padding(1, 2),
	Subtle: lipgloss.NewStyle().Foreground(colorFgMuted),

	StatusBar:     lipgloss.NewStyle().Background(colorBgAlt).Foreground(colorFg),
	StatusMode:    lipgloss.NewStyle().Bold(true).Padding(0, 1),
	StatusNormal:  lipgloss.NewStyle().Background(colorNormal).Foreground(colorBg).Bold(true).Padding(0, 1),
	StatusInsert:  lipgloss.NewStyle().Background(colorInsert).Foreground(colorBg).Bold(true).Padding(0, 1),
	StatusVisual:  lipgloss.NewStyle().Background(colorVisual).Foreground(colorBg).Bold(true).Padding(0, 1),
	StatusInfo:    lipgloss.NewStyle().Background(colorBgHighlight).Foreground(colorFgMuted).Padding(0, 1),
	StatusPending: lipgloss.NewStyle().Foreground(colorYellow).Bold(true),

	EditorBox:     lipgloss.NewStyle().Background(colorBg).Foreground(colorFg),
	LineNumber:    lipgloss.NewStyle().Foreground(colorFgSubtle).Width(4).Align(lipgloss.Right),
	LineNumberCur: lipgloss.NewStyle().Foreground(colorYellow).Width(4).Align(lipgloss.Right).Bold(true),
	CurrentLine:   lipgloss.NewStyle().Background(colorBgHighlight),
	CursorBlock:   lipgloss.NewStyle().Background(colorBlue).Foreground(colorBg),

	Title:       lipgloss.NewStyle().Foreground(colorBlue).Bold(true).MarginBottom(1),
	Subtitle:    lipgloss.NewStyle().Foreground(colorCyan).Bold(true),
	Explanation: lipgloss.NewStyle().Foreground(colorFg),
	Goal:        lipgloss.NewStyle().Foreground(colorYellow).Bold(true),
	Hint:        lipgloss.NewStyle().Foreground(colorOrange).Italic(true),
	Success:     lipgloss.NewStyle().Foreground(colorGreen).Bold(true),
	Failure:     lipgloss.NewStyle().Foreground(colorRed).Bold(true),
	Separator:   lipgloss.NewStyle().Foreground(colorFgSubtle),

	ProgressFill:  lipgloss.NewStyle().Foreground(colorBlue),
	ProgressEmpty: lipgloss.NewStyle().Foreground(colorFgSubtle),
	XPBar:         lipgloss.NewStyle().Foreground(colorFgMuted),
	XPFill:        lipgloss.NewStyle().Foreground(colorXP),

	MenuItem:       lipgloss.NewStyle().Foreground(colorFg).Padding(0, 2),
	MenuItemActive: lipgloss.NewStyle().Foreground(colorBlue).Bold(true).Padding(0, 2).Background(colorBgHighlight),
	MenuTag:        lipgloss.NewStyle().Foreground(colorPurple).Italic(true),
	MenuLocked:     lipgloss.NewStyle().Foreground(colorFgSubtle).Padding(0, 2),

	StatLabel: lipgloss.NewStyle().Foreground(colorFgMuted),
	StatValue: lipgloss.NewStyle().Foreground(colorFg).Bold(true),
	StatBig:   lipgloss.NewStyle().Foreground(colorBlue).Bold(true),
	Streak:    lipgloss.NewStyle().Foreground(colorOrange).Bold(true),
	Level:     lipgloss.NewStyle().Foreground(colorLevel).Bold(true),
}

// ProgressBar renders a simple text progress bar.
func ProgressBar(current, total, width int) string {
	if total == 0 {
		return Styles.ProgressEmpty.Render(repeatChar('░', width))
	}
	filled := current * width / total
	if filled > width {
		filled = width
	}
	bar := Styles.ProgressFill.Render(repeatChar('█', filled))
	bar += Styles.ProgressEmpty.Render(repeatChar('░', width-filled))
	return bar
}

// Sparkline renders a simple bar sparkline from values.
func Sparkline(values []float64, width int) string {
	if len(values) == 0 {
		return ""
	}
	bars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

	maxVal := 0.0
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal == 0 {
		maxVal = 1
	}

	// use up to width values
	vals := values
	if len(vals) > width {
		vals = vals[len(vals)-width:]
	}

	result := ""
	for _, v := range vals {
		idx := int(v / maxVal * float64(len(bars)-1))
		if idx >= len(bars) {
			idx = len(bars) - 1
		}
		bar := string(bars[idx])
		result += Styles.XPFill.Render(bar)
	}
	return result
}

// ColorGreenExport exports the green color for use in other packages.
func ColorGreenExport() lipgloss.Color {
	return colorGreen
}

func repeatChar(ch rune, n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = ch
	}
	return string(b)
}
