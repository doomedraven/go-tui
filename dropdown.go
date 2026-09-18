// Copyright 2026 Serge Smertin
// SPDX-License-Identifier: MIT

package tui

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
)

type dropdown struct {
	Ctx       context.Context
	Label     string
	Items     []any
	displayed []any
	Default   any
	Hide      bool
	OneReturn bool

	LabelTemplate        string
	labelTemplate        *template.Template
	ActiveItemTemplate   string
	activeItemTemplate   *template.Template
	InactiveItemTemplate string
	inactiveItemTemplate *template.Template
	MoreItemsTemplate    string
	moreItemsTemplate    *template.Template
	AnswerTemplate       string
	answerTemplate       *template.Template

	ItemsFn func(prefix string) []any

	selected int
	offset   int

	in  io.Reader
	out io.Writer

	// TODO: special case for testing?..
	makeTermIO func(in io.Reader, out io.Writer) (*termIO, error)
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
	// apparently, there's no other non-reflective way around
	anyItems := make([]any, len(items))
	for i, v := range items {
		anyItems[i] = v
	}
	i, err := DropdownIndex(label, anyItems, opts...)
	if err != nil {
		return zero, err
	}
	// we know i is valid
	return items[i], nil
}

func DropdownIndex(label string, items []any, o ...opt) (int, error) {
	d, err := newDropdown()
	if err != nil {
		return -1, err
	}
	d.Label = label
	d.Items = items
	err = opts(o).Apply(d)
	if err != nil {
		return -1, err
	}
	if d.OneReturn && len(d.Items) == 1 {
		return 0, nil
	}
	err = d.parseTemplates()
	if err != nil {
		return -1, fmt.Errorf("templates: %w", err)
	}
	i, err := d.run()
	if err != nil {
		return -1, err
	}
	if !d.Hide {
		err = d.answerTemplate.Execute(d.out, dropdownAnswer{
			Label:  label,
			Answer: items[i],
		})
		if err != nil {
			return i, fmt.Errorf("answer: %w", err)
		}
		// TODO: append trailing spaces and newline to templates where necessary
		d.out.Write([]byte{'\n'})
	}
	return i, nil
}

var DefaultDropdownLabelTemplate = `{{ "?" | green }} {{ . | bold }}`
var DefaultDropdownActiveItemTemplate = `{{ cyan "→ " . }}`
var DefaultDropdownInactiveItemTemplate = `{{ dim "→ " . }}`
var DefaultDropdownMoreItemsTemplate = ` {{ dim "↓ " .More " more … (" .Total " total)" | italic }}`
var DefaultDropdownAnswerTemplate = `{{ dim "✔ " .Label " …" }} {{ .Answer | bold }}`

type dropdownAnswer struct {
	Label  string
	Answer any
}

func dropdownOpt(o func(d *dropdown) error) opt {
	return func(a any) error {
		// check if a is any dropdown
		d, ok := a.(*dropdown)
		if !ok {
			return fmt.Errorf("need a dropdown, got %v", a)
		}
		return o(d)
	}
}

func WithOneReturn() opt {
	return dropdownOpt(func(d *dropdown) error {
		d.OneReturn = true
		return nil
	})
}

func WithHide() opt {
	return dropdownOpt(func(d *dropdown) error {
		d.Hide = true
		return nil
	})
}

func WithLabelTemplate(tmpl string) opt {
	return dropdownOpt(func(d *dropdown) error {
		d.LabelTemplate = tmpl
		return nil
	})
}

func WithActiveItemTemplate(tmpl string) opt {
	return dropdownOpt(func(d *dropdown) error {
		d.ActiveItemTemplate = tmpl
		return nil
	})
}

func WithInactiveItemTemplate(tmpl string) opt {
	return dropdownOpt(func(d *dropdown) error {
		d.InactiveItemTemplate = tmpl
		return nil
	})
}

func WithAnswerTemplate(tmpl string) opt {
	return dropdownOpt(func(d *dropdown) error {
		d.AnswerTemplate = tmpl
		return nil
	})
}

func newDropdown() (*dropdown, error) {
	d := &dropdown{
		in:                   os.Stdin,
		out:                  os.Stderr,
		Ctx:                  context.Background(),
		Label:                "Select from list",
		makeTermIO:           makeTermIO,
		LabelTemplate:        DefaultDropdownLabelTemplate,
		ActiveItemTemplate:   DefaultDropdownActiveItemTemplate,
		InactiveItemTemplate: DefaultDropdownInactiveItemTemplate,
		MoreItemsTemplate:    DefaultDropdownMoreItemsTemplate,
		AnswerTemplate:       DefaultDropdownAnswerTemplate,
	}
	return d, nil
}

func (d *dropdown) parseTemplates() error {
	var err error
	tmpl := template.New("dropdown").Funcs(colorFns)
	d.labelTemplate, err = tmpl.New("label").Parse(d.LabelTemplate)
	if err != nil {
		return fmt.Errorf("label: %w", err)
	}
	d.activeItemTemplate, err = tmpl.New("active").Parse(d.ActiveItemTemplate)
	if err != nil {
		return fmt.Errorf("active item: %w", err)
	}
	d.inactiveItemTemplate, err = tmpl.New("inactive").Parse(d.InactiveItemTemplate)
	if err != nil {
		return fmt.Errorf("inactive item: %w", err)
	}
	d.moreItemsTemplate, err = tmpl.New("more").Parse(d.MoreItemsTemplate)
	if err != nil {
		return fmt.Errorf("more items: %w", err)
	}
	d.answerTemplate, err = tmpl.New("answer").Parse(d.AnswerTemplate)
	if err != nil {
		return fmt.Errorf("answer: %w", err)
	}
	return nil
}

// implements [withIO]
func (d *dropdown) setReader(r io.Reader) {
	d.in = r
}

// implements [withIO]
func (d *dropdown) setWriter(w io.Writer) {
	tui, ok := w.(*Tui)
	if ok {
		c := tui.prependView()
		c.height = d.space() + 1
		c.next.height -= c.height // TODO: propagate down
		w = c
	}
	d.out = w
}

// implements [withContext]
func (d *dropdown) setContext(ctx context.Context) {
	d.Ctx = ctx
}

// implements [withContext]
func (d *dropdown) getContext() context.Context {
	return d.Ctx
}

// render displays the dropdown
func (d *dropdown) render(io *termIO) error {
	// use buffer to write to io only once
	var buf bytes.Buffer
	var prefix int
	var err error
	total := len(d.Items)
	height := min(total, io.Height/2)
	if len(d.displayed) == 0 {
		d.displayed = d.Items[:height]
	}
	for i, item := range d.displayed {
		fmt.Fprint(&buf, "\r") // ensure we start from the leftmost position
		if i == 0 {
			var labelBuf bytes.Buffer
			err = d.labelTemplate.Execute(&labelBuf, d.Label)
			if err != nil {
				return fmt.Errorf("label: %w", err)
			}
			prefix = width(labelBuf.Bytes()) + 1
			labelBuf.WriteTo(&buf)
			buf.WriteByte(' ')
		} else {
			for range prefix {
				fmt.Fprintf(&buf, " ")
			}
		}
		if i == d.selected {
			err = d.activeItemTemplate.Execute(&buf, item)
			if err != nil {
				return fmt.Errorf("active: %w", err)
			}
			buf.WriteByte('\n')
		} else {
			err = d.inactiveItemTemplate.Execute(&buf, item)
			if err != nil {
				return fmt.Errorf("inactive: %w", err)
			}
			buf.WriteByte('\n')
		}
	}
	if total > len(d.displayed) {
		fmt.Fprint(&buf, "\r") // always display a line to avoid flickering
		if d.offset+height < total {
			for range prefix - 1 { // ???...
				fmt.Fprintf(&buf, " ")
			}
			err = d.moreItemsTemplate.Execute(&buf, dropdownMore{
				More:  total - d.offset - height,
				Total: total,
			})
			if err != nil {
				return fmt.Errorf("more: %w", err)
			}
		}
		buf.WriteByte('\n')
	}
	fmt.Fprint(&buf, "\r")
	_, err = buf.WriteTo(io)
	if err != nil {
		return fmt.Errorf("flush: %w", err)
	}
	return nil
}

type dropdownMore struct {
	More  int
	Total int
}

func (d *dropdown) space() int {
	displayed, total := len(d.displayed), len(d.Items)
	if total > displayed {
		return displayed + 1 // more ... row
	}
	return displayed
}

// Show displays the dropdown and handles user input
func (d *dropdown) run() (int, error) {
	io, err := d.makeTermIO(d.in, d.out)
	if err != nil {
		return -1, fmt.Errorf("raw term: %w", err)
	}
	defer io.Restore()
	for {
		err = d.render(io)
		if err != nil {
			return -1, fmt.Errorf("render: %w", err)
		}
		space := d.space()
		displayed := len(d.displayed)
		select {
		case <-d.Ctx.Done():
			io.clear(space)
			return -1, d.Ctx.Err()
		default:
			key, err := io.ReadRune()
			io.clear(space)
			if err != nil {
				if errors.Is(err, ErrUnknownRune) {
					continue
				}
				// Ctrl+C or Ctrl+D
				return -1, err
			}
			switch key {
			case keyEnter:
				return d.offset + d.selected, nil
			case '↑':
				if d.offset > 0 && d.selected == 0 { // page up
					d.offset--
					d.displayed = d.Items[d.offset : d.offset+displayed]
				} else if d.selected > 0 {
					d.selected--
				}
			case '↓':
				if d.offset+displayed < len(d.Items) { // page down
					d.offset++
					d.displayed = d.Items[d.offset : d.offset+displayed]
				} else if d.selected < displayed-1 {
					d.selected++
				}
			}
		}
	}
}
