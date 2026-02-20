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
	"github.com/charmbracelet/bubbles/list"
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

// --- Timezone Data ---

type zoneItem string

func (z zoneItem) Title() string       { return string(z) }
func (z zoneItem) Description() string { return "" }
func (z zoneItem) FilterValue() string { return string(z) }

var ianaTimezones = []string{
	"Local", "UTC", "Africa/Abidjan", "Africa/Accra", "Africa/Addis_Ababa", "Africa/Algiers", "Africa/Asmara", "Africa/Bamako", "Africa/Bangui", "Africa/Banjul", "Africa/Bissau", "Africa/Blantyre", "Africa/Brazzaville", "Africa/Bujumbura", "Africa/Cairo", "Africa/Casablanca", "Africa/Ceuta", "Africa/Conakry", "Africa/Dakar", "Africa/Dar_es_Salaam", "Africa/Djibouti", "Africa/Douala", "Africa/El_Aaiun", "Africa/Freetown", "Africa/Gaborone", "Africa/Harare", "Africa/Johannesburg", "Africa/Juba", "Africa/Kampala", "Africa/Khartoum", "Africa/Kigali", "Africa/Kinshasa", "Africa/Lagos", "Africa/Libreville", "Africa/Lome", "Africa/Luanda", "Africa/Lubumbashi", "Africa/Lusaka", "Africa/Malabo", "Africa/Maputo", "Africa/Maseru", "Africa/Mbabane", "Africa/Mogadishu", "Africa/Monrovia", "Africa/Nairobi", "Africa/Ndjamena", "Africa/Niamey", "Africa/Nouakchott", "Africa/Ouagadougou", "Africa/Porto-Novo", "Africa/Sao_Tome", "Africa/Tripoli", "Africa/Tunis", "Africa/Windhoek", "America/Adak", "America/Anchorage", "America/Anguilla", "America/Antigua", "America/Araguaina", "America/Argentina/Buenos_Aires", "America/Argentina/Catamarca", "America/Argentina/Cordoba", "America/Argentina/Jujuy", "America/Argentina/La_Rioja", "America/Argentina/Mendoza", "America/Argentina/Rio_Gallegos", "America/Argentina/Salta", "America/Argentina/San_Juan", "America/Argentina/San_Luis", "America/Argentina/Tucuman", "America/Argentina/Ushuaia", "America/Aruba", "America/Asuncion", "America/Atikokan", "America/Bahia", "America/Bahia_Banderas", "America/Barbados", "America/Belem", "America/Belize", "America/Blanc-Sablon", "America/Boa_Vista", "America/Bogota", "America/Boise", "America/Cambridge_Bay", "America/Campo_Grande", "America/Cancun", "America/Caracas", "America/Cayenne", "America/Cayman", "America/Chicago", "America/Chihuahua", "America/Costa_Rica", "America/Creston", "America/Cuiaba", "America/Curacao", "America/Danmarkshavn", "America/Dawson", "America/Dawson_Creek", "America/Denver", "America/Detroit", "America/Dominica", "America/Edmonton", "America/Eirunepe", "America/El_Salvador", "America/Fort_Nelson", "America/Fortaleza", "America/Glace_Bay", "America/Godthab", "America/Goose_Bay", "America/Grand_Turk", "America/Grenada", "America/Guadeloupe", "America/Guatemala", "America/Guayaquil", "America/Guyana", "America/Halifax", "America/Havana", "America/Hermosillo", "America/Indiana/Indianapolis", "America/Indiana/Knox", "America/Indiana/Marengo", "America/Indiana/Petersburg", "America/Indiana/Tell_City", "America/Indiana/Vevay", "America/Indiana/Vincennes", "America/Indiana/Winamac", "America/Inuvik", "America/Iqaluit", "America/Jamaica", "America/Juneau", "America/Kentucky/Louisville", "America/Kentucky/Monticello", "America/Kralendijk", "America/La_Paz", "America/Lima", "America/Los_Angeles", "America/Lower_Princes", "America/Maceio", "America/Managua", "America/Manaus", "America/Marigot", "America/Martinique", "America/Matamoros", "America/Mazatlan", "America/Menominee", "America/Merida", "America/Metlakatla", "America/Mexico_City", "America/Miquelon", "America/Moncton", "America/Monterrey", "America/Montevideo", "America/Montreal", "America/Montserrat", "America/Nassau", "America/New_York", "America/Nipigon", "America/Nome", "America/Noronha", "America/North_Dakota/Beulah", "America/North_Dakota/Center", "America/North_Dakota/New_Salem", "America/Ojinaga", "America/Panama", "America/Pangnirtung", "America/Paramaribo", "America/Phoenix", "America/Port-au-Prince", "America/Port_of_Spain", "America/Porto_Velho", "America/Puerto_Rico", "America/Punta_Arenas", "America/Rainy_River", "America/Rankin_Inlet", "America/Recife", "America/Regina", "America/Resolute", "America/Rio_Branco", "America/Santarem", "America/Santiago", "America/Santo_Domingo", "America/Sao_Paulo", "America/Scoresbysund", "America/Sitka", "America/St_Barthelemy", "America/St_Johns", "America/St_Kitts", "America/St_Lucia", "America/St_Thomas", "America/St_Vincent", "America/Swift_Current", "America/Tegucigalpa", "America/Thule", "America/Thunder_Bay", "America/Tijuana", "America/Toronto", "America/Tortola", "America/Vancouver", "America/Whitehorse", "America/Winnipeg", "America/Yakutat", "America/Yellowknife", "Antarctica/Casey", "Antarctica/Davis", "Antarctica/DumontDUrville", "Antarctica/Macquarie", "Antarctica/Mawson", "Antarctica/McMurdo", "Antarctica/Palmer", "Antarctica/Rothera", "Antarctica/Syowa", "Antarctica/Troll", "Antarctica/Vostok", "Arctic/Longyearbyen", "Asia/Aden", "Asia/Almaty", "Asia/Amman", "Asia/Anadyr", "Asia/Aqtau", "Asia/Aqtobe", "Asia/Ashgabat", "Asia/Atyrau", "Asia/Baghdad", "Asia/Bahrain", "Asia/Baku", "Asia/Bangkok", "Asia/Barnaul", "Asia/Beirut", "Asia/Bishkek", "Asia/Brunei", "Asia/Chita", "Asia/Choibalsan", "Asia/Colombo", "Asia/Damascus", "Asia/Dhaka", "Asia/Dili", "Asia/Dubai", "Asia/Dushanbe", "Asia/Famagusta", "Asia/Gaza", "Asia/Hebron", "Asia/Ho_Chi_Minh", "Asia/Hong_Kong", "Asia/Hovd", "Asia/Irkutsk", "Asia/Jakarta", "Asia/Jayapura", "Asia/Jerusalem", "Asia/Kabul", "Asia/Kamchatka", "Asia/Karachi", "Asia/Kathmandu", "Asia/Khandyga", "Asia/Kolkata", "Asia/Krasnoyarsk", "Asia/Kuala_Lumpur", "Asia/Kuching", "Asia/Kuwait", "Asia/Macau", "Asia/Magadan", "Asia/Makassar", "Asia/Manila", "Asia/Muscat", "Asia/Nicosia", "Asia/Novokuznetsk", "Asia/Novosibirsk", "Asia/Omsk", "Asia/Oral", "Asia/Phnom_Penh", "Asia/Pontianak", "Asia/Pyongyang", "Asia/Qatar", "Asia/Qostanay", "Asia/Qyzylorda", "Asia/Riyadh", "Asia/Sakhalin", "Asia/Samarkand", "Asia/Seoul", "Asia/Shanghai", "Asia/Singapore", "Asia/Srednekolymsk", "Asia/Taipei", "Asia/Tashkent", "Asia/Tbilisi", "Asia/Tehran", "Asia/Thimphu", "Asia/Tokyo", "Asia/Tomsk", "Asia/Ulaanbaatar", "Asia/Urumqi", "Asia/Ust-Nera", "Asia/Vientiane", "Asia/Vladivostok", "Asia/Yakutsk", "Asia/Yangon", "Asia/Yekaterinburg", "Asia/Yerevan", "Atlantic/Azores", "Atlantic/Bermuda", "Atlantic/Canary", "Atlantic/Cape_Verde", "Atlantic/Faroe", "Atlantic/Madeira", "Atlantic/Reykjavik", "Atlantic/South_Georgia", "Atlantic/St_Helena", "Atlantic/Stanley", "Australia/Adelaide", "Australia/Brisbane", "Australia/Broken_Hill", "Australia/Darwin", "Australia/Eucla", "Australia/Hobart", "Australia/Lindeman", "Australia/Lord_Howe", "Australia/Melbourne", "Australia/Perth", "Australia/Sydney", "Europe/Amsterdam", "Europe/Andorra", "Europe/Astrakhan", "Europe/Athens", "Europe/Belgrade", "Europe/Berlin", "Europe/Bratislava", "Europe/Brussels", "Europe/Bucharest", "Europe/Budapest", "Europe/Busingen", "Europe/Chisinau", "Europe/Copenhagen", "Europe/Dublin", "Europe/Gibraltar", "Europe/Guernsey", "Europe/Helsinki", "Europe/Isle_of_Man", "Europe/Istanbul", "Europe/Jersey", "Europe/Kaliningrad", "Europe/Kiev", "Europe/Kirov", "Europe/Lisbon", "Europe/Ljubljana", "Europe/London", "Europe/Luxembourg", "Europe/Madrid", "Europe/Malta", "Europe/Mariehamn", "Europe/Minsk", "Europe/Monaco", "Europe/Moscow", "Europe/Oslo", "Europe/Paris", "Europe/Podgorica", "Europe/Prague", "Europe/Riga", "Europe/Rome", "Europe/Samara", "Europe/San_Marino", "Europe/Sarajevo", "Europe/Saratov", "Europe/Simferopol", "Europe/Skopje", "Europe/Sofia", "Europe/Stockholm", "Europe/Tallinn", "Europe/Tirane", "Europe/Ulyanovsk", "Europe/Uzhgorod", "Europe/Vaduz", "Europe/Vatican", "Europe/Vienna", "Europe/Vilnius", "Europe/Volgograd", "Europe/Warsaw", "Europe/Zagreb", "Europe/Zaporozhye", "Europe/Zurich", "Indian/Antananarivo", "Indian/Chagos", "Indian/Christmas", "Indian/Cocos", "Indian/Comoro", "Indian/Kerguelen", "Indian/Mahe", "Indian/Maldives", "Indian/Mauritius", "Indian/Mayotte", "Indian/Reunion", "Pacific/Apia", "Pacific/Auckland", "Pacific/Bougainville", "Pacific/Chatham", "Pacific/Chuuk", "Pacific/Easter", "Pacific/Efate", "Pacific/Fakaofo", "Pacific/Fiji", "Pacific/Funafuti", "Pacific/Galapagos", "Pacific/Gambier", "Pacific/Guadalcanal", "Pacific/Guam", "Pacific/Honolulu", "Pacific/Kanton", "Pacific/Kiritimati", "Pacific/Kosrae", "Pacific/Kwajalein", "Pacific/Majuro", "Pacific/Marquesas", "Pacific/Midway", "Pacific/Nauru", "Pacific/Niue", "Pacific/Norfolk", "Pacific/Noumea", "Pacific/Pago_Pago", "Pacific/Palau", "Pacific/Pitcairn", "Pacific/Pohnpei", "Pacific/Port_Moresby", "Pacific/Rarotonga", "Pacific/Saipan", "Pacific/Tahiti", "Pacific/Tarawa", "Pacific/Tongatapu", "Pacific/Wake", "Pacific/Wallis",
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
	zoneList   list.Model
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
	Back:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	Quit:   key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Label (e.g. New York)"
	ti.Focus()

	items := make([]list.Item, len(ianaTimezones))
	for i, tz := range ianaTimezones {
		items[i] = zoneItem(tz)
	}
	zl := list.New(items, list.NewDefaultDelegate(), 0, 0)
	zl.Title = "Select Timezone"
	zl.SetShowStatusBar(false)
	zl.SetFilteringEnabled(false) // Disabled filtering
	zl.SetShowFilter(false)       // Removed filter UI
	zl.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#D4AF37"))

	config := loadConfig()
	return model{
		state:     viewList,
		clocks:    config.Clocks,
		textInput: ti,
		zoneList:  zl,
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
		m.zoneList.SetSize(m.width-10, m.height-15)

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

			if m.inputStep == 0 {
				if key.Matches(msg, m.keys.Back) {
					m.state = viewList
					m.textInput.Reset()
					return m, nil
				}
				if key.Matches(msg, m.keys.Enter) {
					val := strings.TrimSpace(m.textInput.Value())
					if val != "" {
						m.newEntry.Label = val
						m.inputStep = 1
						return m, nil
					}
				}
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}

			if m.inputStep == 1 {
				if key.Matches(msg, m.keys.Back) {
					m.inputStep = 0
					return m, nil
				}
				if key.Matches(msg, m.keys.Enter) {
					if i, ok := m.zoneList.SelectedItem().(zoneItem); ok {
						m.newEntry.Location = string(i)
						m.inputStep = 2
						return m, nil
					}
				}
				m.zoneList, cmd = m.zoneList.Update(msg)
				return m, cmd
			}
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
		lipgloss.NewStyle().MarginTop(1).Render(m.help.View(m.keys)),
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
	if m.inputStep == 0 {
		return "STEP 1: ENTER LABEL\n\n" + m.textInput.View()
	} else if m.inputStep == 1 {
		return m.zoneList.View()
	} else {
		return confirmStyle.Render(fmt.Sprintf("CONFIRM ADDING CLOCK?\n\nLabel: %s\nLocation: %s\n\n(y)es / (n)o", 
			labelStyle.Render(m.newEntry.Label), timeStyle.Render(m.newEntry.Location)))
	}
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

var Version = "dev"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Printf("atlas.clock v%s\n", Version)
		return
	}

	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
