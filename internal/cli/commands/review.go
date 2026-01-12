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
	"github.com/pm-assist/pm-assist/internal/ui"
	"github.com/spf13/cobra"
)

// NewReviewCmd returns the review command.
func NewReviewCmd(global *app.GlobalFlags) *cobra.Command {
	var (
		flagInput         string
		flagCase          string
		flagActivity      string
		flagTimestamp     string
		flagTimeFormat    string
		flagMissing       string
		flagDuplicate     string
		flagOrder         string
		flagParse         string
		flagAllowBlocking string
	)
	cmd := &cobra.Command{
		Use:   "review",
		Short: "Run QA checks and summarize issues",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.PrintCommandStart(ui.CommandFrame{
				Title:     "pm-assist review",
				Purpose:   "Run QA checks and produce a summary",
				StepIndex: 7,
				StepTotal: 7,
				Writes:    []string{"outputs/<run-id>/quality"},
				Asks:      []string{"thresholds"},
				Next:      "pm-assist report",
			})
			success := false
			defer func() {
				ui.PrintCommandEnd(ui.CommandFrame{Title: "pm-assist review", Next: "pm-assist report"}, success)
			}()
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

			inputPath, err := resolveString(flagInput, "Input log path", filepath.Join(outputPath, "stage_03_clean_filter", "filtered_log.csv"), true)
			if err != nil {
				return err
			}
			if _, err := os.Stat(inputPath); err != nil {
				return formatPathError(inputPath)
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
			timestampFormat, err := resolveString(flagTimeFormat, "Timestamp format (optional)", "", false)
			if err != nil {
				return err
			}

			missingThreshold, err := resolveString(flagMissing, "Missing value threshold", "0.05", true)
			if err != nil {
				return err
			}
			duplicateThreshold, err := resolveString(flagDuplicate, "Duplicate threshold", "0.02", true)
			if err != nil {
				return err
			}
			orderThreshold, err := resolveString(flagOrder, "Order violation threshold", "0.02", true)
			if err != nil {
				return err
			}
			parseThreshold, err := resolveString(flagParse, "Timestamp parse failure threshold", "0.02", true)
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
				allowBlocking, err := resolveBool(flagAllowBlocking, "Allow blocking issues?", false)
				if err != nil {
					return err
				}
				if global.NonInteractive {
					if allowBlocking {
						goto complete
					}
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

		complete:
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
			success = true

			fmt.Println("[SUCCESS] QA review completed.")
			updated, _ := config.Load(global.ConfigPath)
			ui.PrintSplash(updated, ui.SplashOptions{CompletedCommand: "review", WorkingDir: projectPath})
			return nil
		},
		Example: "  pm-assist review",
	}
	cmd.Flags().StringVar(&flagInput, "input", "", "Input log path")
	cmd.Flags().StringVar(&flagCase, "case", "", "Case ID column")
	cmd.Flags().StringVar(&flagActivity, "activity", "", "Activity column")
	cmd.Flags().StringVar(&flagTimestamp, "timestamp", "", "Timestamp column")
	cmd.Flags().StringVar(&flagTimeFormat, "timestamp-format", "", "Timestamp format")
	cmd.Flags().StringVar(&flagMissing, "missing-threshold", "", "Missing value threshold")
	cmd.Flags().StringVar(&flagDuplicate, "duplicate-threshold", "", "Duplicate threshold")
	cmd.Flags().StringVar(&flagOrder, "order-threshold", "", "Order violation threshold")
	cmd.Flags().StringVar(&flagParse, "parse-threshold", "", "Timestamp parse failure threshold")
	cmd.Flags().StringVar(&flagAllowBlocking, "allow-blocking", "", "Proceed despite blocking issues (true|false)")
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
