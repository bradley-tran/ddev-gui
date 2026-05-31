//go:build !windows

package backend

import (
	"fmt"
	"os/exec"
)

// launchWindowsElevated is a no-op on non-Windows platforms.
func launchWindowsElevated(path string) error {
	return fmt.Errorf("elevation only supported on Windows")
}

// HideWindow is a no-op on non-Windows platforms.
func HideWindow(cmd *exec.Cmd) {}

// HideWSLWindow is a no-op on non-Windows platforms.
func HideWSLWindow(cmd *exec.Cmd) {}

// ShowWindow is a no-op on non-Windows platforms.
func ShowWindow(cmd *exec.Cmd) {}

// BuildWindowsCmd is a dummy implementation for non-Windows platforms.
// It simply creates an exec.Cmd using the provided command and arguments.
func BuildWindowsCmd(command string, args []string) *exec.Cmd {
	return exec.Command(command, args...)
}
