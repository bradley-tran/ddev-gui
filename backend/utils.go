package backend

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// stripAnsi removes ANSI escape sequences from a string.
func stripAnsi(s string) string {
	var builder strings.Builder
	builder.Grow(len(s))

	i := 0
	for i < len(s) {
		idx := strings.IndexByte(s[i:], '\x1b')
		if idx == -1 {
			builder.WriteString(s[i:])
			break
		}

		// Write everything up to the escape character
		builder.WriteString(s[i : i+idx])
		i += idx

		if i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && !((s[j] >= 'A' && s[j] <= 'Z') || (s[j] >= 'a' && s[j] <= 'z')) {
				j++
			}
			if j < len(s) {
				j++ // skip the letter
			}
			i = j
		} else {
			builder.WriteByte(s[i])
			i++
		}
	}
	return strings.TrimSpace(builder.String())
}

// shellEnvKey is the context key for extra env vars to export in the WSL shell.
type shellEnvKey struct{}

// shellEnvVars extracts the env var slice from ctx, or nil if not set.
func shellEnvVars(ctx context.Context) []string {
	if v := ctx.Value(shellEnvKey{}); v != nil {
		if s, ok := v.([]string); ok {
			return s
		}
	}
	return nil
}

// GenerateRandomString returns a cryptographically secure random hex string of length n.
func GenerateRandomString(n int) (string, error) {
	b := make([]byte, n/2+1)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b)[:n], nil
}

// withShellEnv returns a new context carrying the given env vars for the WSL shell.
func withShellEnv(ctx context.Context, vars ...string) context.Context {
	return context.WithValue(ctx, shellEnvKey{}, vars)
}

// createProjectPackage returns the composer create-project package string for a given project type.
func createProjectPackage(ptype string) string {
	switch strings.ToLower(strings.TrimSpace(ptype)) {
	case "drupal", "drupal11":
		return "drupal/recommended-project:^11"
	case "drupal12":
		return "drupal/recommended-project:^12"
	case "drupal10":
		return "drupal/recommended-project:^10"
	case "drupal9":
		return "drupal/recommended-project:^9"
	case "drupal8":
		return "drupal/recommended-project:^8"
	case "drupal7":
		return "drupal/drupal:^7"
	case "laravel":
		return "laravel/laravel:^12"
	case "craftcms":
		return "craftcms/craft"
	case "cakephp":
		return "cakephp/app:~5.0"
	case "codeigniter":
		return "codeigniter4/appstarter"
	case "symfony":
		return "symfony/skeleton"
	case "typo3":
		return "typo3/cms-base-distribution"
	case "shopware6":
		return "shopware/production"
	case "silverstripe":
		return "silverstripe/installer"
	default:
		return ""
	}
}

// cleanForCreateProject removes all files and directories from dirHint
// except the .ddev configuration directory, so `composer create-project` can succeed.
func cleanForCreateProject(dirHint string) {
	entries, err := os.ReadDir(dirHint)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.Name() == ".ddev" {
			continue
		}
		_ = os.RemoveAll(filepath.Join(dirHint, e.Name()))
	}
}

// toWSLPath converts a Windows path to a WSL-compatible path.
func toWSLPath(winPath string) string {
	// Try wslpath first
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "wsl.exe", "wslpath", "-a", winPath)
	HideWSLWindow(cmd)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err == nil {
		p := strings.TrimSpace(out.String())
		if p != "" {
			return p
		}
	}
	// Fallback naive conversion: C:\foo\bar -> /mnt/c/foo/bar
	p := strings.ReplaceAll(winPath, "\\", "/")
	if len(p) >= 2 && p[1] == ':' {
		drive := strings.ToLower(string(p[0]))
		p = "/mnt/" + drive + p[2:]
	}
	return p
}

var errUserCancelled = errors.New("installation cancelled")

// fileSHA256 computes the SHA-256 hex digest of the file at path.
func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// fetchExpectedChecksum downloads the checksums.txt from the given URL and
// returns the expected SHA-256 hex digest for assetName.
// Returns ("", nil) if the asset is not found in the checksum file.
func fetchExpectedChecksum(client *http.Client, checksumURL, assetName string) (string, error) {
	req, err := http.NewRequest("GET", checksumURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "ddev-gui/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("checksum fetch error: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	// Parse lines: <sha256>  <filename>
	for _, line := range strings.Split(string(body), "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) >= 2 {
			fileField := strings.TrimPrefix(fields[len(fields)-1], "*")
			if fileField == assetName {
				return strings.ToLower(fields[0]), nil
			}
		}
	}
	return "", nil
}

var (
	execCommand = exec.Command
	runtimeGOOS = runtime.GOOS
)

// launchWindowsElevated starts an executable via PowerShell with -Verb RunAs so UAC can elevate it.
// It waits for PowerShell to exit and maps a user-cancelled UAC prompt to errUserCancelled.
func launchWindowsElevated(path string) error {
	if runtimeGOOS != "windows" {
		return fmt.Errorf("elevation only supported on Windows")
	}
	safe := strings.ReplaceAll(path, "'", "''")
	// Use -ErrorAction Stop so cancellation produces a non-zero exit code.
	// Exit 1223 (ERROR_CANCELLED) when user cancels UAC if detectable; otherwise inspect output.
	ps := `$ErrorActionPreference='Stop'; try { Start-Process -FilePath '` + safe + `' -Verb RunAs; exit 0 } catch { if ($_.Exception -and ($_.Exception.NativeErrorCode -eq 1223 -or $_.Exception.HResult -eq -2147023675) ) { exit 1223 } else { Write-Output $_.Exception.Message; exit 1 } }`
	cmd := execCommand("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", ps)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		// Check for explicit cancel exit code
		if ee, ok := err.(*exec.ExitError); ok {
			if ee.ExitCode() == 1223 {
				return errUserCancelled
			}
		}
		// Fallback to string matching
		out := strings.ToLower(outBuf.String() + " " + errBuf.String())
		if strings.Contains(out, "canceled by the user") || strings.Contains(out, "cancelled by the user") {
			return errUserCancelled
		}
		return err
	}
	return nil
}
