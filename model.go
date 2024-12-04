package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	Posts         int
	Amulets       int
	TotalAmulets  int
	Rate          float64
	lastPostTimes []time.Time
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
		writeBuffer: make(chan Entry, 1000),
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

	writer := bufio.NewWriterSize(f, 32*1024)
	defer writer.Flush()

	batch := make([]Entry, 0, 100)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case entry := <-m.writeBuffer:
			batch = append(batch, entry)
			if len(batch) >= 100 {
				writeBatch(writer, batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) > 0 {
				writeBatch(writer, batch)
				batch = batch[:0]
			}
		case <-m.done:
			if len(batch) > 0 {
				writeBatch(writer, batch)
			}
			return
		}
	}
}

func writeBatch(writer *bufio.Writer, entries []Entry) {
	for _, entry := range entries {
		data, _ := json.Marshal(entry)
		writer.Write(append(data, '\n'))
	}
	writer.Flush()
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		headerHeight := 2 // stats header height

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
		m.stats.lastPostTimes = append(m.stats.lastPostTimes, now)

		timeSinceStart := now.Sub(m.startTime)
		if timeSinceStart < time.Minute {
			// during first minute calculate rate based on time since start
			m.stats.Rate = float64(len(m.stats.lastPostTimes)) / timeSinceStart.Seconds()
		} else {
			// after first minute, use rolling window
			cutoff := now.Add(-time.Minute)
			for i, t := range m.stats.lastPostTimes {
				if t.After(cutoff) {
					m.stats.lastPostTimes = m.stats.lastPostTimes[i:]
					break
				}
			}
			m.stats.Rate = float64(len(m.stats.lastPostTimes)) / 60.0
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

func (m Model) View() string {
	containerStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("0")).
		Height(m.viewport.Height + 1) // +1 for stats bar

	return containerStyle.Render(fmt.Sprintf("%s\n%s",
		m.renderStats(),
		m.viewport.View()))
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
