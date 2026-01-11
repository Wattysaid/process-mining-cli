package commands

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/pm-assist/pm-assist/internal/config"
	"github.com/pm-assist/pm-assist/internal/logging"
	"github.com/pm-assist/pm-assist/internal/manifest"
)

func defaultRunID() string {
	return fmt.Sprintf("%s", time.Now().UTC().Format("20060102-150405"))
}

func initRunManifest(runID string, outputPath string, cfg *config.Config) (*manifest.Manager, error) {
	if err := logging.SetRunLog(outputPath); err != nil {
		return nil, err
	}
	manager, _, err := manifest.NewManager(runID, outputPath)
	if err != nil {
		return nil, err
	}
	if cfg != nil {
		snapshotPath := filepath.Join(outputPath, "config_snapshot.yaml")
		if err := cfg.WriteSnapshot(snapshotPath); err != nil {
			return nil, err
		}
		if err := manager.SetConfigSnapshot(snapshotPath); err != nil {
			return nil, err
		}
	}
	return manager, nil
}
