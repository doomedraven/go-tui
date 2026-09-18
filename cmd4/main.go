package main

import (
	"fmt"
	"os"
	"time"

	"github.com/nfx/go-tui"
	"golang.org/x/term"
)

func main() {
	// Check if we're in a terminal
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		fmt.Println("This program requires a terminal")
		return
	}

	// Create a spinner manager
	sm := tui.NewSpinnerManager()

	// Create some spinners with update channels
	updates1 := make(chan string)
	updates2 := make(chan string)

	sm.AddSpinner(updates1)
	sm.AddSpinner(updates2)

	// Send updates to spinners
	go func() {
		updates1 <- "Processing task 1..."
		time.Sleep(2 * time.Second)
		updates1 <- "Almost done with task 1..."
		time.Sleep(2 * time.Second)
		close(updates1) // This will remove the spinner
	}()

	go func() {
		updates2 <- "Processing task 2..."
		time.Sleep(3 * time.Second)
		updates2 <- "Finalizing task 2..."
		time.Sleep(2 * time.Second)
		close(updates2) // This will remove the spinner
	}()

	// Wait for all spinners to complete
	time.Sleep(6 * time.Second)
}
