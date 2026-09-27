// Copyright 2026 Serge Smertin
// SPDX-License-Identifier: MIT

package tui

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/nfx/go-tui/internal/assert"
)

func TestProgressbarTickRenders(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)

	cio := &chanIO{
		ctx: ctx,
		In:  make(chan string),
		Out: make(chan string, 4),
	}
	ticks := make(chan time.Time)
	start := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	now := start

	p, err := newStartedProgressBar("download", 20,
		WithInput(cio),
		WithOutput(cio),
		progressbarOpt(func(pb *Progressbar) error {
			pb.now = func() time.Time { return now }
			pb.redrawAt = start
			pb.ticks = ticks
			pb.ticker = time.NewTicker(time.Hour)
			pb.makeTermIO = func(in io.Reader, out io.Writer) (*termIO, error) {
				return &termIO{
					in:      in,
					out:     out,
					Width:   40,
					Height:  1,
					Restore: func() error { return nil },
				}, nil
			}

			return nil
		}),
	)
	assert.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, p.Close())
	})

	p.Add(10)
	now = start.Add(time.Second)

	select {
	case ticks <- now:
	case <-time.After(time.Second):
		t.Fatalf("tick not delivered")
	}

	var output string
	select {
	case output = <-cio.Out:
	case <-time.After(time.Second):
		t.Fatalf("no progress output")
	}

	assert.Contains(t, output, "\x1b[1A\r\x1b[K\r")
	assert.Contains(t, output, "download 50% [>] (10.00/s, 1s remaining)")
}

func TestNewMaxProgressBar(t *testing.T) {
	p, err := NewMaxProgressBar("max", 3)
	assert.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, p.Close()) })
	assert.NotNil(t, p)
	assert.Equal(t, int64(3), p.maxNum)

	// Add should be a no-op without a TTY.
	p.Add(3)
}

type fakeInfo struct {
	size int64
}

func (f fakeInfo) Name() string       { return "fake" }
func (f fakeInfo) Size() int64        { return f.size }
func (f fakeInfo) Mode() os.FileMode  { return 0 }
func (f fakeInfo) ModTime() time.Time { return time.Time{} }
func (f fakeInfo) IsDir() bool        { return false }
func (f fakeInfo) Sys() any           { return nil }

type statReader struct {
	data []byte
	pos  int
}

func (r *statReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n

	return n, nil
}

func (r *statReader) Stat() (os.FileInfo, error) {
	return fakeInfo{size: int64(len(r.data))}, nil
}

func TestNewFileProgressReader(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)

	cio := &chanIO{
		ctx: ctx,
		In:  make(chan string),
		Out: make(chan string, 4),
	}
	ticks := make(chan time.Time, 1)
	applied := false

	r, err := NewFileProgressReader(&statReader{data: []byte("hello world")}, "file",
		WithInput(cio),
		WithOutput(cio),
		progressbarOpt(func(pb *Progressbar) error {
			applied = true
			pb.ticks = ticks
			pb.ticker = time.NewTicker(time.Hour)
			pb.increments = make(chan int64, 8)
			pb.now = func() time.Time { return time.Now() }
			pb.redrawAt = pb.now()
			pb.makeTermIO = func(in io.Reader, out io.Writer) (*termIO, error) {
				return &termIO{
					in:      in,
					out:     out,
					Width:   30,
					Height:  1,
					Restore: func() error { return nil },
				}, nil
			}

			return nil
		}),
	)
	assert.NoError(t, err)
	assert.True(t, applied)
	assert.NotNil(t, r.p.io)
	t.Cleanup(func() { assert.NoError(t, r.Close()) })

	_, err = io.ReadAll(r)
	assert.NoError(t, err)
	for {
		select {
		case n := <-r.p.increments:
			r.p.currentNum += n
		default:
			goto drained
		}
	}
drained:
	assert.Equal(t, int64(11), r.p.currentNum)
	assert.Equal(t, int64(11), r.p.maxNum)
}

type sizedReader struct {
	size int64
}

func (r *sizedReader) Read(p []byte) (int, error) {
	return 0, io.EOF
}

func (r *sizedReader) Size() int64 {
	return r.size
}

func TestWrapReaderSizeFallback(t *testing.T) {
	r := &wrapReader{r: &sizedReader{size: 42}}

	size, err := r.Size()
	assert.NoError(t, err)
	assert.Equal(t, int64(42), size)
}

func TestWrapReaderSizeUnknown(t *testing.T) {
	_, err := NewFileProgressReader(bytes.NewBufferString("no size"), "file")
	assert.ErrorIs(t, err, errNoSize)
}

func TestNewSliceProgressBar(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)

	cio := &chanIO{
		ctx: ctx,
		In:  make(chan string),
		Out: make(chan string, 2),
	}
	ticks := make(chan time.Time, 1)
	var seen *Progressbar

	items := []int{1, 2, 3}
	seq := NewSliceProgressBar("items", items,
		WithInput(cio),
		WithOutput(cio),
		progressbarOpt(func(pb *Progressbar) error {
			seen = pb
			pb.ticks = ticks
			pb.ticker = time.NewTicker(time.Hour)
			pb.increments = make(chan int64, 8)
			pb.now = func() time.Time { return time.Now() }
			pb.redrawAt = pb.now()
			pb.makeTermIO = func(in io.Reader, out io.Writer) (*termIO, error) {
				return &termIO{
					in:      in,
					out:     out,
					Width:   30,
					Height:  1,
					Restore: func() error { return nil },
				}, nil
			}

			return nil
		}),
	)

	var got []int
	for v, err := range seq {
		assert.NoError(t, err)
		got = append(got, v)
	}

	for {
		select {
		case n := <-seen.increments:
			seen.currentNum += n
		default:
			goto drainedSlice
		}
	}
drainedSlice:
	assert.Equal(t, items, got)
	assert.NotNil(t, seen)
	assert.Equal(t, int64(len(items)), seen.currentNum)
	assert.True(t, seen.isDone())
}

func TestProgressStateHelpers(t *testing.T) {
	ps := &progressState{
		maxNum:     100,
		currentNum: 60,
	}

	assert.Equal(t, "4s remaining", ps.remainingTime(10))
	assert.Equal(t, "", ps.remainingTime(0))
	assert.Equal(t, "[==========]", ps.filledBarLine(10, 1.2))
}
