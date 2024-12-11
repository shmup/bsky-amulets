package main

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shmup/amulet.go"
)

type Model struct {
	stats       Stats
	entries     []Entry
	entryMap    map[string]Entry
	minRarity   *int
	newestFirst bool
	startTime   time.Time
	viewport    viewport.Model
	writeBuffer chan Entry
	done        chan struct{}
}

func NewModel(minRarity *int) Model {
	v := viewport.New(80, 20)

	m := Model{
		viewport:    v,
		entryMap:    make(map[string]Entry),
		minRarity:   minRarity,
		startTime:   time.Now(),
		writeBuffer: make(chan Entry, 1000),
		done:        make(chan struct{}),
	}

	go m.bufferWriter()

	entries := loadHistoryFromFile(*minRarity)
	for _, entry := range entries {
		m.entryMap[entry.Text] = entry
	}
	m.entries = mapToSlice(m.entryMap, m.newestFirst)

	m.stats.Amulets = 0
	m.stats.TotalAmulets = len(entries)
	m.entries = entries

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		headerHeight := 2
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - headerHeight

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case ProcessMsg:
		return m.handleProcessMsg(msg)

	case HistoryMsg:
		m.entries = msg.Entries
	}

	m.viewport.SetContent(strings.TrimSpace(m.renderEntries()))

	return m, nil
}

func (m Model) handleProcessMsg(msg ProcessMsg) (tea.Model, tea.Cmd) {
	now := time.Now()
	m.stats.Posts++
	m.stats.updateRate(now, m.startTime)

	if isAmulet, rarity := amulet.IsAmulet(msg.Text); isAmulet {
		if rarity >= *m.minRarity {
			entry := Entry{Text: msg.Text, Rarity: rarity, Time: time.Now()}
			if _, exists := m.entryMap[msg.Text]; !exists {
				m.entryMap[msg.Text] = entry
				m.entries = mapToSlice(m.entryMap, m.newestFirst)
				m.stats.Amulets++
				m.stats.TotalAmulets++
				m.writeBuffer <- entry
				m.viewport.SetContent(m.renderEntries())
				if m.newestFirst {
					m.viewport.GotoTop()
				} else {
					m.viewport.GotoBottom()
				}
			}
		}
	}
	return m, nil
}

func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "[":
		m.scrollToPreviousDay()
	case "]":
		m.scrollToNextDay()
	case "j":
		m.viewport.LineDown(1)
	case "k":
		m.viewport.LineUp(1)
	case "g":
		m.viewport.GotoTop()
	case "G":
		m.viewport.GotoBottom()
	case "1", "2", "3", "4", "5", "6", "7":
		newRarity := int(msg.String()[0] - '0')
		*m.minRarity = newRarity

		m.entryMap = make(map[string]Entry)
		loadedEntries := loadHistoryFromFile(*m.minRarity)
		for _, entry := range loadedEntries {
			m.entryMap[entry.Text] = entry
		}
		m.entries = mapToSlice(m.entryMap, m.newestFirst)

		m.stats.TotalAmulets = len(m.entries)
		m.stats.Amulets = len(m.entries)
		m.viewport.SetContent(m.renderEntries())
		if m.newestFirst {
			m.viewport.GotoTop()
		} else {
			m.viewport.GotoBottom()
		}
	case "ctrl+d":
		m.viewport.LineDown(10)
	case "ctrl+u":
		m.viewport.LineUp(10)
	case "r":
		m.newestFirst = !m.newestFirst
		m.entries = mapToSlice(m.entryMap, m.newestFirst)
		m.viewport.SetContent(m.renderEntries())
		m.viewport.GotoTop()
	case "q", "ctrl+c":
		return m, tea.Quit
	}

	return m, nil
}

func (m *Model) scrollToNextDay() {
	currentOffset := m.viewport.YOffset
	currentLine := 0
	for _, line := range strings.Split(m.viewport.View(), "\n") {
		if strings.HasPrefix(line, "January") || strings.HasPrefix(line, "February") || strings.HasPrefix(line, "March") || strings.HasPrefix(line, "April") || strings.HasPrefix(line, "May") || strings.HasPrefix(line, "June") || strings.HasPrefix(line, "July") || strings.HasPrefix(line, "August") || strings.HasPrefix(line, "September") || strings.HasPrefix(line, "October") || strings.HasPrefix(line, "November") || strings.HasPrefix(line, "December") {
			if currentLine > currentOffset {
				m.viewport.SetYOffset(currentLine)
				return
			}
		}
		currentLine++
	}
}

func (m *Model) scrollToPreviousDay() {
	currentOffset := m.viewport.YOffset
	currentLine := 0
	var previousDayOffset int
	for _, line := range strings.Split(m.viewport.View(), "\n") {
		if strings.HasPrefix(line, "January") || strings.HasPrefix(line, "February") || strings.HasPrefix(line, "March") || strings.HasPrefix(line, "April") || strings.HasPrefix(line, "May") || strings.HasPrefix(line, "June") || strings.HasPrefix(line, "July") || strings.HasPrefix(line, "August") || strings.HasPrefix(line, "September") || strings.HasPrefix(line, "October") || strings.HasPrefix(line, "November") || strings.HasPrefix(line, "December") {
			if currentLine >= currentOffset {
				m.viewport.SetYOffset(previousDayOffset)
				return
			}
			previousDayOffset = currentLine
		}
		currentLine++
	}
	m.viewport.SetYOffset(previousDayOffset)
}
