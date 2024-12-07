package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"time"
)

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

func loadHistoryFromFile(minRarity int) []Entry {
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
		if entry.Rarity >= minRarity {
			entries = append(entries, entry)
		}
	}

	return entries
}
