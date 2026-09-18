package main

import (
	"os"
	"strings"
	"time"

	"math/rand"

	"github.com/nfx/go-tui"
)

func main() {
	raw, _ := os.ReadFile("/usr/share/dict/words")
	words := strings.Split(string(raw), "\n")
	rand.Shuffle(len(words), func(i, j int) {
		words[i], words[j] = words[j], words[i]
	})
	tui.Dropdown("Select a word", words, tui.WithTimeout(30*time.Second))
}
