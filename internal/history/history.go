package history

import (
	"fmt"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/ygryan360/lab-cli/internal/config"
)

type Entry struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Category    string    `json:"category"`
	Profile     string    `json:"profile"`
	OpenCount   int       `json:"open_count"`
	LastOpenedAt time.Time `json:"last_opened_at"`
}

type History struct {
	Entries []Entry `json:"entries"`
}

func historyPath() string {
	return filepath.Join(config.ConfigDir(), "history.json")
}

func Load() (*History, error) {
	data, err := os.ReadFile(historyPath())
	if err != nil {
		if os.IsNotExist(err) {
			return &History{Entries: []Entry{}}, nil
		}
		return nil, err
	}
	var h History
	if err := json.Unmarshal(data, &h); err != nil {
		return nil, err
	}
	return &h, nil
}

func Save(h *History) error {
	if err := os.MkdirAll(config.ConfigDir(), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(historyPath(), data, 0644)
}

func (h *History) Record(name, path, category, profile string) {
	for i, e := range h.Entries {
		if e.Path == path {
			h.Entries[i].OpenCount++
			h.Entries[i].LastOpenedAt = time.Now()
			h.Entries[i].Profile = profile
			return
		}
	}
	h.Entries = append(h.Entries, Entry{
		Name:         name,
		Path:         path,
		Category:     category,
		Profile:      profile,
		OpenCount:    1,
		LastOpenedAt: time.Now(),
	})
}

func (h *History) GetProfile(path string) (string, bool) {
	for _, e := range h.Entries {
		if e.Path == path {
			return e.Profile, true
		}
	}
	return "", false
}

// Recent returns the N most recently opened entries
func (h *History) Recent(n int) []Entry {
	sorted := make([]Entry, len(h.Entries))
	copy(sorted, h.Entries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].LastOpenedAt.After(sorted[j].LastOpenedAt)
	})
	if n > len(sorted) {
		n = len(sorted)
	}
	return sorted[:n]
}

// FormatTimeAgo returns a human-friendly relative time string
func FormatTimeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		mins := int(d.Minutes())
		if mins == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", mins)
	case d < 24*time.Hour:
		hours := int(d.Hours())
		if hours == 1 {
			return "1h ago"
		}
		return fmt.Sprintf("%dh ago", hours)
	case d < 7*24*time.Hour:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1d ago"
		}
		return fmt.Sprintf("%dd ago", days)
	default:
		return t.Format("Jan 2")
	}
}
