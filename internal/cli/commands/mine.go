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

// NewMineCmd returns the mine command.
func NewMineCmd(global *app.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mine",
		Short: "Run process mining analysis",
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

			runEDA, err := prompt.AskBool("Run EDA diagnostics?", true)
			if err != nil {
				return err
			}
			runDiscovery, err := prompt.AskBool("Run discovery models?", true)
			if err != nil {
				return err
			}
			runConformance, err := prompt.AskBool("Run conformance checks?", false)
			if err != nil {
				return err
			}
			runPerformance, err := prompt.AskBool("Run performance analysis?", true)
			if err != nil {
				return err
			}

			venvRunner := &runner.Runner{ProjectPath: projectPath}
			reqPath := filepath.Join(projectPath, ".codex", "skills", "cli-tool-skills", "pm-99-utils-and-standards", "requirements.txt")
			if err := venvRunner.EnsureVenv(reqPath); err != nil {
				return err
			}

			nbPath := filepath.Join(outputPath, "analysis_notebook.ipynb")

			if runEDA {
				edaScript := filepath.Join(projectPath, ".codex", "skills", "cli-tool-skills", "pm-05-eda", "scripts", "03_eda.py")
				edaArgs := []string{"--use-filtered", "--output", outputPath, "--case", caseCol, "--activity", activityCol, "--timestamp", timestampCol}
				if resourceCol != "" {
					edaArgs = append(edaArgs, "--resource", resourceCol)
				}
				fmt.Println("[INFO] Running EDA diagnostics...")
				if err := venvRunner.RunScript(edaScript, edaArgs, nil); err != nil {
					return err
				}
				code := fmt.Sprintf("!python %s --use-filtered --output %s --case %s --activity %s --timestamp %s", edaScript, outputPath, caseCol, activityCol, timestampCol)
				if resourceCol != "" {
					code += fmt.Sprintf(" --resource %s", resourceCol)
				}
				if err := notebook.AppendStep(nbPath, "EDA", "## EDA\nWe generated exploratory diagnostics.", code); err != nil {
					return err
				}
			}

			if runDiscovery {
				miner, err := prompt.AskChoice("Discovery miner selection", []string{"auto", "inductive", "heuristic", "both"}, "auto", true)
				if err != nil {
					return err
				}
				discoverScript := filepath.Join(projectPath, ".codex", "skills", "cli-tool-skills", "pm-06-discovery", "scripts", "04_discover.py")
				argsList := []string{"--use-filtered", "--output", outputPath, "--case", caseCol, "--activity", activityCol, "--timestamp", timestampCol, "--miner-selection", miner}
				if resourceCol != "" {
					argsList = append(argsList, "--resource", resourceCol)
				}
				fmt.Println("[INFO] Running discovery...")
				if err := venvRunner.RunScript(discoverScript, argsList, nil); err != nil {
					return err
				}
				code := fmt.Sprintf("!python %s --use-filtered --output %s --case %s --activity %s --timestamp %s --miner-selection %s", discoverScript, outputPath, caseCol, activityCol, timestampCol, miner)
				if resourceCol != "" {
					code += fmt.Sprintf(" --resource %s", resourceCol)
				}
				if err := notebook.AppendStep(nbPath, "Discovery", "## Discovery\nWe discovered process models.", code); err != nil {
					return err
				}
			}

			if runConformance {
				method, err := prompt.AskChoice("Conformance method", []string{"alignments", "token"}, "alignments", true)
				if err != nil {
					return err
				}
				confScript := filepath.Join(projectPath, ".codex", "skills", "cli-tool-skills", "pm-07-conformance", "scripts", "05_conformance.py")
				argsList := []string{"--use-filtered", "--output", outputPath, "--case", caseCol, "--activity", activityCol, "--timestamp", timestampCol, "--conformance-method", method}
				if resourceCol != "" {
					argsList = append(argsList, "--resource", resourceCol)
				}
				fmt.Println("[INFO] Running conformance...")
				if err := venvRunner.RunScript(confScript, argsList, nil); err != nil {
					return err
				}
				code := fmt.Sprintf("!python %s --use-filtered --output %s --case %s --activity %s --timestamp %s --conformance-method %s", confScript, outputPath, caseCol, activityCol, timestampCol, method)
				if resourceCol != "" {
					code += fmt.Sprintf(" --resource %s", resourceCol)
				}
				if err := notebook.AppendStep(nbPath, "Conformance", "## Conformance\nWe evaluated model fitness and deviations.", code); err != nil {
					return err
				}
			}

			if runPerformance {
				advanced, err := prompt.AskBool("Run advanced performance diagnostics?", false)
				if err != nil {
					return err
				}
				sla, err := prompt.AskString("SLA threshold (hours)", "72", true)
				if err != nil {
					return err
				}
				perfScript := filepath.Join(projectPath, ".codex", "skills", "cli-tool-skills", "pm-08-performance", "scripts", "06_performance.py")
				argsList := []string{"--use-filtered", "--output", outputPath, "--case", caseCol, "--activity", activityCol, "--timestamp", timestampCol, "--sla-hours", sla}
				if advanced {
					argsList = append(argsList, "--advanced")
				}
				if resourceCol != "" {
					argsList = append(argsList, "--resource", resourceCol)
				}
				fmt.Println("[INFO] Running performance analysis...")
				if err := venvRunner.RunScript(perfScript, argsList, nil); err != nil {
					return err
				}
				code := fmt.Sprintf("!python %s --use-filtered --output %s --case %s --activity %s --timestamp %s --sla-hours %s", perfScript, outputPath, caseCol, activityCol, timestampCol, sla)
				if advanced {
					code += " --advanced"
				}
				if resourceCol != "" {
					code += fmt.Sprintf(" --resource %s", resourceCol)
				}
				if err := notebook.AppendStep(nbPath, "Performance", "## Performance\nWe analyzed throughput and bottlenecks.", code); err != nil {
					return err
				}
			}

			fmt.Println("[SUCCESS] Mining steps completed.")
			return nil
		},
		Example: "  pm-assist mine",
	}
	return cmd
}
