package main

import (
	"encoding/json"
	"fmt"
	"os"
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
}

func NewCollector(workers int) *Collector {
	c := &Collector{
		startTime:   time.Now(),
		workChan:    make(chan string, workers*100),
		discoveries: make([]string, 0, 100),
	}

	for i := 0; i < workers; i++ {
		go c.worker()
	}

	go c.updateDisplay()

	return c
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
	entry := struct {
		Text   string
		Rarity int
		Time   time.Time
	}{text, rarity, time.Now()}

	if data, err := json.Marshal(entry); err == nil {
		f, _ := os.OpenFile("amulets.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		f.Write(append(data, '\n'))
		f.Close()
	}

	display := fmt.Sprintf("%s: %s", []string{"C", "U", "R", "E", "L", "M", "?"}[rarity-4], text)
	c.discoveries = append([]string{display}, c.discoveries...)
	if len(c.discoveries) > 100 {
		c.discoveries = c.discoveries[:100]
	}
}

func (c *Collector) updateDisplay() {
	lastCount := int64(0)
	lastCheck := time.Now()

	for range time.Tick(time.Second) {
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
