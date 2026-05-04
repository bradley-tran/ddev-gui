package backend

import (
	"context"
	"errors"
	"fmt"
	stdruntime "runtime"
	"strconv"
	"strings"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// Status returns `ddev describe <project>` output.
func (d *DdevService) Status(project string) (string, error) {
	project = strings.TrimSpace(project)
	if project == "" {
		return "", errors.New("project name is required")
	}
	out, errOut, err := d.run(context.Background(), "describe", project)
	if err != nil {
		if errOut != "" {
			return "", fmt.Errorf("ddev describe error: %s", strings.TrimSpace(errOut))
		}
		return "", err
	}
	return out, nil
}

// Start runs `ddev start <project>` via a fresh process so it doesn't block fast queries.
func (d *DdevService) Start(project string) (string, error) {
	project = strings.TrimSpace(project)
	if project == "" {
		return "", errors.New("project name is required")
	}
	return d.runDirect(context.Background(), "", nil, "start", project)
}

// Stop runs `ddev stop <project>` via a fresh process so it doesn't block fast queries.
func (d *DdevService) Stop(project string) (string, error) {
	project = strings.TrimSpace(project)
	if project == "" {
		return "", errors.New("project name is required")
	}
	return d.runDirect(context.Background(), "", nil, "stop", project)
}

// Restart runs `ddev restart <project>` via a fresh process so it doesn't block fast queries.
func (d *DdevService) Restart(project string) (string, error) {
	project = strings.TrimSpace(project)
	if project == "" {
		return "", errors.New("project name is required")
	}
	return d.runDirect(context.Background(), "", nil, "restart", project)
}

// PowerOff runs `ddev poweroff` to stop all projects and containers.
func (d *DdevService) PowerOff() (string, error) {
	return d.runDirect(context.Background(), "", nil, "poweroff")
}

// DeleteProject runs `ddev delete -O -y <project>` via a fresh process.
func (d *DdevService) DeleteProject(project string) (string, error) {
	project = strings.TrimSpace(project)
	if project == "" {
		return "", errors.New("project name is required")
	}
	return d.runDirect(context.Background(), "", nil, "delete", "-O", "-y", project)
}

// ExportDB runs `ddev export-db -p <project>` and saves the gzipped SQL dump to a user-chosen file.
// A native save-file dialog is presented so the user can pick the destination.
func (d *DdevService) ExportDB(project string) (string, error) {
	if d.activeBackend() == "ssh" {
		return "", errors.New("database export is not supported with the SSH backend - use the remote server's command line instead")
	}
	project = strings.TrimSpace(project)
	if project == "" {
		return "", errors.New("project name is required")
	}

	// Show native save-file dialog
	if d.ctx == nil {
		return "", errors.New("application context not set")
	}
	defaultName := project + "-db.sql.gz"
	savePath, err := wruntime.SaveFileDialog(d.ctx, wruntime.SaveDialogOptions{
		Title:           "Export Database - " + project,
		DefaultFilename: defaultName,
		Filters: []wruntime.FileFilter{
			{DisplayName: "Gzipped SQL (*.sql.gz)", Pattern: "*.sql.gz"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return "", fmt.Errorf("save dialog error: %w", err)
	}
	if savePath == "" {
		return "", errors.New("export cancelled")
	}

	// On Windows, convert the native path to a WSL path so ddev (running in WSL) can find it.
	// e.g. D:\Users\foo\db.sql.gz -> /mnt/d/Users/foo/db.sql.gz
	if stdruntime.GOOS == "windows" && len(savePath) >= 2 && savePath[1] == ':' {
		drive := strings.ToLower(string(savePath[0]))
		rest := strings.ReplaceAll(savePath[2:], "\\", "/")
		savePath = "/mnt/" + drive + rest
	}

	// Run `ddev export-db <project> -f <path>` via a fresh process
	return d.runDirect(context.Background(), "", nil, "export-db", project, "-f", savePath)
}

// ImportDBSelectFile opens a native file picker for a SQL dump and returns the selected file path.
// On Windows the path is converted to a WSL-compatible path.
func (d *DdevService) ImportDBSelectFile(project string) (string, error) {
	if d.activeBackend() == "ssh" {
		return "", errors.New("database import is not supported with the SSH backend - use the remote server's command line instead")
	}
	if d.ctx == nil {
		return "", errors.New("application context not set")
	}
	openPath, err := wruntime.OpenFileDialog(d.ctx, wruntime.OpenDialogOptions{
		Title: "Import Database - " + strings.TrimSpace(project),
		Filters: []wruntime.FileFilter{
			{DisplayName: "SQL Files (*.sql;*.sql.gz;*.sql.bz2;*.sql.xz;*.gz;*.bz2;*.xz;*.zip;*.tar;*.tar.gz;*.tgz)", Pattern: "*.sql;*.sql.gz;*.sql.bz2;*.sql.xz;*.gz;*.bz2;*.xz;*.zip;*.tar;*.tar.gz;*.tgz"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return "", fmt.Errorf("open dialog error: %w", err)
	}
	if openPath == "" {
		return "", errors.New("import cancelled")
	}
	// On Windows, convert the native path to a WSL path so ddev (running in WSL) can find it.
	if stdruntime.GOOS == "windows" && len(openPath) >= 2 && openPath[1] == ':' {
		drive := strings.ToLower(string(openPath[0]))
		rest := strings.ReplaceAll(openPath[2:], "\\", "/")
		openPath = "/mnt/" + drive + rest
	}
	return openPath, nil
}

// ImportDBFromFile runs `ddev import-db <project> --file=<path>` via a fresh process.
func (d *DdevService) ImportDBFromFile(project, filePath string) (string, error) {
	project = strings.TrimSpace(project)
	filePath = strings.TrimSpace(filePath)
	if project == "" || filePath == "" {
		return "", errors.New("project name and file path are required")
	}
	return d.runDirect(context.Background(), "", nil, "import-db", project, "--file="+filePath)
}

// ModifyProject updates selected settings (php version, nodejs version, project type, docroot)
// on an existing project by running `ddev config` inside the project directory.
// Empty string values are skipped - only non-empty fields are passed to the command.
func (d *DdevService) ModifyProject(name, phpVersion, nodejsVersion, projectType, docroot string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("project name is required")
	}
	phpVersion = strings.TrimSpace(phpVersion)
	nodejsVersion = strings.TrimSpace(nodejsVersion)
	projectType = strings.TrimSpace(projectType)
	docroot = strings.TrimSpace(docroot)

	projectDir := d.resolveProjectDir(name)

	args := []string{"config"}
	if phpVersion != "" {
		args = append(args, "--php-version", phpVersion)
	}
	if nodejsVersion != "" {
		args = append(args, "--nodejs-version", nodejsVersion)
	}
	if projectType != "" {
		args = append(args, "--project-type", projectType)
	}
	if docroot != "" {
		args = append(args, "--docroot", docroot)
	}

	if len(args) == 1 {
		return "", errors.New("no settings to update")
	}

	return d.runDirect(context.Background(), projectDir, nil, args...)
}

func validateServicePort(portValue, portLabel string) error {
	if portValue == "" {
		return nil
	}

	portNum, err := strconv.Atoi(portValue)
	if err != nil || portNum < 1 || portNum > 65535 {
		return fmt.Errorf("%s must be a number between 1 and 65535", portLabel)
	}

	return nil
}

func serviceToggle(enabled bool) string {
	if enabled {
		return "on"
	}
	return "off"
}

// ConfigureServices updates web/db host ports and debugger/profiler settings
// for an existing project. Commands are run in the project directory.
func (d *DdevService) ConfigureServices(name, webPort, dbPort string, xdebugEnabled, xhprofEnabled, xhguiEnabled bool) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("project name is required")
	}

	webPort = strings.TrimSpace(webPort)
	dbPort = strings.TrimSpace(dbPort)

	if err := validateServicePort(webPort, "web port"); err != nil {
		return "", err
	}
	if err := validateServicePort(dbPort, "db port"); err != nil {
		return "", err
	}

	// XHGui relies on XHProf collection, so keep the state consistent.
	if xhguiEnabled {
		xhprofEnabled = true
	}

	projectDir := d.resolveProjectDir(name)
	var outputs []string
	appendOutput := func(raw string) {
		raw = strings.TrimSpace(raw)
		if raw != "" {
			outputs = append(outputs, raw)
		}
	}

	// Ports are configured via ddev config flags.
	configArgs := []string{"config", "--auto"}
	if webPort != "" {
		configArgs = append(configArgs, "--host-webserver-port="+webPort)
	}
	if dbPort != "" {
		configArgs = append(configArgs, "--host-db-port="+dbPort)
	}
	if len(configArgs) > 2 {
		out, err := d.runDirect(context.Background(), projectDir, nil, configArgs...)
		if err != nil {
			return "", err
		}
		appendOutput(out)
	}

	// Feature toggles use dedicated ddev commands and are run from projectDir.
	xdebugOut, err := d.runDirect(context.Background(), projectDir, nil, "xdebug", serviceToggle(xdebugEnabled))
	if err != nil {
		return "", err
	}
	appendOutput(xdebugOut)

	xhprofOut, err := d.runDirect(context.Background(), projectDir, nil, "xhprof", serviceToggle(xhprofEnabled))
	if err != nil {
		return "", err
	}
	appendOutput(xhprofOut)

	xhguiOut, err := d.runDirect(context.Background(), projectDir, nil, "xhgui", serviceToggle(xhguiEnabled))
	if err != nil {
		return "", err
	}
	appendOutput(xhguiOut)

	return strings.Join(outputs, "\n"), nil
}

// ProjectLogs returns `ddev logs -s <service>` output for the given project.
func (d *DdevService) ProjectLogs(project string, service string) (string, error) {
	project = strings.TrimSpace(project)
	if project == "" {
		return "", errors.New("project name is required")
	}
	service = strings.TrimSpace(service)
	if service == "" {
		service = "web"
	}

	dirHint := d.resolveProjectDir(project)
	ctx := context.WithValue(context.Background(), "dir", dirHint)

	out, errOut, err := d.run(ctx, "logs", "-s", service)
	if err != nil {
		message := strings.TrimSpace(errOut)
		if message == "" {
			message = strings.TrimSpace(out)
		}
		if message != "" {
			return "", fmt.Errorf("ddev logs error: %s", message)
		}
		return "", err
	}

	return strings.TrimSpace(out), nil
}
