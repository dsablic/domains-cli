package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type CloudflareConfig struct {
	APIToken string `yaml:"api_token"`
}

type Config struct {
	Cloudflare CloudflareConfig `yaml:"cloudflare"`
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}

	if apiToken := os.Getenv("CLOUDFLARE_API_TOKEN"); apiToken != "" {
		cfg.Cloudflare.APIToken = apiToken
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
