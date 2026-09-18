package main

import (
	"log"
	"time"

	"github.com/nfx/go-tui"
)

func main() {
	// raw, _ := os.ReadFile("/usr/share/dict/words")
	// words := strings.Split(string(raw), "\n")
	// rand.Shuffle(len(words), func(i, j int) {
	// 	words[i], words[j] = words[j], words[i]
	// })
	words := []string{
		"this is an extemely long label sentence just for the sake of it being very long lalala and more lalala an more lalala",
		"lorum ipsum dolor sit amet consectetur adipiscing elit sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
		"this is a short label",
		"here is a potentially long label that could be long or short depending on the context but it is not as long as the first one",
	}
	v, err := tui.Dropdown("this is an extemely long label sentence just for the sake of it being very long", words, tui.WithTimeout(30*time.Second))
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println(v)
}
