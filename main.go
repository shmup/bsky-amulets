package main

import (
	"context"
	"flag"
	"log"
	"runtime"

	firehose "github.com/shmup/bluesky-firehose.go"
)

func main() {
	minRarity := flag.Int("r", 1, "Minimum rarity to display (1-7)")
	maxEntries := flag.Int("n", 100, "Maximum number of entries to display")
	loadHistory := flag.Bool("h", false, "Load history from amulets.json")
	flag.Parse()

	client, _ := firehose.New("wss://bsky.network/xrpc/com.atproto.sync.subscribeRepos")
	collector := NewCollector(runtime.NumCPU(), *minRarity, *maxEntries, *loadHistory)

	client.ConsumeJetstream(
		context.Background(),
		func(post firehose.JetstreamPost) error {
			collector.Process(post.Commit.Record.Text)
			return nil
		},
		func(err error) {
			log.Printf("WebSocket error: %v", err)
			client, _ = firehose.New("wss://bsky.network/xrpc/com.atproto.sync.subscribeRepos")
		},
	)
}
