package main

import (
	"context"

	firehose "github.com/shmup/bluesky-firehose.go"
)

var client *firehose.Firehose

func main() {
	client, _ = firehose.New("wss://bsky.network/xrpc/com.atproto.sync.subscribeRepos")

	collector := NewAmuletCollector()
	collector.StartUI()

	client.ConsumeJetstream(context.Background(), func(post firehose.JetstreamPost) error {
		collector.HandlePost(post.Commit.Record.Text)
		return nil
	})
}
