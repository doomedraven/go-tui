// Copyright 2026 Serge Smertin
// SPDX-License-Identifier: MIT

package tui

import (
	"io"
	"sync"
)

type viewport struct {
	width, height int
	lines         [][]byte
	next          *viewport
	mu            sync.Mutex
}

func NewViewport(w, h int) *viewport {
	return &viewport{
		width:  w,
		height: h,
	}
}

func (v *viewport) appendChild() *viewport {
	c := &viewport{
		width:  v.width,
		height: v.height,
	}
	v.next = c
	return c
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
	for curr != nil {
		if len(curr.lines) > curr.height {
			curr.lines = curr.lines[len(curr.lines)-curr.height+1:]
		}
		for _, l := range curr.lines {
			b, err := w.Write(l)
			if err != nil {
				return total, err
			}
			w.Write([]byte{'\n'})
			total += int64(b) + 1
		}
		curr = curr.next
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

// see https://notes.burke.libbey.me/ansi-escape-codes/
// see https://gist.github.com/fnky/458719343aabd01cfb17a3a4f7296797
func (v *viewport) Write(chunk []byte) (n int, err error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	lo, mid, hi := 0, 0, len(chunk)
	var printed int
	var escape bool
	for mid < hi {
		if escape && isEscapeEnd(chunk[mid]) {
			escape = false
		} else if isEscapeStart(chunk[mid]) {
			escape = true
			printed--
		}
		if printed > 0 && printed%v.width == 0 {
			if lo < mid {
				tmp := make([]byte, mid-lo+1)
				copy(tmp, chunk[lo:mid])
				tmp[len(tmp)-1] = '\n'
				v.lines = append(v.lines, tmp)
			}
			lo = mid
		}
		if chunk[mid] == '\n' { // FIXME: windows is \r\n ?..
			lo, mid = v.padded(chunk, lo, mid)
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
	}
	return len(chunk), nil
}
