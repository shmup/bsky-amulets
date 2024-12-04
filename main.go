package main

import (
	"context"
	"log"
	"runtime"

	firehose "github.com/shmup/bluesky-firehose.go"
)

func main() {
	client, _ := firehose.New("wss://bsky.network/xrpc/com.atproto.sync.subscribeRepos")
	collector := NewCollector(runtime.NumCPU())

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
