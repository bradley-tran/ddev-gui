//go:build !windows

package backend

import (
	"os/exec"
	"path/filepath"
	"testing"
)

func TestBuildWindowsCmd(t *testing.T) {
	cmd := BuildWindowsCmd("echo", []string{"hello", "world"})

	if cmd == nil {
		t.Fatal("BuildWindowsCmd returned nil")
	}

	if filepath.Base(cmd.Path) != "echo" {
		t.Errorf("Expected path base to be 'echo', got '%s'", filepath.Base(cmd.Path))
	}

	expectedArgs := []string{"echo", "hello", "world"}
	if len(cmd.Args) != len(expectedArgs) {
		t.Fatalf("Expected %d args, got %d", len(expectedArgs), len(cmd.Args))
	}

	for i, arg := range expectedArgs {
		if cmd.Args[i] != arg {
			t.Errorf("Expected arg[%d] to be '%s', got '%s'", i, arg, cmd.Args[i])
		}
	}
}

func TestHideWindow(t *testing.T) {
	cmd := exec.Command("echo")
	HideWindow(cmd)
	if cmd.SysProcAttr != nil {
		t.Error("Expected SysProcAttr to be nil for non-Windows platforms")
	}
}

func TestHideWSLWindow(t *testing.T) {
	cmd := exec.Command("echo")
	HideWSLWindow(cmd)
	if cmd.SysProcAttr != nil {
		t.Error("Expected SysProcAttr to be nil for non-Windows platforms")
	}
}

func TestShowWindow(t *testing.T) {
	cmd := exec.Command("echo")
	ShowWindow(cmd)
	if cmd.SysProcAttr != nil {
		t.Error("Expected SysProcAttr to be nil for non-Windows platforms")
	}
}
