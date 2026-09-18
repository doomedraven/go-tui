// Copyright 2026 Serge Smertin
// SPDX-License-Identifier: MIT

package assert

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// formatValue recursively formats a reflect.Value with proper indentation
//
//nolint:cyclop // TODO: rewrite this code with proper diff
func formatValue(v reflect.Value, indent int) string {
	indentStr := strings.Repeat("  ", indent)
	switch v.Kind() {
	case reflect.Invalid:
		return "nil"
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%v", v.Float())
	case reflect.Complex64, reflect.Complex128:
		return fmt.Sprintf("%v", v.Complex())
	case reflect.String:
		return fmt.Sprintf("`%s`", v.String())
	case reflect.Interface, reflect.Pointer:
		if v.IsNil() {
			return "nil"
		}
		return formatValue(v.Elem(), indent)
	case reflect.Array, reflect.Slice:
		return formatSlice(v, indentStr, indent)
	case reflect.Map:
		return formatMap(v, indentStr, indent)
	case reflect.Struct:
		return formatStruct(v, indentStr, indent)
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return fmt.Sprintf("%v", v.Interface())
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}

func formatStruct(v reflect.Value, indentStr string, indent int) string {
	t := v.Type()
	if v.NumField() == 0 {
		return "{}"
	}
	var buf bytes.Buffer
	buf.WriteString("{\n")
	for i := range v.NumField() {
		if t.Field(i).PkgPath != "" {
			continue // Skip unexported fields
		}
		buf.WriteString(indentStr + "  ")
		buf.WriteString(t.Field(i).Name)
		buf.WriteString(": ")
		buf.WriteString(formatValue(v.Field(i), indent+1))
		buf.WriteString(",\n")
	}
	buf.WriteString(indentStr + "}")
	return buf.String()
}

func formatMap(v reflect.Value, indentStr string, indent int) string {
	if v.Len() == 0 {
		return "{}"
	}
	var buf bytes.Buffer
	buf.WriteString("{\n")
	// Sort keys for consistent output
	keys := v.MapKeys()
	sort.Slice(keys, func(i, j int) bool {
		return fmt.Sprintf("%v", keys[i].Interface()) < fmt.Sprintf("%v", keys[j].Interface())
	})
	for _, k := range keys {
		buf.WriteString(indentStr + "  ")
		buf.WriteString(formatValue(k, 0))
		buf.WriteString(": ")
		buf.WriteString(formatValue(v.MapIndex(k), indent+1))
		buf.WriteString(",\n")
	}
	buf.WriteString(indentStr + "}")
	return buf.String()
}

func formatSlice(v reflect.Value, indentStr string, indent int) string {
	if v.Len() == 0 {
		return "[]"
	}
	var buf bytes.Buffer
	buf.WriteString("[\n")
	for i := range v.Len() {
		buf.WriteString(indentStr + "  ")
		buf.WriteString(formatValue(v.Index(i), indent+1))
		if i < v.Len()-1 {
			buf.WriteString(",")
		}
		buf.WriteString("\n")
	}
	buf.WriteString(indentStr + "]")
	return buf.String()
}

// Compute the longest common subsequence (LCS) matrix.
func lcsMatrix(a, b []string) [][]int {
	m, n := len(a), len(b)
	lcs := make([][]int, m+1)
	for i := range lcs {
		lcs[i] = make([]int, n+1)
	}
	for i := range m {
		for j := range n {
			if a[i] == b[j] {
				lcs[i+1][j+1] = lcs[i][j] + 1
			} else {
				if lcs[i+1][j] > lcs[i][j+1] {
					lcs[i+1][j+1] = lcs[i+1][j]
				} else {
					lcs[i+1][j+1] = lcs[i][j+1]
				}
			}
		}
	}
	return lcs
}

func multiLineRepr(a any) []string {
	repr := formatValue(reflect.ValueOf(a), 0)
	return strings.Split(repr, "\n")
}

func diff(expected, actual any) string {
	a, b := multiLineRepr(expected), multiLineRepr(actual)
	m, n := len(a), len(b)
	lcs := lcsMatrix(a, b)
	var walk func(i, j int)
	var buf bytes.Buffer
	walk = func(i, j int) {
		switch {
		case i > 0 && j > 0 && a[i-1] == b[j-1]:
			walk(i-1, j-1)
			buf.WriteString("  " + a[i-1] + "\n")
		case j > 0 && (i == 0 || lcs[i][j-1] >= lcs[i-1][j]):
			walk(i, j-1)
			buf.WriteString("+ " + b[j-1] + "\n") // Added
		case i > 0 && (j == 0 || lcs[i][j-1] < lcs[i-1][j]):
			walk(i-1, j)
			buf.WriteString("- " + a[i-1] + "\n") // Removed
		}
	}
	walk(m, n)
	return buf.String()
}
