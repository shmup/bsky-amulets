package main

import (
	"sort"
	"time"
)

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

func mapToSlice(m map[string]Entry, newestFirst bool) []Entry {
	entries := make([]Entry, 0, len(m))
	for _, entry := range m {
		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Time.After(entries[j].Time) == newestFirst
	})

	return entries
}
