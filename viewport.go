// Copyright 2026 Serge Smertin
// SPDX-License-Identifier: MIT

package tui

import (
	"context"
	"io"
	"sync"
)

type viewportChanged int

type viewport struct {
	width, height int
	lines         [][]byte
	next          *viewport
	mu            sync.Mutex
	inner         chan []byte
	notify        chan viewportChanged
	ctx           context.Context
	fixedHeight   bool
	lastLines     int
}

func initViewport(ctx context.Context, notify chan viewportChanged, width, height int) *viewport {
	v := &viewport{
		ctx:    ctx,
		inner:  make(chan []byte),
		notify: notify,
		width:  width,
		height: height,
	}
	go v.loop()
	return v
}

func (v *viewport) combinedHeight() int {
	var n int
	curr := v
	for curr != nil {
		n += curr.height
		curr = curr.next
	}
	return n
}

func (v *viewport) numLines() int {
	var n int
	curr := v
	for curr != nil {
		n += len(curr.lines)
		curr = curr.next
	}
	return n
}

func (v *viewport) WriteTo(w io.Writer) (int64, error) {
	var total int64
	curr := v
	budget := v.height
	for curr != nil {
		if curr.fixedHeight && curr.lastLines > 0 {
			curr.lines = curr.lines[len(curr.lines)-curr.lastLines:]
		} else if len(curr.lines) > curr.height {
			// FIXME: race condition and potential data corruption
			// TODO: definitely need two offsets, as the top fixed viewport will be the first to be trimmed
			curr.lines = curr.lines[len(curr.lines)-curr.height:]
		}
		for _, l := range curr.lines {
			w.Write([]byte{'\r'})
			b, err := w.Write(l)
			if err != nil {
				return total, err
			}
			total += int64(b) + 1
		}
		budget -= len(curr.lines)
		if budget <= 0 {
			break
		}
		curr = curr.next
		if curr != nil {
			// simplified assumption: tail viewport cannot have fixed height
			curr.height = budget
		}
	}
	return total, nil
}

func (v *viewport) padded(chunk []byte, lo, mid int) (int, int) {
	pos, line := lo, []byte{}
	for pos < mid {
		line = append(line, chunk[pos])
		pos++
	}
	pl := v.width - width(chunk[lo:mid])
	if pl < 0 {
		pl = 0
	}
	for range pl {
		line = append(line, ' ')
	}
	line = append(line, '\n') // FIXME: windows is \r\n ?..
	v.lines = append(v.lines, line)
	mid++
	lo = mid
	return lo, mid
}

func (v *viewport) WriteByte(b byte) error {
	_, err := v.Write([]byte{b})
	return err
}

// see https://notes.burke.libbey.me/ansi-escape-codes/
// see https://gist.github.com/fnky/458719343aabd01cfb17a3a4f7296797
func (v *viewport) Write(chunk []byte) (n int, err error) {
	select {
	case <-v.ctx.Done():
		return 0, io.EOF
	case v.inner <- chunk:
		return len(chunk), nil
	}
}

func (v *viewport) appendToLinebuffer(chunk []byte) int {
	lo, mid, hi := 0, 0, len(chunk)
	var printed, addedLines int
	var escape bool
	for mid < hi {
		if escape && isEscapeEnd(chunk[mid]) {
			escape = false
		} else if isEscapeStart(chunk[mid]) {
			escape = true
			printed--
		}
		// TODO: skip \r as well
		if printed > 0 && printed%v.width == 0 {
			if lo < mid {
				tmp := make([]byte, mid-lo+1)
				copy(tmp, chunk[lo:mid])
				tmp[len(tmp)-1] = '\n'
				v.lines = append(v.lines, tmp)
				addedLines++
			}
			lo = mid
		}
		if chunk[mid] == '\n' { // FIXME: windows is \r\n ?..
			lo, mid = v.padded(chunk, lo, mid)
			addedLines++
			printed = 0 // reset printed char count
			continue
		}
		if !escape {
			printed++
		}
		mid++
	}
	if lo < hi { // todo: check for escape seqs
		v.padded(chunk, lo, hi)
		addedLines++
	}
	return addedLines
}

func (v *viewport) loop() {
	for {
		// technically, we can cleanup the old lines here on a time interval,
		// maitaining "append" and "display" offsets
		select {
		case <-v.ctx.Done():
			return
		case chunk := <-v.inner:
			v.lastLines = v.appendToLinebuffer(chunk)
			select {
			case <-v.ctx.Done():
				return
			case v.notify <- viewportChanged(v.lastLines):
			}
		}
	}
}
