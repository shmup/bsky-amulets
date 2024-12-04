package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"sync/atomic"
	"time"

	tm "github.com/buger/goterm"
	amulet "github.com/shmup/amulet.go"
)

type AmuletCollector struct {
	startTime     time.Time
	postCount     atomic.Int64
	amuletCount   atomic.Int64
	lastOpCount   int
	lastCheck     time.Time
	workChan      chan string
	workers       int
	displayBuffer []string
	bufferSize    int
}

type AmuletEntry struct {
	Text      string    `json:"text"`
	Rarity    int       `json:"rarity"`
	Timestamp time.Time `json:"timestamp"`
}

func NewAmuletCollector(workers int) *AmuletCollector {
	backupCheck()

	ac := &AmuletCollector{
		startTime:     time.Now(),
		lastCheck:     time.Now(),
		workChan:      make(chan string, workers*100),
		workers:       workers,
		displayBuffer: loadExistingAmulets(),
		bufferSize:    100,
	}

	return ac
}

var rarityLabels = map[int]string{
	4:  "C",
	5:  "U",
	6:  "R",
	7:  "E",
	8:  "L",
	9:  "M",
	10: "?",
}

func (ac *AmuletCollector) addToDisplay(text string, rarity int) {
	display := fmt.Sprintf("%s: %s", rarityLabels[rarity], text)
	ac.displayBuffer = append([]string{display}, ac.displayBuffer...)
	if len(ac.displayBuffer) > ac.bufferSize {
		ac.displayBuffer = ac.displayBuffer[:ac.bufferSize]
	}
}

func (ac *AmuletCollector) startWorkers() {
	for i := 0; i < ac.workers; i++ {
		go func() {
			for text := range ac.workChan {
				if isAmulet, rarity := amulet.IsAmulet(text); isAmulet {
					ac.amuletCount.Add(1)
					ac.logAmulet(text, rarity)
					ac.displayAmulet(text, rarity)
				}
			}
		}()
	}
}

func (ac *AmuletCollector) calculateOPS() float64 {
	duration := time.Since(ac.lastCheck).Seconds()
	currentCount := ac.postCount.Load()
	ops := float64(currentCount-int64(ac.lastOpCount)) / duration
	ac.lastOpCount = int(currentCount)
	ac.lastCheck = time.Now()
	return ops
}

func (ac *AmuletCollector) HandlePost(text string) error {
	ac.postCount.Add(1)
	ac.workChan <- text
	return nil
}

func (ac *AmuletCollector) StartUI() {
	go func() {
		for {
			tm.Clear()
			tm.MoveCursor(1, 1)
			ops := ac.calculateOPS()
			postCount := ac.postCount.Load()
			amuletCount := ac.amuletCount.Load()

			tm.Printf("Posts/sec: %.2f | Posts: %d | Amulets: %d | Runtime: %s\n\n",
				ops,
				postCount,
				amuletCount,
				time.Since(ac.startTime).Round(time.Second))

			for _, amulet := range ac.displayBuffer {
				tm.Printf("%s\n", amulet)
			}

			tm.Flush()
			time.Sleep(time.Second)
		}
	}()
}

func (ac *AmuletCollector) displayAmulet(text string, rarity int) {
	ac.addToDisplay(text, rarity)
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

func backupCheck() {
	_, errMain := os.Stat("amulets.json")
	bakInfo, errBak := os.Stat("amulets.json.bak")
	mainInfo, _ := os.Stat("amulets.json")

	if errMain == nil && errBak == nil && mainInfo.ModTime().After(bakInfo.ModTime()) {
		os.Remove("amulets.json.bak")
		os.Link("amulets.json", "amulets.json.bak")
	} else if errBak != nil && errMain == nil {
		os.Link("amulets.json", "amulets.json.bak")
	}
}

func loadExistingAmulets() []string {
	file, err := os.Open("amulets.json")
	if err != nil {
		return nil
	}
	defer file.Close()

	var entries []AmuletEntry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var entry AmuletEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}
		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})

	displays := make([]string, 0, len(entries))
	for _, entry := range entries {
		displays = append(displays, fmt.Sprintf("%s: %s", rarityLabels[entry.Rarity], entry.Text))
	}
	return displays
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
