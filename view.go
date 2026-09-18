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
	next          *view
	mu            sync.Mutex
}

func isEscapeEnd(r byte) bool {
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

func (v *view) appendChild() *view {
	c := &view{
		width:  v.width,
		height: v.height,
	}
	v.next = c
	return c
}

func (v *view) combinedHeight() int {
	var n int
	curr := v
	for curr != nil {
		n += curr.height
		curr = curr.next
	}
	return n
}

func (v *view) numLines() int {
	var n int
	curr := v
	for curr != nil {
		n += len(curr.lines)
		curr = curr.next
	}
	return n
}

func (v *view) WriteTo(w io.Writer) (int64, error) {
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

func truncateVisible(chunk []byte, maxLen int, tailer byte) (out []byte) {
	out = []byte(truncateASCII(string(chunk), maxLen))
	if out[len(out)-1] != tailer {
		out = append(out, tailer)
	}
	return
}

func truncateASCII(chunk string, maxLen int) (out string) {
	defer func() {
		if r := recover(); r != nil {
			out = "…" // this is quite a hack
		}
	}()
	lo, hi, w, m := 0, len(chunk), 0, 0
	var escape bool
	for lo < hi {
		if escape && isEscapeEnd(chunk[lo]) {
			escape = false
			if chunk[lo] == 'm' {
				if chunk[lo-1] == '0' {
					m-- // Select Graphics Rendition (SGR) end
				} else {
					m++ // SGR start
				}
			}
		} else if isEscapeStart(chunk[lo]) {
			escape = true
			w--
		}
		if !escape {
			if w >= (maxLen - 1) {
				break
			}
			w++
		}
		lo++
	}
	if lo == hi {
		return chunk
	}
	raw := []rune(chunk)
	raw = append(raw[:lo], '…')
	for range m {
		// reset every SGR
		raw = append(raw, '\x1b', '[', '0', 'm')
	}
	out = string(raw)
	return
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
