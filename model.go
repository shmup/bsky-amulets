package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
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
	Posts        int
	Amulets      int
	TotalAmulets int
	Rate         float64
	recentPosts  [60]time.Time
	postIndex    int
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

		// Update ring buffer
		m.stats.recentPosts[m.stats.postIndex] = now
		m.stats.postIndex = (m.stats.postIndex + 1) % 60

		// Calculate rate based on runtime or ring buffer
		runtime := time.Since(m.startTime).Seconds()
		if runtime < 60 {
			// For first minute, use total posts / runtime
			m.stats.Rate = float64(m.stats.Posts) / runtime
		} else {
			// After first minute, count posts in last 60 seconds
			count := 0
			cutoff := now.Add(-time.Minute)
			for _, t := range m.stats.recentPosts {
				if !t.IsZero() && t.After(cutoff) {
					count++
				}
			}
			m.stats.Rate = float64(count)
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

	baseStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("21")).
		Width(m.viewport.Width).
		Padding(0, 0)

	valueStyle := lipgloss.NewStyle()

	separator := lipgloss.NewStyle().
		SetString(" │ ").
		String()

	stats := baseStyle.Render(fmt.Sprintf("SPS: %s%sSkeets: %s%sFound: %s%sTotal: %s%sRuntime: %s",
		valueStyle.Render(fmt.Sprintf("%6.2f", m.stats.Rate)), separator,
		valueStyle.Render(fmt.Sprintf("%3d", m.stats.Posts)), separator,
		valueStyle.Render(fmt.Sprintf("%d", m.stats.Amulets)), separator,
		valueStyle.Render(fmt.Sprintf("%d", m.stats.TotalAmulets)), separator,
		valueStyle.Render(runtime.String())))

	return stats
}

func (m Model) renderEntries() string {
	var output strings.Builder

	normalRow := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Background(lipgloss.Color("236")).
		Width(m.viewport.Width).
		Padding(0, 0)

	alternateRow := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Background(lipgloss.Color("234")).
		Width(m.viewport.Width).
		Padding(0, 0)

	for i, e := range m.entries {
		style := normalRow
		if i%2 == 0 {
			style = alternateRow
		}

		coloredText := style.Render(e.Text)
		output.WriteString(coloredText + "\n")
	}

	return output.String()
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
