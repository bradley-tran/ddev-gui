package main

import (
	"bytes"
	"context"
	"embed"
	"io/fs"
	"log"
	"strings"
	"sync/atomic"
	"time"

	"ddev-gui/backend"

	"os"
	"os/exec"
	"path/filepath"
	runtime2 "runtime"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

const (
	defaultWidth  = 1024
	defaultHeight = 760
)

//go:embed frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var appIcon []byte

//go:embed build/windows/icon.ico
var appIconICO []byte

func main() {
	cfg := backend.NewConfigService()
	svc := backend.NewDdevService(cfg)

	distFS, err := fs.Sub(assets, "frontend/dist")
	if err != nil {
		log.Fatal(err)
	}

	// Load saved window bounds (falls back to defaults)
	bounds := cfg.GetWindowBounds()
	initWidth := bounds.Width
	initHeight := bounds.Height
	if initWidth <= 0 {
		initWidth = defaultWidth
	}
	if initHeight <= 0 {
		initHeight = defaultHeight
	}

	// Determine Windows backdrop type from saved theme config
	backdropType := application.Acrylic
	windowTranslucent := true
	bgType := application.BackgroundTypeTranslucent
	themeVal, _ := cfg.Get("theme").(string)
	switch themeVal {
	case "tabbed":
		backdropType = application.Tabbed
	case "default":
		backdropType = application.None
		windowTranslucent = false
		bgType = application.BackgroundTypeSolid
	}

	iconData := appIcon
	if runtime2.GOOS == "windows" {
		iconData = appIconICO
	}

	app := application.New(application.Options{
		Name: "DDEV GUI",
		Icon: iconData,
		Services: []application.Service{
			application.NewService(svc),
			application.NewService(cfg),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(distFS),
		},
		Linux: application.LinuxOptions{
			ProgramName: "ddev-gui",
		},
	})

	mainWindow := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:          "DDEV GUI",
		Width:          initWidth,
		Height:         initHeight,
		Frameless:      true,
		URL:            "/",
		BackgroundType: bgType,
		Windows: application.WindowsWindow{
			BackdropType:                      backdropType,
			DisableFramelessWindowDecorations: false,
		},
		Linux: application.LinuxWindow{
			WindowIsTranslucent: windowTranslucent && runtime2.GOOS != "linux",
			WebviewGpuPolicy:    application.WebviewGpuPolicyOnDemand,
		},
	})

	// Install native mouse back/forward handler (Linux only; no-op elsewhere)
	InstallMouseNavHandler(context.Background())
	// Note: DdevService and ConfigService methods are now called
	// directly from the frontend via Go bindings (window.go.backend.*).
	// Only push-notification and fire-and-forget events remain here.

	// Open project folder in system file explorer
	app.Event.On("open:folder", func(e *application.CustomEvent) {
		location := extractEventString(e.Data, "location")

		// Platform-specific open
		if runtime2.GOOS == "windows" {
			distro := svc.WSLDistro()
			// Convert Linux path to UNC: \\wsl.localhost\<distro>\<linux-path>
			// location is a Linux path like /home/user/projects/foo or ~/projects/foo
			linuxPath := location
			if strings.HasPrefix(linuxPath, "~/") {
				// Resolve ~ to actual home dir via whoami
				resolveCtx, resolveCancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer resolveCancel()
				whoamiCmd := exec.CommandContext(resolveCtx, "wsl.exe", "-d", distro, "-e", "whoami")
				backend.HideWindow(whoamiCmd)
				var whoamiOut bytes.Buffer
				whoamiCmd.Stdout = &whoamiOut
				user := "root"
				if err := whoamiCmd.Run(); err == nil {
					if u := strings.TrimSpace(whoamiOut.String()); u != "" {
						user = u
					}
				}
				if user == "root" {
					linuxPath = "/root/" + linuxPath[2:]
				} else {
					linuxPath = "/home/" + user + "/" + linuxPath[2:]
				}
			}
			// Convert forward slashes to backslashes for UNC
			winPath := strings.ReplaceAll(linuxPath, "/", `\`)
			uncPath := `\\wsl.localhost\` + distro + winPath
			cmd := exec.Command("explorer.exe", uncPath)
			_ = cmd.Start()
		} else if runtime2.GOOS == "darwin" {
			path := location
			if strings.HasPrefix(path, "~/") {
				home, _ := os.UserHomeDir()
				path = filepath.Join(home, path[2:])
			}
			cmd := exec.CommandContext(context.Background(), "open", path)
			_ = cmd.Start()
		} else {
			path := location
			if strings.HasPrefix(path, "~/") {
				home, _ := os.UserHomeDir()
				path = filepath.Join(home, path[2:])
			}
			cmd := exec.CommandContext(context.Background(), "xdg-open", path)
			_ = cmd.Start()
		}
	})

	// Open project in the preferred editor
	openEditor := func(e *application.CustomEvent) {
		location := extractEventString(e.Data, "location")
		backend.LaunchPreferredEditor(cfg, svc.WSLDistro(), location)
	}
	app.Event.On("open:editor", openEditor)

	// Open terminal in project directory
	app.Event.On("open:terminal", func(e *application.CustomEvent) {
		location := extractEventString(e.Data, "location")

		if runtime2.GOOS == "windows" {
			distro := svc.WSLDistro()
			// WSL --cd supports ~ paths natively, so pass location as-is
			// Try Windows Terminal first: wt.exe wsl.exe -d <distro> --cd <dir>
			wtCmd := exec.Command("wt.exe", "wsl.exe", "-d", distro, "--cd", location)
			if err := wtCmd.Start(); err != nil {
				// Fallback: open wsl.exe directly in a new cmd window
				cmd := exec.Command("wsl.exe", "-d", distro, "--cd", location)
				backend.ShowWindow(cmd)
				_ = cmd.Start()
			}
		} else if runtime2.GOOS == "darwin" {
			path := location
			if strings.HasPrefix(path, "~/") {
				home, _ := os.UserHomeDir()
				path = filepath.Join(home, path[2:])
			}
			// Pass the path as an argument to the script to avoid injection
			script := "on run argv\ntell application \"Terminal\" to do script \"cd \" & quoted form of item 1 of argv\nend run"
			cmd := exec.Command("osascript", "-e", script, "--", path)
			_ = cmd.Start()
		} else {
			path := location
			if strings.HasPrefix(path, "~/") {
				home, _ := os.UserHomeDir()
				path = filepath.Join(home, path[2:])
			}
			// Try common terminal emulators in order
			terminals := []struct {
				cmd  string
				args []string
			}{
				{"x-terminal-emulator", []string{"--working-directory=" + path}},
				{"konsole", []string{"--workdir", path}},
				{"gnome-terminal", []string{"--working-directory=" + path}},
				{"xfce4-terminal", []string{"--working-directory=" + path}},
				{"mate-terminal", []string{"--working-directory=" + path}},
				{"kitty", []string{"--directory", path}},
				{"alacritty", []string{"--working-directory", path}},
				{"foot", []string{"--working-directory=" + path}},
				{"wezterm", []string{"start", "--cwd", path}},
				{"xterm", []string{"-e", "cd '" + strings.ReplaceAll(path, "'", "'\\''") + "' && $SHELL"}},
			}
			for _, t := range terminals {
				cmd := exec.Command(t.cmd, t.args...)
				if err := cmd.Start(); err == nil {
					break
				}
			}
		}
	})

	// Open URL in system-installed browser
	app.Event.On("open:url", func(e *application.CustomEvent) {
		url := extractEventString(e.Data, "url")
		if url != "" {
			_ = app.Browser.OpenURL(url)
		}
	})

	applyWinIcon := func(_ *application.WindowEvent) {
		backend.ApplyWindowIcon(uintptr(mainWindow.NativeWindow()), iconData)
	}
	mainWindow.OnWindowEvent(events.Windows.WindowShow, applyWinIcon)
	mainWindow.OnWindowEvent(events.Windows.WindowDPIChanged, applyWinIcon)

	mainWindow.OnWindowEvent(events.Common.WindowRuntimeReady, func(e *application.WindowEvent) {
		backend.ApplyWindowIcon(uintptr(mainWindow.NativeWindow()), iconData)

		// Restore saved window position (if we have one)
		if bounds.X != 0 || bounds.Y != 0 {
			mainWindow.SetPosition(bounds.X, bounds.Y)
		}
		// Restore maximized state
		if bounds.Maximized {
			mainWindow.Maximise()
		}

		// Auto-open environment check dialog if ddev is not found
		go func() {
			info, found := svc.VersionInfo()
			if !found {
				app.Event.Emit("ui:info", map[string]string{
					"title":   "Environment Info",
					"message": info,
				})
			}
		}()
	})

	var isQuitting atomic.Bool

	showMainWindow := func() {
		if mainWindow.IsMinimised() {
			mainWindow.UnMinimise()
		}
		mainWindow.Show()
		mainWindow.Focus()
	}

	trayMenu := app.NewMenu()
	trayMenu.Add("Show DDEV GUI").OnClick(func(ctx *application.Context) {
		showMainWindow()
	})
	trayMenu.AddSeparator()
	trayMenu.Add("Quit").OnClick(func(ctx *application.Context) {
		isQuitting.Store(true)
		if mainWindow.IsVisible() {
			saveWindowBounds(mainWindow, cfg)
		}
		app.Quit()
	})

	tray := app.SystemTray.New()
	tray.SetIcon(iconData)
	tray.SetTooltip("DDEV GUI")
	if runtime2.GOOS == "linux" {
		tray.SetLabel("DDEV GUI")
	}
	tray.SetMenu(trayMenu)
	tray.OnClick(func() {
		if mainWindow.IsVisible() && !mainWindow.IsMinimised() {
			saveWindowBounds(mainWindow, cfg)
			mainWindow.Hide()
		} else {
			showMainWindow()
		}
	})

	mainWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		if mainWindow.IsVisible() {
			// Save window bounds right before the window closes
			saveWindowBounds(mainWindow, cfg)
		}
		if !isQuitting.Load() && cfg.GetCloseToTray() {
			e.Cancel()
			mainWindow.Hide()
		}
	})

	err = app.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func extractEventString(data any, key string) string {
	switch v := data.(type) {
	case map[string]any:
		if s, ok := v[key].(string); ok {
			return s
		}
	case []any:
		if len(v) > 0 {
			if m, ok := v[0].(map[string]any); ok {
				if s, ok := m[key].(string); ok {
					return s
				}
			}
		}
	}
	return ""
}

// saveWindowBounds captures the current window geometry and persists it to config.
func saveWindowBounds(win application.Window, cfg *backend.ConfigService) {
	if win == nil {
		return
	}

	isMax := win.IsMaximised()

	// If maximized, un-maximize briefly to capture the "normal" geometry,
	// then re-maximize so the close animation looks correct.
	if isMax {
		win.UnMaximise()
	}

	w, h := win.Size()
	x, y := win.Position()

	if isMax {
		win.Maximise()
	}

	cfg.SetWindowBounds(backend.WindowBounds{
		Width:     w,
		Height:    h,
		X:         x,
		Y:         y,
		Maximized: isMax,
	})
}
