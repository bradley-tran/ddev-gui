package backend

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkResolveProjectDir(b *testing.B) {
	// Create a dummy config service and DdevService
	cfg := NewConfigService()
	svc := NewDdevService(cfg)

	// Create a fake ddev script in a temp dir to mock ddev describe
	tmpDir := b.TempDir()
	ddevScript := filepath.Join(tmpDir, "ddev")

	// Create dummy project directory
	dummyApproot := filepath.Join(tmpDir, "dummy_project")
	os.MkdirAll(dummyApproot, 0755)

	jsonOutput := fmt.Sprintf(`{"raw": {"approot": "%s"}}`, filepath.ToSlash(dummyApproot))

	scriptContent := fmt.Sprintf(`#!/bin/sh
if [ "$1" = "describe" ]; then
	echo '%s'
	exit 0
fi
echo "{}"
exit 0
`, jsonOutput)

	err := os.WriteFile(ddevScript, []byte(scriptContent), 0755)
	if err != nil {
		b.Fatalf("Failed to create dummy ddev script: %v", err)
	}

	// Make sure the directory with the fake ddev script is in PATH
	oldPath := os.Getenv("PATH")
	b.Setenv("PATH", tmpDir+string(os.PathListSeparator)+oldPath)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.resolveProjectDir("dummy_project")
	}
}
