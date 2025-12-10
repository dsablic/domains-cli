package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type CloudflareConfig struct {
	APIKey string `yaml:"api_key"`
	Email  string `yaml:"email"`
}

type Config struct {
	Cloudflare CloudflareConfig `yaml:"cloudflare"`
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}

	if apiKey := os.Getenv("CLOUDFLARE_API_KEY"); apiKey != "" {
		cfg.Cloudflare.APIKey = apiKey
		cfg.Cloudflare.Email = os.Getenv("CLOUDFLARE_EMAIL")
		return cfg, nil
	}

	configPath := configFilePath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	return cfg, nil
}

func configFilePath() string {
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "domains", "config.yaml")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "domains", "config.yaml")
}
