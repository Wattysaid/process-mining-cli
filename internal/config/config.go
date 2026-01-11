package config

import (
	"os"
	"path/filepath"
)

// Config holds resolved configuration. Expand fields as the CLI grows.
type Config struct {
	Path string
}

// Load returns a Config with the resolved path if a config exists.
func Load(path string) (*Config, error) {
	resolved := path
	if resolved == "" {
		if cwd, err := os.Getwd(); err == nil {
			candidate := filepath.Join(cwd, "pm-assist.yaml")
			if _, err := os.Stat(candidate); err == nil {
				resolved = candidate
			}
		}
	}
	return &Config{Path: resolved}, nil
}
