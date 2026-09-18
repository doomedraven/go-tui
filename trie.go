// Copyright 2026 Serge Smertin
// SPDX-License-Identifier: MIT

package tui

type trie struct {
	m    map[byte]*trie
	word string
}

func newTrie() *trie {
	return &trie{
		m: map[byte]*trie{},
	}
}

func (t *trie) Add(word string) {
	r := t
	for _, c := range word {
		b := byte(c)
		s, ok := r.m[b]
		if !ok {
			s = newTrie()
			r.m[b] = s
		}
		r = s
	}
	r.word = word
}

func (t *trie) Search(word string) bool {
	r := t
	for _, c := range word {
		b := byte(c)
		s, ok := r.m[b]
		if !ok {
			return false
		}
		r = s
	}
	return r.word != ""
}
