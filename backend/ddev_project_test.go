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
			"  if \"%2\"==\"errorproject\" (\r\n" +
			"    echo restart error 1>&2\r\n" +
			"    exit /b 1\r\n" +
			"  )\r\n" +
			"  > \"%TEST_DDEV_ARGS_FILE%\" echo %1 %2\r\n" +
			"  echo restarted %2\r\n" +
			"  exit /b 0\r\n" +
			")\r\n" +
			"if \"%1\"==\"stop\" (\r\n" +
			"  if \"%2\"==\"errorproject\" (\r\n" +
			"    echo stop error 1>&2\r\n" +
			"    exit /b 1\r\n" +
			"  )\r\n" +
			"  > \"%TEST_DDEV_ARGS_FILE%\" echo %1 %2\r\n" +
			"  echo stopped %2\r\n" +
			"  exit /b 0\r\n" +
			")\r\n" +
			"if \"%1\"==\"delete\" (\r\n" +
			"  if \"%4\"==\"errorproject\" (\r\n" +
			"    echo delete error 1>&2\r\n" +
			"    exit /b 1\r\n" +
			"  )\r\n" +
			"  > \"%TEST_DDEV_ARGS_FILE%\" echo %1 %2 %3 %4\r\n" +
			"  echo deleted %4\r\n" +
			"  exit /b 0\r\n" +
			")\r\n" +
			"if \"%1\"==\"start\" (\r\n" +
			"  if \"%2\"==\"errorproject\" (\r\n" +
			"    echo start error 1>&2\r\n" +
			"    exit /b 1\r\n" +
			"  )\r\n" +
			"  > \"%TEST_DDEV_ARGS_FILE%\" echo %1 %2\r\n" +
			"  echo started %2\r\n" +
			"  exit /b 0\r\n" +
			")\r\n" +
			"if \"%1\"==\"poweroff\" (\r\n" +
			"  if \"%TEST_ERROR_POWEROFF%\"==\"1\" (\r\n" +
			"    echo poweroff error 1>&2\r\n" +
			"    exit /b 1\r\n" +
			"  )\r\n" +
			"  > \"%TEST_DDEV_ARGS_FILE%\" echo %1\r\n" +
			"  echo powered off\r\n" +
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
		"  if [ \"$2\" = \"errorproject\" ]; then\n" +
		"    echo \"restart error\" >&2\n" +
		"    exit 1\n" +
		"  fi\n" +
		"  printf '%s %s\\n' \"$1\" \"$2\" > \"$TEST_DDEV_ARGS_FILE\"\n" +
		"  printf 'restarted %s\\n' \"$2\"\n" +
		"  exit 0\n" +
		"fi\n" +
		"if [ \"$1\" = \"stop\" ]; then\n" +
		"  if [ \"$2\" = \"errorproject\" ]; then\n" +
		"    echo \"stop error\" >&2\n" +
		"    exit 1\n" +
		"  fi\n" +
		"  printf '%s %s\\n' \"$1\" \"$2\" > \"$TEST_DDEV_ARGS_FILE\"\n" +
		"  printf 'stopped %s\\n' \"$2\"\n" +
		"  exit 0\n" +
		"fi\n" +
		"if [ \"$1\" = \"delete\" ]; then\n" +
		"  if [ \"$4\" = \"errorproject\" ]; then\n" +
		"    echo \"delete error\" >&2\n" +
		"    exit 1\n" +
		"  fi\n" +
		"  printf '%s %s %s %s\\n' \"$1\" \"$2\" \"$3\" \"$4\" > \"$TEST_DDEV_ARGS_FILE\"\n" +
		"  printf 'deleted %s\\n' \"$4\"\n" +
		"  exit 0\n" +
		"fi\n" +
		"if [ \"$1\" = \"start\" ]; then\n" +
		"  if [ \"$2\" = \"errorproject\" ]; then\n" +
		"    echo \"start error\" >&2\n" +
		"    exit 1\n" +
		"  fi\n" +
		"  printf '%s %s\\n' \"$1\" \"$2\" > \"$TEST_DDEV_ARGS_FILE\"\n" +
		"  printf 'started %s\\n' \"$2\"\n" +
		"  exit 0\n" +
		"fi\n" +
		"if [ \"$1\" = \"poweroff\" ]; then\n" +
		"  if [ \"$TEST_ERROR_POWEROFF\" = \"1\" ]; then\n" +
		"    echo \"poweroff error\" >&2\n" +
		"    exit 1\n" +
		"  fi\n" +
		"  printf '%s\\n' \"$1\" > \"$TEST_DDEV_ARGS_FILE\"\n" +
		"  printf 'powered off\\n'\n" +
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

func TestStart(t *testing.T) {
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
	_, err := svc.Start("   ")
	if err == nil {
		t.Fatalf("expected error for empty project name, got nil")
	}
	if err.Error() != "project name is required" {
		t.Fatalf("expected error 'project name is required', got %q", err.Error())
	}

	// Test 2: Valid project name
	output, err := svc.Start("myproject")
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	if strings.TrimSpace(output) != "started myproject" {
		t.Fatalf("expected 'started myproject', got %q", output)
	}

	argsRaw, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("failed to read fake ddev args: %v", err)
	}
	if strings.TrimSpace(string(argsRaw)) != "start myproject" {
		t.Fatalf("expected 'start myproject', got %q", strings.TrimSpace(string(argsRaw)))
	}
}

func TestPowerOff(t *testing.T) {
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

	output, err := svc.PowerOff()
	if err != nil {
		t.Fatalf("PowerOff returned error: %v", err)
	}
	if strings.TrimSpace(output) != "powered off" {
		t.Fatalf("expected 'powered off', got %q", output)
	}

	argsRaw, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("failed to read fake ddev args: %v", err)
	}
	if strings.TrimSpace(string(argsRaw)) != "poweroff" {
		t.Fatalf("expected 'poweroff', got %q", strings.TrimSpace(string(argsRaw)))
	}
}

func TestDeleteProjectExecError(t *testing.T) {
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

	_, err := svc.DeleteProject("errorproject")
	if err == nil {
		t.Fatalf("expected execution error, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "exit status") && !strings.Contains(err.Error(), "error") {
		t.Fatalf("unexpected error format: %v", err)
	}
}

func TestStartExecError(t *testing.T) {
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

	_, err := svc.Start("errorproject")
	if err == nil {
		t.Fatalf("expected execution error, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "exit status") && !strings.Contains(err.Error(), "error") {
		t.Fatalf("unexpected error format: %v", err)
	}
}

func TestStopExecError(t *testing.T) {
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

	_, err := svc.Stop("errorproject")
	if err == nil {
		t.Fatalf("expected execution error, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "exit status") && !strings.Contains(err.Error(), "error") {
		t.Fatalf("unexpected error format: %v", err)
	}
}

func TestRestartExecError(t *testing.T) {
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

	_, err := svc.Restart("errorproject")
	if err == nil {
		t.Fatalf("expected execution error, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "exit status") && !strings.Contains(err.Error(), "error") {
		t.Fatalf("unexpected error format: %v", err)
	}
}

func TestPowerOffExecError(t *testing.T) {
	tempDir := t.TempDir()
	argsFile := filepath.Join(tempDir, "args.txt")
	fakeDdevPath := filepath.Join(tempDir, fakeDdevScriptName())
	if err := os.WriteFile(fakeDdevPath, []byte(fakeDdevScript()), 0755); err != nil {
		t.Fatalf("failed to write fake ddev script: %v", err)
	}

	originalPath := os.Getenv("PATH")
	t.Setenv("PATH", tempDir+string(os.PathListSeparator)+originalPath)
	t.Setenv("TEST_DDEV_ARGS_FILE", argsFile)
	t.Setenv("TEST_ERROR_POWEROFF", "1")

	svc := &DdevService{
		config: &ConfigService{data: map[string]any{"backend": "local"}},
	}

	_, err := svc.PowerOff()
	if err == nil {
		t.Fatalf("expected execution error, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "exit status") && !strings.Contains(err.Error(), "error") {
		t.Fatalf("unexpected error format: %v", err)
	}
}
