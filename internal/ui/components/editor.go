package components

import (
	"fmt"
	"strings"

	"github.com/BRO3886/vimtutor/internal/engine"
	"github.com/BRO3886/vimtutor/internal/ui/theme"
	"github.com/charmbracelet/lipgloss"
)

// EditorOpts configures the editor renderer.
type EditorOpts struct {
	Width         int
	Height        int
	ShowLineNums  bool
	RelativeNums  bool
}

// RenderEditor renders the vim buffer as styled terminal text.
func RenderEditor(buf *engine.Buffer, opts EditorOpts) string {
	if opts.Width == 0 {
		opts.Width = 80
	}
	if opts.Height == 0 {
		opts.Height = 20
	}
	if opts.ShowLineNums == false {
		opts.ShowLineNums = true
	}

	cursor := buf.GetCursor()
	lineCount := buf.LineCount()
	mode := buf.Mode()

	// Determine which lines to show (scroll window)
	startLine := 0
	endLine := lineCount - 1
	editorH := opts.Height
	if lineCount > editorH {
		startLine = cursor.Line - editorH/2
		if startLine < 0 {
			startLine = 0
		}
		endLine = startLine + editorH - 1
		if endLine >= lineCount {
			endLine = lineCount - 1
			startLine = endLine - editorH + 1
			if startLine < 0 {
				startLine = 0
			}
		}
	}

	textWidth := opts.Width - 5 // leave room for line numbers

	var rows []string
	for lineIdx := startLine; lineIdx <= endLine; lineIdx++ {
		line := buf.Line(lineIdx)
		isCursorLine := lineIdx == cursor.Line

		// Line number
		lineNumStr := ""
		if opts.ShowLineNums {
			lineNum := lineIdx + 1
			if isCursorLine {
				lineNumStr = theme.Styles.LineNumberCur.Render(fmt.Sprintf("%3d ", lineNum))
			} else {
				lineNumStr = theme.Styles.LineNumber.Render(fmt.Sprintf("%3d ", lineNum))
			}
		}

		// Render line content with cursor
		content := renderLineContent(line, cursor.Col, isCursorLine, mode, textWidth)

		row := lineNumStr + content
		if isCursorLine && opts.Height > 1 {
			row = theme.Styles.CurrentLine.Render(lineNumStr) + content
		}
		rows = append(rows, row)
	}

	// Fill empty lines
	for len(rows) < editorH {
		rows = append(rows, theme.Styles.Subtle.Render("~"))
	}

	return strings.Join(rows, "\n")
}

func renderLineContent(line string, cursorCol int, isCursorLine bool, mode engine.Mode, maxWidth int) string {
	runes := []rune(line)

	if len(runes) == 0 {
		if isCursorLine && mode == engine.ModeNormal {
			// Show cursor block on empty line
			return theme.Styles.CursorBlock.Render(" ")
		}
		return ""
	}

	var sb strings.Builder
	for i, r := range runes {
		ch := string(r)
		if r == '\t' {
			ch = "    " // expand tabs to 4 spaces for display
		}

		if isCursorLine && i == cursorCol {
			switch mode {
			case engine.ModeInsert:
				// thin cursor in insert mode
				sb.WriteString(lipgloss.NewStyle().
					Underline(true).Foreground(theme.ColorGreenExport()).Render(ch))
			default:
				sb.WriteString(theme.Styles.CursorBlock.Render(ch))
			}
		} else if isCursorLine {
			sb.WriteString(theme.Styles.CurrentLine.Render(ch))
		} else {
			sb.WriteString(ch)
		}
	}

	// If cursor is past end of line (in insert mode), show a trailing cursor
	if isCursorLine && cursorCol >= len(runes) {
		sb.WriteString(theme.Styles.CursorBlock.Render(" "))
	}

	rendered := sb.String()
	// Truncate if too wide (rough, doesn't account for ANSI codes perfectly)
	_ = maxWidth
	return rendered
}

// RenderStatusBar renders the bottom status bar.
func RenderStatusBar(buf *engine.Buffer, pendingKeys string, width int, extra string) string {
	mode := buf.Mode()
	cursor := buf.GetCursor()

	// Mode pill
	modeStr := ""
	switch mode {
	case engine.ModeNormal:
		modeStr = theme.Styles.StatusNormal.Render("NORMAL")
	case engine.ModeInsert:
		modeStr = theme.Styles.StatusInsert.Render("INSERT")
	case engine.ModeVisual:
		modeStr = theme.Styles.StatusVisual.Render("VISUAL")
	}

	// Position
	pos := theme.Styles.StatusInfo.Render(fmt.Sprintf(" %d,%d ", cursor.Line+1, cursor.Col+1))

	// Pending keys
	pending := ""
	if pendingKeys != "" {
		pending = theme.Styles.StatusPending.Render(" " + pendingKeys + "▌")
	}

	// Right side info
	right := ""
	if extra != "" {
		right = theme.Styles.Subtle.Render("  " + extra)
	}

	left := modeStr + pos + pending
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}

	bar := left + strings.Repeat(" ", gap) + right
	return theme.Styles.StatusBar.Width(width).Render(bar)
}
