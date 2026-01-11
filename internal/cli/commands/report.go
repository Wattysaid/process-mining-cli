package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pm-assist/pm-assist/internal/app"
	"github.com/pm-assist/pm-assist/internal/cli/prompt"
	"github.com/pm-assist/pm-assist/internal/notebook"
	"github.com/pm-assist/pm-assist/internal/runner"
	"github.com/spf13/cobra"
)

// NewReportCmd returns the report command.
func NewReportCmd(global *app.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate notebooks and reports",
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

			reportName, err := prompt.AskString("Report filename", "process_mining_report.md", true)
			if err != nil {
				return err
			}
			confirm, err := prompt.AskBool("Generate report now?", true)
			if err != nil {
				return err
			}
			if !confirm {
				fmt.Println("[INFO] Report generation canceled by user.")
				return nil
			}

			venvRunner := &runner.Runner{ProjectPath: projectPath}
			reqPath := filepath.Join(projectPath, ".codex", "skills", "cli-tool-skills", "pm-99-utils-and-standards", "requirements.txt")
			if err := venvRunner.EnsureVenv(reqPath); err != nil {
				return err
			}

			reportScript := filepath.Join(projectPath, ".codex", "skills", "cli-tool-skills", "pm-10-reporting", "scripts", "08_report.py")
			argsList := []string{"--output", outputPath, "--report", reportName}
			fmt.Println("[INFO] Generating report...")
			if err := venvRunner.RunScript(reportScript, argsList, nil); err != nil {
				return err
			}

			nbPath := filepath.Join(outputPath, "analysis_notebook.ipynb")
			code := fmt.Sprintf("!python %s --output %s --report %s", reportScript, outputPath, reportName)
			if err := notebook.AppendStep(nbPath, "Report", "## Report\nWe generated the report artefact.", code); err != nil {
				return err
			}

			fmt.Println("[SUCCESS] Report generated.")
			return nil
		},
		Example: "  pm-assist report",
	}
	return cmd
}
