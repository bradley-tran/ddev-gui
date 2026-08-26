package backend

import (
	"net"
	"path/filepath"
	"testing"

	"golang.org/x/crypto/ssh"
)

type mockSSHConn struct {
	closed bool
}

func (m *mockSSHConn) User() string          { return "" }
func (m *mockSSHConn) SessionID() []byte     { return nil }
func (m *mockSSHConn) ClientVersion() []byte { return nil }
func (m *mockSSHConn) ServerVersion() []byte { return nil }
func (m *mockSSHConn) RemoteAddr() net.Addr  { return &net.TCPAddr{} }
func (m *mockSSHConn) LocalAddr() net.Addr   { return &net.TCPAddr{} }
func (m *mockSSHConn) Close() error {
	m.closed = true
	return nil
}
func (m *mockSSHConn) SendRequest(name string, wantReply bool, payload []byte) (bool, []byte, error) {
	return false, nil, nil
}
func (m *mockSSHConn) OpenChannel(name string, data []byte) (ssh.Channel, <-chan *ssh.Request, error) {
	return nil, nil, nil
}
func (m *mockSSHConn) Wait() error { return nil }

func TestNewSSHShell(t *testing.T) {
	tests := []struct {
		name     string
		cfg      SSHConfig
		expected SSHConfig
	}{
		{
			name: "Port provided",
			cfg: SSHConfig{
				Host: "example.com",
				Port: "2222",
				User: "testuser",
			},
			expected: SSHConfig{
				Host: "example.com",
				Port: "2222",
				User: "testuser",
			},
		},
		{
			name: "Port not provided (fallback to 22)",
			cfg: SSHConfig{
				Host: "example.com",
				User: "testuser",
			},
			expected: SSHConfig{
				Host: "example.com",
				Port: "22",
				User: "testuser",
			},
		},
		{
			name: "Empty config (fallback to 22)",
			cfg:  SSHConfig{},
			expected: SSHConfig{
				Port: "22",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shell := NewSSHShell(tt.cfg)
			if shell.cfg.Port != tt.expected.Port {
				t.Errorf("NewSSHShell() expected Port = %v, got %v", tt.expected.Port, shell.cfg.Port)
			}
			if shell.cfg.Host != tt.expected.Host {
				t.Errorf("NewSSHShell() expected Host = %v, got %v", tt.expected.Host, shell.cfg.Host)
			}
			if shell.cfg.User != tt.expected.User {
				t.Errorf("NewSSHShell() expected User = %v, got %v", tt.expected.User, shell.cfg.User)
			}
		})
	}
}

func TestUpdateConfig_ActiveConnection(t *testing.T) {
	shell := NewSSHShell(SSHConfig{Host: "oldhost", Port: "22"})
	mock := &mockSSHConn{}
	shell.client = ssh.NewClient(mock, nil, nil)
	shell.alive = true

	shell.UpdateConfig(SSHConfig{Host: "newhost", Port: "2222"})

	if shell.alive {
		t.Error("Expected alive to be false")
	}
	if shell.client != nil {
		t.Error("Expected client to be nil")
	}
	if !mock.closed {
		t.Error("Expected underlying connection to be closed")
	}
	if shell.cfg.Host != "newhost" {
		t.Errorf("Expected Host = newhost, got %s", shell.cfg.Host)
	}
	if shell.cfg.Port != "2222" {
		t.Errorf("Expected Port = 2222, got %s", shell.cfg.Port)
	}
}

func TestUpdateConfig(t *testing.T) {
	tests := []struct {
		name     string
		initial  SSHConfig
		update   SSHConfig
		expected SSHConfig
	}{
		{
			name: "Update with port provided",
			initial: SSHConfig{
				Host: "example.com",
				Port: "2222",
			},
			update: SSHConfig{
				Host: "new.example.com",
				Port: "2223",
			},
			expected: SSHConfig{
				Host: "new.example.com",
				Port: "2223",
			},
		},
		{
			name: "Update with no port (fallback to 22)",
			initial: SSHConfig{
				Host: "example.com",
				Port: "2222",
			},
			update: SSHConfig{
				Host: "new.example.com",
			},
			expected: SSHConfig{
				Host: "new.example.com",
				Port: "22",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shell := NewSSHShell(tt.initial)
			shell.UpdateConfig(tt.update)
			if shell.cfg.Port != tt.expected.Port {
				t.Errorf("UpdateConfig() expected Port = %v, got %v", tt.expected.Port, shell.cfg.Port)
			}
			if shell.cfg.Host != tt.expected.Host {
				t.Errorf("UpdateConfig() expected Host = %v, got %v", tt.expected.Host, shell.cfg.Host)
			}
		})
	}
}

func TestExpandHome(t *testing.T) {
	// Setup: Predictable user home directory
	originalHome := t.TempDir()
	t.Setenv("HOME", originalHome)
	t.Setenv("USERPROFILE", originalHome) // Windows compatibility for tests generally relies on USERPROFILE, but x/crypto/ssh and os.UserHomeDir typically check HOME or USERPROFILE. Setenv "HOME" works for this UNIX-centric test setup.

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty path",
			input:    "",
			expected: "",
		},
		{
			name:     "Path without tilde",
			input:    "/var/www/html",
			expected: "/var/www/html",
		},
		{
			name:     "Exact tilde",
			input:    "~",
			expected: originalHome,
		},
		{
			name:     "Tilde with slash",
			input:    "~/",
			expected: originalHome,
		},
		{
			name:     "Tilde with subpath",
			input:    "~/project/test",
			expected: originalHome + "/project/test", // Note: filepath.Join will handle OS-specific separators.
		},
		{
			name:     "Tilde not at start",
			input:    "/path/to/~",
			expected: "/path/to/~",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandHome(tt.input)

			// We need to normalize expected path for Windows using filepath.Join if needed,
			// but since originalHome is from TempDir, it already has the correct separators.
			// However, our expected strings hardcode "/", so let's adjust them.
			expected := tt.expected
			if tt.input == "~/project/test" {
				expected = originalHome + string(filepath.Separator) + "project" + string(filepath.Separator) + "test"
			}

			if result != expected {
				t.Errorf("expandHome(%q) = %q, want %q", tt.input, result, expected)
			}
		})
	}
}
