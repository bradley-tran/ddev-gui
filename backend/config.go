package backend

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// ConfigService manages persistent application configuration stored as JSON.
type ConfigService struct {
	mu   sync.Mutex
	data map[string]any
	path string
}

// NewConfigService creates a new ConfigService and loads existing config from disk.
func NewConfigService() *ConfigService {
	cs := &ConfigService{
		data: map[string]any{
			"openLinksInBrowser": true,
			"showLog":            false,
			"devMode":            false,
			"projects":           map[string]any{},
		},
	}
	cs.data["preferredEditor"] = "vscode"
	cs.path = cs.configPath()
	cs.Load()
	return cs
}

// configPath returns the full path to the config file using OS-standard directories.
func (cs *ConfigService) configPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	return filepath.Join(dir, "ddev-gui", "config.json")
}

// Load reads the config JSON from disk. If the file does not exist, defaults are kept.
func (cs *ConfigService) Load() {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	raw, err := os.ReadFile(cs.path)
	if err != nil {
		return // file doesn't exist yet, use defaults
	}
	var loaded map[string]any
	if err := json.Unmarshal(raw, &loaded); err != nil {
		return
	}
	// Merge loaded values into defaults (so new keys get defaults)
	for k, v := range loaded {
		cs.data[k] = v
	}
}

// Save writes the current config to disk, creating directories as needed.
func (cs *ConfigService) Save() error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	dir := filepath.Dir(cs.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(cs.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cs.path, raw, 0o644)
}

// GetAll returns the entire config as a JSON string.
func (cs *ConfigService) GetAll() string {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	raw, _ := json.Marshal(cs.data)
	return string(raw)
}

// Get returns the value for a top-level config key.
func (cs *ConfigService) Get(key string) any {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	return cs.data[key]
}

// Set sets a top-level config key and saves to disk.
func (cs *ConfigService) Set(key string, value any) error {
	cs.mu.Lock()
	cs.data[key] = value
	cs.mu.Unlock()
	return cs.Save()
}

// GetProjectConfig returns the value of a per-project config key.
func (cs *ConfigService) GetProjectConfig(projectName, key string) any {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	projects, ok := cs.data["projects"].(map[string]any)
	if !ok {
		return nil
	}
	proj, ok := projects[projectName].(map[string]any)
	if !ok {
		return nil
	}
	return proj[key]
}

// SetProjectConfig sets a per-project config key and saves to disk.
func (cs *ConfigService) SetProjectConfig(projectName, key string, value any) error {
	cs.mu.Lock()
	projects, ok := cs.data["projects"].(map[string]any)
	if !ok {
		projects = map[string]any{}
		cs.data["projects"] = projects
	}
	proj, ok := projects[projectName].(map[string]any)
	if !ok {
		proj = map[string]any{}
		projects[projectName] = proj
	}
	proj[key] = value
	cs.mu.Unlock()
	return cs.Save()
}

// WindowBounds holds the saved window position, size, and maximized state.
type WindowBounds struct {
	Width     int  `json:"width"`
	Height    int  `json:"height"`
	X         int  `json:"x"`
	Y         int  `json:"y"`
	Maximized bool `json:"maximized"`
}

// GetWindowBounds returns the saved window bounds from config.
// If no bounds have been saved, a zero-value WindowBounds is returned.
func (cs *ConfigService) GetWindowBounds() WindowBounds {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	raw, ok := cs.data["windowBounds"]
	if !ok {
		return WindowBounds{}
	}

	// The stored value is map[string]any after JSON round-trip.
	m, ok := raw.(map[string]any)
	if !ok {
		return WindowBounds{}
	}

	return WindowBounds{
		Width:     cs.intFromMap(m, "width"),
		Height:    cs.intFromMap(m, "height"),
		X:         cs.intFromMap(m, "x"),
		Y:         cs.intFromMap(m, "y"),
		Maximized: cs.boolFromMap(m, "maximized"),
	}
}

// intFromMap extracts an int from a map, logging a warning on type mismatch.
func (cs *ConfigService) intFromMap(m map[string]any, key string) int {
	v, exists := m[key]
	if !exists || v == nil {
		return 0
	}
	val, ok := intFromJSON(v)
	if !ok {
		log.Printf("[config] invalid type for window %s: %T", key, v)
	}
	return val
}

// boolFromMap extracts a bool from a map, logging a warning on type mismatch.
func (cs *ConfigService) boolFromMap(m map[string]any, key string) bool {
	v, exists := m[key]
	if !exists || v == nil {
		return false
	}
	val, ok := boolFromJSON(v)
	if !ok {
		log.Printf("[config] invalid type for window %s: %T", key, v)
	}
	return val
}

// SetWindowBounds saves the given window bounds to config on disk.
func (cs *ConfigService) SetWindowBounds(b WindowBounds) {
	cs.mu.Lock()
	cs.data["windowBounds"] = map[string]any{
		"width":     b.Width,
		"height":    b.Height,
		"x":         b.X,
		"y":         b.Y,
		"maximized": b.Maximized,
	}
	cs.mu.Unlock()
	_ = cs.Save()
}

// intFromJSON safely converts a JSON-decoded number to int.
func intFromJSON(v any) (int, bool) {
	switch n := v.(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	default:
		return 0, false
	}
}

// boolFromJSON safely converts a JSON-decoded value to bool.
func boolFromJSON(v any) (bool, bool) {
	b, ok := v.(bool)
	return b, ok
}
