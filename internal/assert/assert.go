// Copyright 2026 Serge Smertin
// SPDX-License-Identifier: MIT

// Package assert provides zero-dependency vendored assertions for testing.
package assert

import (
	"errors"
	"reflect"
	"runtime"
	"testing"
)

// Equal compares two values for equality.
func Equal(t *testing.T, expected, actual any) {
	t.Helper()

	expected, actual, res := deepEqual(expected, actual)

	if !res {
		_, file, line, _ := runtime.Caller(1)

		t.Fatalf("%s:%d:\nresults differ:\n%s", file, line, diff(expected, actual))
	}
}

func deepEqual(expected, actual any) (want, got any, ok bool) {
	ev, av := reflect.ValueOf(expected), reflect.ValueOf(actual)

	if ev.Type().ConvertibleTo(av.Type()) {
		expected = ev.Convert(av.Type()).Interface()
	} else if av.Type().ConvertibleTo(ev.Type()) {
		actual = av.Convert(ev.Type()).Interface()
	}

	res := reflect.DeepEqual(expected, actual)
	return expected, actual, res
}

// True asserts that the value is true.
func True(t *testing.T, value bool) { //nolint:revive // ignore
	t.Helper()

	if !value {
		_, file, line, _ := runtime.Caller(1)

		t.Fatalf("%s:%d: expected true, got false", file, line)
	}
}

// Error asserts that the error is not nil.
func Error(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		_, file, line, _ := runtime.Caller(1)

		t.Fatalf("%s:%d: expected error, got: nil", file, line)
	}
}

// ErrorIs asserts that the error is of the expected type.
func ErrorIs(t *testing.T, err, expected error) {
	t.Helper()

	if !errors.Is(err, expected) {
		_, file, line, _ := runtime.Caller(1)

		t.Fatalf("%s:%d: expected error %v, got: %v", file, line, expected, err)
	}
}

// NoError asserts that the error is nil.
func NoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		_, file, line, _ := runtime.Caller(1)

		t.Fatalf("%s:%d: expected no error, got: %v", file, line, err)
	}
}
