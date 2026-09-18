// Copyright 2026 Serge Smertin
// SPDX-License-Identifier: MIT

package tui

import (
	"context"
	"fmt"
	"testing"

	"github.com/nfx/go-tui/internal/assert"
)

func confirmForTest(t *testing.T) (in, out chan string, result chan bool) {
	ctx, cancel := context.WithCancel(context.Background())
	cio := &chanIO{
		ctx: ctx,
		In:  make(chan []byte),
		Out: make(chan []byte),
	}
	ins := make(chan string)
	outs := make(chan string)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case stdin := <-ins:
				select {
				case <-ctx.Done():
					return
				case cio.In <- []byte(stdin):
				}
			case stdout := <-cio.Out:
				select {
				case <-ctx.Done():
					return
				case outs <- string(stdout):
				}
			}
		}
	}()
	t.Cleanup(func() {
		close(ins)
		close(outs)
		close(cio.In)
		close(cio.Out)
		cancel()
	})
	result = make(chan bool)
	go func() {
		defer close(result)
		result <- Confirm("Are you sure?",
			WithInput(cio),
			WithOutput(cio),
			func(a any) error {
				d, ok := a.(*dropdown[string])
				if !ok {
					return fmt.Errorf("not a dropdown")
				}
				// noop the raw term call
				d.makeRawTerm = func() (func() error, error) {
					return func() error {
						return nil
					}, nil
				}
				return nil
			})
	}()
	return ins, outs, result
}

func TestSimpleCase(t *testing.T) {
	in, out, res := confirmForTest(t)
	assert.Equal(t,
		"\rAre you sure? \x1b[36m> Yes\x1b[0m\n\r                No\n\r",
		<-out)
	in <- "\x0d" // enter
	assert.Equal(t, "\x1b[2A\r\x1b[K\x1b[1B\r\x1b[K\x1b[1A\r", <-out)
	assert.Equal(t, "Are you sure?: Yes\n", <-out)
	assert.Equal(t, true, <-res)
}

func TestDenyCase(t *testing.T) {
	in, out, res := confirmForTest(t)
	assert.Equal(t,
		"\rAre you sure? \x1b[36m> Yes\x1b[0m\n\r                No\n\r",
		<-out)
	in <- "\x1b\x5b\x42"
	assert.Equal(t, "\x1b[2A\r\x1b[K\x1b[1B\r\x1b[K\x1b[1A\r", <-out)
	assert.Equal(t,
		"\rAre you sure?   Yes\n\r              \x1b[36m> No\x1b[0m\n\r",
		<-out)
	in <- "\x0d"
	assert.Equal(t, "\x1b[2A\r\x1b[K\x1b[1B\r\x1b[K\x1b[1A\r", <-out)
	assert.Equal(t, "Are you sure?: No\n", <-out)
	assert.Equal(t, false, <-res)
}
