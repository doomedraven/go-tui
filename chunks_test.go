// Copyright 2026 Serge Smertin
// SPDX-License-Identifier: MIT

package tui

import (
	"fmt"
	"testing"

	"github.com/nfx/go-tui/internal/assert"
)

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
