// package main

// import (
// 	"context"
// 	"fmt"
// 	"log"

// 	"github.com/joho/godotenv"
// 	bsky "github.com/shmup/bluesky-firehose.go"
// )

// func main() {
// 	godotenv.Load()
// 	collector := NewAmuletCollector()
// 	firehose := setupFirehose()

// 	collector.StartUI()
// 	firehose.OnPost(context.Background(), collector.HandlePost)

// 	firehose.ConsumeJetstream(context.Background(), func(post firehose.JetstreamPost) error {
// 		fmt.Printf("Post from %s: %s\n", post.DID, post.Commit.Record.Text)
// 		return nil
// 	})
// }

// func setupFirehose() *bsky.Firehose {
// 	firehose, err := bsky.New("wss://bsky.network/xrpc/com.atproto.sync.subscribeRepos")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

//		return firehose
//	}
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
