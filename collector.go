package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	tm "github.com/buger/goterm"
	amulet "github.com/shmup/amulet.go"
)

type AmuletCollector struct {
	startTime   time.Time
	postCount   int
	amuletCount int
}

type AmuletEntry struct {
	Text      string    `json:"text"`
	Rarity    int       `json:"rarity"`
	Timestamp time.Time `json:"timestamp"`
}

func NewAmuletCollector() *AmuletCollector {
	return &AmuletCollector{
		startTime: time.Now(),
	}
}

func (ac *AmuletCollector) HandlePost(text string) error {
	ac.postCount++
	if isAmulet, rarity := amulet.IsAmulet(text); isAmulet {
		ac.amuletCount++
		ac.logAmulet(text, rarity)
		ac.displayAmulet(text, rarity)
	}
	return nil
}

func (ac *AmuletCollector) StartUI() {
	go func() {
		for {
			tm.Clear()
			tm.MoveCursor(1, 1)
			tm.Printf("Posts Processed: %d | Amulets Found: %d | Running for: %s\n",
				ac.postCount,
				ac.amuletCount,
				time.Since(ac.startTime).Round(time.Second))
			tm.Flush()
			time.Sleep(time.Second)
		}
	}()
}

func (ac *AmuletCollector) logAmulet(text string, rarity int) {
	entry := AmuletEntry{
		Text:      text,
		Rarity:    rarity,
		Timestamp: time.Now(),
	}

	if err := appendToLog(entry); err != nil {
		log.Printf("Failed to log amulet: %v", err)
	}
}

func appendToLog(entry AmuletEntry) error {
	f, err := os.OpenFile("amulets.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	jsonEntry, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	if _, err := f.WriteString(string(jsonEntry) + "\n"); err != nil {
		return err
	}
	return nil
}

func (ac *AmuletCollector) displayAmulet(text string, rarity int) {
	fmt.Printf("\n%s", text)
	fmt.Printf("Is Amulet: true Rarity: %d\n", rarity)
}
