// Copyright 2026 Serge Smertin
// SPDX-License-Identifier: MIT

package tui

func isEscapeEnd(r byte) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}

func isEscapeStart(r byte) bool {
	return r == '\x1b' || r == '\x9b'
}

// see https://github.com/rivo/uniseg for a more complete implementation
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
