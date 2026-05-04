package backend

// AppVersionStr and CommitHashStr are set at build time via -ldflags:
//
//	-X ddev-gui/backend.AppVersionStr=v1.0.0
//	-X ddev-gui/backend.CommitHashStr=abc1234
var (
	AppVersionStr = "dev"
	CommitHashStr = "unknown"
)

// AppVersionInfo holds the build-time version metadata returned to the frontend.
type AppVersionInfo struct {
	Version    string `json:"version"`
	CommitHash string `json:"commitHash"`
}

// AppVersion returns the build-time version and commit hash.
func (d *DdevService) AppVersion() AppVersionInfo {
	return AppVersionInfo{
		Version:    AppVersionStr,
		CommitHash: CommitHashStr,
	}
}
