// Copyright 2026 Serge Smertin
// SPDX-License-Identifier: MIT

package tui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nfx/go-tui/internal/assert"
)

func TestFilePickerListFilters(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "keep.txt"), []byte("ok"), 0o644)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(dir, "drop.log"), []byte("noop"), 0o644)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(dir, ".hidden.txt"), []byte("hidden"), 0o644)
	assert.NoError(t, err)
	err = os.Mkdir(filepath.Join(dir, "nested"), 0o755)
	assert.NoError(t, err)

	fp := &filePicker{
		extensions: map[string]bool{".txt": true},
		start:      dir,
	}

	entries, err := fp.list(dir)
	assert.NoError(t, err)

	names := map[string]bool{}
	for _, e := range entries {
		names[e.Name()] = true
	}
	assert.Equal(t, map[string]bool{
		"keep.txt": true,
		"nested":   true,
	}, names)
}

func TestFilePickerListsUpEntry(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "child")
	err := os.Mkdir(child, 0o755)
	assert.NoError(t, err)

	fp := &filePicker{
		extensions: map[string]bool{},
		start:      root,
	}

	entries, err := fp.list(child)
	assert.NoError(t, err)
	assert.True(t, len(entries) > 0)
	assert.Equal(t, "..", entries[0].Name())
}

func TestFilePickerSkipRules(t *testing.T) {
	fp := &filePicker{
		extensions: map[string]bool{},
		showHidden: false,
		ignoreDirs: true,
	}

	assert.True(t, fp.skip(&dirEntry{name: ".hidden"}))
	assert.True(t, fp.skip(&dirEntry{name: "dir", isDir: true}))

	fp.extensions[".txt"] = true
	assert.True(t, fp.skip(&dirEntry{name: "note.md"}))
	assert.True(t, !fp.skip(&dirEntry{name: "keep.txt"}))
}

func TestDirEntryString(t *testing.T) {
	assert.Equal(t, "..", (&dirEntry{name: "..", isDir: true}).String())
	assert.Equal(t, "d folder"+string(os.PathSeparator), (&dirEntry{name: "folder", isDir: true}).String())
	assert.Equal(t, "- file"+string(os.PathSeparator), (&dirEntry{name: "file", isDir: false}).String())
}
