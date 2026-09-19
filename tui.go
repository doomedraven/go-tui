// Copyright 2026 Serge Smertin
// SPDX-License-Identifier: MIT

package tui

import (
	"context"
	"os"
)

func NewTUI(ctx context.Context, opts ...opt) (*Tui, error) {
	cio := NewIO(ctx)
	tio, err := makeTermIO(os.Stdin, cio)
	if err != nil {
		return nil, err
	}
	return &Tui{
		opts:   opts,
		ctx:    ctx,
		termIO: tio,
	}, nil
}

type Tui struct {
	opts
	*termIO // exposes io.ReadWriter
	ctx     context.Context
}

func (t *Tui) prependView() *viewport {
	cio, ok := t.termIO.out.(*chanIO)
	if !ok {
		panic("cannot get view")
	}
	top := &viewport{next: cio.head, width: cio.width}
	cio.head = top
	return top
}

func (t *Tui) view() *viewport {
	cio, ok := t.termIO.out.(*chanIO)
	if ok {
		return cio.head
	}
	return nil
}
