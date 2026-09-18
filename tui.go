// Copyright 2026 Serge Smertin
// SPDX-License-Identifier: MIT

package tui

import "context"

func NewTUI(ctx context.Context, opts ...opt) *Tui {
	return &Tui{
		opts: opts,
		ctx:  ctx,
		termIO: termIO{
			ReadWriter: NewIO(ctx), // TODO: handle errors and stuff
		},
	}
}

type Tui struct {
	opts
	termIO // exposes io.ReadWriter
	ctx    context.Context
}

func (t *Tui) prependView() *view {
	cio, ok := t.termIO.ReadWriter.(*chanIO)
	if !ok {
		panic("cannot get view")
	}
	top := &view{next: cio.head, width: cio.width}
	cio.head = top
	return top
}

func (t *Tui) view() *view {
	cio, ok := t.termIO.ReadWriter.(*chanIO)
	if ok {
		return cio.head
	}
	return nil
}
