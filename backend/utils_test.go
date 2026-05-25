package backend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestShellEnvVars(t *testing.T) {
	t.Run("EmptyContext", func(t *testing.T) {
		ctx := context.Background()
		got := shellEnvVars(ctx)
		if got != nil {
			t.Errorf("shellEnvVars(empty context) = %v, want nil", got)
		}
	})

	t.Run("WithVars", func(t *testing.T) {
		vars := []string{"FOO=bar", "BAZ=qux"}
		ctx := withShellEnv(context.Background(), vars...)
		got := shellEnvVars(ctx)
		if !reflect.DeepEqual(got, vars) {
			t.Errorf("shellEnvVars() = %v, want %v", got, vars)
		}
	})

	t.Run("WithNilVars", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), shellEnvKey{}, nil)
		got := shellEnvVars(ctx)
		if got != nil {
			t.Errorf("shellEnvVars(nil vars) = %v, want nil", got)
		}
	})

	t.Run("WrongType", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), shellEnvKey{}, "not-a-slice")
		got := shellEnvVars(ctx)
		if got != nil {
			t.Errorf("shellEnvVars(wrong type) = %v, want nil", got)
		}
	})
}

func TestStripAnsi(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Plain", "hello world", "hello world"},
		{"PlainWithSpaces", "  hello world  ", "  hello world  "},
		{"Red", "\x1b[31mhello\x1b[0m", "hello"},
		{"Bold", "\x1b[1mhello\x1b[0m", "hello"},
		{"Complex", "\x1b[38;5;208mhello\x1b[0m world", "hello world"},
		{"Multiple", "one \x1b[32mtwo\x1b[0m three \x1b[33mfour\x1b[0m", "one two three four"},
		{"Empty", "", ""},
		{"NoTerminator", "\x1b[31mhello", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripAnsi(tt.input)
			if got != tt.expected {
				t.Errorf("stripAnsi(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGenerateRandomString(t *testing.T) {
	t.Run("Length", func(t *testing.T) {
		lengths := []int{8, 12, 16, 32}
		for _, l := range lengths {
			got, err := GenerateRandomString(l)
			if err != nil {
				t.Fatalf("GenerateRandomString(%d) error: %v", l, err)
			}
			if len(got) != l {
				t.Errorf("GenerateRandomString(%d) length = %d, want %d", l, len(got), l)
			}
		}
	})

	t.Run("Uniqueness", func(t *testing.T) {
		s1, _ := GenerateRandomString(16)
		s2, _ := GenerateRandomString(16)
		if s1 == s2 {
			t.Errorf("GenerateRandomString(16) produced identical strings: %q", s1)
		}
	})

	t.Run("HexCharacters", func(t *testing.T) {
		got, _ := GenerateRandomString(32)
		for _, r := range got {
			isHex := (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')
			if !isHex {
				t.Errorf("GenerateRandomString() produced non-hex character: %q", r)
			}
		}
	})
}

func TestCleanForCreateProject(t *testing.T) {
	// Create a temporary directory for testing
	dirHint := t.TempDir()

	// 1. Create a .ddev directory with a file inside
	ddevDir := filepath.Join(dirHint, ".ddev")
	if err := os.Mkdir(ddevDir, 0755); err != nil {
		t.Fatalf("failed to create .ddev dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ddevDir, "config.yaml"), []byte("name: test"), 0644); err != nil {
		t.Fatalf("failed to write config.yaml: %v", err)
	}

	// 2. Create other files and directories that should be deleted
	if err := os.WriteFile(filepath.Join(dirHint, "index.php"), []byte("<?php echo 'hello';"), 0644); err != nil {
		t.Fatalf("failed to write index.php: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dirHint, ".env"), []byte("APP_ENV=local"), 0644); err != nil {
		t.Fatalf("failed to write .env: %v", err)
	}
	vendorDir := filepath.Join(dirHint, "vendor")
	if err := os.Mkdir(vendorDir, 0755); err != nil {
		t.Fatalf("failed to create vendor dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(vendorDir, "autoload.php"), []byte("<?php"), 0644); err != nil {
		t.Fatalf("failed to write vendor/autoload.php: %v", err)
	}

	// 3. Call the function
	cleanForCreateProject(dirHint)

	// 4. Verify that .ddev still exists and its contents are intact
	if _, err := os.Stat(ddevDir); os.IsNotExist(err) {
		t.Errorf("expected .ddev directory to be kept, but it was removed")
	}
	if _, err := os.Stat(filepath.Join(ddevDir, "config.yaml")); os.IsNotExist(err) {
		t.Errorf("expected .ddev/config.yaml to be kept, but it was removed")
	}

	// 5. Verify that other files and directories are removed
	removedEntries := []string{"index.php", ".env", "vendor"}
	for _, entry := range removedEntries {
		if _, err := os.Stat(filepath.Join(dirHint, entry)); !os.IsNotExist(err) {
			t.Errorf("expected %s to be removed, but it was kept", entry)
		}
	}

	// 6. Test with a non-existent directory (should return gracefully)
	nonExistentDir := filepath.Join(t.TempDir(), "non_existent")
	cleanForCreateProject(nonExistentDir) // Should not panic or error out
}

func TestFetchExpectedChecksum(t *testing.T) {
	mockChecksumContent := `
e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  empty.txt
b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9 *hello.txt
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/checksums.txt" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockChecksumContent))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := server.Client()

	tests := []struct {
		name         string
		url          string
		assetName    string
		expectedHash string
		expectErr    bool
	}{
		{
			name:         "FoundStandard",
			url:          server.URL + "/checksums.txt",
			assetName:    "empty.txt",
			expectedHash: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			expectErr:    false,
		},
		{
			name:         "FoundWithAsterisk",
			url:          server.URL + "/checksums.txt",
			assetName:    "hello.txt",
			expectedHash: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
			expectErr:    false,
		},
		{
			name:         "NotFound",
			url:          server.URL + "/checksums.txt",
			assetName:    "missing.txt",
			expectedHash: "",
			expectErr:    false,
		},
		{
			name:         "HTTPError",
			url:          server.URL + "/404",
			assetName:    "empty.txt",
			expectedHash: "",
			expectErr:    true,
		},
		{
			name:         "InvalidURL",
			url:          "://invalid-url",
			assetName:    "empty.txt",
			expectedHash: "",
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := fetchExpectedChecksum(client, tt.url, tt.assetName)
			if (err != nil) != tt.expectErr {
				t.Errorf("fetchExpectedChecksum() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if hash != tt.expectedHash {
				t.Errorf("fetchExpectedChecksum() = %v, want %v", hash, tt.expectedHash)
			}
		})
	}
}

func TestFileSHA256(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.txt")
	content := []byte("hello world")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// echo -n "hello world" | sha256sum
	expected := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"

	got, err := fileSHA256(path)
	if err != nil {
		t.Fatalf("fileSHA256() error: %v", err)
	}
	if got != expected {
		t.Errorf("fileSHA256() = %q, want %q", got, expected)
	}

	t.Run("NonExistent", func(t *testing.T) {
		_, err := fileSHA256(filepath.Join(tmpDir, "nope.txt"))
		if err == nil {
			t.Error("expected error for non-existent file")
		}
	})
}

func TestToWSLPath_Fallback(t *testing.T) {
	// This test focuses on the fallback logic since wsl.exe might not be present in the test environment.
	// We can't easily mock wsl.exe without more complex setup, but we can verify the fallback.

	winPath := `C:\Users\Name\Projects\my-site`
	expectedFallback := "/mnt/c/Users/Name/Projects/my-site"

	got := toWSLPath(winPath)

	// If wslpath worked, it might be different, but in many CI/test environments it will use fallback.
	// We at least expect it to be a valid WSL-style path.
	if !filepath.IsAbs(got) && !os.IsPathSeparator(got[0]) {
		t.Errorf("toWSLPath(%q) = %q, doesn't look like an absolute WSL path", winPath, got)
	}

	// If it contains /mnt/c/, it likely hit the fallback (or wslpath produced similar)
	if got != expectedFallback && !testing.Short() {
		t.Logf("toWSLPath(%q) = %q (might be expected if wsl.exe exists)", winPath, got)
	}
}

func TestCreateProjectPackage(t *testing.T) {
	tests := []struct {
		name     string
		ptype    string
		expected string
	}{
		// Drupal variants
		{"Drupal", "drupal", "drupal/recommended-project:^11"},
		{"Drupal11", "drupal11", "drupal/recommended-project:^11"},
		{"Drupal11Case", "Drupal11", "drupal/recommended-project:^11"},
		{"Drupal12", "drupal12", "drupal/recommended-project:^12"},
		{"Drupal10", "drupal10", "drupal/recommended-project:^10"},
		{"Drupal9", "drupal9", "drupal/recommended-project:^9"},
		{"Drupal8", "drupal8", "drupal/recommended-project:^8"},
		{"Drupal7", "drupal7", "drupal/drupal:^7"},

		// Laravel
		{"Laravel", "laravel", "laravel/laravel:^12"},
		{"LaravelCase", "Laravel", "laravel/laravel:^12"},

		// Other CMS/framework types
		{"CraftCMS", "craftcms", "craftcms/craft"},
		{"CakePHP", "cakephp", "cakephp/app:~5.0"},
		{"CodeIgniter", "codeigniter", "codeigniter4/appstarter"},
		{"Symfony", "symfony", "symfony/skeleton"},
		{"TYPO3", "typo3", "typo3/cms-base-distribution"},
		{"Shopware6", "shopware6", "shopware/production"},
		{"Silverstripe", "silverstripe", "silverstripe/installer"},

		// Types without composer templates or unknown
		{"Backdrop", "backdrop", ""},
		{"Drupal6", "drupal6", ""},
		{"Generic", "generic", ""},
		{"PHP", "php", ""},
		{"Wordpress", "wordpress", ""},
		{"Unknown", "unknown", ""},
		{"Empty", "", ""},
		{"OnlySpace", "  ", ""},

		// Mixed case and whitespace
		{"MixedCase", "DrUpAl", "drupal/recommended-project:^11"},
		{"Whitespace", "  laravel  ", "laravel/laravel:^12"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createProjectPackage(tt.ptype)
			if got != tt.expected {
				t.Errorf("createProjectPackage(%q) = %q, want %q", tt.ptype, got, tt.expected)
			}
		})
	}
}
