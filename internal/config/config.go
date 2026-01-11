package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds resolved configuration. Expand fields as the CLI grows.
type Config struct {
	Path       string          `yaml:"-"`
	Project    ProjectConfig   `yaml:"project"`
	Profiles   ProfilesConfig  `yaml:"profiles"`
	Business   BusinessConfig  `yaml:"business"`
	LLM        LLMConfig       `yaml:"llm"`
	Connectors []ConnectorSpec `yaml:"connectors"`
}

type ProjectConfig struct {
	Name string `yaml:"name"`
}

type ProfilesConfig struct {
	Active string `yaml:"active"`
}

type BusinessConfig struct {
	Active string `yaml:"active"`
}

type LLMConfig struct {
	Provider string `yaml:"provider"`
}

type ConnectorSpec struct {
	Name     string       `yaml:"name"`
	Type     string       `yaml:"type"`
	File     *FileConfig  `yaml:"file,omitempty"`
	Database *DBConfig    `yaml:"database,omitempty"`
	Options  *ExtraConfig `yaml:"options,omitempty"`
}

type FileConfig struct {
	Paths     []string `yaml:"paths"`
	Format    string   `yaml:"format"`
	Delimiter string   `yaml:"delimiter,omitempty"`
	Encoding  string   `yaml:"encoding,omitempty"`
}

type DBConfig struct {
	Driver string `yaml:"driver"`
	Host   string `yaml:"host"`
	Port   int    `yaml:"port"`
	DBName string `yaml:"database"`
	Schema string `yaml:"schema,omitempty"`
}

type ExtraConfig struct {
	ReadOnly bool `yaml:"read_only"`
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
	cfg := &Config{Path: resolved}
	if resolved == "" {
		return cfg, nil
	}
	data, err := os.ReadFile(resolved)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return nil, err
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	cfg.Path = resolved
	return cfg, nil
}

// Save writes the config to its path.
func (c *Config) Save() error {
	if c.Path == "" {
		return errors.New("config path is required")
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(c.Path, data, 0o644)
}
