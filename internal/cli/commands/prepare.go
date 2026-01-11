package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pm-assist/pm-assist/internal/app"
	"github.com/pm-assist/pm-assist/internal/cli/prompt"
	"github.com/pm-assist/pm-assist/internal/config"
	"github.com/pm-assist/pm-assist/internal/logging"
	"github.com/pm-assist/pm-assist/internal/notebook"
	"github.com/pm-assist/pm-assist/internal/paths"
	"github.com/pm-assist/pm-assist/internal/policy"
	"github.com/pm-assist/pm-assist/internal/runner"
	"github.com/spf13/cobra"
)

// NewPrepareCmd returns the prepare command.
func NewPrepareCmd(global *app.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prepare",
		Short: "Run data preparation pipeline",
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
			policies := policy.FromConfig(cfg)
			if policies.OfflineOnly {
				fmt.Println("[WARN] Offline-only policy is enabled; ensure local data sources are used.")
			}

			manifestManager, err := initRunManifest(runID, outputPath, cfg)
			if err != nil {
				return err
			}
			defer logging.CloseRunLog()
			stepName := "prepare"
			if err := manifestManager.StartStep(stepName); err != nil {
				return err
			}
			stepSuccess := false
			defer func() {
				if !stepSuccess {
					_ = manifestManager.FailStep(stepName, "prepare failed")
					_ = manifestManager.SetStatus("failed")
				}
			}()

			inputPath, err := prompt.AskString("Input log path", filepath.Join(outputPath, "stage_01_ingest_profile", "normalised_log.csv"), true)
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
			resourceCol, err := prompt.AskString("Resource column (optional)", "", false)
			if err != nil {
				return err
			}

			confirm, err := prompt.AskBool("Run data quality checks now?", true)
			if err != nil {
				return err
			}
			if !confirm {
				_ = manifestManager.CompleteStep(stepName)
				_ = manifestManager.SetStatus("completed")
				fmt.Println("[INFO] Data preparation canceled by user.")
				return nil
			}

			if err := manifestManager.AddInputs([]string{inputPath}); err != nil {
				return err
			}

			venvRunner := &runner.Runner{ProjectPath: projectPath}
			skillsRoot, err := paths.SkillsRoot(projectPath)
			if err != nil {
				return err
			}
			reqPath := paths.SkillPath(skillsRoot, "pm-99-utils-and-standards", "requirements.txt")
			if err := venvRunner.EnsureVenv(reqPath); err != nil {
				return err
			}

			qualityScript := paths.SkillPath(skillsRoot, "pm-03-data-quality", "scripts", "02_data_quality.py")
			qualityArgs := []string{
				"--file", inputPath,
				"--case", caseCol,
				"--activity", activityCol,
				"--timestamp", timestampCol,
				"--output", outputPath,
			}
			if resourceCol != "" {
				qualityArgs = append(qualityArgs, "--resource", resourceCol)
			}

			fmt.Println("[INFO] Running data quality checks...")
			logging.Info("running data quality checks", map[string]any{"script": qualityScript})
			if err := venvRunner.RunScript(qualityScript, qualityArgs, nil); err != nil {
				return err
			}

			nbPath := filepath.Join(outputPath, "analysis_notebook.ipynb")
			qualityCode := fmt.Sprintf("!python %s --file %s --case %s --activity %s --timestamp %s --output %s", qualityScript, inputPath, caseCol, activityCol, timestampCol, outputPath)
			if resourceCol != "" {
				qualityCode += fmt.Sprintf(" --resource %s", resourceCol)
			}
			qualityMarkdown := "## Data Quality\nWe profiled data quality and generated recommendations."
			if err := notebook.AppendStep(nbPath, "Data Quality", qualityMarkdown, qualityCode); err != nil {
				return err
			}

			filterChoice, err := prompt.AskChoice("Rare activity filtering", []string{"none", "min-frequency", "top-variants"}, "none", true)
			if err != nil {
				return err
			}
			autoFilter := false
			minFreq := "0.01"
			topVariants := ""
			if filterChoice == "min-frequency" {
				minFreq, err = prompt.AskString("Minimum activity frequency", "0.01", true)
				if err != nil {
					return err
				}
				autoFilter = true
			} else if filterChoice == "top-variants" {
				topVariants, err = prompt.AskString("Top variants to keep", "10", true)
				if err != nil {
					return err
				}
			}
			startActs, err := prompt.AskString("Start activities (comma-separated, optional)", "", false)
			if err != nil {
				return err
			}
			endActs, err := prompt.AskString("End activities (comma-separated, optional)", "", false)
			if err != nil {
				return err
			}

			cleanScript := paths.SkillPath(skillsRoot, "pm-04-clean-filter", "scripts", "02_clean_filter.py")
			cleanInput := filepath.Join(outputPath, "stage_02_data_quality", "cleaned_log.csv")
			cleanArgs := []string{
				"--file", cleanInput,
				"--format", "csv",
				"--case", caseCol,
				"--activity", activityCol,
				"--timestamp", timestampCol,
				"--output", outputPath,
			}
			if resourceCol != "" {
				cleanArgs = append(cleanArgs, "--resource", resourceCol)
			}
			if autoFilter {
				cleanArgs = append(cleanArgs, "--auto-filter-rare-activities", "--min-activity-frequency", minFreq)
			}
			if topVariants != "" {
				cleanArgs = append(cleanArgs, "--top-variants", topVariants)
			}
			if startActs != "" {
				cleanArgs = append(cleanArgs, "--start-activities", startActs)
			}
			if endActs != "" {
				cleanArgs = append(cleanArgs, "--end-activities", endActs)
			}

			fmt.Println("[INFO] Running clean and filter...")
			logging.Info("running clean and filter", map[string]any{"script": cleanScript})
			if err := venvRunner.RunScript(cleanScript, cleanArgs, nil); err != nil {
				return err
			}

			cleanCode := fmt.Sprintf("!python %s --file %s --format csv --case %s --activity %s --timestamp %s --output %s", cleanScript, cleanInput, caseCol, activityCol, timestampCol, outputPath)
			if resourceCol != "" {
				cleanCode += fmt.Sprintf(" --resource %s", resourceCol)
			}
			if autoFilter {
				cleanCode += fmt.Sprintf(" --auto-filter-rare-activities --min-activity-frequency %s", minFreq)
			}
			if topVariants != "" {
				cleanCode += fmt.Sprintf(" --top-variants %s", topVariants)
			}
			if startActs != "" {
				cleanCode += fmt.Sprintf(" --start-activities %s", startActs)
			}
			if endActs != "" {
				cleanCode += fmt.Sprintf(" --end-activities %s", endActs)
			}

			cleanMarkdown := "## Clean and Filter\nWe applied cleaning and filtering rules to produce a filtered log."
			if err := notebook.AppendStep(nbPath, "Clean and Filter", cleanMarkdown, cleanCode); err != nil {
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

			fmt.Println("[SUCCESS] Data preparation completed.")
			return nil
		},
		Example: "  pm-assist prepare",
	}
	return cmd
}
