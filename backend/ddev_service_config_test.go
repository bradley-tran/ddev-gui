package backend

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func newConfigureServicesTestService(t *testing.T) (*DdevService, string, string) {
	t.Helper()

	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "project")
	if err := os.Mkdir(projectDir, 0o755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}

	describePayload, err := json.Marshal(map[string]any{
		"raw": map[string]string{
			"approot": projectDir,
		},
	})
	if err != nil {
		t.Fatalf("failed to marshal describe payload: %v", err)
	}

	describeFile := filepath.Join(tempDir, "describe.json")
	if err := os.WriteFile(describeFile, describePayload, 0o644); err != nil {
		t.Fatalf("failed to write describe payload: %v", err)
	}

	argsFile := filepath.Join(tempDir, "args.txt")
	fakeDdevPath := filepath.Join(tempDir, fakeConfigDdevScriptName())
	if err := os.WriteFile(fakeDdevPath, []byte(fakeConfigDdevScript()), 0o755); err != nil {
		t.Fatalf("failed to write fake ddev script: %v", err)
	}

	originalPath := os.Getenv("PATH")
	t.Setenv("PATH", tempDir+string(os.PathListSeparator)+originalPath)
	t.Setenv("TEST_DDEV_DESCRIBE_FILE", describeFile)
	t.Setenv("TEST_DDEV_ARGS_FILE", argsFile)

	svc := &DdevService{
		config: &ConfigService{data: map[string]any{"backend": "local"}},
	}

	return svc, argsFile, projectDir
}

func readServiceCommandLog(t *testing.T, argsFile string) []string {
	t.Helper()

	argsRaw, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("failed to read args file: %v", err)
	}

	trimmed := strings.TrimSpace(string(argsRaw))
	if trimmed == "" {
		return nil
	}

	return strings.Split(trimmed, "\n")
}

func TestConfigureServices_BuildsExpectedCommandSequence(t *testing.T) {
	svc, argsFile, projectDir := newConfigureServicesTestService(t)

	if _, err := svc.ConfigureServices("demo", "8080", "3307", true, true, false); err != nil {
		t.Fatalf("ConfigureServices returned error: %v", err)
	}

	lines := readServiceCommandLog(t, argsFile)
	expected := []string{
		"config --auto --host-webserver-port=8080 --host-db-port=3307",
		"xdebug on",
		"xhprof on",
		"xhgui off",
	}

	if len(lines) != len(expected) {
		t.Fatalf("expected %d command lines, got %d: %q", len(expected), len(lines), lines)
	}

	for idx, line := range lines {
		parts := strings.SplitN(strings.TrimSpace(line), "|", 2)
		if len(parts) != 2 {
			t.Fatalf("expected command log format '<cwd>|<args>', got %q", line)
		}
		if parts[0] != projectDir {
			t.Fatalf("expected command to run in %q, got %q", projectDir, parts[0])
		}
		if parts[1] != expected[idx] {
			t.Fatalf("expected command %q, got %q", expected[idx], parts[1])
		}
	}
}

func TestConfigureServices_EnablesXhprofWhenXhguiEnabled(t *testing.T) {
	svc, argsFile, _ := newConfigureServicesTestService(t)

	if _, err := svc.ConfigureServices("demo", "", "", false, false, true); err != nil {
		t.Fatalf("ConfigureServices returned error: %v", err)
	}

	lines := readServiceCommandLog(t, argsFile)
	expected := []string{
		"xdebug off",
		"xhprof on",
		"xhgui on",
	}

	if len(lines) != len(expected) {
		t.Fatalf("expected %d command lines, got %d: %q", len(expected), len(lines), lines)
	}

	for idx, line := range lines {
		parts := strings.SplitN(strings.TrimSpace(line), "|", 2)
		if len(parts) != 2 {
			t.Fatalf("expected command log format '<cwd>|<args>', got %q", line)
		}
		if parts[1] != expected[idx] {
			t.Fatalf("expected command %q, got %q", expected[idx], parts[1])
		}
	}
}

func TestConfigureServices_RejectsInvalidPorts(t *testing.T) {
	svc := &DdevService{}

	_, err := svc.ConfigureServices("demo", "abc", "", false, false, false)
	if err == nil {
		t.Fatal("expected validation error for invalid web port")
	}
	if !strings.Contains(err.Error(), "web port") {
		t.Fatalf("expected web port validation error, got %v", err)
	}
}

func fakeConfigDdevScriptName() string {
	if runtime.GOOS == "windows" {
		return "ddev.cmd"
	}
	return "ddev"
}

func fakeConfigDdevScript() string {
	if runtime.GOOS == "windows" {
		return "@echo off\r\n" +
			"if \"%1\"==\"describe\" (\r\n" +
			"  type \"%TEST_DDEV_DESCRIBE_FILE%\"\r\n" +
			"  exit /b 0\r\n" +
			")\r\n" +
			"if \"%1\"==\"config\" (\r\n" +
			"  >> \"%TEST_DDEV_ARGS_FILE%\" echo %CD%^|%*\r\n" +
			"  echo configured\r\n" +
			"  exit /b 0\r\n" +
			")\r\n" +
			"if \"%1\"==\"xdebug\" (\r\n" +
			"  >> \"%TEST_DDEV_ARGS_FILE%\" echo %CD%^|%*\r\n" +
			"  echo xdebug toggled\r\n" +
			"  exit /b 0\r\n" +
			")\r\n" +
			"if \"%1\"==\"xhgui\" (\r\n" +
			"  >> \"%TEST_DDEV_ARGS_FILE%\" echo %CD%^|%*\r\n" +
			"  echo xhgui toggled\r\n" +
			"  exit /b 0\r\n" +
			")\r\n" +
			"if \"%1\"==\"xhprof\" (\r\n" +
			"  >> \"%TEST_DDEV_ARGS_FILE%\" echo %CD%^|%*\r\n" +
			"  echo xhprof toggled\r\n" +
			"  exit /b 0\r\n" +
			")\r\n" +
			"echo unexpected args %* 1>&2\r\n" +
			"exit /b 1\r\n"
	}

	return "#!/bin/sh\n" +
		"if [ \"$1\" = \"describe\" ]; then\n" +
		"  cat \"$TEST_DDEV_DESCRIBE_FILE\"\n" +
		"  exit 0\n" +
		"fi\n" +
		"if [ \"$1\" = \"config\" ]; then\n" +
		"  printf '%s|%s\\n' \"$PWD\" \"$*\" >> \"$TEST_DDEV_ARGS_FILE\"\n" +
		"  echo configured\n" +
		"  exit 0\n" +
		"fi\n" +
		"if [ \"$1\" = \"xdebug\" ]; then\n" +
		"  printf '%s|%s\\n' \"$PWD\" \"$*\" >> \"$TEST_DDEV_ARGS_FILE\"\n" +
		"  echo xdebug toggled\n" +
		"  exit 0\n" +
		"fi\n" +
		"if [ \"$1\" = \"xhgui\" ]; then\n" +
		"  printf '%s|%s\\n' \"$PWD\" \"$*\" >> \"$TEST_DDEV_ARGS_FILE\"\n" +
		"  echo xhgui toggled\n" +
		"  exit 0\n" +
		"fi\n" +
		"if [ \"$1\" = \"xhprof\" ]; then\n" +
		"  printf '%s|%s\\n' \"$PWD\" \"$*\" >> \"$TEST_DDEV_ARGS_FILE\"\n" +
		"  echo xhprof toggled\n" +
		"  exit 0\n" +
		"fi\n" +
		"echo \"unexpected args: $*\" >&2\n" +
		"exit 1\n"
}
