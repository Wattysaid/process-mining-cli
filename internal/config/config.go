package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const CurrentSchemaVersion = 1

// Config holds resolved configuration. Expand fields as the CLI grows.
type Config struct {
	Path       string          `yaml:"-"`
	Version    int             `yaml:"version"`
	Project    ProjectConfig   `yaml:"project"`
	Profiles   ProfilesConfig  `yaml:"profiles"`
	Business   BusinessConfig  `yaml:"business"`
	LLM        LLMConfig       `yaml:"llm"`
	Policy     PolicyConfig    `yaml:"policy"`
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

type PolicyConfig struct {
	LLMEnabled        *bool    `yaml:"llm_enabled,omitempty"`
	OfflineOnly       bool     `yaml:"offline_only,omitempty"`
	AllowedConnectors []string `yaml:"allowed_connectors,omitempty"`
	DeniedConnectors  []string `yaml:"denied_connectors,omitempty"`
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
	Driver  string `yaml:"driver"`
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	DBName  string `yaml:"database"`
	Schema  string `yaml:"schema,omitempty"`
	User    string `yaml:"user,omitempty"`
	SSLMode string `yaml:"ssl_mode,omitempty"`
}

type ExtraConfig struct {
	ReadOnly      bool   `yaml:"read_only"`
	CredentialEnv string `yaml:"credential_env,omitempty"`
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
	cfg.applyDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	cfg.Path = resolved
	return cfg, nil
}

// Validate checks config schema version and basic constraints.
func (c *Config) Validate() error {
	if c.Version == 0 {
		c.Version = CurrentSchemaVersion
	}
	if c.Version != CurrentSchemaVersion {
		return fmt.Errorf("unsupported config schema version: %d", c.Version)
	}
	for _, connector := range c.Connectors {
		if connector.Type == "" {
			return errors.New("connector type is required")
		}
		if connector.Type == "database" && connector.Database == nil {
			return errors.New("database connector missing database config")
		}
		if connector.Type == "file" && connector.File == nil {
			return errors.New("file connector missing file config")
		}
	}
	return nil
}

func (c *Config) applyDefaults() {
	if c.Version == 0 {
		c.Version = CurrentSchemaVersion
	}
	if c.Policy.AllowedConnectors == nil {
		c.Policy.AllowedConnectors = []string{}
	}
	if c.Policy.DeniedConnectors == nil {
		c.Policy.DeniedConnectors = []string{}
	}
	if c.LLM.Provider == "" {
		c.LLM.Provider = "none"
	}
	if c.Policy.LLMEnabled == nil {
		defaultLLM := strings.ToLower(c.LLM.Provider) != "none"
		c.Policy.LLMEnabled = &defaultLLM
	}
}

// Save writes the config to its path.
func (c *Config) Save() error {
	if c.Path == "" {
		return errors.New("config path is required")
	}
	c.applyDefaults()
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(c.Path, data, 0o644)
}
