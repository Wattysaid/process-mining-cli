package config

import "fmt"

// Migrate updates older config versions to the current schema.
func (c *Config) Migrate() error {
	if c == nil {
		return fmt.Errorf("config is nil")
	}
	if c.Version == 0 {
		c.Version = CurrentSchemaVersion
		return nil
	}
	if c.Version > CurrentSchemaVersion {
		return fmt.Errorf("unsupported config schema version: %d", c.Version)
	}
	return nil
}
