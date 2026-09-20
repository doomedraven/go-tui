package main

import "github.com/nfx/go-tui"

func main() {
	password, err := tui.Password()
	if err != nil {
		panic(err)
	}
	println(password)
}
