// Copyright 2026 Serge Smertin
// SPDX-License-Identifier: MIT

package tui

import "sort"

type trie struct {
	m    map[byte]*trie
	word []byte
	idx  []int
}

func newTrie() *trie {
	return &trie{
		m: map[byte]*trie{},
	}
}

func (t *trie) Add(word []byte, i int) {
	r := t
	var escape bool
	for _, b := range word {
		if escape && isEscapeEnd(b) {
			escape = false
			continue
		} else if isEscapeStart(b) {
			escape = true
		}
		if escape {
			continue
		}
		// todo: skip printable non-alphanumeric characters
		s, ok := r.m[b]
		if !ok {
			s = newTrie()
			r.m[b] = s
		}
		r = s
	}
	r.idx = append(r.idx, i)
	r.word = word // TODO: strip escape sequences
}

func (t *trie) Prefix(prefix []byte, limit int) []int {
	r := t
	for _, b := range prefix {
		s, ok := r.m[b]
		if !ok {
			return nil
		}
		r = s
	}
	return r.Indexes(limit)
}

func (t *trie) Indexes(limit int) (out []int) {
	q := []*trie{t}
	for len(q) > 0 {
		r := q[0]
		q = q[1:]
		if len(out) >= limit {
			break
		}
		out = append(out, r.idx...)
		for _, s := range r.m {
			q = append(q, s)
		}
	}
	if len(out) > limit {
		out = out[:limit]
	}
	// keep output in the same order
	sort.Ints(out)
	return
}
