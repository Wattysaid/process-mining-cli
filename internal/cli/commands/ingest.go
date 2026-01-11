package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pm-assist/pm-assist/internal/app"
	"github.com/pm-assist/pm-assist/internal/config"
	"github.com/pm-assist/pm-assist/internal/logging"
	"github.com/pm-assist/pm-assist/internal/notebook"
	"github.com/pm-assist/pm-assist/internal/paths"
	"github.com/pm-assist/pm-assist/internal/policy"
	"github.com/pm-assist/pm-assist/internal/runner"
	"github.com/spf13/cobra"
)

// NewIngestCmd returns the ingest command.
func NewIngestCmd(global *app.GlobalFlags) *cobra.Command {
	var (
		flagConnector string
		flagFile      string
		flagCase      string
		flagActivity  string
		flagTimestamp string
		flagResource  string
		flagConfirm   string
	)
	cmd := &cobra.Command{
		Use:   "ingest",
		Short: "Ingest data into a staging dataset",
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

			if len(cfg.Connectors) == 0 {
				fmt.Println("[WARN] No connectors configured. Run `pm-assist connect` first.")
				return nil
			}

			connectorName, err := resolveString(flagConnector, "Connector name", cfg.Connectors[0].Name, true)
			if err != nil {
				return err
			}
			var selected *config.ConnectorSpec
			for i := range cfg.Connectors {
				if cfg.Connectors[i].Name == connectorName {
					selected = &cfg.Connectors[i]
					break
				}
			}
			if selected == nil {
				return fmt.Errorf("connector not found: %s", connectorName)
			}
			if !policies.AllowsConnector(selected.Type) {
				return fmt.Errorf("connector type blocked by policy: %s", selected.Type)
			}
			if selected.Type != "file" || selected.File == nil || len(selected.File.Paths) == 0 {
				return fmt.Errorf("only file connectors are supported in ingest for now")
			}

			filePath := selected.File.Paths[0]
			if flagFile != "" {
				filePath = flagFile
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
			stepName := "ingest"
			if err := manifestManager.StartStep(stepName); err != nil {
				return err
			}
			stepSuccess := false
			defer func() {
				if !stepSuccess {
					_ = manifestManager.FailStep(stepName, "ingest failed")
					_ = manifestManager.SetStatus("failed")
				}
			}()

			confirm, err := resolveBool(flagConfirm, "Run ingest now?", true)
			if err != nil {
				return err
			}
			if !confirm {
				_ = manifestManager.CompleteStep(stepName)
				_ = manifestManager.SetStatus("completed")
				fmt.Println("[INFO] Ingest canceled by user.")
				return nil
			}

			if err := manifestManager.AddInputs([]string{filePath}); err != nil {
				return err
			}

			venvRunner := &runner.Runner{ProjectPath: projectPath}
			skillsRoot, err := paths.SkillsRoot(projectPath)
			if err != nil {
				return err
			}
			reqPath := paths.SkillPath(skillsRoot, "pm-99-utils-and-standards", "requirements.txt")
			options, err := resolveVenvOptions(projectPath, policies)
			if err != nil {
				return err
			}
			if err := venvRunner.EnsureVenv(reqPath, options); err != nil {
				return err
			}

			scriptPath := paths.SkillPath(skillsRoot, "pm-02-ingest-profile", "scripts", "01_ingest.py")
			argsList := []string{
				"--file", filePath,
				"--format", selected.File.Format,
				"--case", caseCol,
				"--activity", activityCol,
				"--timestamp", timestampCol,
				"--output", outputPath,
			}
			if resourceCol != "" {
				argsList = append(argsList, "--resource", resourceCol)
			}

			fmt.Println("[INFO] Running ingest script...")
			logging.Info("running ingest script", map[string]any{"script": scriptPath})
			if err := venvRunner.RunScript(scriptPath, argsList, nil); err != nil {
				return err
			}

			nbPath := filepath.Join(outputPath, "analysis_notebook.ipynb")
			markdown := "## Ingest\nWe ingested the source file and normalized the log."
			code := fmt.Sprintf("!python %s --file %s --format %s --case %s --activity %s --timestamp %s --output %s", scriptPath, filePath, selected.File.Format, caseCol, activityCol, timestampCol, outputPath)
			if resourceCol != "" {
				code += fmt.Sprintf(" --resource %s", resourceCol)
			}
			if err := notebook.AppendStep(nbPath, "Ingest", markdown, code); err != nil {
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

			fmt.Println("[SUCCESS] Ingest completed.")
			return nil
		},
		Example: "  pm-assist ingest",
	}
	cmd.Flags().StringVar(&flagConnector, "connector", "", "Connector name")
	cmd.Flags().StringVar(&flagFile, "file", "", "Input file path override")
	cmd.Flags().StringVar(&flagCase, "case", "", "Case ID column")
	cmd.Flags().StringVar(&flagActivity, "activity", "", "Activity column")
	cmd.Flags().StringVar(&flagTimestamp, "timestamp", "", "Timestamp column")
	cmd.Flags().StringVar(&flagResource, "resource", "", "Resource column")
	cmd.Flags().StringVar(&flagConfirm, "confirm", "", "Run ingest now (true|false)")
	return cmd
}
