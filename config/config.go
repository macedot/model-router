package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"model-router/models"
)

var Defaults = struct {
	BodyLimit       int
	Port            uint16
	ReadTimeout     time.Duration
	ShutdownTimeout time.Duration
	WriteTimeout    time.Duration
}{
	BodyLimit:       50 * 1024 * 1024, // 50MB
	Port:            12345,
	ReadTimeout:     120 * time.Second,
	ShutdownTimeout: 120 * time.Second,
	WriteTimeout:    120 * time.Second,
}

var defaultConfig = models.Config{
	Port:   Defaults.Port,
	Models: []models.InternalModel{},
}

func Load() (*models.Config, error) {
	path, err := findConfigFile()
	if err != nil {
		if os.IsNotExist(err) {
			return &defaultConfig, nil
		}
		return nil, err
	}

	// Load .env files (optional, non-blocking)
	envPath := loadEnvFiles(path)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	// Expand env vars in the JSON string
	expanded := expandEnvVars(string(data))

	var cfg models.Config
	if err := json.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	log.Printf("config: %s", path)
	if envPath != "" {
		log.Printf("env: %s", envPath)
	}
	for _, im := range cfg.Models {
		names := make([]string, len(im.Externals))
		for i, em := range im.Externals {
			names[i] = em.Name
		}
		log.Printf("model: %s → [%s]", im.Name, strings.Join(names, ", "))
	}

	return &cfg, nil
}

func findConfigFile() (string, error) {
	// 1. Current working directory
	cwd, err := os.Getwd()
	if err == nil {
		candidate := filepath.Join(cwd, "config.json")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	// 2. ~/.config/model-router/config.json
	home, err := os.UserHomeDir()
	if err == nil {
		candidate := filepath.Join(home, ".config", "model-router", "config.json")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", os.ErrNotExist
}

func loadEnvFiles(configPath string) string {
	// Try current directory
	if path := loadEnv(".env"); path != "" {
		return path
	}

	// Try to expand tilde in configPath before getting directory
	if strings.HasPrefix(configPath, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			configPath = filepath.Join(home, configPath[1:])
		}
	}

	// Try same directory as config file
	if dir := filepath.Dir(configPath); dir != "." && dir != "" {
		if path := loadEnv(filepath.Join(dir, ".env")); path != "" {
			return path
		}
	}

	return ""
}

func loadEnv(path string) string {
	file, err := os.Open(path)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("warning: loading env file %s: %v", path, err)
		}
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := parts[1]
		// Remove surrounding quotes if present
		value = strings.Trim(value, `"'`)
		os.Setenv(key, value)
	}

	return path
}

func expandEnvVars(s string) string {
	return os.Expand(s, func(key string) string {
		if strings.HasPrefix(key, "env:") {
			return os.Getenv(key[4:])
		}
		return key
	})
}
