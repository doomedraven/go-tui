// Copyright 2026 Serge Smertin
// SPDX-License-Identifier: MIT

package tui

import (
	"testing"

	"github.com/nfx/go-tui/internal/assert"
)

func TestTrieEscapes(t *testing.T) {
	trie := newTrie()
	trie.Add("\x1b[31mhello\x1b[0m", 1)
	trie.Add("\x1b[31mHELLO\x1b[0m wO\x1b[42mRl\x1b[0md", 2)
	trie.Add("high", 3)
	trie.Add("wo\x1b[42mrl\x1b[0md", 4)
	trie.Add("wonderful", 5)
	trie.Add("привіт", 6)
	trie.Add("\x1b[31mпобут\x1b[0m", 7)
	trie.Add("світ", 7)
	trie.Add("сокіл", 9)
	trie.Add("Н_ОВ_ИЙ    П_Р_ИВ_І_Д", 10)

	assert.Equal(t, []int{1, 2, 3}, trie.Prefix("h"))
	assert.Equal(t, []int{6, 7, 10}, trie.Prefix("_п_"))
	assert.Equal(t, []string{
		"hello", "high", "wonderful", "world",
		"новий", "побут", "привід", "привіт", "світ", "сокіл",
	}, trie.Words())
}
