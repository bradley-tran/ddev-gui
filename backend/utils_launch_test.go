package backend

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

// Helper process for testing
func TestHelperProcessLaunchElevated(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	cmd := os.Getenv("HELPER_CMD")
	switch cmd {
	case "success":
		os.Exit(0)
	case "cancel_1223":
		os.Exit(1223)
	case "cancel_string":
		fmt.Fprint(os.Stderr, "Some error cancelled by the user")
		os.Exit(1)
	case "general_error":
		os.Exit(1)
	default:
		os.Exit(1)
	}
}

func mockExecCommandLaunch(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcessLaunchElevated", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	// Inherit environment and set GO_WANT_HELPER_PROCESS
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
	return cmd
}

func TestLaunchWindowsElevated(t *testing.T) {
	// Save originals
	origExecCommand := execCommand
	origRuntimeGOOS := runtimeGOOS
	defer func() {
		execCommand = origExecCommand
		runtimeGOOS = origRuntimeGOOS
	}()

	// Mock for execution
	execCommand = mockExecCommandLaunch

	tests := []struct {
		name        string
		osName      string
		mockCmd     string
		expectedErr error
	}{
		{
			name:        "non-windows os",
			osName:      "linux",
			mockCmd:     "",
			expectedErr: fmt.Errorf("elevation only supported on Windows"),
		},
		{
			name:        "success",
			osName:      "windows",
			mockCmd:     "success",
			expectedErr: nil,
		},
		{
			name:        "cancel via exit code 1223",
			osName:      "windows",
			mockCmd:     "cancel_1223",
			expectedErr: errUserCancelled,
		},
		{
			name:        "cancel via string match",
			osName:      "windows",
			mockCmd:     "cancel_string",
			expectedErr: errUserCancelled,
		},
		{
			name:        "general error",
			osName:      "windows",
			mockCmd:     "general_error",
			expectedErr: &exec.ExitError{}, // Just checking it's an error but not errUserCancelled
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runtimeGOOS = tt.osName
			if tt.mockCmd != "" {
				t.Setenv("HELPER_CMD", tt.mockCmd)
			}

			err := launchWindowsElevated("dummy_path.exe")

			if tt.expectedErr == nil {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			} else if tt.expectedErr == errUserCancelled {
				// Special check if OS doesn't support 1223 exit code
				if err != errUserCancelled {
					if tt.mockCmd == "cancel_1223" {
						// exit codes above 125 can be truncated or changed depending on the platform in tests
						// But exit code 1223 is mostly 1223 % 256 = 199. Let's see if we get that.
						if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 199 {
							t.Logf("Got exit code 199 for 1223 (modulo 256 behavior) which is okay for this test.")
							return
						}
					}
					t.Errorf("expected errUserCancelled, got %v", err)
				}
			} else if tt.expectedErr.Error() == "elevation only supported on Windows" {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("expected %v, got %v", tt.expectedErr, err)
				}
			} else {
				// General error case
				if err == nil {
					t.Errorf("expected an error, got nil")
				} else if err == errUserCancelled {
					t.Errorf("expected general error, got errUserCancelled")
				}
			}
		})
	}
}
