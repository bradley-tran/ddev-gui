package backend

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

// SSHConfig holds the connection parameters for a remote SSH host.
type SSHConfig struct {
	Host    string
	Port    string
	User    string
	KeyPath string // path to private key; empty = use ssh-agent / default key
}

// SSHShell manages a persistent SSH connection and executes commands via
// successive Session.Run calls. It is the SSH analogue of WSLShell.
type SSHShell struct {
	mu     sync.Mutex
	client *ssh.Client
	cfg    SSHConfig
	alive  bool
}

// NewSSHShell creates a new SSHShell configured with the given settings.
// The underlying SSH connection is not established until the first Exec call.
func NewSSHShell(cfg SSHConfig) *SSHShell {
	if cfg.Port == "" {
		cfg.Port = "22"
	}
	return &SSHShell{cfg: cfg}
}

// SshShellFromConfig reads SSH settings from the config and creates an SSHShell.
func SshShellFromConfig(cfg *ConfigService) *SSHShell {
	raw := cfg.Get("ssh")
	m, ok := raw.(map[string]any)
	if !ok {
		return NewSSHShell(SSHConfig{})
	}
	host, _ := m["host"].(string)
	port, _ := m["port"].(string)
	user, _ := m["user"].(string)
	keyPath, _ := m["keyPath"].(string)
	return NewSSHShell(SSHConfig{Host: host, Port: port, User: user, KeyPath: keyPath})
}

// UpdateConfig replaces the SSH configuration and forces a reconnect on the
// next Exec call. This is safe to call while no command is running.
func (s *SSHShell) UpdateConfig(cfg SSHConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cfg.Port == "" {
		cfg.Port = "22"
	}
	// Close existing connection if config changed
	if s.alive && s.client != nil {
		_ = s.client.Close()
		s.alive = false
		s.client = nil
	}
	s.cfg = cfg
}

// ensureConnected establishes the SSH connection if not already connected.
// Must be called with s.mu held.
func (s *SSHShell) ensureConnected() error {
	if s.alive && s.client != nil {
		// Quick health check: send a keepalive-style request
		_, _, err := s.client.SendRequest("keepalive@ddev-gui", true, nil)
		if err == nil {
			return nil
		}
		log.Println("[sshshell] connection lost, reconnecting")
		_ = s.client.Close()
		s.alive = false
		s.client = nil
	}

	authMethods, err := s.buildAuthMethods()
	if err != nil {
		return fmt.Errorf("sshshell: auth setup: %w", err)
	}

	addr := net.JoinHostPort(s.cfg.Host, s.cfg.Port)

	hostKeyCallback, err := s.buildHostKeyCallback()
	if err != nil {
		log.Printf("[sshshell] known_hosts not available, using InsecureIgnoreHostKey: %v", err)
		hostKeyCallback = ssh.InsecureIgnoreHostKey()
	}

	config := &ssh.ClientConfig{
		User:            s.cfg.User,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         15 * time.Second,
	}

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		// If the error is a known_hosts key mismatch, retry without host key checking
		if strings.Contains(err.Error(), "key mismatch") || strings.Contains(err.Error(), "knownhosts") {
			log.Printf("[sshshell] host key mismatch for %s - retrying without host key verification (update your known_hosts to fix this)", addr)
			config.HostKeyCallback = ssh.InsecureIgnoreHostKey()
			client, err = ssh.Dial("tcp", addr, config)
			if err != nil {
				return fmt.Errorf("sshshell: dial %s: %w", addr, err)
			}
		} else {
			return fmt.Errorf("sshshell: dial %s: %w", addr, err)
		}
	}

	s.client = client
	s.alive = true
	log.Printf("[sshshell] connected to %s@%s", s.cfg.User, addr)
	return nil
}

// EnsureSSHShell creates or updates the SSH shell based on current config.
// Called when settings change so a reconnect happens with new credentials.
func (d *DdevService) EnsureSSHShell() {
	if d.config == nil {
		return
	}
	backend, _ := d.config.Get("backend").(string)
	if backend != "ssh" {
		return
	}
	if d.sshShell == nil {
		d.sshShell = SshShellFromConfig(d.config)
	} else {
		// Update config on existing shell (forces reconnect)
		raw := d.config.Get("ssh")
		m, ok := raw.(map[string]any)
		if ok {
			host, _ := m["host"].(string)
			port, _ := m["port"].(string)
			user, _ := m["user"].(string)
			keyPath, _ := m["keyPath"].(string)
			d.sshShell.UpdateConfig(SSHConfig{Host: host, Port: port, User: user, KeyPath: keyPath})
		}
	}
}

// buildAuthMethods constructs SSH auth methods from the config.
// Priority: explicit key file → default key files → ssh-agent.
func (s *SSHShell) buildAuthMethods() ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod

	// 1. Explicit key file
	if s.cfg.KeyPath != "" {
		expanded := expandHome(s.cfg.KeyPath)
		signer, err := signerFromKeyFile(expanded)
		if err != nil {
			return nil, fmt.Errorf("key file %s: %w", expanded, err)
		}
		methods = append(methods, ssh.PublicKeys(signer))
	}

	// 2. Default key files (~/.ssh/id_rsa, id_ed25519, etc.)
	if s.cfg.KeyPath == "" {
		home, _ := os.UserHomeDir()
		if home != "" {
			for _, name := range []string{"id_ed25519", "id_rsa", "id_ecdsa"} {
				keyPath := filepath.Join(home, ".ssh", name)
				if _, err := os.Stat(keyPath); err == nil {
					signer, err := signerFromKeyFile(keyPath)
					if err == nil {
						methods = append(methods, ssh.PublicKeys(signer))
						break // use the first one that works
					}
				}
			}
		}
	}

	// 3. SSH agent (Windows: named pipe; Unix: SSH_AUTH_SOCK)
	if agentAuth := sshAgentAuth(); agentAuth != nil {
		methods = append(methods, agentAuth)
	}

	if len(methods) == 0 {
		return nil, fmt.Errorf("no SSH authentication methods available - configure a key path or ensure ssh-agent is running")
	}
	return methods, nil
}

// buildHostKeyCallback tries to use the user's known_hosts file for host key verification.
func (s *SSHShell) buildHostKeyCallback() (ssh.HostKeyCallback, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	knownHostsPath := filepath.Join(home, ".ssh", "known_hosts")
	if _, err := os.Stat(knownHostsPath); err != nil {
		return nil, err
	}
	return knownhosts.New(knownHostsPath)
}

// Exec runs a command over the SSH connection. It opens a new SSH session for
// each command (SSH protocol requires this), but reuses the underlying TCP connection.
//
// The parameters and return values mirror WSLShell.Exec exactly:
//   - dir: optional working directory on the remote host
//   - args: command and arguments (e.g. ["ddev", "start", "myproject"])
//   - envVars: optional env var assignments (e.g. ["COMPOSER_ALLOW_SUPERUSER=1"])
//   - timeout: max duration to wait for the command
//   - onLine: called for each line of output as it arrives (may be nil)
//
// Returns combined stdout+stderr, exit code, and any error.
func (s *SSHShell) Exec(dir string, args []string, envVars []string, timeout time.Duration, onLine func(string)) (string, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureConnected(); err != nil {
		return "", -1, err
	}

	session, err := s.client.NewSession()
	if err != nil {
		// Connection may have died - mark for reconnect
		s.alive = false
		return "", -1, fmt.Errorf("sshshell: new session: %w", err)
	}
	defer session.Close()

	// Build the remote command string
	cmdStr := s.buildCommand(dir, args, envVars)

	// Use a unique marker to extract the exit code, same approach as WSLShell
	id := uuid.New().String()
	markerPrefix := "<<<EXIT:"
	markerSuffix := fmt.Sprintf(":%s>>>", id)

	// Wrap: run command, then echo exit code with marker
	wrappedCmd := fmt.Sprintf("%s; echo '<<<EXIT:'$?':%s>>>'", cmdStr, id)

	// Set up stdout pipe (merge stderr into stdout via the shell)
	stdout, err := session.StdoutPipe()
	if err != nil {
		return "", -1, fmt.Errorf("sshshell: stdout pipe: %w", err)
	}
	session.Stderr = session.Stdout // merge stderr

	// Start the command
	if err := session.Start(fmt.Sprintf("bash -c %s", shellQuote(wrappedCmd))); err != nil {
		return "", -1, fmt.Errorf("sshshell: start: %w", err)
	}

	// Read output with timeout
	var outputLines []string
	exitCode := -1

	done := make(chan struct{})
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	go func() {
		defer close(done)
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 0, 256*1024), 1024*1024)
		scanner.Split(scanLinesOrCR) // reuse WSLShell's split function

		for scanner.Scan() {
			line := scanner.Text()

			// Check for exit marker
			if strings.HasPrefix(line, markerPrefix) && strings.HasSuffix(line, markerSuffix) {
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
	}()

	select {
	case <-done:
		// Command completed
	case <-ctx.Done():
		// Timeout - close session to unblock the scanner
		_ = session.Close()
		s.alive = false
		return strings.Join(outputLines, "\n"), -1, fmt.Errorf("sshshell: command timed out after %v", timeout)
	}

	// Wait for the SSH session to finish
	_ = session.Wait()

	output := strings.Join(outputLines, "\n")
	return output, exitCode, nil
}

// buildCommand constructs a bash command string from dir, args, and envVars.
func (s *SSHShell) buildCommand(dir string, args []string, envVars []string) string {
	var parts []string

	if dir != "" {
		expandedDir := dir
		if strings.HasPrefix(expandedDir, "~/") {
			expandedDir = "$HOME/" + expandedDir[2:]
		} else if expandedDir == "~" {
			expandedDir = "$HOME"
		}
		parts = append(parts, fmt.Sprintf("cd %s 2>/dev/null || true", expandedDir))
	}

	for _, ev := range envVars {
		parts = append(parts, "export "+ev)
	}

	quotedArgs := make([]string, len(args))
	for i, a := range args {
		quotedArgs[i] = shellQuote(a)
	}
	parts = append(parts, strings.Join(quotedArgs, " ")+" 2>&1")

	return strings.Join(parts, "; ")
}

// Close shuts down the SSH connection.
func (s *SSHShell) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.alive || s.client == nil {
		return nil
	}

	log.Println("[sshshell] closing SSH connection")
	s.alive = false
	err := s.client.Close()
	s.client = nil
	return err
}

// ── Helpers ──────────────────────────────────────────────────────────────────

// signerFromKeyFile reads a private key file and returns an ssh.Signer.
func signerFromKeyFile(path string) (ssh.Signer, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(raw)
	if err != nil {
		return nil, err
	}
	return signer, nil
}

// sshAgentAuth returns an ssh.AuthMethod from the running SSH agent, or nil.
func sshAgentAuth() ssh.AuthMethod {
	sock := os.Getenv("SSH_AUTH_SOCK")
	if sock == "" {
		return nil
	}
	conn, err := net.Dial("unix", sock)
	if err != nil {
		return nil
	}
	return ssh.PublicKeysCallback(agent.NewClient(conn).Signers)
}

// expandHome expands a leading ~ to the user's home directory.
func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") || path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}
