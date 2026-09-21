package main

import "github.com/nfx/go-tui"

func main() {
	var err error
	// _, err = tui.Dropdown("Select", []string{"a", "b", "c", "d", "e", "f", "g",
	//  "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"})
	emu := func(yield func(string, error) bool) {
		for _, v := range []string{
			"1", "2", "3", "4", "5", "6", "7", "8", "9", "10",
			"11", "12", "13", "14", "15", "16", "17", "18", "19", "20",
			"21", "22", "23", "24", "25", "26", "27", "28", "29", "30"} {
			if !yield(v, nil) {
				return
			}
		}
	}
	_, err = tui.DropdownLazy("Select", emu)
	panic(err)
}
