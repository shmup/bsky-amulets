package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"

	tm "github.com/buger/goterm"
	amulet "github.com/shmup/amulet.go"
	bsky "github.com/shmup/bluesky-firehose.go"
)

type AmuletEntry struct {
	Text      string    `json:"text"`
	Rarity    int       `json:"rarity"`
	Timestamp time.Time `json:"timestamp"`
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

func main() {
	startTime := time.Now()
	var postCount int
	var amuletCount int

	firehose, err := bsky.New("wss://bsky.network/xrpc/com.atproto.sync.subscribeRepos")
	if err != nil {
		log.Fatal(err)
	}
	godotenv.Load()

	if err := firehose.Authenticate(os.Getenv("BSKY_EMAIL"), os.Getenv("BSKY_PASSWORD")); err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			tm.Clear()
			tm.MoveCursor(1, 1)
			tm.Printf("Posts Processed: %d | Amulets Found: %d | Running for: %s\n",
				postCount,
				amuletCount,
				time.Since(startTime).Round(time.Second))
			tm.Flush()
			time.Sleep(time.Second)
		}
	}()

	firehose.OnPost(context.Background(), func(text string) error {
		postCount++
		if isAmulet, rarity := amulet.IsAmulet(text); isAmulet {
			amuletCount++

			entry := AmuletEntry{
				Text:      text,
				Rarity:    rarity,
				Timestamp: time.Now(),
			}

			if err := appendToLog(entry); err != nil {
				log.Printf("Failed to log amulet: %v", err)
			}

			fmt.Printf("\n%s", text)
			fmt.Printf("Is Amulet: %t Rarity: %d\n", isAmulet, rarity)
		}
		return nil
	})
}
