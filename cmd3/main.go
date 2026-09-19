package main

import (
	"context"
	"log/slog"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/lmittmann/tint"
	"github.com/nfx/go-tui"
)

func main() {
	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, 30*time.Second)
	w := tui.NewIO(ctx)

	s, err := tui.NewSpinners(tui.WithOutput(w), tui.WithContext(ctx))
	if err != nil {
		panic(err)
	}

	first := s.MustAddBackground()
	first.Update("Loading...")

	second := s.MustAddBackground()
	second.Update("Also loading...")

	// set global logger with custom options
	slog.SetDefault(slog.New(
		tint.NewHandler(w, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))

	go func() {
		raw, _ := os.ReadFile("/usr/share/dict/words")
		words := strings.Split(string(raw), "\n")
		ticks := time.NewTicker(333 * time.Millisecond)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticks.C:
				word := words[rand.Intn(len(words))]
				slog.Info("word of the second", "word", word)
			}
		}
	}()

	// tui.Confirm("Do you agree?", tui.WithOutput(w), tui.WithContext(ctx))
	time.Sleep(30 * time.Second)
}
