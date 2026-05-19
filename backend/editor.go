package backend

import (
	"bytes"
	"context"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	runtime2 "runtime"
)

type editorLaunchCommand struct {
	command string
	args    []string
	shell   bool
}

var execCommandContext = exec.CommandContext

func LaunchPreferredEditor(cfg *ConfigService, distro, location string) {
	editor := "vscode"
	if cfg != nil {
		if value, ok := cfg.Get("preferredEditor").(string); ok && strings.TrimSpace(value) != "" {
			editor = strings.ToLower(strings.TrimSpace(value))
		}
	}

	var started bool
	switch editor {
	case "vscode":
		started = launchVSCodeEditor(location, distro)
	case "phpstorm":
		started = launchPhpStormEditor(location, distro)
	case "neovim":
		started = launchNeovimEditor(location, distro)
	case "sublime":
		started = launchSublimeEditor(location, distro)
	case "antigravity":
		started = launchAntigravityEditor(location, distro)
	default:
		started = launchVSCodeEditor(location, distro)
	}

	if !started {
		log.Printf("[editor] failed to launch %s for %s", editor, location)
	}
}

func launchVSCodeEditor(location, distro string) bool {
	switch runtime2.GOOS {
	case "windows":
		// For WSL projects, use VS Code's remote WSL protocol
		if distro != "" {
			linuxPath := resolveWSLProjectLocation(location, distro)
			if linuxPath == "" {
				linuxPath = location
			}
			remoteUri := "wsl+" + distro
			return startEditorCandidates(
				editorLaunchCommand{command: "code", args: []string{"--remote", remoteUri, linuxPath}, shell: true},
				editorLaunchCommand{command: "code-insiders", args: []string{"--remote", remoteUri, linuxPath}, shell: true},
				editorLaunchCommand{command: "codium", args: []string{"--remote", remoteUri, linuxPath}, shell: true},
			)
		}
		// For local Windows projects, use UNC path
		path := resolveWindowsEditorPath(location, distro)
		return startEditorCandidates(
			editorLaunchCommand{command: "code", args: []string{path}, shell: true},
			editorLaunchCommand{command: "code-insiders", args: []string{path}, shell: true},
			editorLaunchCommand{command: "codium", args: []string{path}, shell: true},
		)
	case "darwin":
		path := expandProjectPath(location)
		return startEditorCandidates(
			editorLaunchCommand{command: "code", args: []string{path}},
			editorLaunchCommand{command: "open", args: []string{"-a", "Visual Studio Code", path}},
		)
	default:
		path := expandProjectPath(location)
		return startEditorCandidates(
			editorLaunchCommand{command: "code", args: []string{path}},
			editorLaunchCommand{command: "code-insiders", args: []string{path}},
			editorLaunchCommand{command: "codium", args: []string{path}},
		)
	}
}

func launchPhpStormEditor(location, distro string) bool {
	switch runtime2.GOOS {
	case "windows":
		path := resolveWindowsEditorPath(location, distro)
		return startEditorCandidates(
			editorLaunchCommand{command: "phpstorm", args: []string{path}, shell: true},
			editorLaunchCommand{command: "phpstorm64.exe", args: []string{path}, shell: true},
			editorLaunchCommand{command: "storm", args: []string{path}, shell: true},
		)
	case "darwin":
		path := expandProjectPath(location)
		return startEditorCandidates(
			editorLaunchCommand{command: "phpstorm", args: []string{path}},
			editorLaunchCommand{command: "open", args: []string{"-a", "PhpStorm", path}},
		)
	default:
		path := expandProjectPath(location)
		return startEditorCandidates(
			editorLaunchCommand{command: "phpstorm", args: []string{path}},
			editorLaunchCommand{command: "phpstorm.sh", args: []string{path}},
			editorLaunchCommand{command: "storm", args: []string{path}},
		)
	}
}

func launchSublimeEditor(location, distro string) bool {
	switch runtime2.GOOS {
	case "windows":
		path := resolveWindowsEditorPath(location, distro)
		return startEditorCandidates(
			editorLaunchCommand{command: "subl", args: []string{path}, shell: true},
			editorLaunchCommand{command: "sublime_text", args: []string{path}, shell: true},
			editorLaunchCommand{command: "sublime_text.exe", args: []string{path}, shell: true},
		)
	case "darwin":
		path := expandProjectPath(location)
		return startEditorCandidates(
			editorLaunchCommand{command: "subl", args: []string{path}},
			editorLaunchCommand{command: "open", args: []string{"-a", "Sublime Text", path}},
		)
	default:
		path := expandProjectPath(location)
		return startEditorCandidates(
			editorLaunchCommand{command: "subl", args: []string{path}},
			editorLaunchCommand{command: "sublime_text", args: []string{path}},
		)
	}
}

func launchNeovimEditor(location, distro string) bool {
	switch runtime2.GOOS {
	case "windows":
		projectLocation := resolveWSLProjectLocation(location, distro)
		if distro != "" {
			wtCmd := exec.Command("wt.exe", "wsl.exe", "-d", distro, "--cd", projectLocation, "nvim")
			HideWindow(wtCmd)
			if err := wtCmd.Start(); err == nil {
				return true
			}

			fallback := exec.Command("wsl.exe", "-d", distro, "--cd", projectLocation, "nvim")
			ShowWindow(fallback)
			if err := fallback.Start(); err == nil {
				return true
			}
		}

		wtCmd := exec.Command("wt.exe", "nvim")
		HideWindow(wtCmd)
		if err := wtCmd.Start(); err == nil {
			return true
		}
		return false
	case "darwin":
		path := expandProjectPath(location)
		// Pass the path as an argument to the script to avoid injection
		script := "on run argv\ntell application \"Terminal\" to do script \"cd \" & quoted form of item 1 of argv & \" && nvim\"\nend run"
		cmd := exec.Command("osascript", "-e", script, path)
		return cmd.Start() == nil
	default:
		path := expandProjectPath(location)
		return startEditorCandidates(
			editorLaunchCommand{command: "x-terminal-emulator", args: []string{"--working-directory=" + path, "-e", "nvim"}},
			editorLaunchCommand{command: "konsole", args: []string{"--workdir", path, "-e", "nvim"}},
			editorLaunchCommand{command: "gnome-terminal", args: []string{"--working-directory=" + path, "--", "nvim"}},
			editorLaunchCommand{command: "xfce4-terminal", args: []string{"--working-directory=" + path, "-e", "nvim"}},
			editorLaunchCommand{command: "mate-terminal", args: []string{"--working-directory=" + path, "-e", "nvim"}},
			editorLaunchCommand{command: "kitty", args: []string{"--directory", path, "nvim"}},
			editorLaunchCommand{command: "alacritty", args: []string{"--working-directory", path, "-e", "nvim"}},
			editorLaunchCommand{command: "foot", args: []string{"--working-directory=" + path, "nvim"}},
			editorLaunchCommand{command: "wezterm", args: []string{"start", "--cwd", path, "nvim"}},
			editorLaunchCommand{command: "xterm", args: []string{"-e", "sh", "-lc", "cd '" + strings.ReplaceAll(path, "'", "'\\''") + "' && nvim"}},
		)
	}
}

func launchAntigravityEditor(location, distro string) bool {
	switch runtime2.GOOS {
	case "windows":
		projectLocation := resolveWSLProjectLocation(location, distro)
		if distro == "" {
			projectLocation = expandProjectPath(location)
		}
		antigravityPath := filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Antigravity", "bin", "antigravity.cmd")
		args := []string{projectLocation}
		if distro != "" {
			args = []string{"--remote", "wsl+" + distro, projectLocation}
		}
		return startEditorCandidates(
			editorLaunchCommand{command: antigravityPath, args: args, shell: true},
			editorLaunchCommand{command: "antigravity", args: []string{projectLocation}, shell: true},
		)
	case "darwin":
		path := expandProjectPath(location)
		return startEditorCandidates(
			editorLaunchCommand{command: "agy", args: []string{path}},
			editorLaunchCommand{command: "antigravity", args: []string{path}},
		)
	default:
		path := expandProjectPath(location)
		return startEditorCandidates(
			editorLaunchCommand{command: "agy", args: []string{path}},
			editorLaunchCommand{command: "antigravity", args: []string{path}},
		)
	}
}

func startEditorCandidates(candidates ...editorLaunchCommand) bool {
	for _, candidate := range candidates {
		var cmd *exec.Cmd
		if candidate.shell && runtime2.GOOS == "windows" {
			cmd = BuildWindowsCmd(candidate.command, candidate.args)
		} else {
			cmd = exec.Command(candidate.command, candidate.args...)
		}
		if runtime2.GOOS == "windows" {
			HideWindow(cmd)
		}
		if err := cmd.Start(); err == nil {
			return true
		}
	}
	return false
}

func expandProjectPath(location string) string {
	// Expand ~ to user home dir if needed
	if strings.HasPrefix(location, "~/") || strings.HasPrefix(location, "~\\") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, location[2:])
		}
	}
	return location
}

func resolveWSLProjectLocation(location, distro string) string {
	if location == "" || distro == "" || !strings.HasPrefix(location, "~/") {
		return location
	}

	resolveCtx, resolveCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer resolveCancel()
	whoamiCmd := execCommandContext(resolveCtx, "wsl.exe", "-d", distro, "-e", "whoami")
	HideWindow(whoamiCmd)
	var whoamiOut bytes.Buffer
	whoamiCmd.Stdout = &whoamiOut
	user := "root"
	if err := whoamiCmd.Run(); err == nil {
		if u := strings.TrimSpace(whoamiOut.String()); u != "" {
			user = u
		}
	}
	if user == "root" {
		return "/root/" + location[2:]
	}
	return "/home/" + user + "/" + location[2:]
}

func resolveWindowsEditorPath(location, distro string) string {
	if location == "" {
		return location
	}
	if len(location) >= 2 && location[1] == ':' {
		return location
	}
	if strings.HasPrefix(location, `\\`) {
		return location
	}
	if distro == "" {
		return expandProjectPath(location)
	}

	linuxPath := resolveWSLProjectLocation(location, distro)
	if linuxPath == "" {
		linuxPath = location
	}
	return `\\wsl.localhost\` + distro + strings.ReplaceAll(linuxPath, "/", `\\`)
}
