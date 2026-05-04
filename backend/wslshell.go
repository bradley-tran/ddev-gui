package backend

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// WSLShell manages a persistent bash session inside WSL, allowing commands
// to be sent via stdin and their output captured without spawning a new
// wsl.exe process for each invocation.
type WSLShell struct {
	mu      sync.Mutex
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	scanner *bufio.Scanner
	distro  string
	alive   bool
}

// NewWSLShell creates a new WSLShell configured for the given WSL distro.
// The underlying bash process is not started until the first Exec call.
func NewWSLShell(distro string) *WSLShell {
	return &WSLShell{distro: distro}
}

// ensureRunning starts the bash process if it is not already running.
// Must be called with w.mu held.
func (w *WSLShell) ensureRunning() error {
	if w.alive && w.cmd != nil && w.cmd.Process != nil {
		// Quick health check: if the process has already exited, restart.
		// Process.Signal(0) is not available on Windows, so we check ProcessState.
		if w.cmd.ProcessState != nil && w.cmd.ProcessState.Exited() {
			log.Println("[wslshell] process exited, restarting")
			w.alive = false
		} else {
			return nil
		}
	}

	// Start a new bash process inside WSL.
	// --norc --noprofile avoids user shell config that could print unexpected output.
	// If distro is empty, omit -d to use the user's default WSL distro.
	var wslArgs []string
	if w.distro != "" {
		wslArgs = []string{"-d", w.distro, "-e", "bash", "--norc", "--noprofile"}
	} else {
		wslArgs = []string{"-e", "bash", "--norc", "--noprofile"}
	}
	w.cmd = exec.Command("wsl.exe", wslArgs...)
	HideWSLWindow(w.cmd)

	var err error
	w.stdin, err = w.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("wslshell: stdin pipe: %w", err)
	}

	stdoutPipe, err := w.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("wslshell: stdout pipe: %w", err)
	}

	// Merge stderr into stdout so we get a single ordered stream.
	w.cmd.Stderr = w.cmd.Stdout

	w.scanner = bufio.NewScanner(stdoutPipe)
	// Increase scanner buffer for large DDEV output (e.g. JSON lists).
	w.scanner.Buffer(make([]byte, 0, 256*1024), 1024*1024)
	// Use a custom split function that also splits on \r (carriage return).
	// This prevents hangs on Composer progress bars and other tools that
	// use \r to update output in-place without a trailing \n.
	w.scanner.Split(scanLinesOrCR)

	if err := w.cmd.Start(); err != nil {
		return fmt.Errorf("wslshell: start: %w", err)
	}

	w.alive = true
	distroLabel := w.distro
	if distroLabel == "" {
		distroLabel = "(default)"
	}
	log.Printf("[wslshell] started persistent shell (PID %d, distro %s)", w.cmd.Process.Pid, distroLabel)

	// Send a no-op command and wait for its marker to confirm the shell is ready.
	// This drains any login banner or MOTD output.
	readyID := uuid.New().String()
	marker := fmt.Sprintf("<<<EXIT:0:%s>>>", readyID)
	initCmd := fmt.Sprintf("echo '<<<EXIT:0:%s>>>'\n", readyID)
	if _, err := io.WriteString(w.stdin, initCmd); err != nil {
		w.alive = false
		return fmt.Errorf("wslshell: init write: %w", err)
	}
	// Drain until we see the ready marker.
	for w.scanner.Scan() {
		if strings.TrimSpace(w.scanner.Text()) == marker {
			break
		}
	}
	if err := w.scanner.Err(); err != nil {
		w.alive = false
		return fmt.Errorf("wslshell: init drain: %w", err)
	}

	return nil
}

// Exec runs a command inside the persistent WSL shell.
//
// dir is an optional working directory (WSL path like ~/ddev-projects/foo).
// args are the command arguments (e.g. ["ddev", "start", "myproject"]).
// envVars are optional shell env var assignments prepended before the command
// (e.g. ["COMPOSER_ALLOW_SUPERUSER=1"]). They are exported in the bash session.
// timeout controls how long to wait for the command to complete.
// onLine is called for each line of output as it arrives (may be nil).
//
// Returns the combined stdout+stderr, the exit code, and any error.
// If the shell process dies mid-command, it returns an error and marks the
// shell for restart on the next call.
func (w *WSLShell) Exec(dir string, args []string, envVars []string, timeout time.Duration, onLine func(string)) (string, int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.ensureRunning(); err != nil {
		return "", -1, err
	}

	id := uuid.New().String()
	markerPrefix := "<<<EXIT:"
	markerSuffix := fmt.Sprintf(":%s>>>", id)

	// Build the command string.
	// 1. Optionally cd to the working directory.
	// 2. Run the command with stderr merged via 2>&1.
	// 3. Capture $? and echo the delimiter with exit code.
	var cmdParts []string
	if dir != "" {
		// Expand ~ to $HOME before quoting, because single-quoted ~ won't expand in bash.
		var expandedDir string
		if strings.HasPrefix(dir, "~/") {
			expandedDir = "$HOME/" + shellQuote(dir[2:])
		} else if dir == "~" {
			expandedDir = "$HOME"
		} else {
			expandedDir = shellQuote(dir)
		}
		// Use cd with || true so a missing directory doesn't abort the whole line.
		cmdParts = append(cmdParts, fmt.Sprintf("cd %s 2>/dev/null || true", expandedDir))
	}
	// Prepend any env var exports before the command.
	for _, ev := range envVars {
		parts := strings.SplitN(ev, "=", 2)
		if len(parts) == 2 {
			cmdParts = append(cmdParts, fmt.Sprintf("export %s=%s", parts[0], shellQuote(parts[1])))
		} else {
			cmdParts = append(cmdParts, "export "+shellQuote(ev))
		}
	}

	// Shell-quote each argument to handle spaces and special characters.
	quotedArgs := make([]string, len(args))
	for i, a := range args {
		quotedArgs[i] = shellQuote(a)
	}
	cmdParts = append(cmdParts, strings.Join(quotedArgs, " ")+" 2>&1")

	// The final line captures the exit code and echoes the marker.
	fullCmd := strings.Join(cmdParts, "; ") + fmt.Sprintf("; echo '<<<EXIT:'$?':%s>>>'", id)

	// Write the command to stdin.
	if _, err := io.WriteString(w.stdin, fullCmd+"\n"); err != nil {
		w.alive = false
		return "", -1, fmt.Errorf("wslshell: write command: %w", err)
	}

	// Read output lines until we see our marker.
	var outputLines []string
	exitCode := -1

	// Set up a timeout.
	done := make(chan struct{})
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	go func() {
		defer close(done)
		for w.scanner.Scan() {
			line := w.scanner.Text()

			// Check if this line is our exit marker.
			if strings.HasPrefix(line, markerPrefix) && strings.HasSuffix(line, markerSuffix) {
				// Extract exit code: <<<EXIT:<code>:<uuid>>>>
				inner := strings.TrimPrefix(line, markerPrefix)
				inner = strings.TrimSuffix(inner, markerSuffix)
				if code, err := strconv.Atoi(inner); err == nil {
					exitCode = code
				}
				return
			}

			outputLines = append(outputLines, line)
			if onLine != nil {
				onLine(line)
			}
		}
		// If scanner exits without finding the marker, the process died.
		w.alive = false
	}()

	select {
	case <-done:
		// Command completed.
	case <-ctx.Done():
		// Timeout - kill the shell process so we don't leak it.
		w.alive = false
		if w.cmd != nil && w.cmd.Process != nil {
			_ = w.cmd.Process.Kill()
		}
		return strings.Join(outputLines, "\n"), -1, fmt.Errorf("wslshell: command timed out after %v", timeout)
	}

	if err := w.scanner.Err(); err != nil {
		w.alive = false
		return strings.Join(outputLines, "\n"), exitCode, fmt.Errorf("wslshell: scanner error: %w", err)
	}

	output := strings.Join(outputLines, "\n")
	return output, exitCode, nil
}

// Close shuts down the persistent shell process.
func (w *WSLShell) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.alive || w.cmd == nil {
		return nil
	}

	log.Println("[wslshell] shutting down persistent shell")
	w.alive = false

	// Send exit command to terminate bash gracefully.
	_, _ = io.WriteString(w.stdin, "exit\n")
	_ = w.stdin.Close()

	// Give it a moment to exit, then force kill.
	done := make(chan error, 1)
	go func() { done <- w.cmd.Wait() }()

	select {
	case <-done:
		// Clean exit.
	case <-time.After(3 * time.Second):
		if w.cmd.Process != nil {
			_ = w.cmd.Process.Kill()
		}
	}
	return nil
}

// shellQuote wraps a string in single quotes for bash, escaping any
// embedded single quotes using the '\” idiom.
func shellQuote(s string) string {
	// If the string is simple (no special chars), skip quoting for readability.
	if s != "" && !strings.ContainsAny(s, " \t\n'\"\\$`!#&|;(){}[]<>?*~") {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// scanLinesOrCR is a bufio.SplitFunc that splits input on \n, \r\n, or bare \r.
// This handles commands like Composer that use \r for progress bar updates
// without a trailing \n, which would otherwise block Scanner.Scan() indefinitely.
func scanLinesOrCR(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	for i, b := range data {
		if b == '\n' {
			return i + 1, data[:i], nil
		}
		if b == '\r' {
			// \r\n counts as a single line break
			if i+1 < len(data) && data[i+1] == '\n' {
				return i + 2, data[:i], nil
			}
			return i + 1, data[:i], nil
		}
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}
