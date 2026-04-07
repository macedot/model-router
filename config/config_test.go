package config

import (
	"os"
	"path/filepath"
	"testing"
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

	// expandEnvVars uses os.Expand with a custom mapping that handles "env:" prefix
	// Only ${env:VAR_NAME} syntax works; $env:VAR does NOT because os.Expand
	// stops at colons when extracting variable names
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"${env:TEST_VAR}", "test-value"},
		{"${env:UNDEFINED}", ""},
		{"prefix-${env:TEST_VAR}-suffix", "prefix-test-value-suffix"},
		{"$env:TEST_VAR", "env:TEST_VAR"},        // $env doesn't expand (env var not set)
		{"plain ${VAR}", "plain VAR"},              // $VAR doesn't expand (VAR not set)
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