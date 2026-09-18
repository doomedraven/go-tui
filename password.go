// Copyright 2026 Serge Smertin
// SPDX-License-Identifier: MIT

package tui

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

func readPasswordWithStars(prompt string) (string, error) {
	fmt.Print(prompt)

	// Get the file descriptor for standard input
	fd := int(os.Stdin.Fd())

	// Set terminal to raw mode
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return "", err
	}
	defer term.Restore(fd, oldState) // Restore terminal state on function exit

	// Read input character by character
	var password strings.Builder
	for {
		var buf [1]byte
		_, err := os.Stdin.Read(buf[:])
		if err != nil {
			return "", err
		}

		if buf[0] == '\n' || buf[0] == '\r' { // Enter key pressed
			fmt.Println() // Move to the next line
			break
		} else if buf[0] == 127 { // Backspace key pressed
			if password.Len() > 0 {
				// password.Truncate(password.Len() - 1)
				fmt.Print("\b \b") // Erase the last star
			}
		} else {
			password.WriteByte(buf[0])
			fmt.Print("*") // Print a star for each character
		}
	}

	return password.String(), nil
}
