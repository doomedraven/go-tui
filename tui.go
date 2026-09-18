// Copyright 2026 Serge Smertin
// SPDX-License-Identifier: MIT

package tui

func NewTUI(opts ...opt) *Tui {
	return &Tui{
		opts: opts,
	}
}

type Tui struct {
	opts
}
