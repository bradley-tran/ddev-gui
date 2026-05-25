package backend

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNewConfigService(t *testing.T) {
	// Setup a temporary config directory
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	t.Setenv("AppData", tmpDir)
	t.Setenv("LOCALAPPDATA", tmpDir)
	t.Setenv("HOME", tmpDir)

	cs := NewConfigService()

	if cs == nil {
		t.Fatal("NewConfigService() returned nil")
	}

	// Check default values
	if cs.Get("openLinksInBrowser") != true {
		t.Errorf("Expected openLinksInBrowser to be true, got %v", cs.Get("openLinksInBrowser"))
	}
	if cs.Get("showLog") != false {
		t.Errorf("Expected showLog to be false, got %v", cs.Get("showLog"))
	}
	if cs.Get("devMode") != false {
		t.Errorf("Expected devMode to be false, got %v", cs.Get("devMode"))
	}

	projects := cs.Get("projects")
	if projects == nil {
		t.Fatal("projects map is nil")
	}
	projMap, ok := projects.(map[string]any)
	if !ok {
		t.Errorf("projects is not a map[string]any, got %T", projects)
	}
	if len(projMap) != 0 {
		t.Errorf("Expected empty projects map, got length %d", len(projMap))
	}

	// Check path
	expectedPath := filepath.Join(tmpDir, "ddev-gui", "config.json")
	if cs.path != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, cs.path)
	}
}

func TestConfigService_Save(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cs := &ConfigService{
		data: map[string]any{
			"foo": "bar",
		},
		path: configPath,
	}

	err := cs.Save()
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify file exists and has correct content
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("Config file was not created at %s", configPath)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read saved config: %v", err)
	}

	var loaded map[string]any
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Failed to unmarshal saved config: %v", err)
	}

	if loaded["foo"] != "bar" {
		t.Errorf("Expected foo to be bar, got %v", loaded["foo"])
	}
}

func TestConfigService_Load(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Create a dummy config file
	initialData := map[string]any{
		"openLinksInBrowser": false,
		"customKey":          "customValue",
	}
	bytes, _ := json.Marshal(initialData)
	_ = os.WriteFile(configPath, bytes, 0644)

	cs := &ConfigService{
		data: map[string]any{
			"openLinksInBrowser": true, // Default
			"otherKey":           "otherValue",
		},
		path: configPath,
	}

	cs.Load()

	if cs.data["openLinksInBrowser"] != false {
		t.Errorf("Expected openLinksInBrowser to be false (loaded), got %v", cs.data["openLinksInBrowser"])
	}
	if cs.data["customKey"] != "customValue" {
		t.Errorf("Expected customKey to be customValue, got %v", cs.data["customKey"])
	}
	if cs.data["otherKey"] != "otherValue" {
		t.Errorf("Expected otherKey to be otherValue (preserved), got %v", cs.data["otherKey"])
	}
}

func TestConfigService_GetSet(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cs := &ConfigService{
		data: map[string]any{},
		path: configPath,
	}

	err := cs.Set("key1", "value1")
	if err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	if cs.Get("key1") != "value1" {
		t.Errorf("Expected key1 to be value1, got %v", cs.Get("key1"))
	}

	// Verify it saved to disk
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("Config file was not created by Set()")
	}
}

func TestConfigService_GetAll(t *testing.T) {
	cs := &ConfigService{
		data: map[string]any{
			"k1": "v1",
			"k2": 123,
		},
	}

	all := cs.GetAll()
	var decoded map[string]any
	err := json.Unmarshal([]byte(all), &decoded)
	if err != nil {
		t.Fatalf("GetAll() returned invalid JSON: %v", err)
	}

	if decoded["k1"] != "v1" {
		t.Errorf("Expected k1: v1, got %v", decoded["k1"])
	}
	if decoded["k2"] != float64(123) {
		t.Errorf("Expected k2: 123, got %v", decoded["k2"])
	}
}

func TestConfigService_GetProjectConfig(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(cs *ConfigService)
		projectName string
		key         string
		want        any
	}{
		{
			name:        "happy path string",
			projectName: "valid-proj",
			key:         "key1",
			want:        "value1",
		},
		{
			name:        "happy path int",
			projectName: "valid-proj",
			key:         "key2",
			want:        42,
		},
		{
			name:        "missing key",
			projectName: "valid-proj",
			key:         "missing",
			want:        nil,
		},
		{
			name:        "missing project",
			projectName: "missing-proj",
			key:         "key1",
			want:        nil,
		},
		{
			name:        "invalid project type",
			projectName: "invalid-proj",
			key:         "key1",
			want:        nil,
		},
		{
			name: "invalid projects type",
			setup: func(cs *ConfigService) {
				cs.data["projects"] = "not-a-map"
			},
			projectName: "valid-proj",
			key:         "key1",
			want:        nil,
		},
		{
			name: "missing projects",
			setup: func(cs *ConfigService) {
				delete(cs.data, "projects")
			},
			projectName: "valid-proj",
			key:         "key1",
			want:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := &ConfigService{
				data: map[string]any{
					"projects": map[string]any{
						"valid-proj": map[string]any{
							"key1": "value1",
							"key2": 42,
						},
						"invalid-proj": "not-a-map",
					},
					"invalid-projects": "not-a-map",
				},
			}
			if tt.setup != nil {
				tt.setup(cs)
			}
			got := cs.GetProjectConfig(tt.projectName, tt.key)
			if got != tt.want {
				t.Errorf("GetProjectConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigService_ProjectConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cs := &ConfigService{
		data: map[string]any{
			"projects": map[string]any{},
		},
		path: configPath,
	}

	err := cs.SetProjectConfig("my-proj", "color", "red")
	if err != nil {
		t.Fatalf("SetProjectConfig() failed: %v", err)
	}

	val := cs.GetProjectConfig("my-proj", "color")
	if val != "red" {
		t.Errorf("Expected color to be red, got %v", val)
	}

	// Non-existent project
	if cs.GetProjectConfig("ghost", "color") != nil {
		t.Errorf("Expected nil for ghost project")
	}

	// Non-existent key
	if cs.GetProjectConfig("my-proj", "size") != nil {
		t.Errorf("Expected nil for non-existent key")
	}
}

func TestConfigService_WindowBounds(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cs := &ConfigService{
		data: map[string]any{},
		path: configPath,
	}

	// Test default when not set
	defaults := cs.GetWindowBounds()
	if defaults.Width != 0 || defaults.Height != 0 || defaults.X != 0 || defaults.Y != 0 || defaults.Maximized != false {
		t.Errorf("Expected zeroed window bounds when not set, got %+v", defaults)
	}

	bounds := WindowBounds{
		Width:     800,
		Height:    600,
		X:         100,
		Y:         100,
		Maximized: true,
	}

	cs.SetWindowBounds(bounds)

	got := cs.GetWindowBounds()
	if got != bounds {
		t.Errorf("Expected %+v, got %+v", bounds, got)
	}

	// Test persistence and JSON conversion (int vs float64)
	cs.Load() // Reload from disk
	gotAfterLoad := cs.GetWindowBounds()
	if gotAfterLoad != bounds {
		t.Errorf("Expected %+v after load, got %+v", bounds, gotAfterLoad)
	}
}

func TestIntFromJSON(t *testing.T) {
	tests := []struct {
		input    any
		expected int
		ok       bool
	}{
		{float64(123), 123, true},
		{int(456), 456, true},
		{"789", 0, false},
		{nil, 0, false},
		{true, 0, false},
	}

	for _, tt := range tests {
		val, ok := intFromJSON(tt.input)
		if val != tt.expected || ok != tt.ok {
			t.Errorf("intFromJSON(%v) = (%v, %v), want (%v, %v)", tt.input, val, ok, tt.expected, tt.ok)
		}
	}
}

func TestBoolFromJSON(t *testing.T) {
	tests := []struct {
		input    any
		expected bool
		ok       bool
	}{
		{true, true, true},
		{false, false, true},
		{"true", false, false},
		{nil, false, false},
		{1, false, false},
	}

	for _, tt := range tests {
		val, ok := boolFromJSON(tt.input)
		if val != tt.expected || ok != tt.ok {
			t.Errorf("boolFromJSON(%v) = (%v, %v), want (%v, %v)", tt.input, val, ok, tt.expected, tt.ok)
		}
	}
}
