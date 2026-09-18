// Copyright 2026 Serge Smertin
// SPDX-License-Identifier: MIT

package tui

import (
	"context"
	"fmt"
	"io"
	"time"
)

type opt func(any) error

type opts []opt

func (o opts) Apply(d any) error {
	var err error
	for _, o := range o {
		err = o(d)
		if err != nil {
			return err
		}
	}
	return nil
}

type withIO interface {
	setWriter(io.Writer)
	setReader(io.Reader)
}

func WithInput(r io.Reader) opt {
	return func(d any) error {
		io, ok := d.(withIO)
		if !ok {
			return fmt.Errorf("cannot set IO")
		}
		io.setReader(r)
		return nil
	}
}

func WithOutput(w io.Writer) opt {
	return func(d any) error {
		io, ok := d.(withIO)
		if !ok {
			return fmt.Errorf("cannot set IO")
		}
		io.setWriter(w)
		return nil
	}
}

type withContext interface {
	setContext(context.Context)
	getContext() context.Context
}

// design tradeoff - we're not passing context as the first argument, because we don't always need it
func WithContext(ctx context.Context) opt {
	return func(d any) error {
		x, ok := d.(withContext)
		if !ok {
			return fmt.Errorf("cannot set context")
		}
		x.setContext(ctx)
		return nil
	}
}

func WithTimeout(timeout time.Duration) opt {
	return func(d any) error {
		x, ok := d.(withContext)
		if !ok {
			return fmt.Errorf("cannot set context")
		}
		ctx := x.getContext()
		ctx, _ = context.WithTimeout(ctx, timeout)
		x.setContext(ctx)
		return nil
	}
}
