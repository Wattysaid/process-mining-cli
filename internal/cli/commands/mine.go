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
	"github.com/pm-assist/pm-assist/internal/ui"
	"github.com/spf13/cobra"
)

// NewMineCmd returns the mine command.
func NewMineCmd(global *app.GlobalFlags) *cobra.Command {
	var (
		flagCase           string
		flagActivity       string
		flagTimestamp      string
		flagResource       string
		flagRunEDA         string
		flagRunDiscovery   string
		flagRunConformance string
		flagRunPerformance string
		flagMiner          string
		flagConformance    string
		flagAdvanced       string
		flagSLA            string
	)
	cmd := &cobra.Command{
		Use:   "mine",
		Short: "Run process mining analysis",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.PrintCommandStart(ui.CommandFrame{
				Title:     "pm-assist mine",
				Purpose:   "Run discovery, conformance, and performance analyses",
				StepIndex: 5,
				StepTotal: 7,
				Writes:    []string{"outputs/<run-id>/stage_04_discovery", "outputs/<run-id>/stage_05_conformance", "outputs/<run-id>/stage_06_performance"},
				Asks:      []string{"analysis options"},
				Next:      "pm-assist report",
			})
			success := false
			defer func() {
				ui.PrintCommandEnd(ui.CommandFrame{Title: "pm-assist mine", Next: "pm-assist report"}, success)
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
			stepName := "mine"
			if err := manifestManager.StartStep(stepName); err != nil {
				return err
			}
			stepSuccess := false
			defer func() {
				if !stepSuccess {
					_ = manifestManager.FailStep(stepName, "mine failed")
					_ = manifestManager.SetStatus("failed")
				}
			}()

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

			runEDA, err := resolveBool(flagRunEDA, "Run EDA diagnostics?", true)
			if err != nil {
				return err
			}
			runDiscovery, err := resolveBool(flagRunDiscovery, "Run discovery models?", true)
			if err != nil {
				return err
			}
			runConformance, err := resolveBool(flagRunConformance, "Run conformance checks?", false)
			if err != nil {
				return err
			}
			runPerformance, err := resolveBool(flagRunPerformance, "Run performance analysis?", true)
			if err != nil {
				return err
			}

			totalSteps := 1
			if runEDA {
				totalSteps++
			}
			if runDiscovery {
				totalSteps++
			}
			if runConformance {
				totalSteps++
			}
			if runPerformance {
				totalSteps++
			}
			stepIndex := 1

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

			nbPath := filepath.Join(outputPath, "analysis_notebook.ipynb")

			if runEDA {
				printStepProgress(stepIndex, totalSteps, "Running EDA diagnostics")
				stepIndex++
				edaScript := paths.SkillPath(skillsRoot, "pm-05-eda", "scripts", "03_eda.py")
				edaArgs := []string{"--use-filtered", "--output", outputPath, "--case", caseCol, "--activity", activityCol, "--timestamp", timestampCol}
				if resourceCol != "" {
					edaArgs = append(edaArgs, "--resource", resourceCol)
				}
				fmt.Println("[INFO] Running EDA diagnostics...")
				logging.Info("running EDA diagnostics", map[string]any{"script": edaScript})
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
				printStepProgress(stepIndex, totalSteps, "Running discovery models")
				stepIndex++
				miner, err := resolveChoice(flagMiner, "Discovery miner selection", []string{"auto", "inductive", "heuristic", "both"}, "auto", true)
				if err != nil {
					return err
				}
				discoverScript := paths.SkillPath(skillsRoot, "pm-06-discovery", "scripts", "04_discover.py")
				argsList := []string{"--use-filtered", "--output", outputPath, "--case", caseCol, "--activity", activityCol, "--timestamp", timestampCol, "--miner-selection", miner}
				if resourceCol != "" {
					argsList = append(argsList, "--resource", resourceCol)
				}
				fmt.Println("[INFO] Running discovery...")
				logging.Info("running discovery", map[string]any{"script": discoverScript, "miner": miner})
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
				printStepProgress(stepIndex, totalSteps, "Running conformance checks")
				stepIndex++
				method, err := resolveChoice(flagConformance, "Conformance method", []string{"alignments", "token"}, "alignments", true)
				if err != nil {
					return err
				}
				confScript := paths.SkillPath(skillsRoot, "pm-07-conformance", "scripts", "05_conformance.py")
				argsList := []string{"--use-filtered", "--output", outputPath, "--case", caseCol, "--activity", activityCol, "--timestamp", timestampCol, "--conformance-method", method}
				if resourceCol != "" {
					argsList = append(argsList, "--resource", resourceCol)
				}
				fmt.Println("[INFO] Running conformance...")
				logging.Info("running conformance", map[string]any{"script": confScript, "method": method})
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
				printStepProgress(stepIndex, totalSteps, "Running performance analysis")
				stepIndex++
				advanced, err := resolveBool(flagAdvanced, "Run advanced performance diagnostics?", false)
				if err != nil {
					return err
				}
				sla, err := resolveString(flagSLA, "SLA threshold (hours)", "72", true)
				if err != nil {
					return err
				}
				perfScript := paths.SkillPath(skillsRoot, "pm-08-performance", "scripts", "06_performance.py")
				argsList := []string{"--use-filtered", "--output", outputPath, "--case", caseCol, "--activity", activityCol, "--timestamp", timestampCol, "--sla-hours", sla}
				if advanced {
					argsList = append(argsList, "--advanced")
				}
				if resourceCol != "" {
					argsList = append(argsList, "--resource", resourceCol)
				}
				fmt.Println("[INFO] Running performance analysis...")
				logging.Info("running performance analysis", map[string]any{"script": perfScript})
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

			printStepProgress(stepIndex, totalSteps, "Finalizing mining outputs")
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

			fmt.Println("[SUCCESS] Mining steps completed.")
			updated, _ := config.Load(global.ConfigPath)
			ui.PrintSplash(updated, ui.SplashOptions{CompletedCommand: "mine", WorkingDir: projectPath})
			return nil
		},
		Example: "  pm-assist mine",
	}
	cmd.Flags().StringVar(&flagCase, "case", "", "Case ID column")
	cmd.Flags().StringVar(&flagActivity, "activity", "", "Activity column")
	cmd.Flags().StringVar(&flagTimestamp, "timestamp", "", "Timestamp column")
	cmd.Flags().StringVar(&flagResource, "resource", "", "Resource column")
	cmd.Flags().StringVar(&flagRunEDA, "run-eda", "", "Run EDA diagnostics (true|false)")
	cmd.Flags().StringVar(&flagRunDiscovery, "run-discovery", "", "Run discovery models (true|false)")
	cmd.Flags().StringVar(&flagRunConformance, "run-conformance", "", "Run conformance checks (true|false)")
	cmd.Flags().StringVar(&flagRunPerformance, "run-performance", "", "Run performance analysis (true|false)")
	cmd.Flags().StringVar(&flagMiner, "miner", "", "Discovery miner selection (auto|inductive|heuristic|both)")
	cmd.Flags().StringVar(&flagConformance, "conformance-method", "", "Conformance method (alignments|token)")
	cmd.Flags().StringVar(&flagAdvanced, "advanced-performance", "", "Run advanced performance diagnostics (true|false)")
	cmd.Flags().StringVar(&flagSLA, "sla-hours", "", "SLA threshold (hours)")
	return cmd
}
