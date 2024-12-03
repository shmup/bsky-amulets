package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	bsky "github.com/shmup/bluesky-firehose.go"
)

func main() {
	godotenv.Load()
	collector := NewAmuletCollector()
	firehose := setupFirehose()

	collector.StartUI()
	firehose.OnPost(context.Background(), collector.HandlePost)
}

func setupFirehose() *bsky.Firehose {
	firehose, err := bsky.New("wss://bsky.network/xrpc/com.atproto.sync.subscribeRepos")
	if err != nil {
		log.Fatal(err)
	}

	if err := firehose.Authenticate(os.Getenv("BSKY_EMAIL"), os.Getenv("BSKY_PASSWORD")); err != nil {
		log.Fatal(err)
	}

	return firehose
}
