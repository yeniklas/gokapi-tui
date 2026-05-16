package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Defaults struct {
	AllowedDownloads int    `yaml:"allowed_downloads"`
	ExpiryDays       int    `yaml:"expiry_days"`
	Password         string `yaml:"password"`
}

type Config struct {
	ServerURL string   `yaml:"server_url"`
	APIKey    string   `yaml:"api_key"`
	Defaults  Defaults `yaml:"defaults"`
}

func DefaultPath() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("cannot find config dir: %w", err)
	}
	return filepath.Join(cfgDir, "gokapi-tui", "config.yaml"), nil
}

func Load() (*Config, error) {
	path, err := DefaultPath()
	if err != nil {
		return nil, err
	}
	return LoadFrom(path)
}

func LoadFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read config at %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	if cfg.ServerURL == "" {
		return nil, fmt.Errorf("server_url is required in %s", path)
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("api_key is required in %s", path)
	}
	if cfg.Defaults.AllowedDownloads == 0 {
		cfg.Defaults.AllowedDownloads = 1
	}
	if cfg.Defaults.ExpiryDays == 0 {
		cfg.Defaults.ExpiryDays = 7
	}
	return &cfg, nil
}
