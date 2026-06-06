package backend

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func fakeAddonDdevScriptName() string {
	if runtime.GOOS == "windows" {
		return "ddev.cmd"
	}
	return "ddev"
}

func fakeAddonDdevScript() string {
	if runtime.GOOS == "windows" {
		return "@echo off\r\n" +
			"if \"%1\"==\"describe\" (\r\n" +
			"  type \"%TEST_DDEV_DESCRIBE_FILE%\"\r\n" +
			"  exit /b 0\r\n" +
			")\r\n" +
			"if \"%1\"==\"add-on\" (\r\n" +
			"  > \"%TEST_DDEV_ARGS_FILE%\" echo %1 %2 %3 %4\r\n" +
			"  echo [{\"name\":\"test-addon\"}]\r\n" +
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
		"if [ \"$1\" = \"add-on\" ]; then\n" +
		"  echo \"$1 $2 $3 $4\" > \"$TEST_DDEV_ARGS_FILE\"\n" +
		"  echo '[{\"name\":\"test-addon\"}]'\n" +
		"  exit 0\n" +
		"fi\n" +
		"echo \"unexpected args $@\" >&2\n" +
		"exit 1\n"
}

func TestAddonsJSON(t *testing.T) {
	tempDir := t.TempDir()

	projectDir := filepath.Join(tempDir, "project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
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
	if err := os.WriteFile(describeFile, describePayload, 0644); err != nil {
		t.Fatalf("failed to write describe payload: %v", err)
	}

	argsFile := filepath.Join(tempDir, "args.txt")
	fakeDdevPath := filepath.Join(tempDir, fakeAddonDdevScriptName())
	if err := os.WriteFile(fakeDdevPath, []byte(fakeAddonDdevScript()), 0755); err != nil {
		t.Fatalf("failed to write fake ddev script: %v", err)
	}

	originalPath := os.Getenv("PATH")
	t.Setenv("PATH", tempDir+string(os.PathListSeparator)+originalPath)
	t.Setenv("TEST_DDEV_DESCRIBE_FILE", describeFile)
	t.Setenv("TEST_DDEV_ARGS_FILE", argsFile)
	t.Setenv("HOME", tempDir)
	t.Setenv("USERPROFILE", tempDir)

	svc := &DdevService{
		config: &ConfigService{data: map[string]any{"backend": "local"}},
	}

	output, err := svc.AddonsJSON("my-project")
	if err != nil {
		t.Fatalf("AddonsJSON returned error: %v", err)
	}

	if output != `[{"name":"test-addon"}]` {
		t.Fatalf("expected JSON output, got %q", output)
	}

	argsRaw, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("failed to read fake ddev args: %v", err)
	}

	argsString := strings.TrimSpace(string(argsRaw))
	if argsString != "add-on list --installed -j" {
		t.Fatalf("expected args 'add-on list --installed -j', got %q", argsString)
	}
}

func TestAddonsJSON_EmptyProjectName(t *testing.T) {
	svc := &DdevService{}

	_, err := svc.AddonsJSON("   ")
	if err == nil {
		t.Fatal("expected error for empty project name, got nil")
	}

	if err.Error() != "project name is required" {
		t.Fatalf("expected error 'project name is required', got %v", err)
	}
}
