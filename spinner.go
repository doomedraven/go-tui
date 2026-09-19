// Copyright 2026 Serge Smertin
// SPDX-License-Identifier: MIT

package tui

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"
)

// These two styles are taken from cli-spinners (MIT License)
// See https://github.com/sindresorhus/cli-spinners/blob/main/spinners.json for more spinner styles
var DefaultSpinnerStyle = []string{"⠉⠉", "⠈⠙", "⠀⠹", "⠀⢸", "⠀⣰", "⢀⣠", "⣀⣀", "⣄⡀", "⣆⠀", "⡇⠀", "⠏⠀", "⠋⠁"}
var SpinnerStyleDocs = []string{".  ", ".. ", "...", " ..", "  .", "   "}

type Spinners struct {
	config
	cancel context.CancelFunc
	io     *termIO

	creates chan createSpinner
	updates chan updateSpinner
	stops   chan int
	ticker  *time.Ticker
	ticks   <-chan time.Time

	state      []*spinnerState
	active     int
	makeTermIO func(io.Reader, io.Writer) (*termIO, error)
}

func spinnersOpt(o func(s *Spinners) error) opt {
	return func(a any) error {
		s, ok := a.(*Spinners)
		if !ok {
			return nil
		}
		return o(s)
	}
}

func newSpinners() *Spinners {
	ctx, cancel := context.WithCancel(context.Background())
	ticker := time.NewTicker(100 * time.Millisecond)
	return &Spinners{
		config: config{
			ctx: ctx,
			in:  os.Stdin,
			out: os.Stdout,
		},
		ticker:     ticker,
		ticks:      ticker.C,
		cancel:     cancel,
		makeTermIO: makeTermIO,
		creates:    make(chan createSpinner),
		updates:    make(chan updateSpinner),
		stops:      make(chan int),
	}
}

func NewSpinners(opt ...opt) (*Spinners, error) {
	s := newSpinners()
	err := opts(opt).Apply(s)
	if err != nil {
		return nil, err
	}
	s.io, err = s.makeTermIO(s.in, s.out)
	if err != nil {
		return nil, err
	}
	s.io.Restore() // todo: hack, fix this
	go s.start(s.ctx)
	return s, nil
}

func (s *Spinners) setContext(ctx context.Context) {
	s.ctx, s.cancel = context.WithCancel(ctx)
}

func (s *Spinners) start(ctx context.Context) {
	defer s.stop()
	frame := bytes.NewBuffer(make([]byte, 2*s.io.Width))
	frame.Reset()
	var prevActive, currActive int
	for {
		select {
		case <-ctx.Done():
			for _, spinner := range s.state {
				if spinner == nil {
					continue
				}
				spinner.cancel()
			}
			return
		case ns := <-s.creates:
			// TODO: write serially in CI mode, as well as when number of spinners
			// is greater than the height of the terminal
			s.newSpinner(ns)
		case update := <-s.updates:
			if s.state[update.offset] == nil {
				continue // it's already stopped and we don't care
			}
			s.state[update.offset].Message = update.message
		case offset := <-s.stops:
			if offset >= 0 && offset < len(s.state) { // remove spinner at offset
				s.state[offset] = nil // TODO: add concept of "done" spinners, that are still snown
				s.active--
			}
		case <-s.ticks:
			if prevActive > 0 {
				s.io.clear(prevActive, frame)
			}
			currActive = 0
			for _, spinner := range s.state {
				if spinner == nil {
					continue
				}
				spinner.tick = (spinner.tick + 1) % len(spinner.frames)
				frame.WriteByte('\r')
				frame.WriteString(spinner.frames[spinner.tick])
				frame.WriteString(" ")
				frame.WriteString(spinner.Message)
				frame.WriteByte('\n')
				frame.WriteByte('\r')
				currActive++
			}
			prevActive = currActive
			frame.WriteTo(s.io)
		}
	}
}

func (s *Spinners) Close() {
	s.cancel()
}

func (s *Spinners) stop() {
	s.io.clear(s.active, s.io)
	// s.io.Restore()
	s.ticker.Stop()
	close(s.creates)
	close(s.updates)
	close(s.stops)
}

type createSpinner struct {
	ctx         context.Context
	cancel      context.CancelFunc
	frames      []string
	replyOffset chan int
}

func (s *Spinners) newSpinner(ns createSpinner) {
	offset := len(s.state)
	s.state = append(s.state, &spinnerState{
		ctx:    ns.ctx,
		cancel: ns.cancel,
		tick:   (len(s.state) + 1) % len(ns.frames),
		frames: ns.frames,
		active: true,
	})
	s.active++
	select {
	case <-s.ctx.Done():
		return
	case <-ns.ctx.Done():
		return
	case ns.replyOffset <- offset:
	}
}

type spinnerState struct {
	ctx     context.Context
	cancel  context.CancelFunc
	tick    int
	active  bool
	frames  []string
	Message string
}

func (s *Spinners) MustAddBackground() *Spinner {
	spinner, err := s.Add(context.Background())
	if err != nil {
		panic(err)
	}
	return spinner
}

func (s *Spinners) Add(ctx context.Context) (*Spinner, error) {
	// rewrap the context, so that we can cancel the spinner when
	// we don't want to cancel the parent context.
	ctx, cancel := context.WithCancel(ctx)
	replyOffset := make(chan int)
	defer close(replyOffset)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case s.creates <- createSpinner{
		ctx:         ctx,
		cancel:      cancel,
		frames:      SpinnerStyleDocs,
		replyOffset: replyOffset,
	}:
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case offset := <-replyOffset:
			// go close in background
			spinner := &Spinner{
				ctx:    ctx,
				parent: s,
				offset: offset,
			}
			go spinner.monitor()
			return spinner, nil
		}
	}
}

type Spinner struct {
	ctx    context.Context
	parent *Spinners
	offset int
}

func (s *Spinner) monitor() {
	defer s.Close()
	for {
		select {
		case <-s.parent.ctx.Done():
			return
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *Spinner) Close() error {
	for {
		select {
		case <-s.parent.ctx.Done():
			return s.parent.ctx.Err()
		case <-s.ctx.Done():
			return s.ctx.Err()
		case s.parent.stops <- s.offset:
			return nil
		}
	}
}

type updateSpinner struct {
	offset  int
	message string
}

func (s *Spinner) Update(message string) {
	for {
		select {
		case <-s.parent.ctx.Done():
			return
		case <-s.ctx.Done():
			return
		case s.parent.updates <- updateSpinner{
			offset:  s.offset,
			message: message,
		}: // ok
			return
		}
	}
}

func (s *Spinner) Updatef(format string, args ...interface{}) {
	s.Update(fmt.Sprintf(format, args...))
}
