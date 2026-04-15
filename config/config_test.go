package config

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"model-router/models"
)

func TestFindConfigFile_CurrentDir(t *testing.T) {
	// Create a temp dir and change to it
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldCwd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Create config.json in current directory
	configPath := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(configPath, []byte(`{"port": 12345}`), 0644); err != nil {
		t.Fatal(err)
	}

	path, err := findConfigFile()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if path != configPath {
		t.Fatalf("expected %s, got %s", configPath, path)
	}
}

func TestFindConfigFile_HomeDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot get home dir")
	}

	// Create ~/.config/model-router/config.json
	modelRouterDir := filepath.Join(home, ".config", "model-router")
	if err := os.MkdirAll(modelRouterDir, 0755); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(filepath.Join(home, ".config", "model-router"))

	configPath := filepath.Join(modelRouterDir, "config.json")
	if err := os.WriteFile(configPath, []byte(`{"port": 12345}`), 0644); err != nil {
		t.Fatal(err)
	}

	path, err := findConfigFile()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if path != configPath {
		t.Fatalf("expected %s, got %s", configPath, path)
	}
}

func TestFindConfigFile_NotFound(t *testing.T) {
	// Create empty temp dir
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldCwd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Remove home dir config if exists
	home, _ := os.UserHomeDir()
	modelRouterDir := filepath.Join(home, ".config", "model-router")
	os.RemoveAll(modelRouterDir)

	path, err := findConfigFile()
	if err == nil {
		t.Fatalf("expected error, got nil, path=%s", path)
	}
}

func TestExpandEnvVars(t *testing.T) {
	os.Setenv("TEST_VAR", "test-value")
	defer os.Unsetenv("TEST_VAR")

	// expandEnvVars uses os.Expand(os.Getenv) — standard ${VAR} syntax
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"${TEST_VAR}", "test-value"},
		{"${UNDEFINED}", ""},
		{"prefix-${TEST_VAR}-suffix", "prefix-test-value-suffix"},
		{"$TEST_VAR", "test-value"},
		{"plain ${VAR}", "plain "},
	}

	for _, tt := range tests {
		got := expandEnvVars(tt.input)
		if got != tt.expected {
			t.Errorf("expandEnvVars(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestLoadEnvFiles(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")

	content := `TEST_KEY=value
# comment
EMPTY_KEY=
QUOTED_KEY="quoted"
`
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Clear the env var first
	os.Unsetenv("TEST_KEY")
	os.Unsetenv("EMPTY_KEY")
	os.Unsetenv("QUOTED_KEY")

	loadEnv(envPath)

	if os.Getenv("TEST_KEY") != "value" {
		t.Errorf("TEST_KEY = %q, want %q", os.Getenv("TEST_KEY"), "value")
	}
	if os.Getenv("EMPTY_KEY") != "" {
		t.Errorf("EMPTY_KEY = %q, want empty", os.Getenv("EMPTY_KEY"))
	}
	if os.Getenv("QUOTED_KEY") != "quoted" {
		t.Errorf("QUOTED_KEY = %q, want %q", os.Getenv("QUOTED_KEY"), "quoted")
	}
}

func TestLoad_WithConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldCwd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	os.Setenv("TEST_ROUTER_KEY", "secret-123")
	defer os.Unsetenv("TEST_ROUTER_KEY")

	configContent := `{
		"port": 9999,
		"models": [{
			"name": "test",
			"format": "openai",
			"externals": [{
				"name": "ext",
				"url": "https://api.example.com/v1",
				"api_key": "${TEST_ROUTER_KEY}"
			}]
		}]
	}`
	if err := os.WriteFile(filepath.Join(tmpDir, "config.json"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Port != 9999 {
		t.Errorf("Port = %d, want 9999", cfg.Port)
	}
	if len(cfg.Models) != 1 {
		t.Fatalf("Models len = %d, want 1", len(cfg.Models))
	}
	if cfg.Models[0].Name != "test" {
		t.Errorf("Model Name = %q, want %q", cfg.Models[0].Name, "test")
	}
	if len(cfg.Models[0].Externals) != 1 {
		t.Fatalf("Externals len = %d, want 1", len(cfg.Models[0].Externals))
	}
	if cfg.Models[0].Externals[0].APIKey != "secret-123" {
		t.Errorf("APIKey = %q, want %q", cfg.Models[0].Externals[0].APIKey, "secret-123")
	}
}

func TestLoad_NoConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldCwd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Port != Defaults.Port {
		t.Errorf("Port = %d, want default %d", cfg.Port, Defaults.Port)
	}
	if len(cfg.Models) != 0 {
		t.Errorf("Models len = %d, want 0", len(cfg.Models))
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldCwd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "config.json"), []byte(`{invalid}`), 0644); err != nil {
		t.Fatal(err)
	}

	_, err = Load()
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestValidateConfig_EnvPrefixWarning(t *testing.T) {
	cfg := &models.Config{
		Models: []models.InternalModel{
			{
				Name: "test",
				Externals: []models.ExternalModel{
					{Name: "ext", URL: "https://api.example.com", APIKey: "env:MISSING_KEY"},
				},
			},
		},
	}

	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	validateConfig(cfg)

	if !bytes.Contains(buf.Bytes(), []byte(`api_key "env:MISSING_KEY" not expanded`)) {
		t.Errorf("expected warning about env: prefix, got: %s", buf.String())
	}
}

func TestValidateConfig_EmptyAPIKey(t *testing.T) {
	cfg := &models.Config{
		Models: []models.InternalModel{
			{
				Name: "test",
				Externals: []models.ExternalModel{
					{Name: "ext", URL: "https://api.example.com", APIKey: ""},
				},
			},
		},
	}

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	validateConfig(cfg)

	if !bytes.Contains(buf.Bytes(), []byte(`api_key is empty`)) {
		t.Errorf("expected warning about empty api_key, got: %s", buf.String())
	}
}

func TestLoad_ProviderReferences(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldCwd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	os.Setenv("TEST_ROUTER_KEY", "secret-123")
	defer os.Unsetenv("TEST_ROUTER_KEY")

	configContent := `{
		"port": 9999,
		"providers": [
			{
				"id": "p1",
				"name": "external-a",
				"url": "https://api.example.com/v1",
				"api_key": "${TEST_ROUTER_KEY}",
				"format": "openai"
			},
			{
				"id": "p2",
				"name": "external-b",
				"url": "https://api.other.com/v1",
				"api_key": "hardcoded-key",
				"format": "anthropic"
			}
		],
		"models": [
			{
				"name": "test",
				"externals": ["p1", "p2"]
			}
		]
	}`
	if err := os.WriteFile(filepath.Join(tmpDir, "config.json"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Port != 9999 {
		t.Errorf("Port = %d, want 9999", cfg.Port)
	}
	if len(cfg.Providers) != 2 {
		t.Fatalf("Providers len = %d, want 2", len(cfg.Providers))
	}
	if len(cfg.Models) != 1 {
		t.Fatalf("Models len = %d, want 1", len(cfg.Models))
	}
	if len(cfg.Models[0].Externals) != 2 {
		t.Fatalf("Externals len = %d, want 2", len(cfg.Models[0].Externals))
	}

	e1 := cfg.Models[0].Externals[0]
	if e1.Name != "external-a" {
		t.Errorf("External[0].Name = %q, want %q", e1.Name, "external-a")
	}
	if e1.URL != "https://api.example.com/v1" {
		t.Errorf("External[0].URL = %q, want %q", e1.URL, "https://api.example.com/v1")
	}
	if e1.APIKey != "secret-123" {
		t.Errorf("External[0].APIKey = %q, want %q", e1.APIKey, "secret-123")
	}
	if e1.Format != models.FormatOpenAI {
		t.Errorf("External[0].Format = %q, want %q", e1.Format, models.FormatOpenAI)
	}

	e2 := cfg.Models[0].Externals[1]
	if e2.Name != "external-b" {
		t.Errorf("External[1].Name = %q, want %q", e2.Name, "external-b")
	}
	if e2.Format != models.FormatAnthropic {
		t.Errorf("External[1].Format = %q, want %q", e2.Format, models.FormatAnthropic)
	}
}

func TestLoad_ProviderReference_UnknownID(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldCwd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	configContent := `{
		"providers": [],
		"models": [
			{
				"name": "test",
				"externals": ["nonexistent"]
			}
		]
	}`
	if err := os.WriteFile(filepath.Join(tmpDir, "config.json"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	_, err = Load()
	if err == nil {
		t.Fatal("expected error for unknown provider, got nil")
	}
	if !strings.Contains(err.Error(), `unknown provider "nonexistent"`) {
		t.Errorf("error = %q, want unknown provider error", err.Error())
	}
}

func TestLoad_ProviderReference_DuplicateID(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldCwd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	configContent := `{
		"providers": [
			{"id": "dup", "name": "a", "url": "https://a.com", "api_key": "k", "format": "openai"},
			{"id": "dup", "name": "b", "url": "https://b.com", "api_key": "k", "format": "openai"}
		],
		"models": []
	}`
	if err := os.WriteFile(filepath.Join(tmpDir, "config.json"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	_, err = Load()
	if err == nil {
		t.Fatal("expected error for duplicate provider id, got nil")
	}
	if !strings.Contains(err.Error(), `duplicate provider id`) {
		t.Errorf("error = %q, want duplicate provider id error", err.Error())
	}
}

func TestLoad_ProviderReference_MissingID(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldCwd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	configContent := `{
		"providers": [
			{"id": "", "name": "a", "url": "https://a.com", "api_key": "k", "format": "openai"}
		],
		"models": []
	}`
	if err := os.WriteFile(filepath.Join(tmpDir, "config.json"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	_, err = Load()
	if err == nil {
		t.Fatal("expected error for missing provider id, got nil")
	}
	if !strings.Contains(err.Error(), "id is required") {
		t.Errorf("error = %q, want id is required error", err.Error())
	}
}

func TestLoad_MixedFormats(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldCwd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	configContent := `{
		"providers": [
			{"id": "p1", "name": "shared-provider", "url": "https://shared.com", "api_key": "k1", "format": "openai"}
		],
		"models": [
			{
				"name": "model-with-ref",
				"externals": ["p1"]
			},
			{
				"name": "model-with-inline",
				"externals": [
					{"name": "inline-provider", "url": "https://inline.com", "api_key": "k2", "format": "anthropic"}
				]
			}
		]
	}`
	if err := os.WriteFile(filepath.Join(tmpDir, "config.json"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.Models) != 2 {
		t.Fatalf("Models len = %d, want 2", len(cfg.Models))
	}

	// Model with provider reference
	if len(cfg.Models[0].Externals) != 1 {
		t.Fatalf("Models[0].Externals len = %d, want 1", len(cfg.Models[0].Externals))
	}
	if cfg.Models[0].Externals[0].Name != "shared-provider" {
		t.Errorf("Models[0].Externals[0].Name = %q, want %q", cfg.Models[0].Externals[0].Name, "shared-provider")
	}

	// Model with inline external
	if len(cfg.Models[1].Externals) != 1 {
		t.Fatalf("Models[1].Externals len = %d, want 1", len(cfg.Models[1].Externals))
	}
	if cfg.Models[1].Externals[0].Name != "inline-provider" {
		t.Errorf("Models[1].Externals[0].Name = %q, want %q", cfg.Models[1].Externals[0].Name, "inline-provider")
	}
	if cfg.Models[1].Externals[0].Format != models.FormatAnthropic {
		t.Errorf("Models[1].Externals[0].Format = %q, want %q", cfg.Models[1].Externals[0].Format, models.FormatAnthropic)
	}
}

func TestValidateConfig_ValidKeys(t *testing.T) {
	cfg := &models.Config{
		Models: []models.InternalModel{
			{
				Name: "test",
				Externals: []models.ExternalModel{
					{Name: "ext", URL: "https://api.example.com", APIKey: "sk-valid-key"},
				},
			},
		},
	}

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	validateConfig(cfg)

	if buf.Len() > 0 {
		t.Errorf("expected no warnings for valid keys, got: %s", buf.String())
	}
}