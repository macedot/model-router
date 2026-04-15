package config

import (
	"bufio"
	"bytes"
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

type rawConfig struct {
	Port      uint16             `json:"port"`
	Providers []models.Provider  `json:"providers"`
	Models    []rawInternalModel `json:"models"`
}

type rawInternalModel struct {
	Name           string              `json:"name"`
	RequestFormat  models.RequestFormat `json:"request_format"`
	Strategy       models.Strategy     `json:"strategy"`
	RetryDelaySecs uint32              `json:"retry_delay_secs"`
	Externals      []json.RawMessage   `json:"externals"`
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

	var raw rawConfig
	if err := json.Unmarshal([]byte(expanded), &raw); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	providerMap, err := buildProviderMap(raw.Providers)
	if err != nil {
		return nil, err
	}

	cfg := &models.Config{
		Port:      raw.Port,
		Providers: raw.Providers,
		Models:    make([]models.InternalModel, len(raw.Models)),
	}

	for i, rm := range raw.Models {
		externals, err := resolveExternals(rm.Externals, providerMap, rm.Name)
		if err != nil {
			return nil, err
		}
		cfg.Models[i] = models.InternalModel{
			Name:           rm.Name,
			RequestFormat:  rm.RequestFormat,
			Strategy:       rm.Strategy,
			RetryDelaySecs: rm.RetryDelaySecs,
			Externals:      externals,
		}
	}

	log.Printf("config: %s", path)
	if envPath != "" {
		log.Printf("env: %s", envPath)
	}
	for _, p := range cfg.Providers {
		log.Printf("provider: %s (%s)", p.ID, p.Name)
	}
	for _, im := range cfg.Models {
		ids := make([]string, len(im.Externals))
		for i, em := range im.Externals {
			if em.ID != "" {
				ids[i] = em.ID
			} else {
				ids[i] = em.Name
			}
		}
		log.Printf("model: %s → [%s]", im.Name, strings.Join(ids, ", "))
	}

	validateConfig(cfg)

	return cfg, nil
}

func buildProviderMap(providers []models.Provider) (map[string]models.Provider, error) {
	m := make(map[string]models.Provider, len(providers))
	for _, p := range providers {
		if p.ID == "" {
			return nil, fmt.Errorf("provider %q: id is required", p.Name)
		}
		if _, exists := m[p.ID]; exists {
			return nil, fmt.Errorf("duplicate provider id: %q", p.ID)
		}
		m[p.ID] = p
	}
	return m, nil
}

func resolveExternals(rawExternals []json.RawMessage, providerMap map[string]models.Provider, modelName string) ([]models.ExternalModel, error) {
	result := make([]models.ExternalModel, 0, len(rawExternals))
	for _, raw := range rawExternals {
		trimmed := bytes.TrimSpace(raw)

		// String = provider reference
		if len(trimmed) > 0 && trimmed[0] == '"' {
			var id string
			if err := json.Unmarshal(raw, &id); err != nil {
				return nil, fmt.Errorf("model %q: invalid external reference: %w", modelName, err)
			}
			provider, ok := providerMap[id]
			if !ok {
				return nil, fmt.Errorf("model %q: unknown provider %q", modelName, id)
			}
			result = append(result, provider.ToExternal())
		} else {
			// Inline ExternalModel (old format)
			var ext models.ExternalModel
			if err := json.Unmarshal(raw, &ext); err != nil {
				return nil, fmt.Errorf("model %q: invalid external: %w", modelName, err)
			}
			result = append(result, ext)
		}
	}
	return result, nil
}

func validateConfig(cfg *models.Config) {
	for _, im := range cfg.Models {
		for _, em := range im.Externals {
			if strings.HasPrefix(em.APIKey, "env:") {
				log.Printf("warning: model %q external %q: api_key %q not expanded — use ${VAR} syntax", im.Name, em.Name, em.APIKey)
			} else if em.APIKey == "" {
				log.Printf("warning: model %q external %q: api_key is empty", im.Name, em.Name)
			}
		}
	}
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
	return os.Expand(s, os.Getenv)
}
