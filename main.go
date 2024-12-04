package main

import (
	"context"
	"flag"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	firehose "github.com/shmup/bluesky-firehose.go"
)

func main() {
	minRarity := flag.Int("r", 2, "Minimum rarity (1-7)")
	maxEntries := flag.Int("n", 30, "Maximum entries")
	loadHistory := flag.Bool("h", true, "Load history")
	flag.Parse()

	p := tea.NewProgram(NewModel(maxEntries, minRarity, loadHistory))
	go startFirehose(p)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func startFirehose(p *tea.Program) {
	client, _ := firehose.New("wss://bsky.network/xrpc/com.atproto.sync.subscribeRepos")
	client.ConsumeJetstream(
		context.Background(),
		func(post firehose.JetstreamPost) error {
			p.Send(ProcessMsg{Text: post.Commit.Record.Text})
			return nil
		},
		func(err error) {
			client, _ = firehose.New("wss://bsky.network/xrpc/com.atproto.sync.subscribeRepos")
		},
	)
}
