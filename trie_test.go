// Copyright 2026 Serge Smertin
// SPDX-License-Identifier: MIT

package tui

import (
	"testing"

	"github.com/nfx/go-tui/internal/assert"
)

func TestTrieEscapes(t *testing.T) {
	trie := newTrie()
	trie.Add([]byte("\x1b[31mhello\x1b[0m"), 1)
	trie.Add([]byte("\x1b[31mhello\x1b[0m wo\x1b[42mrl\x1b[0md"), 2)
	trie.Add([]byte("high"), 3)
	trie.Add([]byte("wo\x1b[42mrl\x1b[0md"), 4)
	trie.Add([]byte("wonderful"), 5)

	assert.Equal(t, []int{1, 3}, trie.Prefix([]byte("h"), 2))
}
