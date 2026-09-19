// Copyright 2026 Serge Smertin
// SPDX-License-Identifier: MIT

package tui

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/nfx/go-tui/internal/assert"
)

func TestViewport(t *testing.T) {
	for _, tt := range []struct {
		in  string
		out []string
	}{
		{ // explicit NL across lines
			in: "\x1b[1;31m12345\n67890\x1b[0m",
			out: []string{
				"\x1b[1;31m12345     \n",
				"67890\x1b[0m     \n",
			},
		},
		{ // only one line
			in: "this \x1b[1;31mline\x1b[0m has escape sequences.",
			out: []string{
				"this \x1b[1;31mline\x1b[0m \n",
				"has escape\n",
				" sequences\n",
				".         \n",
			},
		},
		{ // across lines
			in: "this \x1b[1;31mline has\x1b[0m escape sequences.",
			out: []string{
				"this \x1b[1;31mline \n",
				"has\x1b[0m escape\n",
				" sequences\n",
				".         \n",
			},
		},
		{
			in: `this line is
without any escaping characters.`,
			out: []string{
				"this line \n",
				"is        \n",
				"without an\n",
				"y escaping\n",
				" character\n",
				"s.        \n",
			},
		},
	} {
		t.Run(fmt.Sprint(tt), func(t *testing.T) {
			v := &viewport{
				width:  10,
				height: 10,
			}
			v.appendToLinebuffer([]byte(tt.in))
			var lines []string
			for _, line := range v.lines {
				lines = append(lines, string(line))
			}
			assert.Equal(t, tt.out, lines)
		})
	}
}

func TestViewportLinkedList(t *testing.T) {
	v := &viewport{
		height: 4,
		lines:  [][]byte{[]byte("a\n"), []byte("b\n")},
		next: &viewport{
			height: 2,
			lines:  [][]byte{[]byte("c\n"), []byte("d\n")},
		},
	}
	assert.Equal(t, 4, v.numLines())

	var buf bytes.Buffer
	v.WriteTo(&buf)
	assert.Equal(t, "\ra\n\rb\n\rc\n\rd\n", buf.String())
}
