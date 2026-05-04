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
