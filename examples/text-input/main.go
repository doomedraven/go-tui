package main

import "github.com/nfx/go-tui"

func main() {
	result, err := tui.Input("Enter something:")
	if err != nil {
		panic(err)
	}
	println("You entered:", result)
}
