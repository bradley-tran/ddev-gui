package backend

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func fakeDdevDrushScriptName() string {
	if runtime.GOOS == "windows" {
		return "ddev.cmd"
	}
	return "ddev"
}

func fakeDdevDrushScript() string {
	if runtime.GOOS == "windows" {
		return "@echo off\r\n" +
			"if \"%1\"==\"describe\" (\r\n" +
			"  type \"%TEST_DDEV_DESCRIBE_FILE%\"\r\n" +
			"  exit /b 0\r\n" +
			")\r\n" +
			"if \"%1\"==\"drush\" (\r\n" +
			"  > \"%TEST_DDEV_ARGS_FILE%\" echo %*\r\n" +
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
		"  echo \"$@\" > \"$TEST_DDEV_ARGS_FILE\"\n" +
		"  exit 0\n" +
		"fi\n" +
		"echo unexpected args \"$@\" 1>&2\n" +
		"exit 1\n"
}

func TestDrushCacheRebuild(t *testing.T) {
	tempDir := t.TempDir()
	describeFile := filepath.Join(tempDir, "describe.json")
	if err := os.WriteFile(describeFile, []byte(`{"status": "running", "approot": "`+filepath.ToSlash(tempDir)+`"}`), 0644); err != nil {
		t.Fatalf("failed to write describe file: %v", err)
	}

	argsFile := filepath.Join(tempDir, "args.txt")
	fakeDdevPath := filepath.Join(tempDir, fakeDdevDrushScriptName())
	if err := os.WriteFile(fakeDdevPath, []byte(fakeDdevDrushScript()), 0755); err != nil {
		t.Fatalf("failed to write fake ddev script: %v", err)
	}

	originalPath := os.Getenv("PATH")
	t.Setenv("PATH", tempDir+string(os.PathListSeparator)+originalPath)
	t.Setenv("TEST_DDEV_ARGS_FILE", argsFile)
	t.Setenv("TEST_DDEV_DESCRIBE_FILE", describeFile)

	svc := &DdevService{
		config: &ConfigService{data: map[string]any{"backend": "local"}},
	}

	_, err := svc.DrushCacheRebuild("")
	if err == nil || !strings.Contains(err.Error(), "project name is required") {
		t.Fatalf("expected error for empty project name, got: %v", err)
	}

	_, err = svc.DrushCacheRebuild("demo")
	if err != nil {
		t.Fatalf("DrushCacheRebuild returned unexpected error: %v", err)
	}

	argsRaw, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("failed to read fake ddev args: %v", err)
	}
	if !strings.Contains(string(argsRaw), "drush cr") {
		t.Fatalf("expected drush cr args, got %q", string(argsRaw))
	}
}
