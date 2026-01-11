package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pm-assist/pm-assist/internal/buildinfo"
	"github.com/pm-assist/pm-assist/internal/cli/prompt"
	"github.com/pm-assist/pm-assist/internal/config"
	"github.com/pm-assist/pm-assist/internal/policy"
	"github.com/spf13/cobra"
)

func runStartup(cmdRoot *cobra.Command) error {
	printBanner()
	printStatus()
	if !hasProjectConfig() {
		confirm, err := prompt.AskBool("It looks like this is your first time using PM Assist. Continue setup?", true)
		if err != nil {
			return err
		}
		if confirm {
			return dispatchCommand(cmdRoot, "init")
		}
	}
	fmt.Println()
	fmt.Println("What would you like to do?")
	fmt.Println()
	fmt.Println("  1) Start a new process mining project")
	fmt.Println("  2) Continue an existing project")
	fmt.Println("  3) Run environment diagnostics (doctor)")
	fmt.Println("  4) Configure LLM integration")
	fmt.Println("  5) Manage user or business profiles")
	fmt.Println("  6) Exit")
	fmt.Println()
	choice, err := prompt.AskChoice("Select an option", []string{"1", "2", "3", "4", "5", "6"}, "1", true)
	if err != nil {
		return err
	}

	switch choice {
	case "1":
		return dispatchCommand(cmdRoot, "init")
	case "2":
		return continueProjectFlow(cmdRoot)
	case "3":
		return dispatchCommand(cmdRoot, "doctor")
	case "4":
		return dispatchCommand(cmdRoot, "agent", "setup")
	case "5":
		manage, err := prompt.AskChoice("Manage which profile", []string{"user", "business"}, "user", true)
		if err != nil {
			return err
		}
		if manage == "business" {
			return dispatchCommand(cmdRoot, "business", "init")
		}
		return dispatchCommand(cmdRoot, "profile", "init")
	case "6":
		fmt.Println("Goodbye.")
		return nil
	default:
		return fmt.Errorf("invalid selection")
	}
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

func continueProjectFlow(cmdRoot *cobra.Command) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	manifestPath, manifest, err := findLatestManifest(filepath.Join(cwd, "outputs"))
	if err != nil {
		return err
	}
	if manifest == nil {
		confirm, err := prompt.AskBool("No previous runs found. Start a new run now?", true)
		if err != nil {
			return err
		}
		if confirm {
			return dispatchCommand(cmdRoot, "connect")
		}
		return nil
	}

	fmt.Printf("[INFO] Latest run: %s (%s)\n", manifest.RunID, manifest.Status)
	if manifest.CompletedAt != "" {
		fmt.Printf("[INFO] Completed at: %s\n", manifest.CompletedAt)
	} else if manifest.StartedAt != "" {
		fmt.Printf("[INFO] Started at: %s\n", manifest.StartedAt)
	}
	fmt.Printf("[INFO] Manifest: %s\n", manifestPath)

	nextStep := nextRecommendedStep(manifest)
	if nextStep == "" {
		fmt.Println("[INFO] All pipeline steps appear complete for the latest run.")
		return nil
	}
	confirm, err := prompt.AskBool(fmt.Sprintf("Continue with next step (%s)?", nextStep), true)
	if err != nil {
		return err
	}
	if confirm {
		return dispatchCommand(cmdRoot, nextStep)
	}
	return nil
}

func findLatestManifest(outputsPath string) (string, *runManifest, error) {
	pattern := filepath.Join(outputsPath, "*", "run_manifest.json")
	candidates, err := filepath.Glob(pattern)
	if err != nil {
		return "", nil, err
	}
	if len(candidates) == 0 {
		return "", nil, nil
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
		return "", nil, err
	}
	var manifest runManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return "", nil, err
	}
	return latest, &manifest, nil
}

func nextRecommendedStep(manifest *runManifest) string {
	if manifest == nil {
		return ""
	}
	pipeline := []string{"ingest", "map", "prepare", "mine", "report", "review"}
	statusByStep := make(map[string]string)
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

func dispatchCommand(cmdRoot *cobra.Command, name string, args ...string) error {
	for _, cmd := range cmdRoot.Commands() {
		if cmd.Name() != name {
			continue
		}
		if len(args) > 0 {
			return dispatchSubcommand(cmd, args[0], args[1:]...)
		}
		if run := cmd.RunE; run != nil {
			return run(cmd, args)
		}
		return nil
	}
	return fmt.Errorf("command not found: %s", name)
}

func dispatchSubcommand(parent *cobra.Command, name string, args ...string) error {
	for _, cmd := range parent.Commands() {
		if cmd.Name() != name {
			continue
		}
		if run := cmd.RunE; run != nil {
			return run(cmd, args)
		}
		return nil
	}
	return fmt.Errorf("subcommand not found: %s", name)
}

func printBanner() {
	banner := `            ____  __  __        ___              __
  (\_/)    |  _ \|  \/  |      / _ \ ___ ___ ___/ _\___
  ( •_•)   | |_) | |\/| |_____/ /_)/ __/ __/ _ \ \ / __|
   />[_]   |  __/| |  | |_____/ ___/ (_| (_|  __/\ \__ \
           |_|   |_|  |_|      \/    \___\___\___\__/___/

            PM Assist · Enterprise Process Mining CLI
            -----------------------------------------`
	fmt.Println(banner)
}

func printStatus() {
	pythonStatus := resolvePythonStatus()
	llmStatus := resolveLLMStatus()
	graphvizStatus := "missing"
	if _, err := lookupPath("dot"); err == nil {
		graphvizStatus = "ready"
	}
	fmt.Printf("\nVersion: %s | Python: %s | LLM: %s | Graphviz: %s\n", buildinfo.Version, pythonStatus, llmStatus, graphvizStatus)

	if hasProjectConfig() {
		fmt.Println("[INFO] Project detected in current directory.")
	}
}

func hasProjectConfig() bool {
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(cwd, "pm-assist.yaml"))
	return err == nil
}

func hasEnv(key string) bool {
	return os.Getenv(key) != ""
}

func lookupPath(binary string) (string, error) {
	return execLookPath(binary)
}

var execLookPath = func(binary string) (string, error) {
	return exec.LookPath(binary)
}

func resolvePythonStatus() string {
	pythonPath := ""
	if hasProjectConfig() {
		cwd, err := os.Getwd()
		if err == nil {
			candidate := filepath.Join(cwd, ".venv", "bin", "python")
			if _, err := os.Stat(candidate); err == nil {
				pythonPath = candidate
			}
		}
	}
	if pythonPath == "" {
		if path, err := lookupPath("python3"); err == nil {
			pythonPath = path
		} else if path, err := lookupPath("python"); err == nil {
			pythonPath = path
		}
	}
	if pythonPath == "" {
		return "missing"
	}
	if err := checkPythonImports(pythonPath); err != nil {
		return "deps missing"
	}
	return "ready"
}

func checkPythonImports(pythonPath string) error {
	cmd := exec.Command(pythonPath, "-c", "import pm4py, pandas, numpy, matplotlib, yaml, openpyxl, pyarrow")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func resolveLLMStatus() string {
	cfg, err := config.Load("")
	if err != nil {
		return "not configured"
	}
	policies := policy.FromConfig(cfg)
	if policies.LLMEnabled != nil && !*policies.LLMEnabled {
		return "disabled by policy"
	}
	if policies.OfflineOnly && cfg.LLM.Provider != "" && strings.ToLower(cfg.LLM.Provider) != "ollama" && strings.ToLower(cfg.LLM.Provider) != "none" {
		return "disabled by policy"
	}
	if strings.ToLower(cfg.LLM.Provider) == "ollama" {
		if hasEnv("OLLAMA_HOST") {
			return "configured"
		}
		return "not configured"
	}
	if hasEnv("OPENAI_API_KEY") || hasEnv("ANTHROPIC_API_KEY") || hasEnv("GEMINI_API_KEY") || hasEnv("GOOGLE_API_KEY") {
		return "configured"
	}
	return "not configured"
}
