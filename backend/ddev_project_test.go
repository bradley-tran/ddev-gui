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
		"echo \"unexpected args: $*\" >&2\n" +
		"exit 1\n"
}
