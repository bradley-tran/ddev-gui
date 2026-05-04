package backend

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// SnapshotListJSON returns `ddev snapshot --list -j <project>` output as a JSON string.
func (d *DdevService) SnapshotListJSON(project string) (string, error) {
	project = strings.TrimSpace(project)
	if project == "" {
		return "", errors.New("project name is required")
	}
	out, errOut, err := d.run(context.Background(), "snapshot", "--list", "-j", project)
	if err != nil {
		if errOut != "" {
			return "", fmt.Errorf("ddev snapshot --list error: %s", strings.TrimSpace(errOut))
		}
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// SnapshotCreate creates a new database snapshot for the given project.
// If name is non-empty it is passed as --name=<name>.
func (d *DdevService) SnapshotCreate(project, name string) (string, error) {
	project = strings.TrimSpace(project)
	if project == "" {
		return "", errors.New("project name is required")
	}
	args := []string{"snapshot", project}
	name = strings.TrimSpace(name)
	if name != "" {
		args = append(args, "--name="+name)
	}
	return d.runDirect(context.Background(), "", nil, args...)
}

// SnapshotRestore restores a named snapshot for the given project via a fresh process.
func (d *DdevService) SnapshotRestore(project, name string) (string, error) {
	project = strings.TrimSpace(project)
	name = strings.TrimSpace(name)
	if project == "" || name == "" {
		return "", errors.New("project name and snapshot name are required")
	}
	dirHint := d.resolveProjectDir(project)
	return d.runDirect(context.Background(), dirHint, nil, "snapshot", "restore", name)
}

// SnapshotDelete removes a named snapshot for the given project via a fresh process.
func (d *DdevService) SnapshotDelete(project, name string) (string, error) {
	project = strings.TrimSpace(project)
	name = strings.TrimSpace(name)
	if project == "" || name == "" {
		return "", errors.New("project name and snapshot name are required")
	}
	dirHint := d.resolveProjectDir(project)
	return d.runDirect(context.Background(), dirHint, nil, "snapshot", "--cleanup", "--name", name, "-y")
}
