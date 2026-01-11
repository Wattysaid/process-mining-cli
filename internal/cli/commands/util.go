package commands

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pm-assist/pm-assist/internal/cli/prompt"
	"github.com/pm-assist/pm-assist/internal/config"
	"github.com/pm-assist/pm-assist/internal/logging"
	"github.com/pm-assist/pm-assist/internal/manifest"
	"github.com/pm-assist/pm-assist/internal/paths"
	"github.com/pm-assist/pm-assist/internal/policy"
	"github.com/pm-assist/pm-assist/internal/runner"
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

func resolveString(flagValue string, question string, defaultValue string, required bool) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}
	return prompt.AskString(question, defaultValue, required)
}

func resolveChoice(flagValue string, question string, options []string, defaultValue string, required bool) (string, error) {
	if flagValue != "" {
		for _, option := range options {
			if strings.EqualFold(option, flagValue) {
				return option, nil
			}
		}
		return "", fmt.Errorf("invalid value %q for %s (options: %s)", flagValue, question, strings.Join(options, ", "))
	}
	return prompt.AskChoice(question, options, defaultValue, required)
}

func resolveBool(flagValue string, question string, defaultValue bool) (bool, error) {
	if flagValue != "" {
		parsed, err := strconv.ParseBool(flagValue)
		if err != nil {
			return false, fmt.Errorf("invalid boolean for %s: %w", question, err)
		}
		return parsed, nil
	}
	return prompt.AskBool(question, defaultValue)
}

func resolveVenvOptions(projectPath string, policies policy.Policy) (runner.VenvOptions, error) {
	options := runner.VenvOptions{Offline: policies.OfflineOnly}
	wheelsRoot, err := paths.WheelsRoot(projectPath)
	if err == nil {
		options.WheelsPath = wheelsRoot
	}
	if options.Offline && options.WheelsPath == "" {
		return options, fmt.Errorf("offline-only mode requires bundled wheels; set %s", "PM_ASSIST_WHEELS_DIR")
	}
	return options, nil
}
