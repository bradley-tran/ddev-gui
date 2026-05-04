//go:build windows

package backend

import (
	"os/exec"
	"strings"
	"syscall"
)

// HideWindow sets SysProcAttr on the command to prevent a visible console
// window from flashing when running CLI tools from a GUI app on Windows.
func HideWindow(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.HideWindow = true
	cmd.SysProcAttr.CreationFlags |= 0x08000000
}

// HideWSLWindow sets SysProcAttr on the command to prevent a visible console
// window from appearing when running wsl.exe as a subprocess of a GUI app.
func HideWSLWindow(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.HideWindow = true
	cmd.SysProcAttr.CreationFlags |= 0x08000000
}

// ShowWindow sets SysProcAttr on the command to spawn it in a new console window.
func ShowWindow(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.CreationFlags |= 0x00000010 // CREATE_NEW_CONSOLE
}

// BuildWindowsCmd constructs an *exec.Cmd safely for cmd.exe /c execution.
// It bypasses Go's automatic argument quoting, which interacts poorly with
// cmd.exe's quote-stripping rules, by manually escaping arguments and
// using SysProcAttr.CmdLine along with the cmd.exe /S switch.
func BuildWindowsCmd(command string, args []string) *exec.Cmd {
	var b strings.Builder
	b.WriteString(`cmd.exe /S /C "`)
	b.WriteString(syscall.EscapeArg(command))
	for _, a := range args {
		b.WriteString(" ")
		b.WriteString(syscall.EscapeArg(a))
	}
	b.WriteString(`"`)

	cmd := exec.Command("cmd.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CmdLine: b.String(),
	}
	return cmd
}
