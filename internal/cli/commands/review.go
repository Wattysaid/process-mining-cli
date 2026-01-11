package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pm-assist/pm-assist/internal/app"
	"github.com/pm-assist/pm-assist/internal/cli/prompt"
	"github.com/pm-assist/pm-assist/internal/config"
	"github.com/pm-assist/pm-assist/internal/logging"
	"github.com/pm-assist/pm-assist/internal/qa"
	"github.com/spf13/cobra"
)

// NewReviewCmd returns the review command.
func NewReviewCmd(global *app.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review",
		Short: "Run QA checks and summarize issues",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectPath := global.ProjectPath
			if projectPath == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				projectPath = cwd
			}

			runID := global.RunID
			if runID == "" {
				runID = defaultRunID()
			}
			outputPath := filepath.Join(projectPath, "outputs", runID)
			if err := os.MkdirAll(outputPath, 0o755); err != nil {
				return err
			}

			cfg, err := config.Load(global.ConfigPath)
			if err != nil {
				return err
			}

			manifestManager, err := initRunManifest(runID, outputPath, cfg)
			if err != nil {
				return err
			}
			defer logging.CloseRunLog()
			stepName := "review"
			if err := manifestManager.StartStep(stepName); err != nil {
				return err
			}
			stepSuccess := false
			defer func() {
				if !stepSuccess {
					_ = manifestManager.FailStep(stepName, "review failed")
					_ = manifestManager.SetStatus("failed")
				}
			}()

			inputPath, err := prompt.AskString("Input log path", filepath.Join(outputPath, "stage_03_clean_filter", "filtered_log.csv"), true)
			if err != nil {
				return err
			}
			caseCol, err := prompt.AskString("Case ID column", "case_id", true)
			if err != nil {
				return err
			}
			activityCol, err := prompt.AskString("Activity column", "activity", true)
			if err != nil {
				return err
			}
			timestampCol, err := prompt.AskString("Timestamp column", "timestamp", true)
			if err != nil {
				return err
			}
			timestampFormat, err := prompt.AskString("Timestamp format (optional)", "", false)
			if err != nil {
				return err
			}

			missingThreshold, err := prompt.AskString("Missing value threshold", "0.05", true)
			if err != nil {
				return err
			}
			duplicateThreshold, err := prompt.AskString("Duplicate threshold", "0.02", true)
			if err != nil {
				return err
			}
			orderThreshold, err := prompt.AskString("Order violation threshold", "0.02", true)
			if err != nil {
				return err
			}
			parseThreshold, err := prompt.AskString("Timestamp parse failure threshold", "0.02", true)
			if err != nil {
				return err
			}

			thresholds, err := parseThresholds(missingThreshold, duplicateThreshold, orderThreshold, parseThreshold)
			if err != nil {
				return err
			}

			if err := manifestManager.AddInputs([]string{inputPath}); err != nil {
				return err
			}

			results, backlog, err := qa.RunCSV(inputPath, caseCol, activityCol, timestampCol, timestampFormat, thresholds)
			if err != nil {
				return err
			}
			if err := qa.WriteOutputs(outputPath, results, backlog); err != nil {
				return err
			}

			if len(results.BlockingIssues) > 0 {
				fmt.Printf("[WARN] Blocking issues detected: %v\n", results.BlockingIssues)
				if global.NonInteractive {
					_ = manifestManager.FailStep(stepName, "blocking QA issues detected")
					_ = manifestManager.SetStatus("failed")
					return fmt.Errorf("blocking QA issues detected")
				}
				proceed, err := prompt.AskBool("Proceed despite blocking issues?", false)
				if err != nil {
					return err
				}
				if !proceed {
					_ = manifestManager.FailStep(stepName, "review halted by user")
					_ = manifestManager.SetStatus("failed")
					return fmt.Errorf("review halted due to blocking issues")
				}
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

			fmt.Println("[SUCCESS] QA review completed.")
			return nil
		},
		Example: "  pm-assist review",
	}
	return cmd
}

func parseThresholds(missing string, duplicate string, order string, parseFail string) (qa.Thresholds, error) {
	missingVal, err := strconv.ParseFloat(missing, 64)
	if err != nil {
		return qa.Thresholds{}, err
	}
	duplicateVal, err := strconv.ParseFloat(duplicate, 64)
	if err != nil {
		return qa.Thresholds{}, err
	}
	orderVal, err := strconv.ParseFloat(order, 64)
	if err != nil {
		return qa.Thresholds{}, err
	}
	parseVal, err := strconv.ParseFloat(parseFail, 64)
	if err != nil {
		return qa.Thresholds{}, err
	}
	return qa.Thresholds{
		MissingValue: missingVal,
		Duplicate:    duplicateVal,
		OrderViol:    orderVal,
		ParseFail:    parseVal,
	}, nil
}
