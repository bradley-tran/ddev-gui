package backend

import (
	"fmt"
	"runtime"
	"testing"
)

func TestLaunchWindowsElevated_Other(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping non-windows test on windows")
	}

	expectedErr := fmt.Errorf("elevation only supported on Windows")
	err := launchWindowsElevated("dummy_path.exe")

	if err == nil || err.Error() != expectedErr.Error() {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
}
