package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync/atomic"
	"time"

	tm "github.com/buger/goterm"
	amulet "github.com/shmup/amulet.go"
)

type Collector struct {
	posts       atomic.Int64
	amulets     atomic.Int64
	startTime   time.Time
	workChan    chan string
	discoveries []string
	minRarity   int
	maxEntries  int
}

func NewCollector(workers, minRarity, maxEntries int, loadHistory bool) *Collector {
	c := &Collector{
		startTime:   time.Now(),
		workChan:    make(chan string, workers*100),
		discoveries: make([]string, 0, maxEntries),
		minRarity:   minRarity,
		maxEntries:  maxEntries,
	}

	// force immediate display
	tm.Clear()
	tm.MoveCursor(1, 1)
	tm.Printf("Initializing...\n")
	tm.Flush()

	for i := 0; i < workers; i++ {
		go c.worker()
	}

	go c.updateDisplay()

	if loadHistory {
		go c.loadHistoryFromFile()
	}

	return c
}

func (c *Collector) updateDisplay() {
	lastCount := int64(0)
	lastCheck := time.Now()

	// immediate first update
	tm.Clear()
	tm.MoveCursor(1, 1)
	tm.Printf("Posts/sec: 0.00 | Posts: 0 | Amulets: 0 | Runtime: 0s\n\n")
	tm.Flush()

	// then regular updates
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		current := c.posts.Load()
		ops := float64(current-lastCount) / time.Since(lastCheck).Seconds()
		lastCount = current
		lastCheck = time.Now()

		tm.Clear()
		tm.MoveCursor(1, 1)
		tm.Printf("Posts/sec: %.2f | Posts: %d | Amulets: %d | Runtime: %s\n\n",
			ops,
			current,
			c.amulets.Load(),
			time.Since(c.startTime).Round(time.Second))

		for _, d := range c.discoveries {
			tm.Printf("%s\n", d)
		}
		tm.Flush()
	}
}

func (c *Collector) loadHistoryFromFile() {
	file, err := os.Open("amulets.json")
	if err != nil {
		return
	}
	defer file.Close()

	reader := bufio.NewReaderSize(file, 32*1024) // 32KB buffer
	decoder := json.NewDecoder(reader)

	var entries []struct {
		Text   string    `json:"Text"`
		Rarity int       `json:"Rarity"`
		Time   time.Time `json:"Time"`
	}

	for decoder.More() {
		var entry struct {
			Text   string    `json:"Text"`
			Rarity int       `json:"Rarity"`
			Time   time.Time `json:"Time"`
		}
		if err := decoder.Decode(&entry); err != nil {
			continue
		}
		if entry.Rarity >= c.minRarity {
			entries = append(entries, entry)
		}
	}

	c.processLoadedEntries(entries)
}

func (c *Collector) processLoadedEntries(entries []struct {
	Text   string    `json:"Text"`
	Rarity int       `json:"Rarity"`
	Time   time.Time `json:"Time"`
}) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Time.After(entries[j].Time)
	})

	if len(entries) > c.maxEntries {
		entries = entries[:c.maxEntries]
	}

	for _, entry := range entries {
		display := fmt.Sprintf("%s: %s", []string{"C", "U", "R", "E", "L", "M", "?"}[entry.Rarity-4], entry.Text)
		c.discoveries = append(c.discoveries, display)
	}
}

func (c *Collector) worker() {
	for text := range c.workChan {
		if isAmulet, rarity := amulet.IsAmulet(text); isAmulet {
			c.amulets.Add(1)
			c.logDiscovery(text, rarity)
		}
	}
}

func (c *Collector) Process(text string) {
	c.posts.Add(1)
	c.workChan <- text
}
func (c *Collector) logDiscovery(text string, rarity int) {
	if rarity < c.minRarity+3 { // Add 3 to align with actual rarity values
		return
	}

	entry := struct {
		Text   string    `json:"Text"`
		Rarity int       `json:"Rarity"`
		Time   time.Time `json:"Time"`
	}{text, rarity, time.Now()}

	if data, err := json.Marshal(entry); err == nil {
		f, _ := os.OpenFile("amulets.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		f.Write(append(data, '\n'))
		f.Close()
	}

	display := fmt.Sprintf("%s: %s", []string{"C", "U", "R", "E", "L", "M", "?"}[rarity-4], text)
	c.discoveries = append([]string{display}, c.discoveries...)
	if len(c.discoveries) > c.maxEntries {
		c.discoveries = c.discoveries[:c.maxEntries]
	}
}
