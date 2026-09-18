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

func NewIO(ctx context.Context) *chanIO {
	w, h, _ := term.GetSize(int(os.Stderr.Fd()))
	head := &view{
		width:  w,
		height: h,
	}
	cio := &chanIO{
		ctx:    ctx,
		In:     make(chan []byte),
		Out:    make(chan []byte),
		width:  w,
		height: h,
		head:   head,
		tail:   head,
	}
	// fmt.Printf("w:%d, h:%d\n", w, h)

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
	head, tail    *view
}

func (i *chanIO) forwardTo(w io.Writer) {
	for {
		select {
		case <-i.ctx.Done():
			return
		case line := <-i.Out:
			var buf bytes.Buffer
			i.tail.Write(line) // fill buffer
			space := i.head.combinedHeight()
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
				// -1 words well for mid scroll, but -1 is good for screen redraw
				fmt.Fprintf(&buf, "\x1b[%dA\r", space-1)
			}
			i.head.WriteTo(&buf)
			x := buf.Bytes()
			w.Write(x[:len(x)-1]) // trim last newline
			// _, _ = buf.WriteTo(w)
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
		copy(p, res)
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
