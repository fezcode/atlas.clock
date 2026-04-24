package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Entry is a single clock on the dashboard.
type Entry struct {
	Label    string `json:"label"`
	Location string `json:"location"` // IANA name ("Europe/Istanbul") or "Local"
}

// Config is the persisted dashboard state.
type Config struct {
	Clocks []Entry `json:"clocks"`
}

// ConfigPath returns the default config path: $HOME/.atlas/clock.json.
func ConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".atlas", "clock.json")
}

// Load reads the config, returning a sensible default if the file doesn't exist.
func Load() Config {
	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		return Config{Clocks: []Entry{
			{Label: "Local", Location: "Local"},
			{Label: "UTC", Location: "UTC"},
			{Label: "Istanbul", Location: "Europe/Istanbul"},
		}}
	}
	var cfg Config
	_ = json.Unmarshal(data, &cfg)
	return cfg
}

// Save atomically persists the config (best-effort; errors are ignored to
// keep the TUI responsive — the config is a convenience, not a source of truth).
func Save(cfg Config) error {
	path := ConfigPath()
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Now returns the current time in the entry's zone. Invalid zones fall back
// to local time so a malformed config never crashes the UI.
func (e Entry) Now() time.Time {
	if e.Location == "" || e.Location == "Local" {
		return time.Now()
	}
	loc, err := time.LoadLocation(e.Location)
	if err != nil {
		return time.Now()
	}
	return time.Now().In(loc)
}
