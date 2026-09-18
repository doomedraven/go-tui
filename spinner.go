// Copyright 2026 Serge Smertin
// SPDX-License-Identifier: MIT

package tui

import (
	"fmt"
	"sync"
	"time"
)

type Spinner struct {
	mu       sync.Mutex
	line     int
	message  string
	frames   []string
	updates  chan string
	stopChan chan struct{}
}

type SpinnerManager struct {
	mu       sync.Mutex
	spinners []*Spinner
	height   int
}

func NewSpinnerManager() *SpinnerManager {
	return &SpinnerManager{
		spinners: make([]*Spinner, 0),
	}
}

func (sm *SpinnerManager) AddSpinner(updates chan string) *Spinner {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	s := &Spinner{
		line:     sm.height,
		frames:   []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		updates:  updates,
		stopChan: make(chan struct{}),
	}

	sm.spinners = append(sm.spinners, s)
	sm.height++

	go s.spin(sm)
	go s.handleUpdates(sm)

	return s
}

func (s *Spinner) handleUpdates(sm *SpinnerManager) {
	for {
		select {
		case msg, ok := <-s.updates:
			if !ok {
				s.stop(sm)
				return
			}
			s.mu.Lock()
			s.message = msg
			s.mu.Unlock()
		case <-s.stopChan:
			return
		}
	}
}

func (s *Spinner) stop(sm *SpinnerManager) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	close(s.stopChan)

	// Remove spinner from the manager
	for i, spinner := range sm.spinners {
		if spinner == s {
			sm.spinners = append(sm.spinners[:i], sm.spinners[i+1:]...)
			break
		}
	}

	// Update lines for remaining spinners
	for i, spinner := range sm.spinners {
		spinner.line = i
	}
	sm.height--

	// Clear the spinner's line
	fmt.Printf("\033[%dH\033[K", s.line+1)

	// Move remaining spinners up
	for _, spinner := range sm.spinners {
		if spinner.line > s.line {
			fmt.Printf("\033[%dH%s %s", spinner.line+1, spinner.frames[0], spinner.message)
		}
	}
}

func (s *Spinner) spin(sm *SpinnerManager) {
	ticker := time.NewTicker(100 * time.Millisecond)
	frameIndex := 0

	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			frame := s.frames[frameIndex]
			message := s.message
			line := s.line
			s.mu.Unlock()

			// Move cursor to line and clear it
			fmt.Printf("\033[%dH\033[K%s %s", line+1, frame, message)

			frameIndex = (frameIndex + 1) % len(s.frames)
		case <-s.stopChan:
			ticker.Stop()
			return
		}
	}
}
