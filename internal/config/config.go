package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// Add type for config
type Config struct {
	Context string `toml:"context"`
}

func getConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	yaduDir := filepath.Join(configDir, "yadu")
	if err := os.MkdirAll(yaduDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(yaduDir, "config.toml"), nil
}

func LoadConfig() (Config, error) {
	var cfg Config
	configPath, err := getConfigPath()
	if err != nil {
		return cfg, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("failed to read config file: %w", err)
	}

	err = toml.Unmarshal(data, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("failed to parse config file: %w", err)
	}
	return cfg, nil
}

func (cfg *Config) SetContext(contextPath string) error {
	// get full path for context path
	absPath, err := filepath.Abs(contextPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for context: %w", err)
	}
	cfg.Context = absPath
	return cfg.save()
}

func (cfg Config) save() error {
	configPath, err := getConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
