// Copyright 2026 Serge Smertin
// SPDX-License-Identifier: MIT

package tui

func isEscapeEnd(r byte) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}

func isEscapeStart(r byte) bool {
	return r == '\x1b' || r == '\x9b'
}

// see https://github.com/rivo/uniseg for a more complete implementation.
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
	out = []byte(truncateAscii(string(chunk), maxLen-1))
	if out[len(out)-1] != tailer {
		out = append(out, tailer)
	}

	return
}

func truncateAscii(chunk string, maxLen int) (out string) {
	defer func() {
		if r := recover(); r != nil {
			out = "…" // this is quite a hack
		}
	}()
	if width([]byte(chunk)) <= maxLen {
		return chunk
	}
	lo, hi, width, m := 0, len(chunk), 0, 0
	var escape bool
	for lo < hi {
		escape, m, width = truncateAsciiEscapes(escape, chunk, lo, m, width)
		if !escape {
			if width >= (maxLen - 1) {
				break
			}
			width++
		}
		lo++
	}
	if lo == hi {
		return chunk
	}

	return truncateAsciiFinish(chunk, lo, m)
}

func truncateAsciiEscapes(escape bool, chunk string, lo, m, width int) (bool, int, int) {
	if escape && isEscapeEnd(chunk[lo]) {
		escape = false
		if chunk[lo] == 'm' {
			if chunk[lo-1] == '0' {
				m-- // Select Graphics Rendition (SGR) end
			} else {
				m++ // SGR start
			}
		}
	}
	if isEscapeStart(chunk[lo]) {
		escape = true
		width--
	}

	return escape, m, width
}

func truncateAsciiFinish(chunk string, lo, m int) string {
	raw := []rune(chunk)
	raw = append(raw[:lo], '…')
	for range m {
		// reset every SGR
		raw = append(raw, '\x1b', '[', '0', 'm')
	}

	return string(raw)
}
