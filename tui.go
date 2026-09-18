// Copyright 2026 Serge Smertin
// SPDX-License-Identifier: MIT

package tui

import "context"

func NewTUI(ctx context.Context, opts ...opt) (*Tui, error) {
	tio, err := makeTermIO(defaultIO.Reader, defaultIO.Writer)
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

func (t *Tui) prependView() *view {
	cio, ok := t.termIO.Writer.(*chanIO)
	if !ok {
		panic("cannot get view")
	}
	top := &view{next: cio.head, width: cio.width}
	cio.head = top
	return top
}

func (t *Tui) view() *view {
	cio, ok := t.termIO.Writer.(*chanIO)
	if ok {
		return cio.head
	}
	return nil
}
