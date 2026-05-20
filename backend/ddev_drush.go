package backend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// DrushUli runs `ddev drush uli` for the given project and returns the one-time login URL.
// If drush is not installed, it auto-installs it via `ddev composer require drush/drush` first.
func (d *DdevService) DrushUli(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("project name is required")
	}
	dirHint := d.resolveProjectDir(name)

	out, err := d.runDirect(context.Background(), dirHint, nil, "drush", "uli")
	combined := strings.TrimSpace(out)

	// Check if drush needs to be installed first
	if err != nil && strings.Contains(strings.ToLower(combined), "drush is not available") {
		if d.ctx != nil {
			wruntime.EventsEmit(d.ctx, "ddev:output", "Drush not found - installing drush/drush via Composer…")
		}
		_, installErr := d.runDirect(context.Background(), dirHint, nil, "composer", "require", "drush/drush")
		if installErr != nil {
			return "", fmt.Errorf("failed to install drush: %v", installErr)
		}
		if d.ctx != nil {
			wruntime.EventsEmit(d.ctx, "ddev:output", "Drush installed. Generating admin login URL…")
		}
		// Retry after installing drush
		out, err = d.runDirect(context.Background(), dirHint, nil, "drush", "uli")
		combined = strings.TrimSpace(out)
	}

	// Try to extract a URL from the output (drush uli outputs a URL)
	for _, line := range strings.Split(combined, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
			return line, nil
		}
	}
	if err != nil {
		cleaned := stripAnsi(combined)
		if cleaned != "" {
			return "", fmt.Errorf("drush uli error: %s", cleaned)
		}
		return "", err
	}
	if strings.TrimSpace(out) != "" {
		return strings.TrimSpace(out), nil
	}
	return "", errors.New("drush uli returned no URL - is this a Drupal project with a working site?")
}

// DrushUliAsUser runs `ddev drush uli --uid=<uid>` for the given project and returns
// a one-time login URL for that specific Drupal user.
// If drush is not installed, it auto-installs it via `ddev composer require drush/drush` first.
func (d *DdevService) DrushUliAsUser(name, uid string) (string, error) {
	name = strings.TrimSpace(name)
	uid = strings.TrimSpace(uid)
	if name == "" {
		return "", errors.New("project name is required")
	}
	if uid == "" {
		return "", errors.New("user ID is required")
	}
	dirHint := d.resolveProjectDir(name)

	// Clear all Drupal sessions server-side so the ULI link works
	// even if a user is already logged in via the browser.
	if d.ctx != nil {
		wruntime.EventsEmit(d.ctx, "ddev:output", "Clearing existing sessions…")
	}
	d.runDirect(context.Background(), dirHint, nil, "drush", "sql:query", "TRUNCATE sessions")

	out, err := d.runDirect(context.Background(), dirHint, nil, "drush", "uli", "--uid="+uid)
	combined := strings.TrimSpace(out)

	// Check if drush needs to be installed first
	if err != nil && strings.Contains(strings.ToLower(combined), "drush is not available") {
		if d.ctx != nil {
			wruntime.EventsEmit(d.ctx, "ddev:output", "Drush not found - installing drush/drush via Composer…")
		}
		_, installErr := d.runDirect(context.Background(), dirHint, nil, "composer", "require", "drush/drush")
		if installErr != nil {
			return "", fmt.Errorf("failed to install drush: %v", installErr)
		}
		if d.ctx != nil {
			wruntime.EventsEmit(d.ctx, "ddev:output", "Drush installed. Generating login URL…")
		}
		// Retry after installing drush
		out, err = d.runDirect(context.Background(), dirHint, nil, "drush", "uli", "--uid="+uid)
		combined = strings.TrimSpace(out)
	}

	// Try to extract a URL from the output (drush uli outputs a URL)
	for _, line := range strings.Split(combined, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
			return line, nil
		}
	}
	if err != nil {
		cleaned := stripAnsi(combined)
		if cleaned != "" {
			return "", fmt.Errorf("drush uli error: %s", cleaned)
		}
		return "", err
	}
	if strings.TrimSpace(out) != "" {
		return strings.TrimSpace(out), nil
	}
	return "", errors.New("drush uli returned no URL - is this a Drupal project with a working site?")
}

// DrushRecentUsers queries the Drupal database for the 20 most recently accessed
// user accounts. Returns a JSON array of objects with uid, name, and mail fields.
func (d *DdevService) DrushRecentUsers(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "[]", errors.New("project name is required")
	}
	dirHint := d.resolveProjectDir(name)

	query := "SELECT uid, name, mail FROM users_field_data WHERE uid > 0 ORDER BY access DESC, uid ASC LIMIT 20"
	out, err := d.runDirect(context.Background(), dirHint, nil, "drush", "sql:query", query)
	if err != nil {
		return "[]", fmt.Errorf("failed to query users: %v", err)
	}

	// Parse tab-separated output into JSON array
	type drupalUser struct {
		UID  string `json:"uid"`
		Name string `json:"name"`
		Mail string `json:"mail"`
	}
	var users []drupalUser
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 2 {
			continue
		}
		u := drupalUser{UID: parts[0], Name: parts[1]}
		if len(parts) >= 3 {
			u.Mail = parts[2]
		}
		users = append(users, u)
	}
	if users == nil {
		users = []drupalUser{}
	}
	b, _ := json.Marshal(users)
	return string(b), nil
}

// DrushSiteInstall runs `ddev drush site:install -y` to initialize the Drupal database.
// If drush is not installed, it auto-installs it via `ddev composer require drush/drush` first.
// All commands run via fresh processes to avoid blocking the persistent shell.
func (d *DdevService) DrushSiteInstall(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("project name is required")
	}
	dirHint := d.resolveProjectDir(name)

	password, pErr := GenerateRandomString(12)
	if pErr != nil {
		return "", fmt.Errorf("failed to generate admin password: %w", pErr)
	}

	if d.ctx != nil {
		wruntime.EventsEmit(d.ctx, "ddev:output", "Running Drupal site install…")
		wruntime.EventsEmit(d.ctx, "ddev:output", "Drupal Admin User: admin")
		wruntime.EventsEmit(d.ctx, "ddev:output", "Drupal Admin Password: "+password)
	}

	// First try running drush si to see if drush is available
	out, err := d.runDirect(context.Background(), dirHint, nil, "drush", "site:install", "--account-name=admin", "--account-pass="+password, "-y")

	// Check if drush needs to be installed first
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "drush is not available") {
		if d.ctx != nil {
			wruntime.EventsEmit(d.ctx, "ddev:output", "Drush not found - installing drush/drush via Composer…")
		}
		_, installErr := d.runDirect(context.Background(), dirHint, nil, "composer", "require", "drush/drush")
		if installErr != nil {
			return "", fmt.Errorf("failed to install drush: %v", installErr)
		}
		if d.ctx != nil {
			wruntime.EventsEmit(d.ctx, "ddev:output", "Drush installed. Running site install…")
		}
		// Retry after installing drush
		out, err = d.runDirect(context.Background(), dirHint, nil, "drush", "site:install", "--account-name=admin", "--account-pass="+password, "-y")
	}

	return out, err
}

// DrushCacheRebuild runs `ddev drush cr` to clear/rebuild the Drupal cache.
func (d *DdevService) DrushCacheRebuild(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("project name is required")
	}
	dirHint := d.resolveProjectDir(name)
	return d.runDirect(context.Background(), dirHint, nil, "drush", "cr")
}
