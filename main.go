package main

import (
	"context"
	"flag"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	firehose "github.com/shmup/bluesky-firehose.go"
)

func main() {
	minRarity := flag.Int("r", 3, "Minimum rarity (1-7)")
	maxEntries := flag.Int("n", 1000, "Maximum entries")
	loadHistory := flag.Bool("h", true, "Load history")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := NewModel(maxEntries, minRarity, loadHistory)
	p := tea.NewProgram(m, tea.WithAltScreen())

	go startFirehose(ctx, p)

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

	cancel()                // stop firehose
	close(m.done)           // stop file writer
	time.Sleep(time.Second) // give goroutines time to clean up
}

func startFirehose(ctx context.Context, p *tea.Program) {
	backoff := time.Second
	for {
		select {
		case <-ctx.Done():
			return
		default:
			client, err := firehose.New("wss://bsky.network/xrpc/com.atproto.sync.subscribeRepos")
			if err != nil {
				time.Sleep(backoff)
				backoff *= 2
				continue
			}
			backoff = time.Second

			streamCtx, cancel := context.WithCancel(ctx)
			go func() {
				<-ctx.Done()
				cancel()
			}()

			if err := client.ConsumeJetstream(
				streamCtx,
				func(post firehose.JetstreamPost) error {
					select {
					case <-streamCtx.Done():
						return context.Canceled
					default:
						p.Send(ProcessMsg{Text: post.Commit.Record.Text})
						return nil
					}
				},
			); err != nil {
				cancel()
				if ctx.Err() != nil {
					return
				}
				continue
			}
		}
	}
}
