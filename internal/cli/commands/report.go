package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pm-assist/pm-assist/internal/app"
	"github.com/pm-assist/pm-assist/internal/config"
	"github.com/pm-assist/pm-assist/internal/logging"
	"github.com/pm-assist/pm-assist/internal/notebook"
	"github.com/pm-assist/pm-assist/internal/paths"
	"github.com/pm-assist/pm-assist/internal/policy"
	"github.com/pm-assist/pm-assist/internal/reporting"
	"github.com/pm-assist/pm-assist/internal/runner"
	"github.com/pm-assist/pm-assist/internal/ui"
	"github.com/spf13/cobra"
)

// NewReportCmd returns the report command.
func NewReportCmd(global *app.GlobalFlags) *cobra.Command {
	var (
		flagReport  string
		flagConfirm string
		flagHTML    string
		flagPDF     string
	)
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate notebooks and reports",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.PrintCommandStart(ui.CommandFrame{
				Title:     "pm-assist report",
				Purpose:   "Generate report artifacts and bundle",
				StepIndex: 6,
				StepTotal: 7,
				Writes:    []string{"outputs/<run-id>/stage_09_report", "outputs/<run-id>/bundle"},
				Asks:      []string{"report options"},
				Next:      "pm-assist review",
			})
			success := false
			defer func() {
				ui.PrintCommandEnd(ui.CommandFrame{Title: "pm-assist report", Next: "pm-assist review"}, success)
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
			policies := policy.FromConfig(cfg)
			if policies.OfflineOnly {
				fmt.Println("[WARN] Offline-only policy is enabled; ensure local data sources are used.")
			}

			manifestManager, err := initRunManifest(runID, outputPath, cfg)
			if err != nil {
				return err
			}
			defer logging.CloseRunLog()
			stepName := "report"
			if err := manifestManager.StartStep(stepName); err != nil {
				return err
			}
			stepSuccess := false
			defer func() {
				if !stepSuccess {
					_ = manifestManager.FailStep(stepName, "report failed")
					_ = manifestManager.SetStatus("failed")
				}
			}()

			reportName, err := resolveString(flagReport, "Report filename", "process_mining_report.md", true)
			if err != nil {
				return err
			}
			confirm, err := resolveBool(flagConfirm, "Generate report now?", true)
			if err != nil {
				return err
			}
			if !confirm {
				_ = manifestManager.CompleteStep(stepName)
				_ = manifestManager.SetStatus("completed")
				fmt.Println("[INFO] Report generation canceled by user.")
				return nil
			}
			exportHTML, err := resolveBool(flagHTML, "Export HTML report?", false)
			if err != nil {
				return err
			}
			exportPDF, err := resolveBool(flagPDF, "Export PDF report?", false)
			if err != nil {
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
			printDependencyNotice(options)
			if err := ensureVenvWithSpinner(venvRunner, reqPath, options); err != nil {
				return err
			}

			printStepProgress(1, 2, "Generating report content")
			reportScript := paths.SkillPath(skillsRoot, "pm-10-reporting", "scripts", "08_report.py")
			argsList := []string{"--output", outputPath, "--report", reportName}
			fmt.Println("[INFO] Generating report...")
			logging.Info("generating report", map[string]any{"script": reportScript, "report": reportName})
			if err := venvRunner.RunScript(reportScript, argsList, nil); err != nil {
				return err
			}

			nbPath := filepath.Join(outputPath, "analysis_notebook.ipynb")
			code := fmt.Sprintf("!python %s --output %s --report %s", reportScript, outputPath, reportName)
			if err := notebook.AppendStep(nbPath, "Report", "## Report\nWe generated the report artefact.", code); err != nil {
				return err
			}

			reportPath := filepath.Join(outputPath, "stage_09_report", reportName)
			if exportHTML {
				htmlPath := strings.TrimSuffix(reportPath, filepath.Ext(reportPath)) + ".html"
				if err := reporting.MarkdownToHTML(reportPath, htmlPath, "PM Assist Report"); err != nil {
					fmt.Printf("[WARN] HTML export failed: %v\n", err)
				} else {
					fmt.Printf("[SUCCESS] HTML report generated: %s\n", htmlPath)
				}
			}
			if exportPDF {
				pdfPath := strings.TrimSuffix(reportPath, filepath.Ext(reportPath)) + ".pdf"
				if err := reporting.MarkdownToPDF(reportPath, pdfPath); err != nil {
					fmt.Printf("[WARN] PDF export failed: %v\n", err)
				} else {
					fmt.Printf("[SUCCESS] PDF report generated: %s\n", pdfPath)
				}
			}

			printStepProgress(2, 2, "Bundling report outputs")
			bundlePath := filepath.Join(outputPath, "bundle", fmt.Sprintf("report_bundle_%s.zip", runID))
			entries := map[string]string{
				"report/report.md":                 reportPath,
				"notebook/analysis_notebook.ipynb": nbPath,
				"manifest/run_manifest.json":       filepath.Join(outputPath, "run_manifest.json"),
				"manifest/config_snapshot.yaml":    filepath.Join(outputPath, "config_snapshot.yaml"),
				"quality/qa_summary.md":            filepath.Join(outputPath, "quality", "qa_summary.md"),
				"quality/qa_results.json":          filepath.Join(outputPath, "quality", "qa_results.json"),
			}
			htmlCandidate := strings.TrimSuffix(reportPath, filepath.Ext(reportPath)) + ".html"
			pdfCandidate := strings.TrimSuffix(reportPath, filepath.Ext(reportPath)) + ".pdf"
			if _, err := os.Stat(htmlCandidate); err == nil {
				entries["report/report.html"] = htmlCandidate
			}
			if _, err := os.Stat(pdfCandidate); err == nil {
				entries["report/report.pdf"] = pdfCandidate
			}
			if err := reporting.BuildReportBundle(bundlePath, entries); err != nil {
				fmt.Printf("[WARN] Report bundle creation failed: %v\n", err)
			} else {
				fmt.Printf("[SUCCESS] Report bundle created: %s\n", bundlePath)
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
			success = true

			fmt.Println("[SUCCESS] Report generated.")
			updated, _ := config.Load(global.ConfigPath)
			ui.PrintSplash(updated, ui.SplashOptions{CompletedCommand: "report", WorkingDir: projectPath})
			return nil
		},
		Example: "  pm-assist report",
	}
	cmd.Flags().StringVar(&flagReport, "report", "", "Report filename")
	cmd.Flags().StringVar(&flagConfirm, "confirm", "", "Generate report now (true|false)")
	cmd.Flags().StringVar(&flagHTML, "html", "", "Export HTML report (true|false)")
	cmd.Flags().StringVar(&flagPDF, "pdf", "", "Export PDF report (true|false)")
	return cmd
}
