package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	sepGlyph     = "·"
	sectionOpen  = "┤"
	sectionClose = "├"
)

// section renders a rounded box with the title inlined into the top border.
func section(num, title, body string, width int) string {
	if width < 20 {
		width = 20
	}
	titleInner := sSectionKey.Render(fmt.Sprintf(" §%s  ", num)) +
		sSectionTitle.Render(title) +
		" "
	label := sBorder.Render(sectionOpen) + titleInner + sBorder.Render(sectionClose)
	labelW := lipgloss.Width(label)

	lead := 2
	fill := width - 2 - lead - labelW
	if fill < 1 {
		lead = 1
		fill = width - 2 - lead - labelW
		if fill < 1 {
			fill = 1
		}
	}
	top := sBorder.Render("╭"+strings.Repeat("─", lead)) +
		label +
		sBorder.Render(strings.Repeat("─", fill)+"╮")

	inner := width - 4
	var rows []string
	for _, ln := range strings.Split(body, "\n") {
		w := lipgloss.Width(ln)
		if w < inner {
			ln = ln + strings.Repeat(" ", inner-w)
		} else if w > inner {
			ln = truncateVisible(ln, inner)
		}
		rows = append(rows,
			sBorder.Render("│")+" "+ln+" "+sBorder.Render("│"))
	}
	bottom := sBorder.Render("╰" + strings.Repeat("─", width-2) + "╯")

	return top + "\n" + strings.Join(rows, "\n") + "\n" + bottom
}

// card is a fixed-width mini-box used for the grid layout on the dashboard.
// `title` may contain ANSI styling; callers are responsible for sizing it to
// fit `width-4` visible cells (we pad only).
func card(width int, selected bool, title, timeStr, meta string) string {
	if width < 18 {
		width = 18
	}
	inner := width - 4

	borderStyle := sBorder
	if selected {
		borderStyle = sAmber
	}

	top := borderStyle.Render("╭" + strings.Repeat("─", width-2) + "╮")
	bot := borderStyle.Render("╰" + strings.Repeat("─", width-2) + "╯")

	titleLn := padLeft(title, inner)
	timeLn := padLeft(sBigDigit.Render(timeStr), inner)
	metaLn := padLeft(sDim.Render(meta), inner)

	row := func(content string) string {
		return borderStyle.Render("│") + " " + content + " " + borderStyle.Render("│")
	}
	return strings.Join([]string{
		top,
		row(titleLn),
		row(timeLn),
		row(metaLn),
		bot,
	}, "\n")
}

func truncateVisible(s string, n int) string {
	if n <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

func labelValue(label, value string, labelCol int) string {
	lbl := sLabel.Render(label)
	pad := labelCol - lipgloss.Width(lbl)
	if pad < 1 {
		pad = 1
	}
	return lbl + strings.Repeat(" ", pad) + value
}

func horiz(parts ...string) string {
	return strings.Join(parts, "  "+sBorder.Render(sepGlyph)+"  ")
}

func padLeft(s string, n int) string {
	w := lipgloss.Width(s)
	if w >= n {
		return s
	}
	return s + strings.Repeat(" ", n-w)
}

func padRight(s string, n int) string {
	w := lipgloss.Width(s)
	if w >= n {
		return s
	}
	return strings.Repeat(" ", n-w) + s
}

func nonempty(s, fb string) string {
	if strings.TrimSpace(s) == "" {
		return fb
	}
	return s
}

// joinH joins rendered blocks horizontally with a gap.
func joinH(gutter int, parts ...string) string {
	if gutter <= 0 {
		return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
	}
	gutterStr := strings.Repeat(" ", gutter)
	spaced := make([]string, 0, 2*len(parts)-1)
	for i, p := range parts {
		if i > 0 {
			spaced = append(spaced, gutterStr)
		}
		spaced = append(spaced, p)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, spaced...)
}
