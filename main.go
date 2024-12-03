package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	tm "github.com/buger/goterm"
	"github.com/joho/godotenv"
	amulet "github.com/shmup/amulet.go"
	bsky "github.com/shmup/bluesky-firehose.go"
)

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

	// Update status line every second
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
			fmt.Printf("\n%s", text)
			fmt.Printf("Is Amulet: %t Rarity: %d\n", isAmulet, rarity)
		}
		return nil
	})
}
