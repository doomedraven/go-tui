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
		height: 2,
		lines:  [][]byte{[]byte("a"), []byte("b")},
		next: &view{
			height: 2,
			lines:  [][]byte{[]byte("c"), []byte("d")},
		},
	}
	assert.Equal(t, 4, v.numLines())

	var buf bytes.Buffer
	v.WriteTo(&buf)
	assert.Equal(t, "a\nb\nc\nd\n", buf.String())
}

func TestTruncateASCII(t *testing.T) {
	for _, tt := range []struct {
		in  string
		out string
	}{
		{"12345\x1b[1;31m6\x1b[0m789", "123…"},
		{"\x1b[1;31m12\x1b[0m\x1b[1;32m345\x1b[0m", "\x1b[1;31m12\x1b[0m\x1b[1;32m3…\x1b[0m"},
		{"1\x1b[1;31m2\x1b[0m345", "1\x1b[1;31m2\x1b[0m3…"},
		{"1\x1b[1;31m23\x1b[0m45", "1\x1b[1;31m23\x1b[0m…"},
		{"1\x1b[1;31m234\x1b[0m5", "1\x1b[1;31m23…\x1b[0m"},
		{"1\x1b[1;31m2\x1b[1;32m3\x1b[0m45\x1b[0m6", "1\x1b[1;31m2\x1b[1;32m3\x1b[0m…\x1b[0m"},
		{"1\x1b[1;31m2\x1b[1;32m34\x1b[0m5\x1b[0m6", "1\x1b[1;31m2\x1b[1;32m3…\x1b[0m\x1b[0m"},
		{"", ""},
		{"a", "a"},
		{"ab", "ab"},
		{"abc", "abc"},
		{"abcd", "abc…"},
		{"abcdef", "abc…"},
		{"abcde", "abc…"},
	} {
		t.Run(fmt.Sprint(tt), func(t *testing.T) {
			assert.Equal(t, tt.out, string(truncateASCII(tt.in, 4)))
		})
	}
}
