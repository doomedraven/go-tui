package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/nfx/go-tui"
	"golang.org/x/term"
)

func main() {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		panic(err)
	}
	slog.Info("got terminal", "w", width, "h", height)

	if !tui.Confirm("Proceed?") {
		return
	}

	// Example usage
	items := []string{
		"Option 1 ...",
		"Option 2 ... .....",
		"Option 3 ...",
		"Option 4 ...",
		"Option 5",
		"Option 6",
		// "Option 7",
		// "Option 8",
		// "Option 9",
		// "Option 10",
		// "Option 11",
		// "Option 12",
		// "Option 13",
		// "Option 24",
		// "Option 34",
		// "Option 45",
		// "Option 56",
		// "Option 67",
		// "Option 78",
		// "Option 83",
		// "Option 94",
		// "Option 102",
		// "Option 113",
		// "Option 123",
	}

	selected, err := tui.Dropdown("Select your option", items,
		tui.WithTimeout(10*time.Second))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("\nYou selected: %s\n", selected)
}
