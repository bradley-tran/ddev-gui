package backend

import (
	"bufio"
	"bytes"
	"context"

	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	stdruntime "runtime"
	"strings"
	"sync"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type DdevService struct {
	mu        sync.Mutex
	ctx       context.Context
	config    *ConfigService
	shell     *WSLShell // persistent WSL shell for fast read-only queries (Windows only)
	fileShell *WSLShell // dedicated shell for file reads (avoids blocking the main shell)
	sshShell  *SSHShell // persistent SSH shell for remote execution
}

func NewDdevService(cfg *ConfigService) *DdevService {
	svc := &DdevService{config: cfg}
	backend, _ := cfg.Get("backend").(string)
	switch backend {
	case "ssh":
		svc.sshShell = SshShellFromConfig(cfg)
	default:
		if stdruntime.GOOS == "windows" {
			distro := resolveWSLDistro(cfg)
			svc.shell = NewWSLShell(distro)
			svc.fileShell = NewWSLShell(distro)
		}
	}
	return svc
}

// activeBackend returns the currently configured backend type ("wsl", "ssh", or "local").
func (d *DdevService) activeBackend() string {
	if d.config != nil {
		if b, ok := d.config.Get("backend").(string); ok && b != "" {
			return b
		}
	}
	if stdruntime.GOOS == "windows" {
		return "wsl"
	}
	return "local"
}

// WSLDistro returns the resolved WSL distribution name for use by external
// callers (e.g. main.go). It applies the same priority logic as wslDistro:
// 1) explicit config value, 2) "DDEV" if installed, 3) default distro ("").
func (d *DdevService) WSLDistro() string {
	return resolveWSLDistro(d.config)
}

var (
	ddevDistroExistsCache bool
	ddevDistroExistsOnce  sync.Once
)

// resolveWSLDistro determines which WSL distro to use.
// Priority: 1) explicit config value, 2) "DDEV" if it exists, 3) user's default distro.
func resolveWSLDistro(cfg *ConfigService) string {
	if cfg != nil {
		if v, ok := cfg.Get("wslDistro").(string); ok && v != "" {
			return v
		}
	}

	// No explicit config - check if "DDEV" distro is installed (cached).
	ddevDistroExistsOnce.Do(func() {
		ddevDistroExistsCache = distroExists("DDEV")
	})

	if ddevDistroExistsCache {
		return "DDEV"
	}

	// "DDEV" distro not found - fall back to user's default distro.
	log.Println("[wsl] DDEV distro not found, using default WSL distro")
	return ""
}

// distroExists checks whether a named WSL distribution is installed.
func distroExists(name string) bool {
	if stdruntime.GOOS != "windows" {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "wsl.exe", "-l", "-q")
	HideWSLWindow(cmd)
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	raw := strings.ReplaceAll(string(out), "\x00", "")
	for len(raw) > 0 {
		var line string
		if idx := strings.IndexByte(raw, '\n'); idx >= 0 {
			line, raw = raw[:idx], raw[idx+1:]
		} else {
			line, raw = raw, ""
		}
		if strings.EqualFold(strings.TrimSpace(line), name) {
			return true
		}
	}
	return false
}

// ListWSLDistros returns the names of all installed WSL distributions.
// On non-Windows systems it returns an empty slice.
func (d *DdevService) ListWSLDistros() []string {
	if stdruntime.GOOS != "windows" {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "wsl.exe", "-l", "-q")
	HideWSLWindow(cmd)
	out, err := cmd.Output()
	if err != nil {
		log.Printf("[wsl] failed to list distros: %v", err)
		return nil
	}
	// wsl -l -q outputs UTF-16LE on some Windows versions - normalise
	raw := string(out)
	// Strip NUL bytes from UTF-16LE output
	raw = strings.ReplaceAll(raw, "\x00", "")
	var distros []string
	scanner := bufio.NewScanner(strings.NewReader(raw))
	for scanner.Scan() {
		name := strings.TrimSpace(scanner.Text())
		if name != "" {
			distros = append(distros, name)
		}
	}
	return distros
}

// WSLExists returns whether WSL is available for the current runtime backend.
// To simplify, the actual check only happens for WSL backend on Windows.
// TODO: create a more robust backend check and allows user to select when auto-detection fails (e.g. WSL installed but not working correctly).
func (d *DdevService) WSLExists() bool {
	if stdruntime.GOOS != "windows" {
		return true
	}

	if d.activeBackend() != "wsl" {
		return true
	}

	return len(d.ListWSLDistros()) > 0
}

// ActiveBackend returns the current backend type as a string for the frontend.
func (d *DdevService) ActiveBackend() string {
	return d.activeBackend()
}

// ReloadBackend tears down existing shells and reinitializes the correct
// backend based on current config. Call this after changing the backend
// or WSL distro setting so changes take effect without restarting the app.
func (d *DdevService) ReloadBackend() {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Tear down existing shells
	if d.shell != nil {
		_ = d.shell.Close()
		d.shell = nil
	}
	if d.fileShell != nil {
		_ = d.fileShell.Close()
		d.fileShell = nil
	}
	if d.sshShell != nil {
		_ = d.sshShell.Close()
		d.sshShell = nil
	}

	// Re-initialize based on current config
	backend := d.activeBackend()
	log.Printf("[ddev] reloading backend: %s", backend)
	switch backend {
	case "ssh":
		d.sshShell = SshShellFromConfig(d.config)
	case "wsl":
		if stdruntime.GOOS == "windows" {
			distro := d.WSLDistro()
			d.shell = NewWSLShell(distro)
			d.fileShell = NewWSLShell(distro)
			log.Printf("[ddev] WSL shell reloaded with distro: %s", distro)
		}
	case "local":
		// No persistent shell needed for local execution
	}
}

// SetContext sets the Wails application context for event emitting.
func (d *DdevService) SetContext(ctx context.Context) {
	d.ctx = ctx

	if telemetryOptIn, ok := d.telemetryOptInPreference(); ok {
		go func() {
			if err := d.applyTelemetryPreference(context.Background(), telemetryOptIn); err != nil {
				log.Printf("[ddev] telemetry preference apply failed (will retry next launch): %v", err)
			}
		}()
	}
}

func (d *DdevService) telemetryOptInPreference() (bool, bool) {
	if d.config == nil {
		return false, false
	}

	telemetryOptIn, ok := d.config.Get("ddevTelemetryOptIn").(bool)
	return telemetryOptIn, ok
}

func (d *DdevService) applyTelemetryPreference(ctx context.Context, telemetryOptIn bool) error {
	return applyTelemetryPreference(ctx, telemetryOptIn, d.run)
}

func applyTelemetryPreference(
	ctx context.Context,
	telemetryOptIn bool,
	runner func(context.Context, ...string) (string, string, error),
) error {
	_, _, err := runner(ctx, "config", "global", fmt.Sprintf("--instrumentation-opt-in=%t", telemetryOptIn))
	if err != nil {
		return err
	}

	log.Printf("[ddev] telemetry preference applied: opt-in=%t", telemetryOptIn)
	return nil
}

// ListJSON returns the output of `ddev list --json-output`.
func (d *DdevService) ListJSON() (string, error) {
	out, errOut, err := d.run(context.Background(), "list", "--json-output")
	if err != nil {
		if errOut != "" {
			return "", fmt.Errorf("ddev list --json-output error: %s", strings.TrimSpace(errOut))
		}
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// DescribeJSON returns `ddev describe -j -p <project>` output as a JSON string.
func (d *DdevService) DescribeJSON(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("project name is required")
	}
	// `ddev describe` takes the project name positionally and JSON via -j/--json-output
	out, errOut, err := d.run(context.Background(), "describe", "-j", name)
	if err != nil {
		if errOut != "" {
			return "", fmt.Errorf("ddev describe -j error: %s", strings.TrimSpace(errOut))
		}
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// resolveProjectDir returns the actual project directory by querying `ddev describe -j`.
// On Windows (WSL) it returns a Linux path (e.g. /home/user/mysite).
// On macOS/Linux it returns the native path.
// Falls back to ~/ddev-projects/<name> if the lookup fails.
func (d *DdevService) resolveProjectDir(name string) string {
	fallback := "~/ddev-projects/" + name
	if stdruntime.GOOS != "windows" {
		if home, err := os.UserHomeDir(); err == nil && strings.TrimSpace(home) != "" {
			fallback = filepath.Join(home, "ddev-projects", name)
		}
	}
	raw, err := d.DescribeJSON(name)
	if err != nil {
		return fallback
	}
	// Parse the JSON to extract approot. The output may be wrapped in {"raw": {...}}.
	var wrapper struct {
		Raw struct {
			Approot string `json:"approot"`
		} `json:"raw"`
		Approot string `json:"approot"`
	}
	if err := json.Unmarshal([]byte(raw), &wrapper); err != nil {
		return fallback
	}
	if wrapper.Raw.Approot != "" {
		return wrapper.Raw.Approot
	}
	if wrapper.Approot != "" {
		return wrapper.Approot
	}
	return fallback
}

// FileEntry represents a single file or directory returned by ListDir.
type FileEntry struct {
	Name     string `json:"name"`
	IsDir    bool   `json:"isDir"`
	Size     string `json:"size"`
	Modified string `json:"modified"`
}

// execShellCmd runs a bash command string using the most efficient available
// transport: persistent WSL shell, SSH shell, or local exec.
// This avoids spawning a fresh wsl.exe process for each call (~5-10x faster).
func (d *DdevService) execShellCmd(cmd string) (string, error) {
	switch d.activeBackend() {
	case "ssh":
		if d.sshShell != nil {
			output, exitCode, execErr := d.sshShell.Exec("", []string{"bash", "-c", cmd}, nil, 30*time.Second, nil)
			if execErr != nil {
				return "", execErr
			}
			if exitCode != 0 {
				return "", fmt.Errorf("exit status %d", exitCode)
			}
			return output, nil
		}
		return "", errors.New("SSH shell not available")
	case "wsl":
		// Use the persistent WSL shell if available (avoids WSL startup latency)
		if d.shell != nil {
			output, exitCode, execErr := d.shell.Exec("", []string{"bash", "-c", cmd}, nil, 30*time.Second, nil)
			if execErr != nil {
				return "", execErr
			}
			if exitCode != 0 {
				return "", fmt.Errorf("exit status %d", exitCode)
			}
			return output, nil
		}
		// Fallback: spawn a fresh wsl.exe process
		execCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		wslArgs := []string{"-d", d.WSLDistro(), "-e", "bash", "-c", cmd}
		c := exec.CommandContext(execCtx, "wsl.exe", wslArgs...)
		HideWSLWindow(c)
		raw, execErr := c.Output()
		if execErr != nil {
			return "", execErr
		}
		return string(raw), nil
	default:
		execCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		c := exec.CommandContext(execCtx, "bash", "-c", cmd)
		raw, execErr := c.Output()
		if execErr != nil {
			return "", execErr
		}
		return string(raw), nil
	}
}

// ListDir lists files and directories inside a project path.
// relPath is relative to the project root (use "" or "." for the root).
// Returns a JSON array of FileEntry objects.
func (d *DdevService) ListDir(project, relPath string) (string, error) {
	project = strings.TrimSpace(project)
	if project == "" {
		return "[]", errors.New("project name is required")
	}
	dirHint := d.resolveProjectDir(project)

	// Sanitise the relative path to prevent directory traversal
	relPath = strings.TrimSpace(relPath)
	if relPath == "" || relPath == "." {
		relPath = "."
	}
	// Block obvious traversal attempts
	if strings.Contains(relPath, "..") {
		return "[]", errors.New("relative path must not contain '..'")
	}

	target := dirHint
	if relPath != "." {
		target = dirHint + "/" + relPath
	}

	// Run ls -la with machine-readable timestamps and --group-directories-first
	// Using find instead for more reliable cross-platform parsing
	// Format: type|size|date|name
	cmd := fmt.Sprintf(
		"find %s -maxdepth 1 -mindepth 1 -printf '%%y|%%s|%%TY-%%Tm-%%Td %%TH:%%TM|%%f\\n' 2>/dev/null | sort -t'|' -k1,1r -k4,4",
		shellQuote(target),
	)

	out, err := d.execShellCmd(cmd)
	if err != nil {
		return "[]", err
	}

	// Parse output lines: type|size|date|name
	var entries []FileEntry
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 4)
		if len(parts) < 4 {
			continue
		}
		ftype := parts[0]
		size := parts[1]
		modified := parts[2]
		name := parts[3]
		if name == "." || name == ".." {
			continue
		}
		entries = append(entries, FileEntry{
			Name:     name,
			IsDir:    ftype == "d",
			Size:     size,
			Modified: modified,
		})
	}
	if entries == nil {
		entries = []FileEntry{}
	}
	b, _ := json.Marshal(entries)
	return string(b), nil
}

// ReadFile reads the content of a file inside the project directory.
// relPath is relative to the project root.
// Returns the raw file content as a string (capped at 1 MB).
func (d *DdevService) ReadFile(project, relPath string) (string, error) {
	project = strings.TrimSpace(project)
	if project == "" {
		return "", errors.New("project name is required")
	}
	relPath = strings.TrimSpace(relPath)
	if relPath == "" {
		return "", errors.New("file path is required")
	}
	if strings.Contains(relPath, "..") {
		return "", errors.New("relative path must not contain '..'")
	}

	dirHint := d.resolveProjectDir(project)
	target := dirHint + "/" + relPath

	// Pipe through base64 for safe transport through the persistent shell.
	// Raw binary content can overflow the scanner's 1MB line buffer and kill the shell.
	cmd := fmt.Sprintf("head -c 1048576 %s 2>/dev/null | base64", shellQuote(target))

	out, err := d.execSpawnCmd(cmd, 30*time.Second)
	if err != nil {
		return "", err
	}
	// Strip line-wrapping whitespace and decode
	out = strings.ReplaceAll(out, "\n", "")
	out = strings.ReplaceAll(out, "\r", "")
	decoded, err := base64.StdEncoding.DecodeString(out)
	if err != nil {
		return "", fmt.Errorf("base64 decode: %w", err)
	}
	return string(decoded), nil
}

// ReadFileBase64 reads a binary file and returns its content as base64.
// Useful for images and other binary files that can't be transferred as text.
// Capped at 5 MB.
func (d *DdevService) ReadFileBase64(project, relPath string) (string, error) {
	project = strings.TrimSpace(project)
	if project == "" {
		return "", errors.New("project name is required")
	}
	relPath = strings.TrimSpace(relPath)
	if relPath == "" {
		return "", errors.New("file path is required")
	}
	if strings.Contains(relPath, "..") {
		return "", errors.New("relative path must not contain '..'")
	}

	dirHint := d.resolveProjectDir(project)
	target := dirHint + "/" + relPath

	// Read up to 5MB and base64-encode it (default line wrapping).
	cmd := fmt.Sprintf("head -c 5242880 %s 2>/dev/null | base64", shellQuote(target))

	out, err := d.execSpawnCmd(cmd, 30*time.Second)
	if err != nil {
		return "", err
	}
	// Strip all whitespace (newlines from line wrapping) to produce a clean base64 string
	out = strings.ReplaceAll(out, "\n", "")
	out = strings.ReplaceAll(out, "\r", "")
	return out, nil
}

// execSpawnCmd runs a bash command via the dedicated file-read shell.
// Unlike execShellCmd (which uses the main persistent shell), this uses
// a separate persistent shell (fileShell) with its own mutex, so file
// reads don't block directory listings and vice versa.
func (d *DdevService) execSpawnCmd(cmd string, timeout time.Duration) (string, error) {
	switch d.activeBackend() {
	case "wsl":
		// Use the dedicated file shell if available
		if d.fileShell != nil {
			output, exitCode, execErr := d.fileShell.Exec("", []string{"bash", "-c", cmd}, nil, timeout, nil)
			if execErr != nil {
				return "", execErr
			}
			if exitCode != 0 {
				return "", fmt.Errorf("exit status %d", exitCode)
			}
			return output, nil
		}
		// Fallback: spawn a fresh process
		execCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		wslArgs := []string{"-d", d.WSLDistro(), "-e", "bash", "-c", cmd}
		c := exec.CommandContext(execCtx, "wsl.exe", wslArgs...)
		HideWSLWindow(c)
		raw, execErr := c.Output()
		if execErr != nil {
			return "", execErr
		}
		return string(raw), nil
	case "ssh":
		if d.sshShell != nil {
			output, exitCode, execErr := d.sshShell.Exec("", []string{"bash", "-c", cmd}, nil, timeout, nil)
			if execErr != nil {
				return "", execErr
			}
			if exitCode != 0 {
				return "", fmt.Errorf("exit status %d", exitCode)
			}
			return output, nil
		}
		return "", errors.New("SSH shell not available")
	default:
		execCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		c := exec.CommandContext(execCtx, "bash", "-c", cmd)
		raw, execErr := c.Output()
		if execErr != nil {
			return "", execErr
		}
		return string(raw), nil
	}
}

// WpCoreInstall runs `ddev wp core download` followed by `ddev wp core install`
// to set up a WordPress site with default admin credentials.
// All commands run via fresh processes to avoid blocking the persistent shell.
func (d *DdevService) WpCoreInstall(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("project name is required")
	}
	dirHint := d.resolveProjectDir(name)

	// Step 1: Download WordPress core files
	if d.ctx != nil {
		wruntime.EventsEmit(d.ctx, "ddev:output", "Downloading WordPress core…")
	}
	_, err := d.runDirect(context.Background(), dirHint, nil, "wp", "core", "download")
	if err != nil {
		errMsg := err.Error()
		// If WP is already downloaded, that's okay - continue
		if !strings.Contains(strings.ToLower(errMsg), "already exists") {
			return "", fmt.Errorf("wp core download error: %s", stripAnsi(errMsg))
		}
	}

	// Step 2: Install WordPress with a random admin password
	password, err := GenerateRandomString(12)
	if err != nil {
		return "", fmt.Errorf("failed to generate admin password: %w", err)
	}

	if d.ctx != nil {
		wruntime.EventsEmit(d.ctx, "ddev:output", "Installing WordPress…")
		wruntime.EventsEmit(d.ctx, "ddev:output", "WordPress Admin User: admin")
		wruntime.EventsEmit(d.ctx, "ddev:output", "WordPress Admin Password: "+password)
	}
	siteURL := fmt.Sprintf("https://%s.ddev.site", name)
	return d.runDirect(context.Background(), dirHint, nil,
		"wp", "core", "install",
		"--url="+siteURL,
		"--title="+name,
		"--admin_user=admin",
		"--admin_password="+password,
		"--admin_email=admin@"+name+".ddev.site",
	)
}

// LaravelInit runs post-install artisan commands to set up a Laravel project:
// - php artisan key:generate
// - php artisan migrate --force
// All commands run via fresh processes to avoid blocking the persistent shell.
func (d *DdevService) LaravelInit(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("project name is required")
	}
	dirHint := d.resolveProjectDir(name)

	// Step 1: Generate application key
	if d.ctx != nil {
		wruntime.EventsEmit(d.ctx, "ddev:output", "Generating Laravel app key…")
	}
	_, err := d.runDirect(context.Background(), dirHint, nil, "exec", "php", "artisan", "key:generate")
	if err != nil {
		return "", fmt.Errorf("artisan key:generate error: %s", stripAnsi(err.Error()))
	}

	// Step 2: Run database migrations
	if d.ctx != nil {
		wruntime.EventsEmit(d.ctx, "ddev:output", "Running Laravel migrations…")
	}
	return d.runDirect(context.Background(), dirHint, nil, "exec", "php", "artisan", "migrate", "--force")
}

// ExecCommand runs a command inside the project's ddev web container via
// `ddev exec` and streams output line-by-line via Wails events:
//   - "terminal:output:<project>" for each line of output
//   - "terminal:done:<project>" when the command finishes (payload: exit code)
//
// This powers the embedded terminal in the project detail view, providing
// the same environment as a `ddev ssh` session.
func (d *DdevService) ExecCommand(project, command string) (string, error) {
	project = strings.TrimSpace(project)
	command = strings.TrimSpace(command)
	if project == "" {
		return "", errors.New("project name is required")
	}
	if command == "" {
		return "", errors.New("command is required")
	}

	dirHint := d.resolveProjectDir(project)
	timeout := 5 * time.Minute

	outputEvent := "terminal:output:" + project
	doneEvent := "terminal:done:" + project

	onLine := func(line string) {
		if d.ctx != nil {
			wruntime.EventsEmit(d.ctx, outputEvent, line)
		}
	}

	emitDone := func(exitCode int) {
		if d.ctx != nil {
			wruntime.EventsEmit(d.ctx, doneEvent, exitCode)
		}
	}

	switch d.activeBackend() {
	case "ssh":
		if d.sshShell == nil {
			emitDone(-1)
			return "", errors.New("SSH shell not available")
		}
		output, exitCode, execErr := d.sshShell.Exec(dirHint, []string{"ddev", "exec", "--", "bash", "-c", command}, nil, timeout, onLine)
		emitDone(exitCode)
		if execErr != nil {
			return "", execErr
		}
		if exitCode != 0 {
			return strings.TrimSpace(output), fmt.Errorf("exit status %d", exitCode)
		}
		return strings.TrimSpace(output), nil

	case "wsl":
		if d.shell != nil {
			output, exitCode, execErr := d.shell.Exec(dirHint, []string{"ddev", "exec", "--", "bash", "-c", command}, nil, timeout, onLine)
			emitDone(exitCode)
			if execErr != nil {
				return "", execErr
			}
			if exitCode != 0 {
				return strings.TrimSpace(output), fmt.Errorf("exit status %d", exitCode)
			}
			return strings.TrimSpace(output), nil
		}
		// Fallback: spawn a fresh wsl.exe process
		execCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		wslArgs := []string{"-d", d.WSLDistro(), "--cd", dirHint, "-e", "bash", "-c", "ddev exec -- bash -c \"$1\" 2>&1", "--", command}
		cmd := exec.CommandContext(execCtx, "wsl.exe", wslArgs...)
		HideWSLWindow(cmd)
		stdoutPipe, _ := cmd.StdoutPipe()
		if err := cmd.Start(); err != nil {
			emitDone(-1)
			return "", err
		}
		var lines []string
		scanner := bufio.NewScanner(stdoutPipe)
		scanner.Buffer(make([]byte, 0, 256*1024), 1024*1024)
		scanner.Split(scanLinesOrCR)
		for scanner.Scan() {
			line := scanner.Text()
			lines = append(lines, line)
			onLine(line)
		}
		exitCode := 0
		if err := cmd.Wait(); err != nil {
			exitCode = -1
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			}
		}
		emitDone(exitCode)
		output := strings.TrimSpace(strings.Join(lines, "\n"))
		if exitCode != 0 {
			return output, fmt.Errorf("exit status %d", exitCode)
		}
		return output, nil

	default:
		// Local backend
		execCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		cmd := exec.CommandContext(execCtx, "ddev", "exec", "--", "bash", "-c", command)
		cmd.Dir = dirHint
		cmd.Env = os.Environ()
		stdoutPipe, _ := cmd.StdoutPipe()
		cmd.Stderr = cmd.Stdout
		if err := cmd.Start(); err != nil {
			emitDone(-1)
			return "", err
		}
		var lines []string
		scanner := bufio.NewScanner(stdoutPipe)
		scanner.Buffer(make([]byte, 0, 256*1024), 1024*1024)
		scanner.Split(scanLinesOrCR)
		for scanner.Scan() {
			line := scanner.Text()
			lines = append(lines, line)
			onLine(line)
		}
		exitCode := 0
		if err := cmd.Wait(); err != nil {
			exitCode = -1
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			}
		}
		emitDone(exitCode)
		output := strings.TrimSpace(strings.Join(lines, "\n"))
		if exitCode != 0 {
			return output, fmt.Errorf("exit status %d", exitCode)
		}
		return output, nil
	}
}

// run is a wrapper around runStream for non-streaming calls.
func (d *DdevService) run(ctx context.Context, args ...string) (string, string, error) {
	return d.runStream(ctx, false, args...)
}

// Shutdown cleans up the persistent WSL/SSH shells. Should be called on app exit.
func (d *DdevService) Shutdown() {
	if d.shell != nil {
		_ = d.shell.Close()
	}
	if d.fileShell != nil {
		_ = d.fileShell.Close()
	}
	if d.sshShell != nil {
		_ = d.sshShell.Close()
	}
}

// runStream executes ddev commands using the main shell. If stream is true and
// SetContext has been called, it emits "ddev:output" events for each line.
func (d *DdevService) runStream(ctx context.Context, stream bool, args ...string) (stdout string, stderr string, err error) {
	sh := d.shell

	// Choose timeout: streaming commands (start, restart, etc.) can be very slow.
	timeout := 60 * time.Second
	if stream {
		timeout = 10 * time.Minute
	}

	// optional working directory from context
	var dir string
	if v := ctx.Value("dir"); v != nil {
		if s, ok := v.(string); ok {
			dir = strings.TrimSpace(s)
		}
	}

	// ── SSH backend ──
	if d.activeBackend() == "ssh" && d.sshShell != nil {
		fullArgs := append([]string{"ddev"}, args...)

		var onLine func(string)
		if stream && d.ctx != nil {
			onLine = func(line string) {
				wruntime.EventsEmit(d.ctx, "ddev:output", line)
			}
		}

		output, exitCode, execErr := d.sshShell.Exec(dir, fullArgs, shellEnvVars(ctx), timeout, onLine)
		if execErr != nil {
			return output, "", execErr
		}
		if exitCode != 0 {
			return output, "", fmt.Errorf("exit status %d", exitCode)
		}
		return output, "", nil
	}

	// ── WSL backend (Windows) ──
	if d.activeBackend() == "wsl" && sh != nil {
		// Prepend "ddev" to args since the shell runs bare bash.
		fullArgs := append([]string{"ddev"}, args...)

		var onLine func(string)
		if stream && d.ctx != nil {
			onLine = func(line string) {
				wruntime.EventsEmit(d.ctx, "ddev:output", line)
			}
		}

		output, exitCode, execErr := sh.Exec(dir, fullArgs, shellEnvVars(ctx), timeout, onLine)
		// The persistent shell merges stderr into stdout, so stderr is always empty.
		if execErr != nil {
			return output, "", execErr
		}
		if exitCode != 0 {
			return output, "", fmt.Errorf("exit status %d", exitCode)
		}
		return output, "", nil
	}

	// ── Local backend: standard exec ──
	c, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(c, "ddev", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = os.Environ()

	var outBuf, errBuf bytes.Buffer

	// If streaming is requested and we have a Wails context to emit to
	if stream && d.ctx != nil {
		stdoutPipe, _ := cmd.StdoutPipe()
		stderrPipe, _ := cmd.StderrPipe()

		if err := cmd.Start(); err != nil {
			return "", "", err
		}

		var wg sync.WaitGroup
		wg.Add(2)

		streamReader := func(pipe io.Reader, buf *bytes.Buffer) {
			defer wg.Done()
			tee := io.TeeReader(pipe, buf)
			scanner := bufio.NewScanner(tee)
			for scanner.Scan() {
				line := scanner.Text()
				wruntime.EventsEmit(d.ctx, "ddev:output", line)
			}
		}

		go streamReader(stdoutPipe, &outBuf)
		go streamReader(stderrPipe, &errBuf)

		err := cmd.Wait()
		wg.Wait()
		return outBuf.String(), errBuf.String(), err
	}

	// Standard non-streaming capture
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	runErr := cmd.Run()
	return outBuf.String(), errBuf.String(), runErr
}

// runDirect spawns a fresh wsl.exe process (Windows/WSL), SSH session, or local exec
// for a single ddev command. Unlike runStream (which uses the persistent shell),
// this does NOT hold the shared WSL mutex, so the persistent shell stays free for
// quick queries (list, describe) while long-running operations execute in parallel.
//
// dir is an optional working directory (WSL path like ~/ddev-projects/foo).
// envVars are optional environment variable assignments (e.g. "COMPOSER_ALLOW_SUPERUSER=1").
// args are the ddev subcommand and arguments (e.g. "start", "myproject").
//
// Output is streamed line-by-line via "ddev:output" events when a Wails context is set.
// Returns combined stdout+stderr and any error.
func (d *DdevService) runDirect(ctx context.Context, dir string, envVars []string, args ...string) (string, error) {
	timeout := 15 * time.Minute

	// ── SSH backend ──
	if d.activeBackend() == "ssh" && d.sshShell != nil {
		fullArgs := append([]string{"ddev"}, args...)
		var onLine func(string)
		if d.ctx != nil {
			onLine = func(line string) {
				wruntime.EventsEmit(d.ctx, "ddev:output", line)
			}
		}
		output, exitCode, execErr := d.sshShell.Exec(dir, fullArgs, envVars, timeout, onLine)
		if execErr != nil {
			if strings.TrimSpace(output) != "" {
				return "", fmt.Errorf("%s", strings.TrimSpace(output))
			}
			return "", execErr
		}
		if exitCode != 0 {
			if strings.TrimSpace(output) != "" {
				return "", fmt.Errorf("%s", strings.TrimSpace(output))
			}
			return "", fmt.Errorf("exit status %d", exitCode)
		}
		return strings.TrimSpace(output), nil
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// ── WSL backend (Windows): spawn a fresh wsl.exe ──
	var cmd *exec.Cmd
	if stdruntime.GOOS == "windows" && d.activeBackend() == "wsl" {
		wslArgs := []string{"-d", d.WSLDistro()}
		if dir != "" {
			wslArgs = append(wslArgs, "--cd", dir)
		}
		// Build a bash -c command with 2>&1 so stderr is merged inside Linux.
		// Using Go's cmd.Stderr = cmd.Stdout after StdoutPipe() can keep wsl.exe
		// alive when ddev child processes inherit the duplicated pipe fd.
		var parts []string
		for _, ev := range envVars {
			parts = append(parts, "export "+ev)
		}
		var sb strings.Builder
		sb.WriteString("ddev")
		for _, a := range args {
			sb.WriteByte(' ')
			sb.WriteString(shellQuote(a))
		}
		sb.WriteString(" 2>&1")
		parts = append(parts, sb.String())
		wslArgs = append(wslArgs, "-e", "bash", "-c", strings.Join(parts, "; "))
		cmd = exec.CommandContext(execCtx, "wsl.exe", wslArgs...)
		HideWSLWindow(cmd)
	} else {
		// ── Local backend ──
		cmd = exec.CommandContext(execCtx, "ddev", args...)
		if dir != "" {
			cmd.Dir = dir
		}
		cmd.Env = os.Environ()
		for _, ev := range envVars {
			cmd.Env = append(cmd.Env, ev)
		}
	}

	stdoutPipe, _ := cmd.StdoutPipe()
	// For local backend only: merge stderr into stdout at the Go pipe level.
	// WSL backend already merges via 2>&1 inside bash to avoid pipe fd leaks.
	if !(stdruntime.GOOS == "windows" && d.activeBackend() == "wsl") {
		cmd.Stderr = cmd.Stdout
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}

	var lines []string
	scanner := bufio.NewScanner(stdoutPipe)
	scanner.Buffer(make([]byte, 0, 256*1024), 1024*1024)
	scanner.Split(scanLinesOrCR)

	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
		if d.ctx != nil {
			wruntime.EventsEmit(d.ctx, "ddev:output", line)
		}
	}

	if err := cmd.Wait(); err != nil {
		out := strings.TrimSpace(strings.Join(lines, "\n"))
		if out != "" {
			return "", fmt.Errorf("%s", out)
		}
		return "", err
	}
	return strings.TrimSpace(strings.Join(lines, "\n")), nil
}

// VersionInfo returns a human-readable summary of ddev and (on Windows) WSL versions.
func (d *DdevService) VersionInfo() (string, bool) {
	var sections []string
	ddevFound := true

	// DDEV section
	var ddevLines []string
	{
		out, errOut, err := d.run(context.Background(), "--version")
		if err != nil {
			errText := strings.ToLower(strings.TrimSpace(err.Error() + " " + errOut))
			if strings.Contains(errText, "executable file not found") ||
				strings.Contains(errText, "file not found") ||
				strings.Contains(errText, "not recognized as an internal or external command") ||
				strings.Contains(errText, "no such file or directory") {
				ddevFound = false
				ddevLines = append(ddevLines, "Status: not found ⚠️")
				if stdruntime.GOOS != "windows" {
					ddevLines = append(ddevLines, "Install: https://ddev.com/get-started/")
				}
			} else {
				em := strings.TrimSpace(errOut)
				if em == "" {
					em = err.Error()
				}
				ddevLines = append(ddevLines, "Status: error: "+strings.TrimSpace(em))
			}
		} else {
			out = strings.TrimSpace(out)
			if out == "" {
				ddevLines = append(ddevLines, "Status: (no output)")
			} else {
				lines := strings.Split(out, "\n")
				ddevLines = append(ddevLines, "Status: "+strings.TrimSpace(lines[0]))
			}
		}
	}
	sections = append(sections, "DDEV\n"+strings.Join(ddevLines, "\n"))

	// WSL section (Windows only)
	if stdruntime.GOOS == "windows" {
		var wslLines []string
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "wsl.exe", "--version")
		HideWSLWindow(cmd)
		var outBuf, errBuf bytes.Buffer
		cmd.Stdout = &outBuf
		cmd.Stderr = &errBuf
		err := cmd.Run()
		if err == nil && strings.TrimSpace(outBuf.String()) != "" {
			line := strings.Split(strings.ReplaceAll(outBuf.String(), "\r\n", "\n"), "\n")[0]
			wslLines = append(wslLines, "Status: "+strings.TrimSpace(line))
			// Determine list of distro and warn if there is no Ubuntu-based one
			outBuf.Reset()
			errBuf.Reset()
			cmd = exec.CommandContext(ctx, "wsl.exe", "-l", "-v")
			HideWSLWindow(cmd)
			cmd.Stdout = &outBuf
			cmd.Stderr = &errBuf
			if err3 := cmd.Run(); err3 == nil {
				lines := strings.Split(strings.ReplaceAll(outBuf.String(), "\r\n", "\n"), "\n")
				ubuntuInstalled := false
				for _, ln := range lines {
					l := strings.TrimSpace(ln)
					if l == "" || strings.HasPrefix(strings.ToUpper(l), "NAME") {
						continue
					}
					if strings.Contains(strings.ToLower(l), "ubuntu") || strings.Contains(l, "DDEV") {
						ubuntuInstalled = true
					}
					wslLines = append(wslLines, l)
				}
				if !ubuntuInstalled {
					wslLines = append(wslLines, "Warning: ddev requires a ubuntu-based wsl distro ⚠️")
				}
			}
		} else {
			outBuf.Reset()
			errBuf.Reset()
			cmd = exec.CommandContext(ctx, "wsl.exe", "-l", "-v")
			HideWSLWindow(cmd)
			cmd.Stdout = &outBuf
			cmd.Stderr = &errBuf
			if err2 := cmd.Run(); err2 == nil && strings.TrimSpace(outBuf.String()) != "" {
				lines := strings.Split(strings.ReplaceAll(outBuf.String(), "\r\n", "\n"), "\n")
				header := strings.TrimSpace(lines[0])
				if header == "" && len(lines) > 1 {
					header = strings.TrimSpace(lines[1])
				}
				if header == "" {
					header = "WSL is installed (details via `wsl -l -v`)"
				}
				wslLines = append(wslLines, "Status: "+header)
				// Determine default distro and warn if not Ubuntu-based
				var defaultName string
				for _, ln := range lines {
					l := strings.TrimSpace(ln)
					if l == "" || strings.HasPrefix(strings.ToUpper(l), "NAME") {
						continue
					}
					isDefault := strings.HasPrefix(l, "*")
					if isDefault {
						l = strings.TrimSpace(strings.TrimPrefix(l, "*"))
						parts := strings.Fields(l)
						if len(parts) > 0 {
							defaultName = parts[0]
							break
						}
					}
				}
				if defaultName != "" {
					wslLines = append(wslLines, "Default distro: "+defaultName)
					if !strings.Contains(strings.ToLower(defaultName), "ubuntu") {
						wslLines = append(wslLines, "Warning: ddev requires a ubuntu-based wsl distro ⚠️")
					}
				}
			} else {
				// Consider unavailable as not installed
				wslLines = append(wslLines, "Status: not installed or not available ⚠️")
				wslLines = append(wslLines, "Install: https://learn.microsoft.com/windows/wsl/install")
			}
		}
		sections = append(sections, "WSL\n"+strings.Join(wslLines, "\n"))
	}

	return strings.Join(sections, "\n\n"), ddevFound
}

// DdevInstalledVersion returns the first line of `ddev version` if available.
// If ddev is not installed or returns an error, an error is returned.
func (d *DdevService) DdevInstalledVersion() (string, error) {
	out, errOut, err := d.run(context.Background(), "--version")
	if err != nil {
		if strings.TrimSpace(errOut) != "" {
			return "", errors.New(strings.TrimSpace(errOut))
		}
		return "", err
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return "", errors.New("no output")
	}
	line := strings.Split(out, "\n")[0]
	return strings.TrimSpace(line), nil
}

// ComposerInstall runs a composer step for the project directory under ddev:
//   - If composer.json exists, runs `ddev composer install`.
//   - Else runs `ddev composer create-project` with a package appropriate to the project type,
//     followed by `ddev composer require drush/drush` for Drupal projects.
//
// All commands run via runDirect (fresh processes) to avoid blocking the persistent shell.
func (d *DdevService) ComposerInstall(name, ptype string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("project name is required")
	}
	// Build directory hint
	dirHint := d.resolveProjectDir(name)

	composerEnv := []string{"COMPOSER_ALLOW_SUPERUSER=1", "COMPOSER_NO_INTERACTION=1"}

	// Check for composer.json
	checkCtx := context.WithValue(context.Background(), "dir", dirHint)
	_, _, testErr := d.run(checkCtx, "exec", "test", "-f", "composer.json")

	if testErr == nil {
		// composer.json already exists - just install dependencies
		if d.ctx != nil {
			wruntime.EventsEmit(d.ctx, "ddev:output", "Running composer install…")
		}
		return d.runDirect(context.Background(), dirHint, composerEnv, "composer", "install", "--no-interaction")
	}

	// No composer.json - scaffold with create-project
	pkg := createProjectPackage(ptype)
	if pkg == "" {
		return "", fmt.Errorf("no create-project template for project type %q - create your project files manually and use Init Site", ptype)
	}

	if d.ctx != nil {
		wruntime.EventsEmit(d.ctx, "ddev:output", fmt.Sprintf("Running composer create-project %s …", pkg))
	}
	// ddev composer create-project is a special DDEV command that handles the
	// current directory automatically and runs post-install scripts. No target
	// dir argument - adding "." triggers interactive "directory not empty" prompts.
	out, err := d.runDirect(context.Background(), dirHint, composerEnv, "composer", "create-project", pkg)
	if err != nil {
		return out, err
	}

	// For Drupal projects, also require drush (official quickstart step)
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(ptype)), "drupal") {
		if d.ctx != nil {
			wruntime.EventsEmit(d.ctx, "ddev:output", "Installing drush/drush…")
		}
		_, _ = d.runDirect(context.Background(), dirHint, composerEnv, "composer", "require", "drush/drush")
	}

	// For TYPO3, create the FIRST_INSTALL file so the install wizard starts
	if strings.ToLower(strings.TrimSpace(ptype)) == "typo3" {
		if d.ctx != nil {
			wruntime.EventsEmit(d.ctx, "ddev:output", "Creating FIRST_INSTALL file…")
		}
		_, _ = d.runDirect(context.Background(), dirHint, nil, "exec", "touch", "public/FIRST_INSTALL")
	}

	return out, nil
}

// ConfigureProject runs `ddev config` in the given directory with the supplied options.
// Uses runDirect (fresh process) for the config command itself.
func (d *DdevService) ConfigureProject(dir, name, ptype, docroot, php string) (string, error) {
	// Directory is intentionally ignored; we always use ~/ddev-projects
	name = strings.TrimSpace(name)
	ptype = strings.TrimSpace(ptype)
	docroot = strings.TrimSpace(docroot)
	php = strings.TrimSpace(php)
	if name == "" || ptype == "" || php == "" {
		return "", errors.New("fields (name, type, php) are required")
	}

	// Make sure the target directory exists.
	targetDir := "~/ddev-projects/" + name
	if runtime.GOOS != "windows" {
		home, herr := os.UserHomeDir()
		if herr != nil || strings.TrimSpace(home) == "" {
			return "", errors.New("could not resolve user home directory")
		}
		targetDir = filepath.Join(home, "ddev-projects", name)
		_ = os.MkdirAll(targetDir, 0o755)
	} else {
		// On Windows, create directory via the persistent WSL shell (quick op).
		if d.shell != nil {
			// Use $HOME instead of ~ to allow safe quoting of the project name.
			cmd := "mkdir -p $HOME/ddev-projects/" + shellQuote(name)
			_, _, _ = d.shell.Exec("", []string{"bash", "-c", cmd}, nil, 15*time.Second, nil)
		} else {
			c, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			// exec.Command handles arguments safely, so ddev-projects/name is fine here
			// as long as it's not passed to a shell.
			mkdir := exec.CommandContext(c, "wsl.exe", "-d", d.WSLDistro(), "--cd", "~", "-e", "mkdir", "-p", "ddev-projects/"+name)
			HideWSLWindow(mkdir)
			_ = mkdir.Run()
		}
	}

	// Run config via a fresh process
	return d.runDirect(context.Background(), targetDir, nil, "config",
		"--project-type", ptype,
		"--docroot", docroot,
		"--php-version", php,
		"--project-name", name,
	)
}

// CreateProject creates a new DDEV project under ~/ddev-projects.
// It delegates to ConfigureProject with the home directory as the base path.
func (d *DdevService) CreateProject(name, ptype, docroot, php string) (string, error) {
	return d.ConfigureProject("~", name, ptype, docroot, php)
}

// CloneRepo clones a git repository into ~/ddev-projects/<name>.
// On Windows/WSL it spawns a fresh wsl.exe process (like ComposerInstall)
// so the main shell stays free for start/stop/list during the clone.
func (d *DdevService) CloneRepo(name, repoURL string) (string, error) {
	name = strings.TrimSpace(name)
	repoURL = strings.TrimSpace(repoURL)
	if name == "" || repoURL == "" {
		return "", errors.New("project name and git repo URL are required")
	}

	targetDir := "~/ddev-projects/" + name

	// Ensure the parent directory exists before cloning.
	if runtime.GOOS != "windows" {
		home, herr := os.UserHomeDir()
		if herr != nil || strings.TrimSpace(home) == "" {
			return "", errors.New("could not resolve user home directory")
		}
		_ = os.MkdirAll(filepath.Join(home, "ddev-projects"), 0o755)
		targetDir = filepath.Join(home, "ddev-projects", name)
	} else if d.shell != nil {
		_, _, _ = d.shell.Exec("", []string{"bash", "-c", "mkdir -p ~/ddev-projects"}, nil, 15*time.Second, nil)
	}

	if d.ctx != nil {
		wruntime.EventsEmit(d.ctx, "ddev:output", fmt.Sprintf("Cloning %s …", repoURL))
	}

	timeout := 10 * time.Minute

	// ── SSH backend ──
	if d.activeBackend() == "ssh" && d.sshShell != nil {
		// Use $HOME instead of ~ to allow safe quoting of the target directory.
		quotedTargetDir := "$HOME/ddev-projects/" + shellQuote(name)
		cloneCmd := fmt.Sprintf("git clone %s %s", shellQuote(repoURL), quotedTargetDir)
		args := []string{"bash", "-c", cloneCmd}
		var onLine func(string)
		if d.ctx != nil {
			onLine = func(line string) {
				wruntime.EventsEmit(d.ctx, "ddev:output", line)
			}
		}
		output, exitCode, execErr := d.sshShell.Exec("", args, nil, timeout, onLine)
		if execErr != nil {
			return "", execErr
		}
		if exitCode != 0 {
			// Check if the repo was actually cloned despite the non-zero exit
			_, checkExit, checkErr := d.sshShell.Exec("", []string{"test", "-d", targetDir + "/.git"}, nil, 15*time.Second, nil)
			if checkErr == nil && checkExit == 0 {
				log.Printf("[ddev] git clone for %q exited with status %d but repo exists, treating as success", name, exitCode)
				return strings.TrimSpace(output), nil
			}
			if strings.TrimSpace(output) != "" {
				return "", fmt.Errorf("%s", strings.TrimSpace(output))
			}
			return "", fmt.Errorf("git clone failed with exit status %d", exitCode)
		}
		return strings.TrimSpace(output), nil
	}

	// ── WSL backend (Windows): spawn a fresh wsl.exe process ──
	// This avoids the shared WSL shell mutex so the main shell stays
	// free for start/stop/list during long-running clones.
	if stdruntime.GOOS == "windows" && d.activeBackend() == "wsl" {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// Use bash -c so $HOME expands properly inside WSL.
		// Merge stderr via 2>&1 inside bash to avoid pipe fd leaks that keep wsl.exe alive.
		// Use $HOME instead of ~ to allow safe quoting of the target directory.
		quotedTargetDir := "$HOME/ddev-projects/" + shellQuote(name)
		cloneCmd := fmt.Sprintf("mkdir -p $HOME/ddev-projects && git clone %s %s 2>&1", shellQuote(repoURL), quotedTargetDir)
		wslArgs := []string{"-d", d.WSLDistro(), "-e", "bash", "-c", cloneCmd}
		cmd := exec.CommandContext(ctx, "wsl.exe", wslArgs...)
		HideWSLWindow(cmd)

		stdoutPipe, _ := cmd.StdoutPipe()

		if err := cmd.Start(); err != nil {
			return "", err
		}

		var mu sync.Mutex
		var lines []string
		scanner := bufio.NewScanner(stdoutPipe)
		scanner.Buffer(make([]byte, 0, 256*1024), 1024*1024)
		scanner.Split(scanLinesOrCR)

		for scanner.Scan() {
			line := scanner.Text()
			mu.Lock()
			lines = append(lines, line)
			mu.Unlock()
			if d.ctx != nil {
				wruntime.EventsEmit(d.ctx, "ddev:output", line)
			}
		}

		if err := cmd.Wait(); err != nil {
			// Check if the repo was actually cloned despite the non-zero exit
			checkCmd := exec.CommandContext(context.Background(), "wsl.exe", "-d", d.WSLDistro(), "-e", "test", "-d", targetDir+"/.git")
			HideWSLWindow(checkCmd)
			if checkErr := checkCmd.Run(); checkErr == nil {
				log.Printf("[ddev] git clone for %q exited with error but repo exists, treating as success: %v", name, err)
				mu.Lock()
				out := strings.Join(lines, "\n")
				mu.Unlock()
				return strings.TrimSpace(out), nil
			}
			mu.Lock()
			out := strings.Join(lines, "\n")
			mu.Unlock()
			combined := strings.TrimSpace(out)
			if combined != "" {
				return "", fmt.Errorf("%s", combined)
			}
			return "", err
		}
		mu.Lock()
		out := strings.Join(lines, "\n")
		mu.Unlock()
		return strings.TrimSpace(out), nil
	}

	// ── Local backend ──
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "clone", repoURL, targetDir)
	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf

	if err := cmd.Run(); err != nil {
		// Check if the repo was actually cloned despite the non-zero exit
		if _, statErr := os.Stat(filepath.Join(targetDir, ".git")); statErr == nil {
			log.Printf("[ddev] git clone for %q exited with error but repo exists, treating as success: %v", name, err)
			return strings.TrimSpace(outBuf.String()), nil
		}
		combined := strings.TrimSpace(outBuf.String())
		if combined != "" {
			return "", fmt.Errorf("%s", combined)
		}
		return "", err
	}
	return strings.TrimSpace(outBuf.String()), nil
}

// InstallDdev downloads the latest DDEV Windows installer from GitHub and launches it.
// On non-Windows platforms, returns an error instructing manual install.
func (d *DdevService) InstallDdev() (string, error) {
	if runtime.GOOS != "windows" {
		return "", errors.New("automatic installer is only supported on Windows")
	}

	// 1) Query latest release
	client := &http.Client{Timeout: 30 * time.Second}
	rel, err := getLatestDdevRelease(client)
	if err != nil {
		return "", err
	}

	// 2) Determine target path
	fname := filepath.Join(os.TempDir(), "ddev-installer.exe")

	// 3) Check cache
	var expectedHash string
	if rel.ChecksumURL != "" {
		expectedHash, _ = fetchExpectedChecksum(client, rel.ChecksumURL, rel.AssetName)
	}

	if _, err := os.Stat(fname); err == nil && expectedHash != "" {
		if d.verifyInstallerChecksum(fname, expectedHash) {
			if err := launchWindowsElevated(fname); err != nil {
				if errors.Is(err, errUserCancelled) {
					return "", errUserCancelled
				}
				return "", fmt.Errorf("failed to launch installer: %w", err)
			}
			return fmt.Sprintf("Launched DDEV installer (%s). Follow the on-screen prompts.", rel.TagName), nil
		}
	}

	// 4) Download if not cached or cache invalid
	if err := d.downloadInstaller(rel.URL, fname); err != nil {
		return "", err
	}

	// 5) Verify checksum
	if expectedHash != "" {
		if d.ctx != nil {
			wruntime.EventsEmit(d.ctx, "ddev:output", "Verifying installer checksum…")
		}
		if !d.verifyInstallerChecksum(fname, expectedHash) {
			_ = os.Remove(fname) // delete corrupt file
			return "", errors.New("checksum mismatch - download may be corrupted, please try again")
		}
		if d.ctx != nil {
			wruntime.EventsEmit(d.ctx, "ddev:output", "Checksum verified ✓")
		}
	}

	// 6) Launch installer (elevated)
	if err := launchWindowsElevated(fname); err != nil {
		if errors.Is(err, errUserCancelled) {
			return "", errUserCancelled
		}
		return "", fmt.Errorf("failed to launch installer: %w", err)
	}
	return fmt.Sprintf("Launched DDEV installer (%s). If prompted by Windows UAC, click Yes and follow the on-screen prompts.", rel.TagName), nil
}

// ddevRelease represents a DDEV release from GitHub.
type ddevRelease struct {
	TagName     string
	URL         string
	AssetName   string
	ChecksumURL string
}

func preferredWindowsInstallerArch() string {
	if stdruntime.GOARCH == "arm64" {
		return "arm64"
	}
	return "amd64"
}

func windowsInstallerArchFromName(name string) string {
	if strings.Contains(name, "arm64") {
		return "arm64"
	}
	if strings.Contains(name, "amd64") || strings.Contains(name, "x86_64") {
		return "amd64"
	}
	return ""
}

func isWindowsInstallerAsset(name string) bool {
	return strings.HasSuffix(name, ".exe") &&
		strings.Contains(name, "windows") &&
		strings.Contains(name, "installer") &&
		strings.Contains(name, "ddev")
}

// getLatestDdevRelease queries the GitHub API for the latest DDEV release
// and selects the appropriate Windows installer asset.
func getLatestDdevRelease(client *http.Client) (*ddevRelease, error) {
	return getLatestDdevReleaseFromURL(client, "https://api.github.com/repos/drud/ddev/releases/latest")
}

func getLatestDdevReleaseFromURL(client *http.Client, url string) (*ddevRelease, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "ddev-gui/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query GitHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var payload struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read GitHub response: %w", err)
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse GitHub response: %w", err)
	}

	var rel ddevRelease
	rel.TagName = payload.TagName
	preferredArch := preferredWindowsInstallerArch()
	var fallbackURL string
	var fallbackAssetName string

	for _, a := range payload.Assets {
		name := strings.ToLower(a.Name)
		if isWindowsInstallerAsset(name) {
			assetArch := windowsInstallerArchFromName(name)
			if assetArch == preferredArch && rel.URL == "" {
				rel.URL = a.BrowserDownloadURL
				rel.AssetName = a.Name
			}
			if assetArch == "" && fallbackURL == "" {
				fallbackURL = a.BrowserDownloadURL
				fallbackAssetName = a.Name
			}
		}
		if (strings.Contains(name, "checksums") || strings.Contains(name, "checksum")) && strings.HasSuffix(name, ".txt") {
			rel.ChecksumURL = a.BrowserDownloadURL
		}
	}

	if rel.URL == "" && fallbackURL != "" {
		rel.URL = fallbackURL
		rel.AssetName = fallbackAssetName
	}

	if rel.URL == "" {
		return nil, fmt.Errorf("could not find Windows %s installer asset in latest release", preferredArch)
	}

	return &rel, nil
}

// downloadInstaller downloads the installer from the given URL to the destination path,
// reporting progress via Wails events.
func (d *DdevService) downloadInstaller(url, dest string) error {
	client := &http.Client{} // No strict timeout for large file download
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "ddev-gui/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("download error: %s", resp.Status)
	}

	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("save failed: %w", err)
	}
	defer f.Close()

	totalBytes := resp.ContentLength
	var downloaded int64
	buf := make([]byte, 32*1024)
	lastEmit := time.Now().Add(-time.Second)

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, wErr := f.Write(buf[:n]); wErr != nil {
				return fmt.Errorf("write failed: %w", wErr)
			}
			downloaded += int64(n)
			if d.ctx != nil && time.Since(lastEmit) >= 250*time.Millisecond {
				if totalBytes > 0 {
					pct := float64(downloaded) / float64(totalBytes) * 100
					wruntime.EventsEmit(d.ctx, "ddev:output",
						fmt.Sprintf("Downloading installer… %.1f MB / %.1f MB (%.0f%%)",
							float64(downloaded)/(1024*1024),
							float64(totalBytes)/(1024*1024),
							pct))
				} else {
					wruntime.EventsEmit(d.ctx, "ddev:output",
						fmt.Sprintf("Downloading installer… %.1f MB",
							float64(downloaded)/(1024*1024)))
				}
				lastEmit = time.Now()
			}
		}
		if readErr != nil {
			if readErr != io.EOF {
				return fmt.Errorf("download failed: %w", readErr)
			}
			break
		}
	}

	// Final progress message
	if d.ctx != nil {
		wruntime.EventsEmit(d.ctx, "ddev:output",
			fmt.Sprintf("Download complete - %.1f MB", float64(downloaded)/(1024*1024)))
	}

	return nil
}

// verifyInstallerChecksum computes the SHA-256 hash of fname and compares it with expectedHash.
func (d *DdevService) verifyInstallerChecksum(fname, expectedHash string) bool {
	if expectedHash == "" {
		return false
	}
	localSum, err := fileSHA256(fname)
	if err != nil {
		return false
	}
	return strings.EqualFold(expectedHash, localSum)
}
