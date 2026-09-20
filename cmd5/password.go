package main

import "github.com/nfx/go-tui"

func main() {
	password, err := tui.Password("Enter password:")
	if err != nil {
		panic(err)
	}
	println(password)
}
