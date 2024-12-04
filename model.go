// model.go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	amulet "github.com/shmup/amulet.go"
)

type Model struct {
	stats      Stats
	entries    []Entry
	maxEntries *int
	minRarity  *int
	startTime  time.Time
	viewport   viewport.Model
}

type Stats struct {
	Posts   int
	Amulets int
	Rate    float64
}

type Entry struct {
	Text   string
	Rarity int
	Time   time.Time
}

type ProcessMsg struct {
	Text string
}

type HistoryMsg struct {
	Entries []Entry
}

func NewModel(maxEntries, minRarity *int, loadHistory *bool) Model {
	v := viewport.New(80, 20)

	m := Model{
		viewport:   v,
		maxEntries: maxEntries,
		minRarity:  minRarity,
		startTime:  time.Now(),
	}

	if *loadHistory {
		entries := loadHistoryFromFile(*minRarity, *maxEntries)
		m.stats.Amulets = len(entries)
		m.entries = entries
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case ProcessMsg:
		m.stats.Posts++
		if isAmulet, rarity := amulet.IsAmulet(msg.Text); isAmulet {
			m.stats.Amulets++
			if rarity >= *m.minRarity+3 {
				entry := Entry{Text: msg.Text, Rarity: rarity, Time: time.Now()}
				m.logEntry(entry)
				m.entries = append([]Entry{entry}, m.entries...)
				if len(m.entries) > *m.maxEntries {
					m.entries = m.entries[:*m.maxEntries]
				}
			}
		}
	case HistoryMsg:
		m.entries = msg.Entries
	}
	return m, nil
}

func (m Model) View() string {
	runtime := time.Since(m.startTime).Round(time.Second)
	rate := float64(m.stats.Posts) / runtime.Seconds()

	output := fmt.Sprintf("Posts/sec: %.2f | Posts: %d | Amulets: %d | Runtime: %s\n\n",
		rate, m.stats.Posts, m.stats.Amulets, runtime)

	for _, e := range m.entries {
		rarityLabel := []string{"C", "U", "R", "E", "L", "M", "?"}[e.Rarity-4]
		output += fmt.Sprintf("%s: %s\n", rarityLabel, e.Text)
	}

	m.viewport.SetContent(output)
	return m.viewport.View()
}

func (m Model) logEntry(entry Entry) {
	data, _ := json.Marshal(entry)
	f, _ := os.OpenFile("amulets.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	f.Write(append(data, '\n'))
}

func loadHistoryFromFile(minRarity, maxEntries int) []Entry {
	file, err := os.Open("amulets.json")
	if err != nil {
		return nil
	}
	defer file.Close()

	var entries []Entry
	decoder := json.NewDecoder(file)
	for decoder.More() {
		var entry Entry
		if err := decoder.Decode(&entry); err != nil {
			continue
		}
		if entry.Rarity >= minRarity+3 {
			entries = append(entries, entry)
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Time.After(entries[j].Time)
	})

	if len(entries) > maxEntries {
		entries = entries[:maxEntries]
	}

	return entries
}
