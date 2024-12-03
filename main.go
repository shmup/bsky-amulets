package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"

	amulet "github.com/shmup/amulet.go"
	bsky "github.com/shmup/bluesky-firehose.go"

	"github.com/bluesky-social/indigo/api/atproto"
)

func main() {
	firehose, err := bsky.New("wss://bsky.network/xrpc/com.atproto.sync.subscribeRepos")
	if err != nil {
		log.Fatal(err)
	}
	godotenv.Load()

	email := os.Getenv("BSKY_EMAIL")
	password := os.Getenv("BSKY_PASSWORD")

	err = firehose.Authenticate(email, password)
	if err != nil {
		log.Fatal(err)
	}

	handler := &PostHandler{firehose: firehose}
	firehose.Subscribe(context.Background(), handler)
}

type PostHandler struct {
	firehose *bsky.Firehose
}

func (h *PostHandler) HandleEvent(evt *atproto.SyncSubscribeRepos_Commit) error {
	for _, op := range evt.Ops {
		if op.Action == "create" && strings.HasPrefix(op.Path, "app.bsky.feed.post") {
			uri := fmt.Sprintf("at://%s/app.bsky.feed.post/%s",
				evt.Repo, strings.TrimPrefix(op.Path, "app.bsky.feed.post/"))
			if text, err := h.firehose.FetchPost(uri); err == nil {
				isAmulet, rarity := amulet.IsAmulet(text)
				if isAmulet {
					fmt.Printf("%s", text)
					fmt.Printf("Is Amulet: %t Rarity: %d\n", isAmulet, rarity)
				}
			}
		}
	}
	return nil
}
