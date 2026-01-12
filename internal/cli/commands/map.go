package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pm-assist/pm-assist/internal/app"
	"github.com/pm-assist/pm-assist/internal/config"
	"github.com/pm-assist/pm-assist/internal/logging"
	"github.com/pm-assist/pm-assist/internal/policy"
	"github.com/pm-assist/pm-assist/internal/preview"
	"github.com/spf13/cobra"
)

// NewMapCmd returns the map command.
func NewMapCmd(global *app.GlobalFlags) *cobra.Command {
	var (
		flagInput      string
		flagCase       string
		flagActivity   string
		flagTimestamp  string
		flagResource   string
		flagTimeFormat string
		flagTimezone   string
		flagDelimiter  string
		flagEncoding   string
		flagPreview    string
	)
	cmd := &cobra.Command{
		Use:   "map",
		Short: "Map columns to process mining schema",
		RunE: func(cmd *cobra.Command, args []string) error {
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
			if policies.OfflineOnly {
				fmt.Println("[WARN] Offline-only policy is enabled; ensure local data sources are used.")
			}
			if cfg.Path == "" {
				cfg.Path = filepath.Join(projectPath, "pm-assist.yaml")
			}

			runID := global.RunID
			if runID == "" {
				runID = defaultRunID()
			}
			outputPath := filepath.Join(projectPath, "outputs", runID)
			if err := os.MkdirAll(outputPath, 0o755); err != nil {
				return err
			}
			manifestManager, err := initRunManifest(runID, outputPath, cfg)
			if err != nil {
				return err
			}
			defer logging.CloseRunLog()
			stepName := "map"
			if err := manifestManager.StartStep(stepName); err != nil {
				return err
			}
			stepSuccess := false
			defer func() {
				if !stepSuccess {
					_ = manifestManager.FailStep(stepName, "map failed")
					_ = manifestManager.SetStatus("failed")
				}
			}()

			defaultInput := filepath.Join(outputPath, "stage_01_ingest_profile", "normalised_log.csv")
			inputPath, err := resolveString(flagInput, "Input log path", defaultInput, true)
			if err != nil {
				return err
			}
			caseCol, err := resolveString(flagCase, "Case ID column", "case_id", true)
			if err != nil {
				return err
			}
			activityCol, err := resolveString(flagActivity, "Activity column", "activity", true)
			if err != nil {
				return err
			}
			timestampCol, err := resolveString(flagTimestamp, "Timestamp column", "timestamp", true)
			if err != nil {
				return err
			}
			resourceCol, err := resolveString(flagResource, "Resource column (optional)", "", false)
			if err != nil {
				return err
			}
			timestampFormat, err := resolveString(flagTimeFormat, "Timestamp format (optional)", "", false)
			if err != nil {
				return err
			}
			timezone, err := resolveString(flagTimezone, "Timezone (optional)", "", false)
			if err != nil {
				return err
			}

			if _, err := os.Stat(inputPath); err != nil {
				return formatPathError(inputPath)
			}

			previewNow, err := resolveBool(flagPreview, "Preview CSV headers and sample rows?", true)
			if err != nil {
				return err
			}
			if previewNow && strings.HasSuffix(strings.ToLower(inputPath), ".csv") {
				delimiter, err := resolveString(flagDelimiter, "CSV delimiter", ",", true)
				if err != nil {
					return err
				}
				encoding, err := resolveString(flagEncoding, "CSV encoding", "utf-8", true)
				if err != nil {
					return err
				}
				if strings.ToLower(encoding) != "utf-8" {
					fmt.Printf("[WARN] Preview skipped for encoding %s (only utf-8 supported).\n", encoding)
				} else {
					sample, err := preview.PreviewCSV(inputPath, delimiter, 5, false)
					if err != nil {
						fmt.Printf("[WARN] Preview failed: %v\n", err)
					} else {
						fmt.Println(preview.FormatSample(sample))
					}
				}
			}

			cfg.Mapping = &config.MappingConfig{
				InputPath:       inputPath,
				CaseID:          caseCol,
				Activity:        activityCol,
				Timestamp:       timestampCol,
				Resource:        resourceCol,
				TimestampFormat: timestampFormat,
				Timezone:        timezone,
			}
			if err := cfg.Save(); err != nil {
				return err
			}

			if err := manifestManager.AddInputs([]string{inputPath}); err != nil {
				return err
			}
			if err := manifestManager.AddOutputs([]string{outputPath}); err != nil {
				return err
			}
			if err := manifestManager.CompleteStep(stepName); err != nil {
				return err
			}
			if err := manifestManager.SetStatus("completed"); err != nil {
				return err
			}
			stepSuccess = true

			fmt.Println("[SUCCESS] Mapping saved.")
			return nil
		},
		Example: "  pm-assist map",
	}
	cmd.Flags().StringVar(&flagInput, "input", "", "Input log path")
	cmd.Flags().StringVar(&flagCase, "case", "", "Case ID column")
	cmd.Flags().StringVar(&flagActivity, "activity", "", "Activity column")
	cmd.Flags().StringVar(&flagTimestamp, "timestamp", "", "Timestamp column")
	cmd.Flags().StringVar(&flagResource, "resource", "", "Resource column")
	cmd.Flags().StringVar(&flagTimeFormat, "timestamp-format", "", "Timestamp format")
	cmd.Flags().StringVar(&flagTimezone, "timezone", "", "Timezone")
	cmd.Flags().StringVar(&flagDelimiter, "delimiter", "", "CSV delimiter")
	cmd.Flags().StringVar(&flagEncoding, "encoding", "", "CSV encoding")
	cmd.Flags().StringVar(&flagPreview, "preview", "", "Preview CSV headers and sample rows (true|false)")
	return cmd
}
