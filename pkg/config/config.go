package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	CacheRules   []CacheRule `yaml:"cache_rules"`
	Cache       CacheConfig  `yaml:"cache"`
	Server      ServerConfig `yaml:"server"`
	ParentProxy string
	ParentAuth  string
	BypassCache bool
}

type CacheRule struct {
	Pattern string `yaml:"pattern"`
	Enabled bool   `yaml:"enabled"`
}

type CacheConfig struct {
	Directory string `yaml:"directory"`
	MaxSizeMB int    `yaml:"max_size_mb"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Set defaults
	if cfg.Cache.Directory == "" {
		cfg.Cache.Directory = "/cache"
	}
	if cfg.Cache.MaxSizeMB == 0 {
		cfg.Cache.MaxSizeMB = 10240
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}

	// Override with environment variables
	if v := os.Getenv("CACHE_DIR"); v != "" {
		cfg.Cache.Directory = v
	}
	if v := os.Getenv("CACHE_SIZE_MB"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Cache.MaxSizeMB)
	}
	if v := os.Getenv("PORT"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Server.Port)
	}
	cfg.ParentProxy = os.Getenv("PARENT_PROXY")
	cfg.ParentAuth = os.Getenv("PARENT_PROXY_AUTH")
	if v := os.Getenv("BYPASS_CACHE"); v == "true" {
		cfg.BypassCache = true
	}

	return &cfg, nil
}
