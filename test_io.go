// Copyright 2026 Serge Smertin
// SPDX-License-Identifier: MIT

package tui

import (
	"os"

	"golang.org/x/term"
)

func NewTestIO() *chanIO {
	return &chanIO{
		In:  make(chan string),
		Out: make(chan string),
	}
}
func getTerminalWidth() (int, error) {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 0, err
	}
	return width, nil
}
