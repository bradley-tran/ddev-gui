//go:build windows

package backend

import (
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

var (
	user32DLL                    = syscall.NewLazyDLL("user32.dll")
	procSendMessageW             = user32DLL.NewProc("SendMessageW")
	procSetClassLongPtrW         = user32DLL.NewProc("SetClassLongPtrW")
	procCreateIconFromResourceEx = user32DLL.NewProc("CreateIconFromResourceEx")
	procDestroyIcon              = user32DLL.NewProc("DestroyIcon")
	procGetDpiForWindow          = user32DLL.NewProc("GetDpiForWindow")
	procGetSystemMetrics         = user32DLL.NewProc("GetSystemMetrics")
	procGetSystemMetricsForDpi   = user32DLL.NewProc("GetSystemMetricsForDpi")

	windowIconsMu   sync.Mutex
	activeBigIcon   uintptr
	activeSmallIcon uintptr
)

const (
	wmSetIcon  = 0x0080
	iconSmall  = 0
	iconBig    = 1
	smCxIcon   = 11
	smCxSmIcon = 49
)

var (
	gclpHIcon   = -14
	gclpHIconSm = -34
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

// ApplyWindowIcon sets both the large (ICON_BIG / GCLP_HICON) and small
// (ICON_SMALL / GCLP_HICONSM) icons on a Win32 window handle using the exact
// DPI-scaled frames from a multi-size .ico byte slice.
//
// Wails v3 only sends WM_SETICON(ICON_BIG) and registers its window class with
// IDI_APPLICATION (resource ID 32512, which is not present in wails3 syso),
// leaving ICON_SMALL, GCLP_HICON, and GCLP_HICONSM unset (NULL). That causes
// Windows Taskbar live preview headers, Alt+Tab, Task View, and Task Manager
// to fall back to the generic placeholder icon.
func ApplyWindowIcon(hwnd uintptr, icoData []byte) {
	if hwnd == 0 || len(icoData) == 0 {
		return
	}

	dpi := getWindowDPI(hwnd)
	bigSize := getIconMetricForDPI(smCxIcon, dpi, 32)
	smallSize := getIconMetricForDPI(smCxSmIcon, dpi, 16)

	bigIcon := createHIconFromICO(icoData, bigSize)
	smallIcon := createHIconFromICO(icoData, smallSize)
	if bigIcon == 0 && smallIcon == 0 {
		return
	}

	windowIconsMu.Lock()
	oldBig := activeBigIcon
	oldSmall := activeSmallIcon
	if bigIcon != 0 {
		activeBigIcon = bigIcon
	}
	if smallIcon != 0 {
		activeSmallIcon = smallIcon
	}
	windowIconsMu.Unlock()

	if bigIcon != 0 {
		_, _, _ = procSendMessageW.Call(hwnd, wmSetIcon, iconBig, bigIcon)
		_, _, _ = procSetClassLongPtrW.Call(hwnd, uintptr(gclpHIcon), bigIcon)
	}
	if smallIcon != 0 {
		_, _, _ = procSendMessageW.Call(hwnd, wmSetIcon, iconSmall, smallIcon)
		_, _, _ = procSetClassLongPtrW.Call(hwnd, uintptr(gclpHIconSm), smallIcon)
	}

	if oldBig != 0 && oldBig != bigIcon {
		_, _, _ = procDestroyIcon.Call(oldBig)
	}
	if oldSmall != 0 && oldSmall != smallIcon {
		_, _, _ = procDestroyIcon.Call(oldSmall)
	}
}

func getWindowDPI(hwnd uintptr) uint32 {
	if err := procGetDpiForWindow.Find(); err == nil {
		ret, _, _ := procGetDpiForWindow.Call(hwnd)
		if ret > 0 {
			return uint32(ret)
		}
	}
	return 96
}

func getIconMetricForDPI(index int, dpi uint32, fallback int) int {
	if dpi > 0 {
		if err := procGetSystemMetricsForDpi.Find(); err == nil {
			ret, _, _ := procGetSystemMetricsForDpi.Call(uintptr(index), uintptr(dpi))
			if int(ret) > 0 {
				return int(ret)
			}
		}
	}
	if err := procGetSystemMetrics.Find(); err == nil {
		ret, _, _ := procGetSystemMetrics.Call(uintptr(index))
		if int(ret) > 0 {
			return int(ret)
		}
	}
	return fallback
}

func createHIconFromICO(icoData []byte, targetSize int) uintptr {
	frame, err := SelectICOFrame(icoData, targetSize)
	if err != nil || len(frame) == 0 {
		return 0
	}
	ret, _, _ := procCreateIconFromResourceEx.Call(
		uintptr(unsafe.Pointer(&frame[0])),
		uintptr(len(frame)),
		1,
		0x00030000,
		uintptr(targetSize),
		uintptr(targetSize),
		0,
	)
	return ret
}
