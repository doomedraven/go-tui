// Copyright 2026 Serge Smertin
// SPDX-License-Identifier: MIT

package tui

import (
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
)

// Browserf opens the specified URL in the default browser.
func Browserf(addr string, args ...any) error {
	var cmd *exec.Cmd
	for i := range args {
		if s, ok := args[i].(string); ok {
			args[i] = url.QueryEscape(s)
		}
	}
	addr = fmt.Sprintf(addr, args...)
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", addr)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", addr)
	case "darwin": // macOS
		cmd = exec.Command("open", addr)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}
