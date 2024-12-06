package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
	firehose "github.com/shmup/bluesky-firehose.go"
)

type Config struct {
	MinRarity  int
	MaxEntries int
}

var defaultConfig = Config{
	MinRarity:  1,
	MaxEntries: 1000,
}

func loadConfig() Config {
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found, using defaults")
	}

	config := defaultConfig

	if minRarity := os.Getenv("MIN_RARITY"); minRarity != "" {
		if val, err := strconv.Atoi(minRarity); err == nil {
			config.MinRarity = val
		}
	}

	if maxEntries := os.Getenv("MAX_ENTRIES"); maxEntries != "" {
		if val, err := strconv.Atoi(maxEntries); err == nil {
			config.MaxEntries = val
		}
	}

	return config
}

func main() {
	config := loadConfig()

	minRarity := flag.Int("r", config.MinRarity, "Minimum rarity (1-7)")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := NewModel(minRarity)
	p := tea.NewProgram(m, tea.WithAltScreen())

	go startFirehose(ctx, p)

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

	cancel()
	close(m.done)
	time.Sleep(time.Second)
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
