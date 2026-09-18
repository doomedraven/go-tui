package main

import (
	"log"
	"math/rand/v2"
	"os"
	"strings"
	"time"

	"github.com/nfx/go-tui"
)

func main() {
	raw, _ := os.ReadFile("/usr/share/dict/words")
	words := strings.Split(string(raw), "\n")
	rand.Shuffle(len(words), func(i, j int) {
		words[i], words[j] = words[j], words[i]
	})
	v, err := tui.Dropdown("Select a word", words, tui.WithTimeout(30*time.Second))
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println(v)
}
