package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	Location string `json:"location"` // e.g., "America/New_York"
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

// --- TUI ---

type viewState int

const (
	viewList viewState = iota
	viewDetail
	viewAdd
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
	inputStep  int // 0 for label, 1 for location
	newEntry   ClockEntry
	help       help.Model
	keys       keyMap
	err        error
}

type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Enter  key.Binding
	Add    key.Binding
	Delete key.Binding
	Back   key.Binding
	Quit   key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Add, k.Delete, k.Back, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter},
		{k.Add, k.Delete, k.Back, k.Quit},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Add: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add clock"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc", "backspace"),
		key.WithHelp("esc", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
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

func (m model) Init() tea.Cmd {
	return tick()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		return m, tick()

	case tea.KeyMsg:
		if m.state == viewAdd {
			switch {
			case key.Matches(msg, m.keys.Back):
				m.state = viewList
				m.inputStep = 0
				m.textInput.Reset()
				return m, nil
			case key.Matches(msg, m.keys.Enter):
				if m.inputStep == 0 {
					m.newEntry.Label = m.textInput.Value()
					m.inputStep = 1
					m.textInput.Reset()
					m.textInput.Placeholder = "Location (e.g. America/New_York)"
					return m, nil
				} else {
					m.newEntry.Location = m.textInput.Value()
					if m.newEntry.Location != "Local" && m.newEntry.Location != "UTC" {
						_, err := time.LoadLocation(m.newEntry.Location)
						if err != nil {
							m.err = fmt.Errorf("Invalid location: %s", m.newEntry.Location)
							return m, nil
						}
					}
					m.clocks = append(m.clocks, m.newEntry)
					saveConfig(ClockConfig{Clocks: m.clocks})
					m.state = viewList
					m.inputStep = 0
					m.textInput.Reset()
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
			if m.state == viewList && m.cursor > 0 {
				m.cursor--
			}

		case key.Matches(msg, m.keys.Down):
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
				m.state = viewAdd
				m.inputStep = 0
				m.textInput.Placeholder = "Label (e.g. New York)"
				m.textInput.Focus()
			}

		case key.Matches(msg, m.keys.Delete):
			if m.state == viewList && len(m.clocks) > 0 {
				m.clocks = append(m.clocks[:m.cursor], m.clocks[m.cursor+1:]...)
				if m.cursor >= len(m.clocks) && m.cursor > 0 {
					m.cursor--
				}
				saveConfig(ClockConfig{Clocks: m.clocks})
			}
		}
	}

	return m, nil
}

// --- Styling ---

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#D4AF37")).
			Padding(0, 1).
			MarginBottom(1)

	clockBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#555555")).
			Padding(0, 2).
			MarginRight(2)

	selectedClockStyle = clockBoxStyle.Copy().
				BorderForeground(lipgloss.Color("#D4AF37"))

	timeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF"))

	dateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA"))

	labelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#D4AF37"))

	detailTimeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#D4AF37")).
			Padding(1, 3).
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("#D4AF37"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			MarginTop(1)
)

func (m model) View() string {
	s := titleStyle.Render("ATLAS CLOCK") + "\n\n"

	switch m.state {
	case viewList:
		s += m.listView()
	case viewDetail:
		s += m.detailView()
	case viewAdd:
		s += m.addView()
	}

	helpView := m.help.View(m.keys)
	s += lipgloss.NewStyle().MarginTop(2).Render(helpView)

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

		timeStr := t.Format("15:04:05")
		dateStr := t.Format("Mon, Jan 02")

		content := lipgloss.JoinVertical(
			lipgloss.Left,
			labelStyle.Render(entry.Label),
			timeStyle.Render(timeStr),
			dateStyle.Render(dateStr),
		)

		style := clockBoxStyle
		if i == m.cursor {
			style = selectedClockStyle
		}

		currentRow = append(currentRow, style.Render(content))

		if len(currentRow) == 3 {
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, currentRow...))
			currentRow = nil
		}
	}

	if len(currentRow) > 0 {
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, currentRow...))
	}

	if len(rows) == 0 {
		return "No clocks added. Press 'a' to add one."
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m model) detailView() string {
	if m.selected >= len(m.clocks) {
		return "Error: Clock not found."
	}
	entry := m.clocks[m.selected]
	t := time.Now()
	if entry.Location != "Local" && entry.Location != "" {
		loc, err := time.LoadLocation(entry.Location)
		if err == nil {
			t = t.In(loc)
		}
	}

	timeStr := t.Format("15:04:05.000")
	dateStr := t.Format("Monday, January 02, 2006")
	tzName, offset := t.Zone()

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		labelStyle.Copy().Render(entry.Label),
		"\n",
		detailTimeStyle.Render(timeStr),
		"\n",
		dateStyle.Render(dateStr),
		dateStyle.Render(fmt.Sprintf("%s (UTC%s)", tzName, formatOffset(offset))),
	)

	return content
}

func (m model) addView() string {
	s := "Add New Clock\n\n"
	if m.inputStep == 0 {
		s += "Step 1: Enter Label\n"
	} else {
		s += "Step 2: Enter Location (e.g., UTC, Local, Europe/Istanbul)\n"
	}
	s += "\n" + m.textInput.View()
	if m.err != nil {
		s += errorStyle.Render(m.err.Error())
	}
	return s
}

func formatOffset(offset int) string {
	hours := offset / 3600
	mins := (offset % 3600) / 60
	if offset >= 0 {
		return fmt.Sprintf("+%02d:%02d", hours, mins)
	}
	return fmt.Sprintf("-%02d:%02d", -hours, -mins)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
