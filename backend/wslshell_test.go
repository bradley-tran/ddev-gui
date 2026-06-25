package backend

import (
	"testing"
)

func TestNewWSLShell(t *testing.T) {
	tests := []struct {
		name   string
		distro string
	}{
		{
			name:   "empty distro",
			distro: "",
		},
		{
			name:   "specific distro",
			distro: "Ubuntu",
		},
		{
			name:   "debian distro",
			distro: "Debian",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shell := NewWSLShell(tt.distro)

			if shell == nil {
				t.Fatalf("NewWSLShell returned nil")
			}

			if shell.distro != tt.distro {
				t.Errorf("expected distro %q, got %q", tt.distro, shell.distro)
			}

			if shell.alive {
				t.Errorf("expected alive to be false, got %t", shell.alive)
			}

			if shell.cmd != nil {
				t.Errorf("expected cmd to be nil, got %v", shell.cmd)
			}

			if shell.stdin != nil {
				t.Errorf("expected stdin to be nil, got %v", shell.stdin)
			}

			if shell.scanner != nil {
				t.Errorf("expected scanner to be nil, got %v", shell.scanner)
			}
		})
	}
}

func TestBuildWSLCommand(t *testing.T) {
	tests := []struct {
		name     string
		dir      string
		args     []string
		envVars  []string
		id       string
		expected string
	}{
		{
			name:     "empty dir, no args, no env",
			dir:      "",
			args:     []string{},
			envVars:  []string{},
			id:       "123",
			expected: " 2>&1; echo '<<<EXIT:'$?':123>>>'",
		},
		{
			name:     "simple command",
			dir:      "",
			args:     []string{"ls", "-la"},
			envVars:  []string{},
			id:       "123",
			expected: "ls -la 2>&1; echo '<<<EXIT:'$?':123>>>'",
		},
		{
			name:     "with specific directory",
			dir:      "/var/www/html",
			args:     []string{"pwd"},
			envVars:  []string{},
			id:       "123",
			expected: "cd /var/www/html 2>/dev/null || true; pwd 2>&1; echo '<<<EXIT:'$?':123>>>'",
		},
		{
			name:     "with tilde directory",
			dir:      "~",
			args:     []string{"ls"},
			envVars:  []string{},
			id:       "123",
			expected: "cd $HOME 2>/dev/null || true; ls 2>&1; echo '<<<EXIT:'$?':123>>>'",
		},
		{
			name:     "with tilde prefix directory",
			dir:      "~/projects/ddev",
			args:     []string{"ls"},
			envVars:  []string{},
			id:       "123",
			expected: "cd $HOME/projects/ddev 2>/dev/null || true; ls 2>&1; echo '<<<EXIT:'$?':123>>>'",
		},
		{
			name:     "with environment variables",
			dir:      "",
			args:     []string{"env"},
			envVars:  []string{"FOO=bar", "BAZ=qux quux"},
			id:       "123",
			expected: "export FOO=bar; export BAZ='qux quux'; env 2>&1; echo '<<<EXIT:'$?':123>>>'",
		},
		{
			name:     "with env var no value",
			dir:      "",
			args:     []string{"env"},
			envVars:  []string{"FOO"},
			id:       "123",
			expected: "export FOO; env 2>&1; echo '<<<EXIT:'$?':123>>>'",
		},
		{
			name:     "with special characters in args",
			dir:      "",
			args:     []string{"echo", "hello world", "it's me"},
			envVars:  []string{},
			id:       "123",
			expected: "echo 'hello world' 'it'\\''s me' 2>&1; echo '<<<EXIT:'$?':123>>>'",
		},
		{
			name:     "complex combination",
			dir:      "~/my project",
			args:     []string{"ddev", "start", "--all"},
			envVars:  []string{"DDEV_DEBUG=true", "PATH=/usr/local/bin:$PATH"},
			id:       "xyz",
			expected: "cd $HOME/'my project' 2>/dev/null || true; export DDEV_DEBUG=true; export PATH='/usr/local/bin:$PATH'; ddev start --all 2>&1; echo '<<<EXIT:'$?':xyz>>>'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := buildWSLCommand(tt.dir, tt.args, tt.envVars, tt.id)
			if actual != tt.expected {
				t.Errorf("buildWSLCommand()\nexpected: %q\nactual:   %q", tt.expected, actual)
			}
		})
	}
}
