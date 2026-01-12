package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/pm-assist/pm-assist/internal/app"
	"github.com/pm-assist/pm-assist/internal/config"
	"github.com/pm-assist/pm-assist/internal/policy"
	"github.com/pm-assist/pm-assist/internal/ui"
	"github.com/spf13/cobra"
)

// NewStatusCmd returns the status command.
func NewStatusCmd(global *app.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show project dashboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.PrintCommandStart(ui.CommandFrame{
				Title:   "pm-assist status",
				Purpose: "Show current project status and next steps",
				Next:    "pm-assist connect",
			})
			success := false
			defer func() {
				ui.PrintCommandEnd(ui.CommandFrame{Title: "pm-assist status"}, success)
			}()

			projectPath := global.ProjectPath
			if projectPath == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				projectPath = cwd
			}

			cfg, err := config.Load(global.ConfigPath)
			if err != nil {
				return err
			}
			policies := policy.FromConfig(cfg)

			fmt.Printf("Project: %s\n", cfg.Project.Name)
			fmt.Printf("Policy: offline_only=%t, llm_enabled=%t\n", policies.OfflineOnly, boolValue(policies.LLMEnabled))
			if cfg.Profiles.Active != "" {
				fmt.Printf("Active profile: %s\n", cfg.Profiles.Active)
			}
			if cfg.Business.Active != "" {
				fmt.Printf("Active business: %s\n", cfg.Business.Active)
			}

			fmt.Printf("Connectors: %d\n", len(cfg.Connectors))
			rows := [][]string{}
			for _, connector := range cfg.Connectors {
				status := "unknown"
				if connector.Type == "file" && connector.File != nil && len(connector.File.Paths) > 0 {
					if _, err := os.Stat(connector.File.Paths[0]); err == nil {
						status = "ok"
					} else {
						status = "missing"
					}
				}
				rows = append(rows, []string{connector.Name, connector.Type, status})
			}
			if len(rows) > 0 {
				if err := ui.RenderTable([]string{"Name", "Type", "Status"}, rows); err != nil {
					for _, row := range rows {
						fmt.Printf("  - %s (%s) [%s]\n", row[0], row[1], row[2])
					}
				}
			}

			manifest, _ := latestManifest(projectPath)
			if manifest != nil {
				fmt.Printf("Last run: %s (%s)\n", manifest.RunID, manifest.Status)
				next := nextRecommendedStep(manifest)
				if next != "" {
					fmt.Printf("Next recommended: pm-assist %s\n", next)
				}
			} else {
				fmt.Println("Last run: none")
				fmt.Println("Next recommended: pm-assist connect")
			}
			success = true
			return nil
		},
	}
	return cmd
}

type runManifest struct {
	RunID       string `json:"run_id"`
	Status      string `json:"status"`
	StartedAt   string `json:"started_at"`
	CompletedAt string `json:"completed_at"`
	Steps       []struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	} `json:"steps"`
}

func latestManifest(projectPath string) (*runManifest, error) {
	pattern := filepath.Join(projectPath, "outputs", "*", "run_manifest.json")
	candidates, err := filepath.Glob(pattern)
	if err != nil || len(candidates) == 0 {
		return nil, err
	}
	sort.Slice(candidates, func(i, j int) bool {
		infoI, errI := os.Stat(candidates[i])
		infoJ, errJ := os.Stat(candidates[j])
		if errI != nil || errJ != nil {
			return candidates[i] < candidates[j]
		}
		return infoI.ModTime().After(infoJ.ModTime())
	})
	latest := candidates[0]
	data, err := os.ReadFile(latest)
	if err != nil {
		return nil, err
	}
	var manifest runManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}
	return &manifest, nil
}

func boolValue(value *bool) bool {
	if value == nil {
		return false
	}
	return *value
}

func nextRecommendedStep(manifest *runManifest) string {
	if manifest == nil {
		return ""
	}
	pipeline := []string{"ingest", "map", "prepare", "mine", "report", "review"}
	statusByStep := map[string]string{}
	for _, step := range manifest.Steps {
		statusByStep[step.Name] = step.Status
	}
	for _, step := range pipeline {
		status := statusByStep[step]
		if status == "" || status == "failed" || status == "started" {
			return step
		}
	}
	return ""
}
