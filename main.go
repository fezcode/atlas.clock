package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- Config ---

type ClockConfig struct {
	Clocks []ClockEntry `json:"clocks"`
}

type ClockEntry struct {
	Label    string `json:"label"`
	Location string `json:"location"`
}

func getConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".atlas", "clock.json")
}

func loadConfig() ClockConfig {
	path := getConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return ClockConfig{
			Clocks: []ClockEntry{
				{Label: "Local", Location: "Local"},
				{Label: "UTC", Location: "UTC"},
				{Label: "Istanbul", Location: "Europe/Istanbul"},
			},
		}
	}
	var config ClockConfig
	json.Unmarshal(data, &config)
	return config
}

func saveConfig(config ClockConfig) {
	path := getConfigPath()
	os.MkdirAll(filepath.Dir(path), 0755)
	data, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(path, data, 0644)
}

// --- Boxy Big Font Renderer ---

var bigDigits = map[rune][]string{
	'0': {" █████ ", " █   █ ", " █   █ ", " █   █ ", " █████ "},
	'1': {"   ██  ", "    █  ", "    █  ", "    █  ", "  █████"},
	'2': {" █████ ", "     █ ", " █████ ", " █     ", " █████ "},
	'3': {" █████ ", "     █ ", "  ████ ", "     █ ", " █████ "},
	'4': {" █   █ ", " █   █ ", " █████ ", "     █ ", "     █ "},
	'5': {" █████ ", " █     ", " █████ ", "     █ ", " █████ "},
	'6': {" █████ ", " █     ", " █████ ", " █   █ ", " █████ "},
	'7': {" █████ ", "     █ ", "    █  ", "   █   ", "   █   "},
	'8': {" █████ ", " █   █ ", " █████ ", " █   █ ", " █████ "},
	'9': {" █████ ", " █   █ ", " █████ ", "     █ ", " █████ "},
	':': {"       ", "   █   ", "       ", "   █   ", "       "},
	'.': {"       ", "       ", "       ", "       ", "   █   "},
	' ': {"       ", "       ", "       ", "       ", "       "},
}

func renderBigText(input string) string {
	lines := make([]string, 5)
	for _, r := range input {
		digit, ok := bigDigits[r]
		if !ok {
			digit = []string{"       ", "       ", "       ", "       ", "       "}
		}
		for i := 0; i < 5; i++ {
			lines[i] += digit[i]
		}
	}
	return strings.Join(lines, "\n")
}

// --- TUI ---

type viewState int

const (
	viewList viewState = iota
	viewDetail
	viewAdd
	viewDeleteConfirm
)

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Every(time.Millisecond*10, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type model struct {
	state      viewState
	clocks     []ClockEntry
	cursor     int
	selected   int
	width      int
	height     int
	textInput  textinput.Model
	inputStep  int // 0: label, 1: location, 2: confirm
	newEntry   ClockEntry
	help       help.Model
	keys       keyMap
	err        error
}

type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Enter  key.Binding
	Add    key.Binding
	Delete key.Binding
	Back   key.Binding
	Quit   key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Left, k.Right, k.Enter, k.Add, k.Delete, k.Back, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right, k.Enter},
		{k.Add, k.Delete, k.Back, k.Quit},
	}
}

var keys = keyMap{
	Up:     key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:   key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Left:   key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "left")),
	Right:  key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "right")),
	Enter:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	Add:    key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add")),
	Delete: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "del")),
	Back:   key.NewBinding(key.WithKeys("esc", "backspace"), key.WithHelp("esc", "back")),
	Quit:   key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Label (e.g. New York)"
	ti.Focus()
	config := loadConfig()
	return model{
		state:     viewList,
		clocks:    config.Clocks,
		textInput: ti,
		help:      help.New(),
		keys:      keys,
	}
}

func (m model) Init() tea.Cmd { return tick() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tickMsg:
		return m, tick()
	case tea.KeyMsg:
		// Handle Delete Confirmation
		if m.state == viewDeleteConfirm {
			switch msg.String() {
			case "y", "Y", "enter":
				m.clocks = append(m.clocks[:m.cursor], m.clocks[m.cursor+1:]...)
				if m.cursor >= len(m.clocks) && m.cursor > 0 {
					m.cursor--
				}
				saveConfig(ClockConfig{Clocks: m.clocks})
				m.state = viewList
			case "n", "N", "esc":
				m.state = viewList
			}
			return m, nil
		}

		// Handle Add View
		if m.state == viewAdd {
			if m.inputStep == 2 { // Confirmation step
				switch msg.String() {
				case "y", "Y", "enter":
					m.clocks = append(m.clocks, m.newEntry)
					saveConfig(ClockConfig{Clocks: m.clocks})
					m.state, m.inputStep = viewList, 0
					m.textInput.Reset()
					m.err = nil
				case "n", "N", "esc":
					m.state, m.inputStep = viewList, 0
					m.textInput.Reset()
				}
				return m, nil
			}

			switch {
			case key.Matches(msg, m.keys.Back):
				m.state = viewList
				m.textInput.Reset()
				return m, nil
			case key.Matches(msg, m.keys.Enter):
				if m.inputStep == 0 {
					val := strings.TrimSpace(m.textInput.Value())
					if val == "" { return m, nil }
					m.newEntry.Label = val
					m.inputStep = 1
					m.textInput.Reset()
					m.textInput.Placeholder = "Location (e.g. Europe/Istanbul)"
					return m, nil
				} else if m.inputStep == 1 {
					locStr := strings.TrimSpace(m.textInput.Value())
					if locStr == "" { return m, nil }
					m.newEntry.Location = locStr
					if m.newEntry.Location != "Local" && m.newEntry.Location != "UTC" {
						_, err := time.LoadLocation(m.newEntry.Location)
						if err != nil {
							m.err = fmt.Errorf("Invalid timezone location")
							return m, nil
						}
					}
					m.inputStep = 2 // Move to confirmation
					m.textInput.Blur()
					m.err = nil
					return m, nil
				}
			}
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Back):
			if m.state == viewDetail {
				m.state = viewList
			}
		case key.Matches(msg, m.keys.Up):
			if m.state == viewList && m.cursor >= 3 {
				m.cursor -= 3
			}
		case key.Matches(msg, m.keys.Down):
			if m.state == viewList && m.cursor+3 < len(m.clocks) {
				m.cursor += 3
			}
		case key.Matches(msg, m.keys.Left):
			if m.state == viewList && m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, m.keys.Right):
			if m.state == viewList && m.cursor < len(m.clocks)-1 {
				m.cursor++
			}
		case key.Matches(msg, m.keys.Enter):
			if m.state == viewList && len(m.clocks) > 0 {
				m.selected = m.cursor
				m.state = viewDetail
			}
		case key.Matches(msg, m.keys.Add):
			if m.state == viewList {
				m.state, m.inputStep = viewAdd, 0
				m.textInput.Reset()
				m.textInput.Placeholder = "Label (e.g. New York)"
				m.textInput.Focus()
			}
		case key.Matches(msg, m.keys.Delete):
			if m.state == viewList && len(m.clocks) > 0 {
				m.state = viewDeleteConfirm
			}
		}
	}
	return m, nil
}

// --- Styling ---

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#D4AF37")).MarginBottom(1)
	
	clockBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#555555")).
			Padding(1, 4).
			Margin(0, 1)

	selectedClockStyle = clockBoxStyle.Copy().BorderForeground(lipgloss.Color("#D4AF37"))
	
	timeStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	dateStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
	labelStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#D4AF37"))
	
	bigTimeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#D4AF37"))
	
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).MarginTop(1)

	confirmStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("#D4AF37")).
			Padding(1, 4).
			MarginTop(1)
)

func (m model) View() string {
	var content string
	switch m.state {
	case viewList:
		content = m.listView()
	case viewDetail:
		content = m.detailView()
	case viewAdd:
		content = m.addView()
	case viewDeleteConfirm:
		content = m.deleteConfirmView()
	}

	s := lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render("ATLAS CLOCK"),
		content,
		lipgloss.NewStyle().MarginTop(2).Render(m.help.View(m.keys)),
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, s)
}

func (m model) listView() string {
	var rows []string
	var currentRow []string
	for i, entry := range m.clocks {
		t := time.Now()
		if entry.Location != "Local" && entry.Location != "" {
			loc, err := time.LoadLocation(entry.Location)
			if err == nil {
				t = t.In(loc)
			}
		}
		c := lipgloss.JoinVertical(lipgloss.Left,
			labelStyle.Render(entry.Label),
			timeStyle.Render(t.Format("15:04:05")),
			dateStyle.Render(t.Format("Mon, Jan 02")),
		)
		style := clockBoxStyle
		if i == m.cursor {
			style = selectedClockStyle
		}
		currentRow = append(currentRow, style.Render(c))
		if len(currentRow) == 3 {
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, currentRow...))
			currentRow = nil
		}
	}
	if len(currentRow) > 0 {
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, currentRow...))
	}
	if len(rows) == 0 {
		return "No clocks. Press 'a' to add."
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m model) detailView() string {
	if m.selected >= len(m.clocks) { return "Error" }
	entry := m.clocks[m.selected]
	t := time.Now()
	if entry.Location != "Local" && entry.Location != "" {
		loc, err := time.LoadLocation(entry.Location)
		if err == nil { t = t.In(loc) }
	}

	bigTime := renderBigText(t.Format("15:04:05.00"))
	tzName, offset := t.Zone()

	return lipgloss.JoinVertical(lipgloss.Center,
		labelStyle.Render(strings.ToUpper(entry.Label)),
		"",
		bigTimeStyle.Render(bigTime),
		"",
		dateStyle.Render(t.Format("Monday, January 02, 2006")),
		dateStyle.Render(fmt.Sprintf("%s (UTC%s)", tzName, formatOffset(offset))),
	)
}

func (m model) addView() string {
	var s string
	if m.inputStep == 0 {
		s = "STEP 1: ENTER LABEL\n\n" + m.textInput.View()
	} else if m.inputStep == 1 {
		s = fmt.Sprintf("LABEL: %s\n\nSTEP 2: ENTER LOCATION\n(e.g. UTC, Local, America/New_York)\n\n%s", 
			labelStyle.Render(m.newEntry.Label), m.textInput.View())
	} else {
		s = confirmStyle.Render(fmt.Sprintf("CONFIRM ADDING CLOCK?\n\nLabel: %s\nLocation: %s\n\n(y)es / (n)o", 
			labelStyle.Render(m.newEntry.Label), timeStyle.Render(m.newEntry.Location)))
	}
	
	if m.err != nil { s += errorStyle.Render("\n" + m.err.Error()) }
	return s
}

func (m model) deleteConfirmView() string {
	entry := m.clocks[m.cursor]
	return confirmStyle.Render(fmt.Sprintf("ARE YOU SURE YOU WANT TO DELETE?\n\n%s (%s)\n\n(y)es / (n)o", 
		labelStyle.Render(entry.Label), dateStyle.Render(entry.Location)))
}

func formatOffset(offset int) string {
	h, m := offset/3600, (offset%3600)/60
	if offset >= 0 { return fmt.Sprintf("+%02d:%02d", h, m) }
	return fmt.Sprintf("-%02d:%02d", -h, -m)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
