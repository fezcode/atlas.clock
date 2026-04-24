package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/fezcode/atlas.clock/pkg/store"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- states -----------------------------------------------------------------

type viewState int

const (
	viewDashboard viewState = iota
	viewDetail
	viewLabelInput
	viewZonePicker
	viewConfirmAdd
	viewConfirmDelete
)

// --- messages ---------------------------------------------------------------

type tickMsg time.Time

// --- zone item --------------------------------------------------------------

type zoneItem string

func (z zoneItem) Title() string       { return string(z) }
func (z zoneItem) Description() string { return "" }
func (z zoneItem) FilterValue() string { return string(z) }

// --- model ------------------------------------------------------------------

type model struct {
	version string
	state   viewState

	clocks []store.Entry
	cursor int

	textInput textinput.Model
	zoneList  list.Model
	newEntry  store.Entry

	width, height int
	blink         bool
	frame         int
	started       time.Time
}

// Config bundles launch parameters.
type Config struct {
	Version string
}

func newModel(cfg Config) model {
	ti := textinput.New()
	ti.Placeholder = "label (e.g. Tokyo Office)"
	ti.CharLimit = 64
	ti.Prompt = ""
	ti.TextStyle = sPaper
	ti.PlaceholderStyle = sDim

	items := make([]list.Item, len(store.IANAZones))
	for i, tz := range store.IANAZones {
		items[i] = zoneItem(tz)
	}
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(ColAmber).
		Bold(true).
		BorderForeground(ColAmber)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(ColAmber).
		BorderForeground(ColAmber)
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.Foreground(ColText)
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.Foreground(ColDim)
	delegate.Styles.DimmedTitle = delegate.Styles.DimmedTitle.Foreground(ColDim)
	delegate.Styles.DimmedDesc = delegate.Styles.DimmedDesc.Foreground(ColDim)

	zl := list.New(items, delegate, 0, 0)
	zl.Title = "SELECT TIMEZONE — type to filter"
	zl.SetShowStatusBar(false)
	zl.SetFilteringEnabled(true)
	zl.Styles.Title = lipgloss.NewStyle().Foreground(ColPaper).Bold(true)
	zl.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(ColAmber).Bold(true)
	zl.FilterInput.TextStyle = sPaper

	cfgData := store.Load()
	return model{
		version: cfg.Version,
		state:   viewDashboard,
		clocks:  cfgData.Clocks,
		textInput: ti,
		zoneList: zl,
		started:  time.Now(),
	}
}

// --- tea.Model --------------------------------------------------------------

func (m model) Init() tea.Cmd { return tick() }

func tick() tea.Cmd {
	// Tick at 20 Hz so milliseconds in detail view feel live without wasting CPU.
	return tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		listW := m.width - 8
		listH := m.height - 10
		if listW < 20 {
			listW = 20
		}
		if listH < 6 {
			listH = 6
		}
		m.zoneList.SetSize(listW, listH)
		return m, tea.ClearScreen

	case tickMsg:
		m.frame++
		m.blink = m.frame%10 == 0
		return m, tick()

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case viewLabelInput:
		return m.keyLabelInput(msg)
	case viewZonePicker:
		return m.keyZonePicker(msg)
	case viewConfirmAdd:
		return m.keyConfirmAdd(msg)
	case viewConfirmDelete:
		return m.keyConfirmDelete(msg)
	case viewDetail:
		return m.keyDetail(msg)
	default:
		return m.keyDashboard(msg)
	}
}

func (m model) keyDashboard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	gridCols := m.gridCols()
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.cursor >= gridCols {
			m.cursor -= gridCols
		}
	case "down", "j":
		if m.cursor+gridCols < len(m.clocks) {
			m.cursor += gridCols
		}
	case "left", "h":
		if m.cursor > 0 {
			m.cursor--
		}
	case "right", "l":
		if m.cursor < len(m.clocks)-1 {
			m.cursor++
		}
	case "enter":
		if len(m.clocks) > 0 {
			m.state = viewDetail
		}
	case "a":
		m.state = viewLabelInput
		m.textInput.Reset()
		m.textInput.Focus()
		return m, textinput.Blink
	case "d":
		if len(m.clocks) > 0 {
			m.state = viewConfirmDelete
		}
	case "K", "shift+up":
		if m.cursor >= gridCols {
			m.clocks[m.cursor], m.clocks[m.cursor-gridCols] =
				m.clocks[m.cursor-gridCols], m.clocks[m.cursor]
			m.cursor -= gridCols
			_ = store.Save(store.Config{Clocks: m.clocks})
		}
	case "J", "shift+down":
		if m.cursor+gridCols < len(m.clocks) {
			m.clocks[m.cursor], m.clocks[m.cursor+gridCols] =
				m.clocks[m.cursor+gridCols], m.clocks[m.cursor]
			m.cursor += gridCols
			_ = store.Save(store.Config{Clocks: m.clocks})
		}
	case "H", "shift+left":
		if m.cursor > 0 {
			m.clocks[m.cursor], m.clocks[m.cursor-1] =
				m.clocks[m.cursor-1], m.clocks[m.cursor]
			m.cursor--
			_ = store.Save(store.Config{Clocks: m.clocks})
		}
	case "L", "shift+right":
		if m.cursor < len(m.clocks)-1 {
			m.clocks[m.cursor], m.clocks[m.cursor+1] =
				m.clocks[m.cursor+1], m.clocks[m.cursor]
			m.cursor++
			_ = store.Save(store.Config{Clocks: m.clocks})
		}
	}
	return m, nil
}

func (m model) keyDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc", "enter":
		m.state = viewDashboard
	case "left", "h":
		if m.cursor > 0 {
			m.cursor--
		}
	case "right", "l":
		if m.cursor < len(m.clocks)-1 {
			m.cursor++
		}
	}
	return m, nil
}

func (m model) keyLabelInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.state = viewDashboard
		return m, nil
	case "enter":
		val := strings.TrimSpace(m.textInput.Value())
		if val == "" {
			return m, nil
		}
		m.newEntry.Label = val
		m.state = viewZonePicker
		return m, nil
	}
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) keyZonePicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		// If filter is active, let the list handle esc to clear it.
		if m.zoneList.FilterState() != list.Unfiltered {
			break
		}
		m.state = viewLabelInput
		return m, nil
	case "enter":
		if m.zoneList.FilterState() == list.Filtering {
			// Let the list commit the filter first.
			break
		}
		if sel, ok := m.zoneList.SelectedItem().(zoneItem); ok {
			m.newEntry.Location = string(sel)
			m.state = viewConfirmAdd
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.zoneList, cmd = m.zoneList.Update(msg)
	return m, cmd
}

func (m model) keyConfirmAdd(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
		m.clocks = append(m.clocks, m.newEntry)
		_ = store.Save(store.Config{Clocks: m.clocks})
		m.state = viewDashboard
		m.textInput.Reset()
	case "n", "N", "esc", "ctrl+c":
		m.state = viewDashboard
		m.textInput.Reset()
	}
	return m, nil
}

func (m model) keyConfirmDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
		if m.cursor >= 0 && m.cursor < len(m.clocks) {
			m.clocks = append(m.clocks[:m.cursor], m.clocks[m.cursor+1:]...)
			if m.cursor >= len(m.clocks) && m.cursor > 0 {
				m.cursor--
			}
			_ = store.Save(store.Config{Clocks: m.clocks})
		}
		m.state = viewDashboard
	case "n", "N", "esc", "ctrl+c":
		m.state = viewDashboard
	}
	return m, nil
}

// --- View -------------------------------------------------------------------

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	if m.width < 64 {
		return sCrit.Render(" terminal too narrow — resize to ≥ 64 columns ")
	}

	var body string
	switch m.state {
	case viewDashboard:
		body = m.renderDashboard()
	case viewDetail:
		body = m.renderDetail()
	case viewLabelInput:
		body = m.renderLabelInput()
	case viewZonePicker:
		body = m.renderZonePicker()
	case viewConfirmAdd:
		body = m.renderConfirmAdd()
	case viewConfirmDelete:
		body = m.renderConfirmDelete()
	}

	full := m.renderMasthead() + "\n" + body + "\n" + m.renderFooter()

	lines := strings.Split(full, "\n")
	if len(lines) < m.height {
		blank := strings.Repeat(" ", m.width)
		for len(lines) < m.height {
			lines = append(lines, blank)
		}
	} else if len(lines) > m.height {
		lines = lines[:m.height]
	}
	return strings.Join(lines, "\n")
}

// --- Masthead ---------------------------------------------------------------

func (m model) renderMasthead() string {
	w := m.width
	rule := sBorder.Render(strings.Repeat("━", w))

	title := sMastTitle.Render("A T L A S") +
		sDim.Render("  ·  ") +
		sMastTitle.Render("C L O C K")

	local := sMastClock.Render(time.Now().Format("15:04:05"))
	var rec string
	if m.blink {
		rec = sRec.Render("● SYNC")
	} else {
		rec = sDim.Render("● SYNC")
	}
	ver := sDim.Render("v" + m.version)
	right := horiz(local, rec, ver)

	titleW := lipgloss.Width(title)
	rightW := lipgloss.Width(right)
	pad := w - 2 - titleW - rightW
	if pad < 1 {
		pad = 1
	}
	line1 := "  " + title + strings.Repeat(" ", pad) + right

	_, off := time.Now().Zone()
	zoneName, _ := time.Now().Zone()
	meta := horiz(
		sDim.Render("CLOCKS ")+sValue.Render(fmt.Sprintf("%d", len(m.clocks))),
		sDim.Render("LOCAL ")+sValue.Render(zoneName+" "+formatOffset(off)),
		sDim.Render("DATE ")+sValue.Render(time.Now().Format("Mon 02 Jan 2006")),
	)
	line2 := "  " + meta
	if lipgloss.Width(line2) > w {
		line2 = "  " + horiz(
			sDim.Render("CLOCKS ")+sValue.Render(fmt.Sprintf("%d", len(m.clocks))),
			sDim.Render("LOCAL ")+sValue.Render(zoneName),
		)
	}

	return strings.Join([]string{rule, line1, line2, rule}, "\n")
}

// --- Dashboard (§01 grid of clock cards) -----------------------------------

func (m model) gridCols() int {
	cardW := 26
	gap := 2
	cols := (m.width - 4 + gap) / (cardW + gap)
	if cols < 1 {
		cols = 1
	}
	if cols > 5 {
		cols = 5
	}
	return cols
}

func (m model) renderDashboard() string {
	if len(m.clocks) == 0 {
		return section("01", "DASHBOARD", sDim.Render("no clocks — press A to add your first"), m.width)
	}

	cardW := 26
	gap := 2
	cols := m.gridCols()

	// Snap cardW up so cols * (cardW+gap) - gap ≤ width-4.
	available := m.width - 4
	if cols > 0 {
		cardW = (available - gap*(cols-1)) / cols
		if cardW < 20 {
			cardW = 20
		}
	}

	var rows []string
	i := 0
	for i < len(m.clocks) {
		end := i + cols
		if end > len(m.clocks) {
			end = len(m.clocks)
		}
		var cards []string
		innerW := cardW - 4
		for j := i; j < end; j++ {
			entry := m.clocks[j]
			t := entry.Now()
			dnGlyph, dnStyle := daynightStyle(t.Hour())

			_, off := t.Zone()

			// Compose the title from raw pieces so we can size the label
			// against the card's inner width without tripping over the glyph's
			// ANSI escapes.
			const glyphSlot = 3 // "X  "
			labelBudget := innerW - glyphSlot
			if labelBudget < 3 {
				labelBudget = 3
			}
			labelStyle := sPaper
			if j == m.cursor {
				labelStyle = sAmber
			}
			label := labelStyle.Render(truncateVisible(entry.Label, labelBudget))
			title := dnStyle.Render(dnGlyph) + "  " + label

			timeStr := t.Format("15:04:05")

			// Meta: zone + offset, sized to the inner width.
			offStr := formatOffset(off)
			zoneBudget := innerW - lipgloss.Width(offStr) - 2
			if zoneBudget < 3 {
				zoneBudget = 3
			}
			meta := truncateVisible(entry.Location, zoneBudget) + "  " + offStr

			cards = append(cards, card(cardW, j == m.cursor, title, timeStr, meta))
		}
		rows = append(rows, joinH(gap, cards...))
		i = end
	}

	return section("01", "DASHBOARD", strings.Join(rows, "\n"), m.width)
}

// --- Detail view ------------------------------------------------------------

func (m model) renderDetail() string {
	if m.cursor < 0 || m.cursor >= len(m.clocks) {
		return section("01", "DETAIL", sCrit.Render("invalid selection"), m.width)
	}
	entry := m.clocks[m.cursor]
	t := entry.Now()

	// Compose: label · zone · offset · day/night on one line; big time; big ms; date.
	zoneName, off := t.Zone()
	dnGlyph, dnStyle := daynightStyle(t.Hour())

	header := horiz(
		sAmber.Render(strings.ToUpper(entry.Label)),
		sValue.Render(entry.Location),
		sValue.Render(zoneName+" "+formatOffset(off)),
		dnStyle.Render(dnGlyph),
	)

	timeStr := t.Format("15:04:05")
	big := renderBigText(timeStr)
	ms := sMs.Render(fmt.Sprintf(".%03d", t.Nanosecond()/int(time.Millisecond)))

	lines := []string{
		header,
		"",
		big,
		"",
		sPaper.Render(t.Format("Monday, 02 January 2006")) + "   " + ms,
	}
	body := strings.Join(lines, "\n")
	return section(fmt.Sprintf("%02d", m.cursor+1), "DETAIL", body, m.width)
}

// --- Add flow ---------------------------------------------------------------

func (m model) renderLabelInput() string {
	inner := m.width - 4
	m.textInput.Width = inner - 4
	body := sPromptMark.Render("❯ ") + m.textInput.View() + "\n\n" +
		sDim.Render("Type a label for the clock, then press ↵ to pick a timezone. ") +
		sDim.Render("Esc to cancel.")
	return section("01", "ADD CLOCK · LABEL", body, m.width)
}

func (m model) renderZonePicker() string {
	return section("01", "ADD CLOCK · TIMEZONE", m.zoneList.View(), m.width)
}

func (m model) renderConfirmAdd() string {
	body := strings.Join([]string{
		sPaper.Render("Add this clock?"),
		"",
		labelValue("LABEL", sValue.Render(m.newEntry.Label), 12),
		labelValue("ZONE", sValue.Render(m.newEntry.Location), 12),
		"",
		sFooterKey.Render("[Y]") + sFooterText.Render(" confirm   ") +
			sFooterKey.Render("[N]") + sFooterText.Render(" cancel"),
	}, "\n")
	return section("01", "ADD CLOCK · CONFIRM", body, m.width)
}

func (m model) renderConfirmDelete() string {
	if m.cursor < 0 || m.cursor >= len(m.clocks) {
		return section("01", "DELETE", sCrit.Render("invalid selection"), m.width)
	}
	entry := m.clocks[m.cursor]
	body := strings.Join([]string{
		sCrit.Render("Delete this clock?"),
		"",
		labelValue("LABEL", sValue.Render(entry.Label), 12),
		labelValue("ZONE", sValue.Render(entry.Location), 12),
		"",
		sFooterKey.Render("[Y]") + sFooterText.Render(" delete   ") +
			sFooterKey.Render("[N]") + sFooterText.Render(" cancel"),
	}, "\n")
	return section("01", "DELETE CLOCK", body, m.width)
}

// --- Footer -----------------------------------------------------------------

func (m model) renderFooter() string {
	var keys []string
	switch m.state {
	case viewLabelInput:
		keys = []string{
			sFooterKey.Render("[↵]") + sFooterText.Render("·NEXT"),
			sFooterKey.Render("[ESC]") + sFooterText.Render("·CANCEL"),
		}
	case viewZonePicker:
		keys = []string{
			sFooterKey.Render("[/]") + sFooterText.Render("·FILTER"),
			sFooterKey.Render("[↵]") + sFooterText.Render("·PICK"),
			sFooterKey.Render("[ESC]") + sFooterText.Render("·BACK"),
		}
	case viewConfirmAdd, viewConfirmDelete:
		keys = []string{
			sFooterKey.Render("[Y/N]") + sFooterText.Render("·CONFIRM"),
		}
	case viewDetail:
		keys = []string{
			sFooterKey.Render("[← →]") + sFooterText.Render("·SWITCH"),
			sFooterKey.Render("[ESC]") + sFooterText.Render("·BACK"),
			sFooterKey.Render("[Q]") + sFooterText.Render("·QUIT"),
		}
	default:
		keys = []string{
			sFooterKey.Render("[↑↓← →]") + sFooterText.Render("·NAV"),
			sFooterKey.Render("[SHIFT+ARR]") + sFooterText.Render("·REORDER"),
			sFooterKey.Render("[↵]") + sFooterText.Render("·DETAIL"),
			sFooterKey.Render("[A]") + sFooterText.Render("·ADD"),
			sFooterKey.Render("[D]") + sFooterText.Render("·DEL"),
			sFooterKey.Render("[Q]") + sFooterText.Render("·QUIT"),
		}
	}
	left := " " + strings.Join(keys, "   ")
	right := sDim.Render(fmt.Sprintf(" uptime · %s ", time.Since(m.started).Truncate(time.Second)))
	pad := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if pad < 1 {
		pad = 1
	}
	return left + strings.Repeat(" ", pad) + right
}

func formatOffset(off int) string {
	sign := "+"
	if off < 0 {
		sign = "-"
		off = -off
	}
	h := off / 3600
	min := (off % 3600) / 60
	return fmt.Sprintf("UTC%s%02d:%02d", sign, h, min)
}

// --- entry point ------------------------------------------------------------

// Start launches the TUI.
func Start(cfg Config) error {
	p := tea.NewProgram(newModel(cfg), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
