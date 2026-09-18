// Copyright 2026 Serge Smertin
// SPDX-License-Identifier: MIT

package tui

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"text/template"
)

type dropdown struct {
	Ctx       context.Context
	Label     string
	Items     []any
	Default   any
	Hide      bool
	OneReturn bool

	LabelTemplate        string
	labelTemplate        *template.Template
	ActiveItemTemplate   string
	activeItemTemplate   *template.Template
	InactiveItemTemplate string
	inactiveItemTemplate *template.Template
	AnswerTemplate       string
	answerTemplate       *template.Template

	ItemsFn func(prefix string) []any
	In      io.Reader
	Out     io.Writer

	selected int
	io       *termIO
	isView   bool

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
		err = d.answerTemplate.Execute(d.io, dropdownAnswer{
			Label:  label,
			Answer: items[i],
		})
		if err != nil {
			return i, fmt.Errorf("answer: %w", err)
		}
		d.io.Write([]byte{'\n'})
	}
	return i, nil
}

var DefaultDropdownLabelTemplate = `{{ "?" | green }} {{ . | bold }} `
var DefaultDropdownActiveItemTemplate = `{{ "→" | cyan }} {{ . | cyan }}`
var DefaultDropdownInactiveItemTemplate = `{{ "→" | dim }} {{ . | dim }}`
var DefaultDropdownAnswerTemplate = `{{ "✔" | dim }} {{ .Label | dim }} {{ "…" | dim }} {{ .Answer | bold }}`

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
		Ctx:         context.Background(),
		Label:       "Select from list",
		makeRawTerm: makeRawTerm,
		io: &termIO{
			ReadWriter: defaultIO,
		},
		LabelTemplate:        DefaultDropdownLabelTemplate,
		ActiveItemTemplate:   DefaultDropdownActiveItemTemplate,
		InactiveItemTemplate: DefaultDropdownInactiveItemTemplate,
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
	d.answerTemplate, err = tmpl.New("answer").Parse(d.AnswerTemplate)
	if err != nil {
		return fmt.Errorf("answer: %w", err)
	}
	return nil
}

func (d *dropdown) getTIO() *tio {
	io, ok := d.io.ReadWriter.(*tio)
	if !ok {
		return nil
	}
	return io
}

// implements [withWriter]
func (d *dropdown) setWriter(w io.Writer) {
	tui, ok := w.(*Tui)
	if ok {
		d.isView = true
		c := tui.prependView()
		c.height = len(d.Items) + 1
		c.next.height -= c.height // TODO: propagate down
		w = c
	}
	tio, ok := d.io.ReadWriter.(*tio)
	if !ok {
		return
	}
	tio.Writer = w
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
func (d *dropdown) render() error {
	// use buffer to write to io only once
	var buf bytes.Buffer
	var prefix int
	var err error
	for i, item := range d.Items {
		fmt.Fprint(&buf, "\r") // ensure we start from the leftmost position
		if i == 0 {
			var labelBuf bytes.Buffer
			err = d.labelTemplate.Execute(&labelBuf, d.Label)
			if err != nil {
				return fmt.Errorf("label: %w", err)
			}
			prefix = width(labelBuf.Bytes())
			labelBuf.WriteTo(&buf)
		} else {
			for range prefix {
				fmt.Fprintf(&buf, " ")
			}
		}
		// TODO: print spaces till the end of the terminal width except for the last line
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
	fmt.Fprint(&buf, "\r")
	_, err = buf.WriteTo(d.io)
	return err
}

// Show displays the dropdown and handles user input
func (d *dropdown) run() (int, error) {
	restore, err := d.makeRawTerm()
	if err != nil {
		return -1, fmt.Errorf("raw term: %w", err)
	}
	defer restore()
	for {
		err = d.render()
		if err != nil {
			return -1, fmt.Errorf("render: %w", err)
		}
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
				}
			case '↓':
				if d.selected < len(d.Items)-1 {
					d.selected++
					d.io.clear(space)
				}
			}
		}
	}
}
