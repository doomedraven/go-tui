package main

import "github.com/nfx/go-tui"

func main() {
	result, err := tui.Password("Enter password:")
	if err != nil {
		panic(err)
	}
	println("You entered:", result)
}
