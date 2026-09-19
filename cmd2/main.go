package main

import (
	"fmt"
	"os"

	"github.com/nfx/go-tui"
)

func main() {
	sample := "this \x1b[1;31mline has\x1b[0m escape sequences."
	println(sample)

	v := tui.NewViewport(10, 10)
	v.Write([]byte(sample))

	v.WriteTo(os.Stdout)

	lo, hi := '@', '~'
	for lo <= hi {
		fmt.Printf("%v (\\x%x): %s\n", lo, lo, string([]rune{lo}))
		lo++
	}
}
