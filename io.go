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

func newUnstartedIO(ctx context.Context, width, height int) *chanIO {
	cio := &chanIO{
		ctx:    ctx,
		In:     make(chan string),
		Out:    make(chan string),
		notify: make(chan viewportChanged, 1), // buffered to avoid blocking
		width:  width,
		height: height,
	}
	cio.head = initViewport(cio.ctx, cio.notify, cio.width, cio.height)
	cio.tail = cio.head
	return cio
}

func NewIO(ctx context.Context) *chanIO {
	w, h, _ := term.GetSize(int(os.Stderr.Fd()))
	cio := newUnstartedIO(ctx, w, h)
	go cio.forwardTo(os.Stderr)
	// go io.Copy(cio, os.Stdin) // FIXME: stdin forwarding is not working
	return cio
}

// implements [io.ReadWriter]
type chanIO struct {
	In  chan string
	Out chan string

	ctx context.Context

	width, height int
	head, tail    *viewport
	notify        chan viewportChanged
}

func (i *chanIO) pushViewport() *viewport {
	prev := i.head // TODO: data race
	// TODO: height is not really relevant anymore?..
	i.head = initViewport(i.ctx, i.notify, i.width, i.height)
	i.head.fixedHeight = true
	i.head.next = prev
	return i.head
}

func (i *chanIO) forwardTo(w io.Writer) {
	var prevH, currH int
	for {
		select {
		case <-i.ctx.Done():
			return
		case line := <-i.Out: // deadlocks here
			i.tail.Write([]byte(line)) // fill buffer
		case <-i.notify:
			var buf bytes.Buffer
			if prevH > 0 { // todo: separate thread for flushing all viewports and viewports have to notify it
				space := prevH
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
	select { // don't send on a closed channel
	case <-i.ctx.Done():
		return 0, io.EOF
	default:
	}
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
