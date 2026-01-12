package commands

import (
	"fmt"
	"os"
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

func printWalkthrough(title string, steps []string) {
	if len(steps) == 0 {
		return
	}
	fmt.Println()
	fmt.Println(title)
	for i, step := range steps {
		fmt.Printf("  %d) %s\n", i+1, step)
	}
	fmt.Println()
}

func printStepProgress(step int, total int, label string) {
	bar := progressBar(step, total, 20)
	fmt.Printf("[%s] %d/%d %s\n", bar, step, total, label)
}

func progressBar(step int, total int, width int) string {
	if total <= 0 {
		return strings.Repeat("-", width)
	}
	if step < 0 {
		step = 0
	}
	if step > total {
		step = total
	}
	filled := int(float64(step) / float64(total) * float64(width))
	if filled > width {
		filled = width
	}
	return strings.Repeat("#", filled) + strings.Repeat("-", width-filled)
}

func confirmSummary(title string, lines []string) (bool, error) {
	fmt.Println()
	fmt.Println(title)
	for _, line := range lines {
		fmt.Printf("  - %s\n", line)
	}
	fmt.Println()
	return prompt.AskBool("Continue?", true)
}

func isWindowsPath(path string) bool {
	if len(path) < 2 {
		return false
	}
	drive := path[0]
	if drive < 'A' || (drive > 'Z' && drive < 'a') || drive > 'z' {
		return false
	}
	return len(path) > 2 && path[1] == ':' && (path[2] == '\\' || path[2] == '/')
}

func inWSL() bool {
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(data)), "microsoft")
}

func normalizePathInput(path string) string {
	trimmed := strings.Trim(path, "\"'")
	if inWSL() && isWindowsPath(trimmed) {
		drive := strings.ToLower(string(trimmed[0]))
		rest := strings.ReplaceAll(trimmed[2:], "\\", "/")
		return fmt.Sprintf("/mnt/%s/%s", drive, strings.TrimPrefix(rest, "/"))
	}
	return trimmed
}

func formatPathError(path string) error {
	if inWSL() && isWindowsPath(path) {
		converted := normalizePathInput(path)
		return fmt.Errorf("path not found: %s. You are running under WSL; try %s", path, converted)
	}
	return fmt.Errorf("path not found: %s", path)
}
