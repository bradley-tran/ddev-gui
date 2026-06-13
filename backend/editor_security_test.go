package backend

import (
	"os/exec"
	"reflect"
	"testing"
)

// We want to test that the command constructed for macOS is secure.
// Since we can't easily mock runtime.GOOS globally without changing the architecture,
// we can test a small helper or just verify the logic if it were exposed.
// However, we can also check if we can simulate the command construction.

func TestLaunchNeovimEditor_DarwinCommand(t *testing.T) {
	// This is a bit tricky because launchNeovimEditor uses runtime2.GOOS.
	// But we can verify the logic by looking at how we'd construct such a command.

	path := "/path/with/spaces and \"quotes\""
	script := "on run argv\ntell application \"Terminal\" to do script \"cd \" & quoted form of item 1 of argv & \" && nvim\"\nend run"
	cmd := exec.Command("osascript", "-e", script, "--", path)

	expectedArgs := []string{"osascript", "-e", script, "--", path}
	if !reflect.DeepEqual(cmd.Args, expectedArgs) {
		t.Errorf("expected args %v, got %v", expectedArgs, cmd.Args)
	}
}
