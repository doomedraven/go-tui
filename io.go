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

type bbuf []byte

func (b *bbuf) String() string {
	return string(*b)
}

func (b *bbuf) Write(p []byte) (n int, err error) {
	*b = append(*b, p...)
	return len(p), nil
}

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
	head := &viewport{
		width:  w,
		height: h,
	}
	cio := &chanIO{
		ctx:    ctx,
		In:     make(chan string),
		Out:    make(chan string),
		width:  w,
		height: h,
		head:   head,
		tail:   head,
	}
	go cio.forwardTo(os.Stderr)
	go io.Copy(cio, os.Stdin)
	return cio
}

// implements [io.ReadWriter]
type chanIO struct {
	In  chan string
	Out chan string

	ctx context.Context

	width, height int
	head, tail    *viewport
}

func (i *chanIO) NewViewport() *viewport {
	return i.head.appendChild()
}

func (i *chanIO) forwardTo(w io.Writer) {
	var prevH, currH int
	for {
		select {
		case <-i.ctx.Done():
			return
		case line := <-i.Out:
			var buf bytes.Buffer
			i.tail.Write([]byte(line)) // fill buffer
			if prevH > 0 {             // todo: separate thread for flushing all viewports and viewports have to notify it
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
					// space-1 words well for mid scroll, but space-1 is good for screen redraw
					fmt.Fprintf(&buf, "\x1b[%dA\r", space)
				}
			}
			currH = i.head.combinedHeight()
			prevH = currH
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
	case i.Out <- string(p):
		return len(p), nil
	}
}

func newWriteC(ctx context.Context) *writeC {
	return &writeC{
		Context: ctx,
		C:       make(chan string),
	}
}

type writeC struct {
	context.Context
	C chan string
}

func (x *writeC) Write(p []byte) (n int, err error) {
	select {
	case <-x.Done():
		return 0, io.EOF
	case x.C <- string(p):
		return len(p), nil
	}
}
