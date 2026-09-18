// Copyright 2026 Serge Smertin
// SPDX-License-Identifier: MIT

package tui

import (
	"io"
	"sync"
)

type view struct {
	width, height int
	lines         [][]byte
	mu            sync.Mutex
}

func isEscapeEnd(r byte) bool {
	// TODO: is '\x40'-'\x5A' is a subset of '@'-'~' or not?..
	// return (r >= '@' && r <= '~') || (r >= '\x40' && r <= '\x5a')
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}

func isEscapeStart(r byte) bool {
	return r == '\x1b' || r == '\x9b'
}

func NewView(w, h int) *view {
	return &view{
		width:  w,
		height: h,
	}
}

func (v *view) WriteTo(w io.Writer) (int64, error) {
	var total int64
	for _, l := range v.lines {
		b, _ := w.Write(l)
		total += int64(b)
		w.Write([]byte{'\n'})
		total++
	}
	return total, nil
}

func width(chunk []byte) int {
	lo, hi, w := 0, len(chunk), 0
	var escape bool
	for lo < hi {
		if escape && isEscapeEnd(chunk[lo]) {
			escape = false
		} else if isEscapeStart(chunk[lo]) {
			escape = true
			w--
		}
		if !escape {
			w++
		}
		lo++
	}
	return w
}

func (v *view) padded(chunk []byte, lo, mid int) (int, int) {
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
	v.lines = append(v.lines, line)
	mid++
	lo = mid
	return lo, mid
}

// see https://notes.burke.libbey.me/ansi-escape-codes/
// see https://gist.github.com/fnky/458719343aabd01cfb17a3a4f7296797
func (v *view) Write(chunk []byte) (n int, err error) {
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
