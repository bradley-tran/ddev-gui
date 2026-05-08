package backend

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// stripAnsi removes ANSI escape sequences from a string.
func stripAnsi(s string) string {
	var out []byte
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			// Skip until we find the terminating letter
			j := i + 2
			for j < len(s) && !((s[j] >= 'A' && s[j] <= 'Z') || (s[j] >= 'a' && s[j] <= 'z')) {
				j++
			}
			if j < len(s) {
				j++ // skip the letter
			}
			i = j
		} else {
			out = append(out, s[i])
			i++
		}
	}
	return strings.TrimSpace(string(out))
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

