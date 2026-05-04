package backend

import (
	"testing"
)

func TestAppVersion(t *testing.T) {
	// Save the original values and restore them after the test
	originalAppVersionStr := AppVersionStr
	originalCommitHashStr := CommitHashStr
	defer func() {
		AppVersionStr = originalAppVersionStr
		CommitHashStr = originalCommitHashStr
	}()

	// Set known values for testing
	AppVersionStr = "v1.2.3-test"
	CommitHashStr = "abcdef123456"

	// Create a dummy DdevService instance
	d := &DdevService{}

	// Call AppVersion
	info := d.AppVersion()

	// Verify the result
	if info.Version != "v1.2.3-test" {
		t.Errorf("Expected Version to be 'v1.2.3-test', got '%s'", info.Version)
	}
	if info.CommitHash != "abcdef123456" {
		t.Errorf("Expected CommitHash to be 'abcdef123456', got '%s'", info.CommitHash)
	}
}
