package backend

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// AddonsJSON returns the output of `ddev add-on list --installed -j`.
func (d *DdevService) AddonsJSON(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("project name is required")
	}
	// Provide working directory hint similar to ComposerInstall
	dirHint := d.resolveProjectDir(name)
	ctx := context.WithValue(context.Background(), "dir", dirHint)
	out, errOut, err := d.run(ctx, "add-on", "list", "--installed", "-j")
	if err != nil {
		if errOut != "" {
			return "", fmt.Errorf("ddev add-on list error: %s", strings.TrimSpace(errOut))
		}
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// AddonsAvailableJSON returns the output of `ddev add-on list -j` for available add-ons.
func (d *DdevService) AddonsAvailableJSON(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("project name is required")
	}
	dirHint := d.resolveProjectDir(name)
	ctx := context.WithValue(context.Background(), "dir", dirHint)
	out, errOut, err := d.run(ctx, "add-on", "list", "-j")
	if err != nil {
		if errOut != "" {
			return "", fmt.Errorf("ddev add-on list error: %s", strings.TrimSpace(errOut))
		}
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// AddonInstall installs a ddev add-on into the given project directory.
// It runs `ddev add-on get <addon>` inside the project via a fresh process.
func (d *DdevService) AddonInstall(name, addon string) (string, error) {
	name = strings.TrimSpace(name)
	addon = strings.TrimSpace(addon)
	if name == "" || addon == "" {
		return "", errors.New("project name and addon are required")
	}
	dirHint := d.resolveProjectDir(name)
	return d.runDirect(context.Background(), dirHint, nil, "add-on", "get", addon)
}

// AddonRemove removes/uninstalls a ddev add-on from the given project directory.
// It runs `ddev add-on remove <addon>` inside the project via a fresh process.
func (d *DdevService) AddonRemove(name, addon string) (string, error) {
	name = strings.TrimSpace(name)
	addon = strings.TrimSpace(addon)
	if name == "" || addon == "" {
		return "", errors.New("project name and addon are required")
	}
	dirHint := d.resolveProjectDir(name)
	return d.runDirect(context.Background(), dirHint, nil, "add-on", "remove", addon)
}
