// Copyright 2026 Serge Smertin
// SPDX-License-Identifier: MIT

package tui

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

type tio struct {
	io.Reader
	io.Writer
}

var defaultIO = &tio{
	Reader: os.Stdin,
	Writer: os.Stderr,
}

type incr struct {
	chunk []byte
	lines int
}

func NewIO() *chanIO {
	w, h, _ := term.GetSize(int(os.Stderr.Fd()))
	w = 25
	h = 10
	cio := &chanIO{
		ctx:    context.Background(),
		In:     make(chan []byte),
		Out:    make(chan []byte),
		width:  w,
		height: h,
		frozen: &view{
			width:  w,
			height: h,
		},
	}
	fmt.Printf("w:%d, h:%d\n", w, h)

	go cio.forwardTo(os.Stderr)
	go io.Copy(cio, os.Stdin)
	return cio
}

// implements [io.ReadWriter]
type chanIO struct {
	In  chan []byte
	Out chan []byte

	ctx context.Context

	width, height int
	frozen        *view
}

func (i *chanIO) forwardTo(w io.Writer) {
	for {
		select {
		case <-i.ctx.Done():
			return
		case line := <-i.Out:
			i.frozen.Write(line) // fill buffer
			if len(i.frozen.lines) < i.height {
				w.Write(line)
			} else {
				space := i.height
				var buf bytes.Buffer
				// Move cursor up to the beginning of the dropdown
				fmt.Fprintf(&buf, "\033[%dA", space)
				// Clear each line
				for i := 0; i < space; i++ {
					fmt.Fprint(&buf, "\r")     // return to start of line
					fmt.Fprint(&buf, "\033[K") // clear current line
					if i < space-1 {
						fmt.Fprint(&buf, "\033[1B") // move cursor down if not last line
					}
				}
				// Move cursor back up to the beginning and to the start of the line
				if space > 1 {
					fmt.Fprintf(&buf, "\033[%dA\r", space-1)
				}
				i.frozen.lines = i.frozen.lines[len(i.frozen.lines)-i.height:]
				for _, line := range i.frozen.lines {
					buf.Write(line)
				}
				_, _ = buf.WriteTo(w)
			}
		}
	}
}

func (i *chanIO) Read(p []byte) (n int, err error) {
	select {
	case <-i.ctx.Done():
		return 0, io.EOF
	case res, ok := <-i.In:
		if !ok {
			return 0, io.EOF
		}
		p = res
		return len(res), nil
	}
}

func (i *chanIO) Write(p []byte) (n int, err error) {
	select {
	case <-i.ctx.Done():
		return 0, io.EOF
	case i.Out <- p:
		return len(p), nil
	}
}
