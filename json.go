// Copyright 2026 Serge Smertin
// SPDX-License-Identifier: MIT

package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type jsonState int

const (
	jsonRoot jsonState = iota
	jsonObject
	jsonArray
	jsonQuoted
	jsonKey
	jsonConst
	jsonValue
	jsonEscape
)

var keyDepthShades = []string{ // shades of blue
	"\x1b[38;5;27m",
	"\x1b[38;5;33m",
	"\x1b[38;5;39m",
	"\x1b[38;5;45m",
	"\x1b[38;5;51m",
}

var valDepthShades = []string{ // shades of green
	"\x1b[38;5;46m",
	"\x1b[38;5;40m",
	"\x1b[38;5;34m",
	"\x1b[38;5;28m",
	"\x1b[38;5;22m",
}

// PrettyJSON pretty-prints JSON data with depth-aware coloring.
func PrettyJSON(w io.Writer, src any) error {
	var identBuf, outBuf bytes.Buffer
	var err error
	var toIndent []byte
	switch src := src.(type) {
	case string:
		toIndent = []byte(src)
	case []byte:
		toIndent = src
	default:
		toIndent, err = json.Marshal(src)
		if err != nil {
			return fmt.Errorf("marshal: %w", err)
		}
	}
	// standard library already indents JSON data, so we're only coloring things here.
	err = json.Indent(&identBuf, toIndent, "", "  ")
	if err != nil {
		return fmt.Errorf("indent: %w", err)
	}
	pretty := identBuf.Bytes()
	stack := []jsonState{jsonRoot}
	lo, hi, depth := 0, identBuf.Len(), -1
	for lo < hi {
		ch := pretty[lo]
		lo++
		curr := stack[len(stack)-1]
		if curr == jsonEscape {
			outBuf.WriteByte(ch)
			stack = stack[:len(stack)-1]

			continue
		}
		if ch != '"' && curr == jsonQuoted {
			outBuf.WriteByte(ch)

			continue
		}
		switch ch {
		case '{':
			depth++
			stack = append(stack, jsonObject)
			outBuf.WriteString(bold)
			outBuf.WriteByte(ch)
			outBuf.WriteString(reset)
		case '[':
			depth++
			stack = append(stack, jsonArray)
			outBuf.WriteString(bold)
			outBuf.WriteByte(ch)
			outBuf.WriteString(reset)
		case '}', ']':
			depth--
			stack = stack[:len(stack)-1]
			outBuf.WriteString(bold)
			outBuf.WriteByte(ch)
			outBuf.WriteString(reset)
		case ':':
			if curr == jsonObject {
				stack = append(stack, jsonValue)
			}
			outBuf.WriteByte(ch)
		case ',':
			if curr == jsonValue {
				stack = stack[:len(stack)-1]
			}
			outBuf.WriteByte(ch)
		case '\\':
			stack = append(stack, jsonEscape)
			outBuf.WriteByte(ch)
		case '"':
			if curr == jsonObject { // open key
				outBuf.WriteString(keyDepthShades[abs(depth)%len(keyDepthShades)])
				outBuf.WriteString(bold)
				outBuf.WriteByte(ch)
				stack = append(stack, jsonQuoted)
			} else if jsonQuoted == curr { // close key
				outBuf.WriteByte(ch)
				outBuf.WriteString(reset)
				stack = stack[:len(stack)-1]
			} else if jsonArray == curr { // open const
				outBuf.WriteString(valDepthShades[abs(depth)%len(valDepthShades)])
				outBuf.WriteByte(ch)
				stack = append(stack, jsonQuoted)
			} else if jsonValue == curr { // open const
				outBuf.WriteString(valDepthShades[abs(depth)%len(valDepthShades)])
				outBuf.WriteByte(ch)
				stack = stack[:len(stack)-1]
				stack = append(stack, jsonQuoted)
			} else {
				outBuf.WriteByte(ch)
			}
		default:
			outBuf.WriteByte(ch)
		}
	}
	outBuf.WriteByte('\n')
	outBuf.WriteTo(w)

	return nil
}

func abs(n int) int {
	if n < 0 {
		return -n
	}

	return n
}
