package backend

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"
)

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	cmd := os.Getenv("HELPER_CMD")
	switch cmd {
	case "whoami_root":
		fmt.Print("root\n")
		os.Exit(0)
	case "whoami_jules":
		fmt.Print("jules\n")
		os.Exit(0)
	case "whoami_error":
		os.Exit(1)
	default:
		os.Exit(1)
	}
}

func mockExecCommandContext(ctx context.Context, command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.CommandContext(ctx, os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestResolveWSLProjectLocation(t *testing.T) {
	originalExecCommandContext := execCommandContext
	defer func() {
		execCommandContext = originalExecCommandContext
	}()

	tests := []struct {
		name       string
		location   string
		distro     string
		helperCmd  string
		expected   string
	}{
		{
			name:     "Empty location",
			location: "",
			distro:   "Ubuntu",
			expected: "",
		},
		{
			name:     "Empty distro",
			location: "~/projects/my-site",
			distro:   "",
			expected: "~/projects/my-site",
		},
		{
			name:     "Location not starting with ~/",
			location: "/mnt/c/projects/my-site",
			distro:   "Ubuntu",
			expected: "/mnt/c/projects/my-site",
		},
		{
			name:      "Root user",
			location:  "~/projects/my-site",
			distro:    "Ubuntu",
			helperCmd: "whoami_root",
			expected:  "/root/projects/my-site",
		},
		{
			name:      "Standard user",
			location:  "~/projects/my-site",
			distro:    "Ubuntu",
			helperCmd: "whoami_jules",
			expected:  "/home/jules/projects/my-site",
		},
		{
			name:      "Command fails",
			location:  "~/projects/my-site",
			distro:    "Ubuntu",
			helperCmd: "whoami_error",
			expected:  "/root/projects/my-site",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.helperCmd != "" {
				execCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
					cmd := mockExecCommandContext(ctx, name, args...)
					cmd.Env = append(cmd.Env, "HELPER_CMD="+tt.helperCmd)
					return cmd
				}
			} else {
				execCommandContext = originalExecCommandContext
			}

			actual := resolveWSLProjectLocation(tt.location, tt.distro)
			if actual != tt.expected {
				t.Errorf("resolveWSLProjectLocation(%q, %q) = %q, expected %q", tt.location, tt.distro, actual, tt.expected)
			}
		})
	}
}
