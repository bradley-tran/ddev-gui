package backend

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func fakeDdevSnapshotScriptName() string {
	if runtime.GOOS == "windows" {
		return "ddev.cmd"
	}
	return "ddev"
}

func fakeDdevSnapshotScript() string {
	if runtime.GOOS == "windows" {
		return "@echo off\r\n" +
			"if \"%1\"==\"describe\" (\r\n" +
			"  type \"%TEST_DDEV_DESCRIBE_FILE%\"\r\n" +
			"  exit /b 0\r\n" +
			")\r\n" +
			"if \"%1\"==\"snapshot\" (\r\n" +
			"  > \"%TEST_DDEV_ARGS_FILE%\" echo %*\r\n" +
			"  if \"%2\"==\"--list\" (\r\n" +
			"    echo {\"status\": \"ok\"}\r\n" +
			"  )\r\n" +
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
		"if [ \"$1\" = \"snapshot\" ]; then\n" +
		"  echo \"$*\" > \"$TEST_DDEV_ARGS_FILE\"\n" +
		"  if [ \"$2\" = \"--list\" ]; then\n" +
		"    echo '{\"status\": \"ok\"}'\n" +
		"  fi\n" +
		"  exit 0\n" +
		"fi\n" +
		"echo \"unexpected args: $*\" >&2\n" +
		"exit 1\n"
}

func TestSnapshotListJSON(t *testing.T) {
	tempDir := t.TempDir()

	argsFile := filepath.Join(tempDir, "args.txt")
	fakeDdevPath := filepath.Join(tempDir, fakeDdevSnapshotScriptName())
	if err := os.WriteFile(fakeDdevPath, []byte(fakeDdevSnapshotScript()), 0755); err != nil {
		t.Fatalf("failed to write fake ddev script: %v", err)
	}

	originalPath := os.Getenv("PATH")
	t.Setenv("PATH", tempDir+string(os.PathListSeparator)+originalPath)
	t.Setenv("TEST_DDEV_ARGS_FILE", argsFile)

	svc := &DdevService{
		config: &ConfigService{data: map[string]any{"backend": "local"}},
	}

	_, err := svc.SnapshotListJSON("")
	if err == nil {
		t.Fatal("expected error for empty project name, got nil")
	}

	output, err := svc.SnapshotListJSON("demo")
	if err != nil {
		t.Fatalf("SnapshotListJSON returned error: %v", err)
	}
	if strings.TrimSpace(output) != `{"status": "ok"}` {
		t.Fatalf("expected JSON output, got %q", output)
	}

	argsRaw, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("failed to read fake ddev args: %v", err)
	}
	if !strings.Contains(string(argsRaw), "snapshot --list -j demo") {
		t.Fatalf("expected snapshot list args, got %q", string(argsRaw))
	}
}

func TestSnapshotCreate(t *testing.T) {
	tempDir := t.TempDir()

	argsFile := filepath.Join(tempDir, "args.txt")
	fakeDdevPath := filepath.Join(tempDir, fakeDdevSnapshotScriptName())
	if err := os.WriteFile(fakeDdevPath, []byte(fakeDdevSnapshotScript()), 0755); err != nil {
		t.Fatalf("failed to write fake ddev script: %v", err)
	}

	originalPath := os.Getenv("PATH")
	t.Setenv("PATH", tempDir+string(os.PathListSeparator)+originalPath)
	t.Setenv("TEST_DDEV_ARGS_FILE", argsFile)

	svc := &DdevService{
		config: &ConfigService{data: map[string]any{"backend": "local"}},
	}

	_, err := svc.SnapshotCreate("", "")
	if err == nil {
		t.Fatal("expected error for empty project name, got nil")
	}

	_, err = svc.SnapshotCreate("demo", "")
	if err != nil {
		t.Fatalf("SnapshotCreate returned error: %v", err)
	}

	argsRaw, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("failed to read fake ddev args: %v", err)
	}
	if !strings.Contains(string(argsRaw), "snapshot demo") {
		t.Fatalf("expected snapshot args, got %q", string(argsRaw))
	}

	_, err = svc.SnapshotCreate("demo", "my-snapshot")
	if err != nil {
		t.Fatalf("SnapshotCreate returned error: %v", err)
	}

	argsRaw, err = os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("failed to read fake ddev args: %v", err)
	}
	if !strings.Contains(string(argsRaw), "snapshot demo --name=my-snapshot") {
		t.Fatalf("expected snapshot --name args, got %q", string(argsRaw))
	}
}

func TestSnapshotRestore(t *testing.T) {
	tempDir := t.TempDir()

	projectDir := filepath.Join(tempDir, "project")
	if err := os.Mkdir(projectDir, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}

	describePayload := `{"raw": {"approot": "` + filepath.ToSlash(projectDir) + `"}}`
	describeFile := filepath.Join(tempDir, "describe.json")
	if err := os.WriteFile(describeFile, []byte(describePayload), 0644); err != nil {
		t.Fatalf("failed to write describe payload: %v", err)
	}

	argsFile := filepath.Join(tempDir, "args.txt")
	fakeDdevPath := filepath.Join(tempDir, fakeDdevSnapshotScriptName())
	if err := os.WriteFile(fakeDdevPath, []byte(fakeDdevSnapshotScript()), 0755); err != nil {
		t.Fatalf("failed to write fake ddev script: %v", err)
	}

	originalPath := os.Getenv("PATH")
	t.Setenv("PATH", tempDir+string(os.PathListSeparator)+originalPath)
	t.Setenv("TEST_DDEV_ARGS_FILE", argsFile)
	t.Setenv("TEST_DDEV_DESCRIBE_FILE", describeFile)

	svc := &DdevService{
		config: &ConfigService{data: map[string]any{"backend": "local"}},
	}

	_, err := svc.SnapshotRestore("", "my-snapshot")
	if err == nil {
		t.Fatal("expected error for empty project name, got nil")
	}

	_, err = svc.SnapshotRestore("demo", "")
	if err == nil {
		t.Fatal("expected error for empty snapshot name, got nil")
	}

	_, err = svc.SnapshotRestore("demo", "my-snapshot")
	if err != nil {
		t.Fatalf("SnapshotRestore returned error: %v", err)
	}

	argsRaw, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("failed to read fake ddev args: %v", err)
	}
	if !strings.Contains(string(argsRaw), "snapshot restore my-snapshot") {
		t.Fatalf("expected snapshot restore args, got %q", string(argsRaw))
	}
}

func TestSnapshotDelete(t *testing.T) {
	tempDir := t.TempDir()

	projectDir := filepath.Join(tempDir, "project")
	if err := os.Mkdir(projectDir, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}

	describePayload := `{"raw": {"approot": "` + filepath.ToSlash(projectDir) + `"}}`
	describeFile := filepath.Join(tempDir, "describe.json")
	if err := os.WriteFile(describeFile, []byte(describePayload), 0644); err != nil {
		t.Fatalf("failed to write describe payload: %v", err)
	}

	argsFile := filepath.Join(tempDir, "args.txt")
	fakeDdevPath := filepath.Join(tempDir, fakeDdevSnapshotScriptName())
	if err := os.WriteFile(fakeDdevPath, []byte(fakeDdevSnapshotScript()), 0755); err != nil {
		t.Fatalf("failed to write fake ddev script: %v", err)
	}

	originalPath := os.Getenv("PATH")
	t.Setenv("PATH", tempDir+string(os.PathListSeparator)+originalPath)
	t.Setenv("TEST_DDEV_ARGS_FILE", argsFile)
	t.Setenv("TEST_DDEV_DESCRIBE_FILE", describeFile)

	svc := &DdevService{
		config: &ConfigService{data: map[string]any{"backend": "local"}},
	}

	_, err := svc.SnapshotDelete("", "my-snapshot")
	if err == nil {
		t.Fatal("expected error for empty project name, got nil")
	}

	_, err = svc.SnapshotDelete("demo", "")
	if err == nil {
		t.Fatal("expected error for empty snapshot name, got nil")
	}

	_, err = svc.SnapshotDelete("demo", "my-snapshot")
	if err != nil {
		t.Fatalf("SnapshotDelete returned error: %v", err)
	}

	argsRaw, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("failed to read fake ddev args: %v", err)
	}
	if !strings.Contains(string(argsRaw), "snapshot --cleanup --name my-snapshot -y") {
		t.Fatalf("expected snapshot cleanup args, got %q", string(argsRaw))
	}
}
