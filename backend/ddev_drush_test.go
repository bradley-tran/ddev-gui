package backend

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func fakeDdevScriptNameDrush() string {
	if runtime.GOOS == "windows" {
		return "ddev.cmd"
	}
	return "ddev"
}

func fakeDdevScriptDrush() string {
	if runtime.GOOS == "windows" {
		return "@echo off\r\n" +
			"if \"%1\"==\"describe\" (\r\n" +
			"  type \"%TEST_DDEV_DESCRIBE_FILE%\"\r\n" +
			"  exit /b 0\r\n" +
			")\r\n" +
			"if \"%1\"==\"drush\" (\r\n" +
			"  > \"%TEST_DDEV_ARGS_FILE%\" echo %1 %2 %3\r\n" +
			"  type \"%TEST_DDEV_MOCK_OUTPUT_FILE%\"\r\n" +
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
		"if [ \"$1\" = \"drush\" ]; then\n" +
		"  printf '%s %s %s\\n' \"$1\" \"$2\" \"$3\" > \"$TEST_DDEV_ARGS_FILE\"\n" +
		"  cat \"$TEST_DDEV_MOCK_OUTPUT_FILE\"\n" +
		"  exit 0\n" +
		"fi\n" +
		"echo \"unexpected args: $*\" >&2\n" +
		"exit 1\n"
}

func TestDrushRecentUsers(t *testing.T) {
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
	mockOutputFile := filepath.Join(tempDir, "mock_output.txt")
	fakeDdevPath := filepath.Join(tempDir, fakeDdevScriptNameDrush())

	mockOutput := "1\tadmin\tadmin@example.com\n2\tuser1\tuser1@example.com"
	if err := os.WriteFile(mockOutputFile, []byte(mockOutput), 0644); err != nil {
		t.Fatalf("failed to write mock output file: %v", err)
	}
	if err := os.WriteFile(fakeDdevPath, []byte(fakeDdevScriptDrush()), 0755); err != nil {
		t.Fatalf("failed to write fake ddev script: %v", err)
	}

	originalPath := os.Getenv("PATH")
	t.Setenv("PATH", tempDir+string(os.PathListSeparator)+originalPath)
	t.Setenv("TEST_DDEV_DESCRIBE_FILE", describeFile)
	t.Setenv("TEST_DDEV_ARGS_FILE", argsFile)
	t.Setenv("TEST_DDEV_MOCK_OUTPUT_FILE", mockOutputFile)

	svc := &DdevService{
		config: &ConfigService{data: map[string]any{"backend": "local"}},
	}

	output, err := svc.DrushRecentUsers("demo")
	if err != nil {
		t.Fatalf("DrushRecentUsers returned error: %v", err)
	}

	expectedJSON := `[{"uid":"1","name":"admin","mail":"admin@example.com"},{"uid":"2","name":"user1","mail":"user1@example.com"}]`
	if strings.TrimSpace(output) != expectedJSON {
		t.Fatalf("expected JSON output %q, got %q", expectedJSON, output)
	}

	argsRaw, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("failed to read fake ddev args: %v", err)
	}
	if !strings.Contains(string(argsRaw), "drush sql:query") {
		t.Fatalf("expected drush command to include sql:query, got %q", strings.TrimSpace(string(argsRaw)))
	}
}

func TestDrushRecentUsers_EmptyProject(t *testing.T) {
	svc := &DdevService{
		config: &ConfigService{data: map[string]any{"backend": "local"}},
	}

	output, err := svc.DrushRecentUsers(" ")
	if err == nil {
		t.Fatal("expected error for empty project name, got nil")
	}

	if output != "[]" {
		t.Fatalf("expected output '[]', got %q", output)
	}

	if err.Error() != "project name is required" {
		t.Fatalf("expected error 'project name is required', got %q", err.Error())
	}
}

func TestDrushRecentUsers_QueryError(t *testing.T) {
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

	fakeDdevPath := filepath.Join(tempDir, fakeDdevScriptNameDrush())

	// mock output that causes an error by exiting 1
	var script string
	if runtime.GOOS == "windows" {
		script = "@echo off\nexit /b 1"
	} else {
		script = "#!/bin/sh\nexit 1"
	}

	if err := os.WriteFile(fakeDdevPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to write fake ddev script: %v", err)
	}

	originalPath := os.Getenv("PATH")
	t.Setenv("PATH", tempDir+string(os.PathListSeparator)+originalPath)
	t.Setenv("TEST_DDEV_DESCRIBE_FILE", describeFile)

	svc := &DdevService{
		config: &ConfigService{data: map[string]any{"backend": "local"}},
	}

	output, err := svc.DrushRecentUsers("demo")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if output != "[]" {
		t.Fatalf("expected output '[]', got %q", output)
	}

	if !strings.Contains(err.Error(), "failed to query users") {
		t.Fatalf("expected error to contain 'failed to query users', got %q", err.Error())
	}
}

func TestDrushRecentUsers_Parsing(t *testing.T) {
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
	mockOutputFile := filepath.Join(tempDir, "mock_output.txt")
	fakeDdevPath := filepath.Join(tempDir, fakeDdevScriptNameDrush())

	// Mock output with some malformed lines and missing mail
	mockOutput := "1\tadmin\tadmin@example.com\n\n2\tuser1\nmalformed"
	if err := os.WriteFile(mockOutputFile, []byte(mockOutput), 0644); err != nil {
		t.Fatalf("failed to write mock output file: %v", err)
	}
	if err := os.WriteFile(fakeDdevPath, []byte(fakeDdevScriptDrush()), 0755); err != nil {
		t.Fatalf("failed to write fake ddev script: %v", err)
	}

	originalPath := os.Getenv("PATH")
	t.Setenv("PATH", tempDir+string(os.PathListSeparator)+originalPath)
	t.Setenv("TEST_DDEV_DESCRIBE_FILE", describeFile)
	t.Setenv("TEST_DDEV_ARGS_FILE", argsFile)
	t.Setenv("TEST_DDEV_MOCK_OUTPUT_FILE", mockOutputFile)

	svc := &DdevService{
		config: &ConfigService{data: map[string]any{"backend": "local"}},
	}

	output, err := svc.DrushRecentUsers("demo")
	if err != nil {
		t.Fatalf("DrushRecentUsers returned error: %v", err)
	}

	expectedJSON := `[{"uid":"1","name":"admin","mail":"admin@example.com"},{"uid":"2","name":"user1","mail":""}]`
	if strings.TrimSpace(output) != expectedJSON {
		t.Fatalf("expected JSON output %q, got %q", expectedJSON, output)
	}
}

func TestDrushRecentUsers_NoUsers(t *testing.T) {
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
	mockOutputFile := filepath.Join(tempDir, "mock_output.txt")
	fakeDdevPath := filepath.Join(tempDir, fakeDdevScriptNameDrush())

	// Mock output with no users
	mockOutput := ""
	if err := os.WriteFile(mockOutputFile, []byte(mockOutput), 0644); err != nil {
		t.Fatalf("failed to write mock output file: %v", err)
	}
	if err := os.WriteFile(fakeDdevPath, []byte(fakeDdevScriptDrush()), 0755); err != nil {
		t.Fatalf("failed to write fake ddev script: %v", err)
	}

	originalPath := os.Getenv("PATH")
	t.Setenv("PATH", tempDir+string(os.PathListSeparator)+originalPath)
	t.Setenv("TEST_DDEV_DESCRIBE_FILE", describeFile)
	t.Setenv("TEST_DDEV_ARGS_FILE", argsFile)
	t.Setenv("TEST_DDEV_MOCK_OUTPUT_FILE", mockOutputFile)

	svc := &DdevService{
		config: &ConfigService{data: map[string]any{"backend": "local"}},
	}

	output, err := svc.DrushRecentUsers("demo")
	if err != nil {
		t.Fatalf("DrushRecentUsers returned error: %v", err)
	}

	expectedJSON := `[]`
	if strings.TrimSpace(output) != expectedJSON {
		t.Fatalf("expected JSON output %q, got %q", expectedJSON, output)
	}
}
