// Copyright 2026 Serge Smertin
// SPDX-License-Identifier: MIT

package tui

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

type termIO struct {
	io.ReadWriter
}

func (t *termIO) clear(space int) error {
	// use buffer to write to io only once
	var buf bytes.Buffer
	// Move cursor up to the beginning of the dropdown
	fmt.Fprintf(&buf, "\x1b[%dA", space)
	// Clear each line
	for i := 0; i < space; i++ {
		fmt.Fprint(&buf, "\r")     // return to start of line
		fmt.Fprint(&buf, "\x1b[K") // clear current line
		if i < space-1 {
			fmt.Fprint(&buf, "\x1b[1B") // move cursor down if not last line
		}
	}
	// Move cursor back up to the beginning and to the start of the line
	if space > 1 {
		fmt.Fprintf(&buf, "\x1b[%dA\r", space-1)
	}
	_, err := buf.WriteTo(t)
	return err
}

const (
	keyCtrlC = 0x03
	keyCtrlD = 0x04
	keyEnter = 0x0d
)

var ErrUnknownRune = fmt.Errorf("unknown rune")

func (t *termIO) ReadRune() (rune, error) {
	buf := make([]byte, 4)
	n, err := t.Read(buf) // todo: fixme
	if err == io.EOF {
		return keyCtrlD, io.EOF
	}
	if n >= 3 && buf[0] == 0x1b && buf[1] == 0x5b {
		switch buf[2] {
		case 0x41: // Up arrow.
			return '↑', nil
		case 0x42: // Down arrow.
			return '↓', nil
		}
	}
	if n > 1 {
		return 0, fmt.Errorf("%w: %x", ErrUnknownRune, buf)
	}
	switch buf[0] {
	case keyCtrlC, keyCtrlD:
		return 0, io.EOF
	default:
		return rune(buf[0]), nil
	}
}

func isTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// see https://stackoverflow.com/a/37014283/277035
func isPrintable(r rune) bool {
	isSurrogate := r >= 0xd800 && r <= 0xdbff
	return r >= 32 && !isSurrogate
}

func makeRawTerm() (func() error, error) {
	// Switch to raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return nil, fmt.Errorf("raw term: %w", err)
	}
	return func() error {
		return term.Restore(int(os.Stdin.Fd()), oldState)
	}, nil
}
