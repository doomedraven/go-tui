// Copyright 2026 Serge Smertin
// SPDX-License-Identifier: MIT

package tui

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/nfx/go-tui/internal/assert"
)

func TestView(t *testing.T) {
	for _, tt := range []struct {
		in  string
		out []string
	}{
		{ // explicit NL across lines
			in: "\x1b[1;31m12345\n67890\x1b[0m",
			out: []string{
				"\x1b[1;31m12345     ",
				"67890\x1b[0m     ",
			},
		},
		{ // only one line
			in: "this \x1b[1;31mline\x1b[0m has escape sequences.",
			out: []string{
				"this \x1b[1;31mline\x1b[0m ",
				"has escape",
				" sequences",
				".         ",
			},
		},
		{ // across lines
			in: "this \x1b[1;31mline has\x1b[0m escape sequences.",
			out: []string{
				"this \x1b[1;31mline ",
				"has\x1b[0m escape",
				" sequences",
				".         ",
			},
		},
		{
			in: `this line is
without any escaping characters.`,
			out: []string{
				"this line ",
				"is        ",
				"without an",
				"y escaping",
				" character",
				"s.        ",
			},
		},
	} {
		t.Run(fmt.Sprint(tt), func(t *testing.T) {
			v := &view{
				width:  10,
				height: 10,
			}
			v.Write([]byte(tt.in))
			var lines []string
			for _, line := range v.lines {
				lines = append(lines, string(line))
			}
			assert.Equal(t, tt.out, lines)
		})
	}
}

func TestViewLinkedList(t *testing.T) {
	v := &view{
		lines: [][]byte{[]byte("a"), []byte("b")},
		next: &view{
			lines: [][]byte{[]byte("c"), []byte("d")},
		},
	}
	assert.Equal(t, 4, v.numLines())

	var buf bytes.Buffer
	v.WriteTo(&buf)
	assert.Equal(t, "a\nb\nc\nd\n", buf.String())
}
