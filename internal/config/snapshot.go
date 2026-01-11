package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// WriteSnapshot writes a redacted config snapshot to the provided path.
func (c *Config) WriteSnapshot(path string) error {
	if path == "" {
		return nil
	}
	snapshot := c.RedactedCopy()
	data, err := yaml.Marshal(snapshot)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// RedactedCopy returns a copy suitable for audit logs.
func (c *Config) RedactedCopy() *Config {
	if c == nil {
		return &Config{}
	}
	copyCfg := *c
	copyCfg.Path = ""
	copyCfg.Connectors = make([]ConnectorSpec, len(c.Connectors))
	for i, connector := range c.Connectors {
		copyCfg.Connectors[i] = connector
		if connector.Options != nil {
			options := *connector.Options
			copyCfg.Connectors[i].Options = &options
		}
		if connector.Database != nil {
			dbCfg := *connector.Database
			copyCfg.Connectors[i].Database = &dbCfg
		}
		if connector.File != nil {
			fileCfg := *connector.File
			copyCfg.Connectors[i].File = &fileCfg
		}
	}
	return &copyCfg
}
