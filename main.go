package main

import (
	"bytes"
	"context"
	"embed"
	"io/fs"
	"log"
	"strings"
	"time"

	"ddev-gui/backend"

	"os"
	"os/exec"
	"path/filepath"
	runtime2 "runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	defaultWidth  = 1024
	defaultHeight = 760
)

//go:embed frontend/dist
var assets embed.FS

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
	backdropType := windows.Acrylic
	windowTranslucent := true
	themeVal, _ := cfg.Get("theme").(string)
	switch themeVal {
	case "tabbed":
		backdropType = windows.Tabbed
	case "default":
		backdropType = windows.None
		windowTranslucent = false
	}

	err = wails.Run(&options.App{
		Title:     "DDEV GUI",
		Width:     initWidth,
		Height:    initHeight,
		Frameless: true,
		AssetServer: &assetserver.Options{
			Assets: distFS,
		},
		OnStartup: func(ctx context.Context) {
			svc.SetContext(ctx)
			// Install native mouse back/forward handler (Linux only; no-op elsewhere)
			InstallMouseNavHandler(ctx)
			// Note: DdevService and ConfigService methods are now called
			// directly from the frontend via Go bindings (window.go.backend.*).
			// Only push-notification and fire-and-forget events remain here.

			// Open project folder in system file explorer
			runtime.EventsOn(ctx, "open:folder", func(optionalData ...any) {
				var location string
				if len(optionalData) > 0 {
					if m, ok := optionalData[0].(map[string]any); ok {
						if v, ok := m["location"].(string); ok {
							location = v
						}
					}
				}

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
					cmd := exec.CommandContext(ctx, "open", path)
					_ = cmd.Start()
				} else {
					path := location
					if strings.HasPrefix(path, "~/") {
						home, _ := os.UserHomeDir()
						path = filepath.Join(home, path[2:])
					}
					cmd := exec.CommandContext(ctx, "xdg-open", path)
					_ = cmd.Start()
				}
			})

			// Open project in the preferred editor
			openEditor := func(optionalData ...any) {
				var location string
				if len(optionalData) > 0 {
					if m, ok := optionalData[0].(map[string]any); ok {
						if v, ok := m["location"].(string); ok {
							location = v
						}
					}
				}
				backend.LaunchPreferredEditor(cfg, svc.WSLDistro(), location)
			}
			runtime.EventsOn(ctx, "open:editor", openEditor)

			// Open terminal in project directory
			runtime.EventsOn(ctx, "open:terminal", func(optionalData ...any) {
				var location string
				if len(optionalData) > 0 {
					if m, ok := optionalData[0].(map[string]any); ok {
						if v, ok := m["location"].(string); ok {
							location = v
						}
					}
				}

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
					cmd := exec.Command("osascript", "-e", script, path)
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
			runtime.EventsOn(ctx, "open:url", func(optionalData ...any) {
				var url string
				if len(optionalData) > 0 {
					if m, ok := optionalData[0].(map[string]any); ok {
						if v, ok := m["url"].(string); ok {
							url = v
						}
					}
				}
				if url != "" {
					runtime.BrowserOpenURL(ctx, url)
				}
			})
		},
		OnDomReady: func(ctx context.Context) {
			// Restore saved window position (if we have one)
			if bounds.X != 0 || bounds.Y != 0 {
				runtime.WindowSetPosition(ctx, bounds.X, bounds.Y)
			}
			// Restore maximized state
			if bounds.Maximized {
				runtime.WindowMaximise(ctx)
			}

			// Auto-open environment check dialog if ddev is not found
			info, found := svc.VersionInfo()
			if !found {
				runtime.EventsEmit(ctx, "ui:info", map[string]string{
					"title":   "Environment Info",
					"message": info,
				})
			}
		},
		OnBeforeClose: func(ctx context.Context) (prevent bool) {
			// Save window bounds right before the window closes
			saveWindowBounds(ctx, cfg)
			return false
		},
		OnShutdown: func(ctx context.Context) {
			svc.Shutdown()
		},
		Menu: nil,
		Bind: []any{svc, cfg},
		Windows: &windows.Options{
			WebviewIsTransparent:              true,
			WindowIsTranslucent:               windowTranslucent,
			BackdropType:                      backdropType,
			DisableFramelessWindowDecorations: false,
		},
		Linux: &linux.Options{
			WindowIsTranslucent: false,
			WebviewGpuPolicy:    linux.WebviewGpuPolicyOnDemand,
			ProgramName:         "ddev-gui",
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}

// saveWindowBounds captures the current window geometry and persists it to config.
func saveWindowBounds(ctx context.Context, cfg *backend.ConfigService) {
	if ctx == nil {
		return
	}

	isMax := runtime.WindowIsMaximised(ctx)

	// If maximized, un-maximize briefly to capture the "normal" geometry,
	// then re-maximize so the close animation looks correct.
	if isMax {
		runtime.WindowUnmaximise(ctx)
	}

	w, h := runtime.WindowGetSize(ctx)
	x, y := runtime.WindowGetPosition(ctx)

	if isMax {
		runtime.WindowMaximise(ctx)
	}

	cfg.SetWindowBounds(backend.WindowBounds{
		Width:     w,
		Height:    h,
		X:         x,
		Y:         y,
		Maximized: isMax,
	})
}
