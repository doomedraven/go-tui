package main

import (
	"log/slog"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/lmittmann/tint"
	"github.com/nfx/go-tui"
)

func main() {
	w := tui.NewIO()

	// set global logger with custom options
	slog.SetDefault(slog.New(
		tint.NewHandler(w, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))

	raw, _ := os.ReadFile("/usr/share/dict/words")
	words := strings.Split(string(raw), "\n")
	ticks := time.NewTicker(333 * time.Millisecond)
	for {
		select {
		case <-ticks.C:
			word := words[rand.Intn(len(words))]
			slog.Info("word of the second", "word", word)
		}
	}
}
