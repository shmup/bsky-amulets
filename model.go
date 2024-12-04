package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	amulet "github.com/shmup/amulet.go"
)

type Model struct {
	stats       Stats
	entries     []Entry
	maxEntries  *int
	minRarity   *int
	startTime   time.Time
	viewport    viewport.Model
	writeBuffer chan Entry
	done        chan struct{}
}

type Stats struct {
	Posts        int
	Amulets      int
	TotalAmulets int
	Rate         float64
	recentPosts  []time.Time
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
		viewport:    v,
		maxEntries:  maxEntries,
		minRarity:   minRarity,
		startTime:   time.Now(),
		writeBuffer: make(chan Entry, 100),
		done:        make(chan struct{}),
	}

	go m.bufferWriter()

	if *loadHistory {
		entries := loadHistoryFromFile(*minRarity, *maxEntries)
		m.stats.Amulets = 0
		m.stats.TotalAmulets = len(entries)
		m.entries = entries
	}

	return m
}

func (m *Model) bufferWriter() {
	f, err := os.OpenFile("amulets.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open file: %v", err)
		return
	}
	defer f.Close()

	for {
		select {
		case entry := <-m.writeBuffer:
			data, _ := json.Marshal(entry)
			f.Write(append(data, '\n'))
		case <-m.done:
			return
		}
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		headerHeight := 2 // Stats header height

		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - headerHeight

	case tea.KeyMsg:
		switch msg.String() {
		case "j":
			m.viewport.LineDown(1)
			return m, nil
		case "k":
			m.viewport.LineUp(1)
			return m, nil
		case "g":
			m.viewport.GotoTop()
			return m, nil
		case "G":
			m.viewport.GotoBottom()
			return m, nil
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case ProcessMsg:
		now := time.Now()
		m.stats.Posts++
		m.stats.recentPosts = append(m.stats.recentPosts, now)

		cutoff := now.Add(-time.Minute)
		for i, t := range m.stats.recentPosts {
			if t.After(cutoff) {
				m.stats.recentPosts = m.stats.recentPosts[i:]
				break
			}
		}

		runtime := time.Since(m.startTime).Seconds()
		if runtime < 60 {
			m.stats.Rate = float64(m.stats.Posts) / runtime
		} else {
			m.stats.Rate = float64(len(m.stats.recentPosts)) / 60.0
		}

		if isAmulet, rarity := amulet.IsAmulet(msg.Text); isAmulet {
			if rarity >= *m.minRarity+3 {
				m.stats.Amulets++
				m.stats.TotalAmulets++
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

	m.viewport.SetContent(m.renderEntries())

	return m, nil
}

func (m Model) renderStats() string {
	runtime := time.Since(m.startTime).Round(time.Second)
	return fmt.Sprintf("SPS: %.2f | Skeets: %d | Found: %d | Total: %d | Runtime: %s\n",
		m.stats.Rate, m.stats.Posts, m.stats.Amulets, m.stats.TotalAmulets, runtime)
}

func (m Model) renderEntries() string {
	var output string
	for _, e := range m.entries {
		rarityLabel := []string{"C", "U", "R", "E", "L", "M", "?"}[e.Rarity-4]
		output += fmt.Sprintf("%s: %s\n", rarityLabel, e.Text)
	}
	return output
}

func (m Model) View() string {
	return fmt.Sprintf("%s\n%s",
		m.renderStats(),
		m.viewport.View())
}

func (m Model) logEntry(entry Entry) {
	m.writeBuffer <- entry
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
