package backend

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestProjectLogsUsesSelectedService(t *testing.T) {
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "project")
	if err := os.Mkdir(projectDir, 0755); err != nil {
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
	fakeDdevPath := filepath.Join(tempDir, fakeDdevScriptName())
	if err := os.WriteFile(fakeDdevPath, []byte(fakeDdevScript()), 0755); err != nil {
		t.Fatalf("failed to write fake ddev script: %v", err)
	}

	originalPath := os.Getenv("PATH")
	t.Setenv("PATH", tempDir+string(os.PathListSeparator)+originalPath)
	t.Setenv("TEST_DDEV_DESCRIBE_FILE", describeFile)
	t.Setenv("TEST_DDEV_ARGS_FILE", argsFile)

	svc := &DdevService{
		config: &ConfigService{data: map[string]any{"backend": "local"}},
	}

	output, err := svc.ProjectLogs("demo", "db")
	if err != nil {
		t.Fatalf("ProjectLogs returned error: %v", err)
	}
	if strings.TrimSpace(output) != "service=db" {
		t.Fatalf("expected service output, got %q", output)
	}

	argsRaw, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("failed to read fake ddev args: %v", err)
	}
	if strings.TrimSpace(string(argsRaw)) != "logs -s db" {
		t.Fatalf("expected logs command to include selected service, got %q", strings.TrimSpace(string(argsRaw)))
	}
}

func fakeDdevScriptName() string {
	if runtime.GOOS == "windows" {
		return "ddev.cmd"
	}
	return "ddev"
}

func fakeDdevScript() string {
	if runtime.GOOS == "windows" {
		return "@echo off\r\n" +
			"if \"%1\"==\"describe\" (\r\n" +
			"  type \"%TEST_DDEV_DESCRIBE_FILE%\"\r\n" +
			"  exit /b 0\r\n" +
			")\r\n" +
			"if \"%1\"==\"logs\" (\r\n" +
			"  > \"%TEST_DDEV_ARGS_FILE%\" echo %1 %2 %3\r\n" +
			"  echo service=%3\r\n" +
			"  exit /b 0\r\n" +
			")\r\n" +
			"if \"%1\"==\"restart\" (\r\n" +
			"  > \"%TEST_DDEV_ARGS_FILE%\" echo %1 %2\r\n" +
			"  echo restarted %2\r\n" +
			"  exit /b 0\r\n" +
			")\r\n" +
			"if \"%1\"==\"stop\" (\r\n" +
			"  > \"%TEST_DDEV_ARGS_FILE%\" echo %1 %2\r\n" +
			"  echo stopped %2\r\n" +
			"  exit /b 0\r\n" +
			")\r\n" +
			"if \"%1\"==\"delete\" (\r\n" +
			"  > \"%TEST_DDEV_ARGS_FILE%\" echo %1 %2 %3 %4\r\n" +
			"  echo deleted %4\r\n" +
			"  exit /b 0\r\n" +
			")\r\n" +
			"if \"%1\"==\"import-db\" (\r\n" +
			"  > \"%TEST_DDEV_ARGS_FILE%\" echo %1 %2 %3\r\n" +
			"  echo imported db for %2\r\n" +
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
		"if [ \"$1\" = \"logs\" ]; then\n" +
		"  printf '%s %s %s\\n' \"$1\" \"$2\" \"$3\" > \"$TEST_DDEV_ARGS_FILE\"\n" +
		"  printf 'service=%s\\n' \"$3\"\n" +
		"  exit 0\n" +
		"fi\n" +
		"if [ \"$1\" = \"restart\" ]; then\n" +
		"  printf '%s %s\\n' \"$1\" \"$2\" > \"$TEST_DDEV_ARGS_FILE\"\n" +
		"  printf 'restarted %s\\n' \"$2\"\n" +
		"  exit 0\n" +
		"fi\n" +
		"if [ \"$1\" = \"stop\" ]; then\n" +
		"  printf '%s %s\\n' \"$1\" \"$2\" > \"$TEST_DDEV_ARGS_FILE\"\n" +
		"  printf 'stopped %s\\n' \"$2\"\n" +
		"  exit 0\n" +
		"fi\n" +
		"if [ \"$1\" = \"delete\" ]; then\n" +
		"  printf '%s %s %s %s\\n' \"$1\" \"$2\" \"$3\" \"$4\" > \"$TEST_DDEV_ARGS_FILE\"\n" +
		"  printf 'deleted %s\\n' \"$4\"\n" +
		"  exit 0\n" +
		"fi\n" +
		"if [ \"$1\" = \"import-db\" ]; then\n" +
		"  printf '%s %s %s\\n' \"$1\" \"$2\" \"$3\" > \"$TEST_DDEV_ARGS_FILE\"\n" +
		"  printf 'imported db for %s\\n' \"$2\"\n" +
		"  exit 0\n" +
		"fi\n" +
		"echo \"unexpected args: $*\" >&2\n" +
		"exit 1\n"
}

func TestRestart(t *testing.T) {
	tempDir := t.TempDir()
	argsFile := filepath.Join(tempDir, "args.txt")
	fakeDdevPath := filepath.Join(tempDir, fakeDdevScriptName())
	if err := os.WriteFile(fakeDdevPath, []byte(fakeDdevScript()), 0755); err != nil {
		t.Fatalf("failed to write fake ddev script: %v", err)
	}

	originalPath := os.Getenv("PATH")
	t.Setenv("PATH", tempDir+string(os.PathListSeparator)+originalPath)
	t.Setenv("TEST_DDEV_ARGS_FILE", argsFile)

	svc := &DdevService{
		config: &ConfigService{data: map[string]any{"backend": "local"}},
	}

	// Test 1: Empty project name
	_, err := svc.Restart("   ")
	if err == nil {
		t.Fatalf("expected error for empty project name, got nil")
	}
	if err.Error() != "project name is required" {
		t.Fatalf("expected error 'project name is required', got %q", err.Error())
	}

	// Test 2: Valid project name
	output, err := svc.Restart("myproject")
	if err != nil {
		t.Fatalf("Restart returned error: %v", err)
	}
	if strings.TrimSpace(output) != "restarted myproject" {
		t.Fatalf("expected 'restarted myproject', got %q", output)
	}

	argsRaw, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("failed to read fake ddev args: %v", err)
	}
	if strings.TrimSpace(string(argsRaw)) != "restart myproject" {
		t.Fatalf("expected 'restart myproject', got %q", strings.TrimSpace(string(argsRaw)))
	}
}

func TestStop(t *testing.T) {
	tempDir := t.TempDir()
	argsFile := filepath.Join(tempDir, "args.txt")
	fakeDdevPath := filepath.Join(tempDir, fakeDdevScriptName())
	if err := os.WriteFile(fakeDdevPath, []byte(fakeDdevScript()), 0755); err != nil {
		t.Fatalf("failed to write fake ddev script: %v", err)
	}

	originalPath := os.Getenv("PATH")
	t.Setenv("PATH", tempDir+string(os.PathListSeparator)+originalPath)
	t.Setenv("TEST_DDEV_ARGS_FILE", argsFile)

	svc := &DdevService{
		config: &ConfigService{data: map[string]any{"backend": "local"}},
	}

	// Test 1: Empty project name
	_, err := svc.Stop("   ")
	if err == nil {
		t.Fatalf("expected error for empty project name, got nil")
	}
	if err.Error() != "project name is required" {
		t.Fatalf("expected error 'project name is required', got %q", err.Error())
	}

	// Test 2: Valid project name
	output, err := svc.Stop("myproject")
	if err != nil {
		t.Fatalf("Stop returned error: %v", err)
	}
	if strings.TrimSpace(output) != "stopped myproject" {
		t.Fatalf("expected 'stopped myproject', got %q", output)
	}

	argsRaw, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("failed to read fake ddev args: %v", err)
	}
	if strings.TrimSpace(string(argsRaw)) != "stop myproject" {
		t.Fatalf("expected 'stop myproject', got %q", strings.TrimSpace(string(argsRaw)))
	}
}

func TestDeleteProject(t *testing.T) {
	tempDir := t.TempDir()
	argsFile := filepath.Join(tempDir, "args.txt")
	fakeDdevPath := filepath.Join(tempDir, fakeDdevScriptName())
	if err := os.WriteFile(fakeDdevPath, []byte(fakeDdevScript()), 0755); err != nil {
		t.Fatalf("failed to write fake ddev script: %v", err)
	}

	originalPath := os.Getenv("PATH")
	t.Setenv("PATH", tempDir+string(os.PathListSeparator)+originalPath)
	t.Setenv("TEST_DDEV_ARGS_FILE", argsFile)

	svc := &DdevService{
		config: &ConfigService{data: map[string]any{"backend": "local"}},
	}

	// Test 1: Empty project name
	_, err := svc.DeleteProject("   ")
	if err == nil {
		t.Fatalf("expected error for empty project name, got nil")
	}
	if err.Error() != "project name is required" {
		t.Fatalf("expected error 'project name is required', got %q", err.Error())
	}

	// Test 2: Valid project name
	output, err := svc.DeleteProject("myproject")
	if err != nil {
		t.Fatalf("DeleteProject returned error: %v", err)
	}
	if strings.TrimSpace(output) != "deleted myproject" {
		t.Fatalf("expected 'deleted myproject', got %q", output)
	}

	argsRaw, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("failed to read fake ddev args: %v", err)
	}
	if strings.TrimSpace(string(argsRaw)) != "delete -O -y myproject" {
		t.Fatalf("expected 'delete -O -y myproject', got %q", strings.TrimSpace(string(argsRaw)))
	}
}

func TestImportDBFromFile(t *testing.T) {
	tempDir := t.TempDir()
	argsFile := filepath.Join(tempDir, "args.txt")
	fakeDdevPath := filepath.Join(tempDir, fakeDdevScriptName())
	if err := os.WriteFile(fakeDdevPath, []byte(fakeDdevScript()), 0755); err != nil {
		t.Fatalf("failed to write fake ddev script: %v", err)
	}

	originalPath := os.Getenv("PATH")
	t.Setenv("PATH", tempDir+string(os.PathListSeparator)+originalPath)
	t.Setenv("TEST_DDEV_ARGS_FILE", argsFile)

	svc := &DdevService{
		config: &ConfigService{data: map[string]any{"backend": "local"}},
	}

	t.Run("success", func(t *testing.T) {
		output, err := svc.ImportDBFromFile("myproject", "/path/to/db.sql")
		if err != nil {
			t.Fatalf("ImportDBFromFile returned error: %v", err)
		}

		if !strings.Contains(output, "imported db for myproject") {
			t.Fatalf("expected output to contain 'imported db for myproject', got %q", output)
		}

		argsRaw, err := os.ReadFile(argsFile)
		if err != nil {
			t.Fatalf("failed to read fake ddev args: %v", err)
		}

		expectedArgs := "import-db myproject --file=/path/to/db.sql"
		if strings.TrimSpace(string(argsRaw)) != expectedArgs {
			t.Fatalf("expected '%s', got %q", expectedArgs, strings.TrimSpace(string(argsRaw)))
		}
	})

	t.Run("empty project", func(t *testing.T) {
		_, err := svc.ImportDBFromFile("   ", "/path/to/db.sql")
		if err == nil {
			t.Fatalf("expected error for empty project, got nil")
		}
		if err.Error() != "project name and file path are required" {
			t.Fatalf("expected specific error message, got %q", err.Error())
		}
	})

	t.Run("empty file path", func(t *testing.T) {
		_, err := svc.ImportDBFromFile("myproject", "   ")
		if err == nil {
			t.Fatalf("expected error for empty file path, got nil")
		}
		if err.Error() != "project name and file path are required" {
			t.Fatalf("expected specific error message, got %q", err.Error())
		}
	})
}
