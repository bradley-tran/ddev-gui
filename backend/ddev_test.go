package backend

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	stdruntime "runtime"
	"strings"
	"testing"
)

func TestGetLatestDdevRelease(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload := `{
			"tag_name": "v1.23.4",
			"assets": [
				{"name": "ddev_windows_amd64_installer.v1.23.4.exe", "browser_download_url": "http://example.com/ddev-amd64.exe"},
				{"name": "ddev_windows_arm64_installer.v1.23.4.exe", "browser_download_url": "http://example.com/ddev-arm64.exe"},
				{"name": "checksums.txt", "browser_download_url": "http://example.com/checksums.txt"}
			]
		}`
		fmt.Fprintln(w, payload)
	}))
	defer ts.Close()

	client := ts.Client()
	rel, err := getLatestDdevReleaseFromURL(client, ts.URL)
	if err != nil {
		t.Fatalf("getLatestDdevReleaseFromURL failed: %v", err)
	}

	if rel.TagName != "v1.23.4" {
		t.Errorf("expected tag v1.23.4, got %s", rel.TagName)
	}
	wantURL := "http://example.com/ddev-amd64.exe"
	if stdruntime.GOARCH == "arm64" {
		wantURL = "http://example.com/ddev-arm64.exe"
	}
	if rel.URL != wantURL {
		t.Errorf("expected URL %s, got %s", wantURL, rel.URL)
	}
	if rel.ChecksumURL != "http://example.com/checksums.txt" {
		t.Errorf("expected checksum URL http://example.com/checksums.txt, got %s", rel.ChecksumURL)
	}
}

func TestGetLatestDdevReleaseFallbackToGenericInstaller(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload := `{
			"tag_name": "v1.23.4",
			"assets": [
				{"name": "ddev_windows_installer.v1.23.4.exe", "browser_download_url": "http://example.com/ddev.exe"},
				{"name": "checksums.txt", "browser_download_url": "http://example.com/checksums.txt"}
			]
		}`
		fmt.Fprintln(w, payload)
	}))
	defer ts.Close()

	client := ts.Client()
	rel, err := getLatestDdevReleaseFromURL(client, ts.URL)
	if err != nil {
		t.Fatalf("getLatestDdevReleaseFromURL failed: %v", err)
	}

	if rel.URL != "http://example.com/ddev.exe" {
		t.Errorf("expected fallback URL http://example.com/ddev.exe, got %s", rel.URL)
	}
}

func TestGetLatestDdevReleaseRejectsWrongArchWithoutFallback(t *testing.T) {
	onlyArch := "arm64"
	if stdruntime.GOARCH == "arm64" {
		onlyArch = "amd64"
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload := fmt.Sprintf(`{
			"tag_name": "v1.23.4",
			"assets": [
				{"name": "ddev_windows_%s_installer.v1.23.4.exe", "browser_download_url": "http://example.com/ddev.exe"}
			]
		}`, onlyArch)
		fmt.Fprintln(w, payload)
	}))
	defer ts.Close()

	client := ts.Client()
	_, err := getLatestDdevReleaseFromURL(client, ts.URL)
	if err == nil {
		t.Fatal("expected an error when only non-matching architecture installers are available")
	}
	if !strings.Contains(err.Error(), "could not find Windows") {
		t.Fatalf("expected architecture filtering error, got: %v", err)
	}
}

func TestVerifyInstallerChecksum(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "test-installer.exe")
	err := os.WriteFile(tmp, []byte("fake content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	d := &DdevService{}
	// SHA256 of "fake content"
	expected := "98b1ae45059b004178a8eee0c1f6179dcea139c0fd8a69ee47a6f02d97af1f17"

	if !d.verifyInstallerChecksum(tmp, expected) {
		t.Error("checksum verification failed for valid file")
	}

	if d.verifyInstallerChecksum(tmp, "wrong") {
		t.Error("checksum verification succeeded for invalid hash")
	}
}

func TestNewDdevService(t *testing.T) {
	// Create a clean config service without loading from disk
	cfg := &ConfigService{
		data: map[string]any{},
	}

	// Default config
	svc := NewDdevService(cfg)
	if svc.config != cfg {
		t.Errorf("expected config to be %p, got %p", cfg, svc.config)
	}
	if svc.sshShell != nil {
		t.Errorf("expected sshShell to be nil, got %v", svc.sshShell)
	}

	if stdruntime.GOOS == "windows" {
		if svc.shell == nil {
			t.Errorf("expected shell to be non-nil on Windows")
		}
		if svc.fileShell == nil {
			t.Errorf("expected fileShell to be non-nil on Windows")
		}
	} else {
		if svc.shell != nil {
			t.Errorf("expected shell to be nil on non-Windows")
		}
		if svc.fileShell != nil {
			t.Errorf("expected fileShell to be nil on non-Windows")
		}
	}

	// SSH config
	cfg.Set("backend", "ssh")
	cfg.Set("ssh", map[string]any{
		"host": "localhost",
		"port": "22",
		"user": "test",
	})
	svcSsh := NewDdevService(cfg)
	if svcSsh.sshShell == nil {
		t.Errorf("expected sshShell to be non-nil when backend is ssh")
	}
}

func TestTelemetryOptInPreference(t *testing.T) {
	tests := []struct {
		name     string
		config   map[string]any
		expected bool
		ok       bool
	}{
		{
			name:   "unset",
			config: map[string]any{},
			ok:     false,
		},
		{
			name: "opt in",
			config: map[string]any{
				"ddevTelemetryOptIn": true,
			},
			expected: true,
			ok:       true,
		},
		{
			name: "opt out",
			config: map[string]any{
				"ddevTelemetryOptIn": false,
			},
			expected: false,
			ok:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DdevService{config: &ConfigService{data: tt.config}}

			got, ok := d.telemetryOptInPreference()

			if ok != tt.ok {
				t.Fatalf("telemetryOptInPreference() ok = %v, want %v", ok, tt.ok)
			}
			if got != tt.expected {
				t.Fatalf("telemetryOptInPreference() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestApplyTelemetryPreference(t *testing.T) {
	var gotArgs []string

	err := applyTelemetryPreference(
		context.Background(),
		true,
		func(_ context.Context, args ...string) (string, string, error) {
			gotArgs = append([]string(nil), args...)
			return "", "", nil
		},
	)
	if err != nil {
		t.Fatalf("applyTelemetryPreference() failed: %v", err)
	}

	want := []string{"config", "global", "--instrumentation-opt-in=true"}
	if len(gotArgs) != len(want) {
		t.Fatalf("applyTelemetryPreference() arg count = %d, want %d (%v)", len(gotArgs), len(want), gotArgs)
	}
	for idx, expected := range want {
		if gotArgs[idx] != expected {
			t.Fatalf("applyTelemetryPreference() args[%d] = %q, want %q (all args: %v)", idx, gotArgs[idx], expected, gotArgs)
		}
	}
}

func TestExecSpawnCmdSignature(t *testing.T) {
	// Simple test to ensure the new signature is implemented correctly and runs a bash command
	d := &DdevService{}

	// Create a script that echoes its first parameter
	script := "echo \"$1\""
	expected := "hello security\n"

	out, err := d.execSpawnCmd(script, []string{"hello security"}, 5 * 1000 * 1000 * 1000) // 5 seconds

	// If bash is not installed or available, we might get an exec error, which is fine for this environment check
	if err != nil {
		t.Skipf("Skipping test, bash may not be available: %v", err)
	}

	if out != expected {
		t.Errorf("execSpawnCmd() = %v, want %v", out, expected)
	}
}

func TestReadFileBase64Errors(t *testing.T) {
	d := &DdevService{}

	tests := []struct {
		name    string
		project string
		relPath string
		wantErr string
	}{
		{
			name:    "empty project name",
			project: "   ",
			relPath: "some/path.txt",
			wantErr: "project name is required",
		},
		{
			name:    "empty relative path",
			project: "myproject",
			relPath: "   ",
			wantErr: "file path is required",
		},
		{
			name:    "path containing ..",
			project: "myproject",
			relPath: "some/../path.txt",
			wantErr: "relative path must not contain '..'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := d.ReadFileBase64(tt.project, tt.relPath)
			if err == nil {
				t.Fatalf("ReadFileBase64() expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("ReadFileBase64() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != "" {
				t.Errorf("ReadFileBase64() got = %v, want empty string", got)
			}
		})
	}
}
