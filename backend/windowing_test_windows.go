//go:build windows

package backend

import (
	"errors"
	"testing"

	"golang.org/x/sys/windows"
)

func TestLaunchWindowsElevated(t *testing.T) {
	origShellExecute := shellExecute
	defer func() { shellExecute = origShellExecute }()

	tests := []struct {
		name        string
		mockErr     error
		expectedErr error
	}{
		{
			name:        "success",
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "cancel",
			mockErr:     windows.ERROR_ACCESS_DENIED,
			expectedErr: errUserCancelled,
		},
		{
			name:        "other error",
			mockErr:     errors.New("other error"),
			expectedErr: errors.New("ShellExecute failed: other error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shellExecute = func(hwnd windows.Handle, verb, file, args, cwd *uint16, showCmd int32) error {
				return tt.mockErr
			}

			err := launchWindowsElevated("dummy.exe")

			if tt.expectedErr == nil {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			} else if tt.expectedErr == errUserCancelled {
				if err != errUserCancelled {
					t.Errorf("expected errUserCancelled, got %v", err)
				}
			} else {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("expected %v, got %v", tt.expectedErr, err)
				}
			}
		})
	}
}
