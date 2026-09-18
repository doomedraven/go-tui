// Copyright 2026 Serge Smertin
// SPDX-License-Identifier: MIT

package tui

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
)

type dropdown[T any] struct {
	Ctx     context.Context
	Label   string
	Items   []T
	Default T
	Hide    bool

	ItemsFn func(prefix string) []T
	In      io.Reader
	Out     io.Writer

	selected int
	io       *termIO

	makeRawTerm func() (func() error, error)
}

func Confirm(action string, opts ...opt) bool {
	res, err := Dropdown(action, []string{"Yes", "No"}, opts...)
	if err != nil {
		return false
	}
	return strings.ToLower(res) == "yes"
}

func Dropdown[T any](label string, items []T, opts ...opt) (T, error) {
	var zero T
	i, err := DropdownIndex(label, items, opts...)
	if err != nil {
		return zero, err
	}
	// we know i is valid
	return items[i], nil
}

func DropdownIndex[T any](label string, items []T, opts ...opt) (int, error) {
	d, err := newDropdown[T](opts)
	if err != nil {
		return -1, err
	}
	d.Label = label
	d.Items = items
	i, err := d.run()
	if err != nil {
		return -1, err
	}
	if !d.Hide {
		fmt.Fprintf(d.io, "%s: %s\n", d.Label, fmt.Sprint(items[i]))
	}
	return i, nil
}

func newDropdown[T any](o opts) (*dropdown[T], error) {
	d := &dropdown[T]{
		Ctx:         context.Background(),
		Label:       "Select from list",
		makeRawTerm: makeRawTerm,
		io: &termIO{
			ReadWriter: defaultIO,
		},
	}
	err := o.Apply(d)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (d *dropdown[T]) getTIO() *tio {
	io, ok := d.io.ReadWriter.(*tio)
	if !ok {
		return nil
	}
	return io
}

// implements [withContext]
func (d *dropdown[T]) setContext(ctx context.Context) {
	d.Ctx = ctx
}

// implements [withContext]
func (d *dropdown[T]) getContext() context.Context {
	return d.Ctx
}

// render displays the dropdown
func (d *dropdown[T]) render() {
	// use buffer to write to io only once
	var buf bytes.Buffer
	var prefix int
	for i, item := range d.Items {
		fmt.Fprint(&buf, "\r") // ensure we start from the leftmost position
		if i == 0 {
			prefix, _ = fmt.Fprintf(&buf, "%s ", d.Label)
		} else {
			for range prefix {
				fmt.Fprintf(&buf, " ")
			}
		}
		// TODO: print spaces till the end of the terminal width
		if i == d.selected {
			fmt.Fprintf(&buf, "\033[36m> %s\033[0m\n", fmt.Sprint(item)) // cyan color for selected item
		} else {
			fmt.Fprintf(&buf, "  %s\n", fmt.Sprint(item))
		}
	}
	fmt.Fprint(&buf, "\r")
	buf.WriteTo(d.io)
}

// Show displays the dropdown and handles user input
func (d *dropdown[T]) run() (int, error) {
	restore, err := d.makeRawTerm()
	if err != nil {
		return -1, err
	}
	defer restore()
	d.render()
	for {
		space := len(d.Items)
		select {
		case <-d.Ctx.Done():
			d.io.clear(space)
			return -1, d.Ctx.Err()
		default:
			key, err := d.io.ReadRune()
			if err != nil { // Ctrl+C or Ctrl+D
				d.io.clear(space)
				return -1, err
			}
			switch key {
			case keyEnter:
				d.io.clear(space)
				return d.selected, nil
			case '↑':
				if d.selected > 0 {
					d.selected--
					d.io.clear(space)
					d.render()
				}
			case '↓':
				if d.selected < len(d.Items)-1 {
					d.selected++
					d.io.clear(space)
					d.render()
				}
			}
		}
	}
}
