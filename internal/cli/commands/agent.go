package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/pm-assist/pm-assist/internal/app"
	"github.com/pm-assist/pm-assist/internal/cli/prompt"
	"github.com/pm-assist/pm-assist/internal/config"
	"github.com/pm-assist/pm-assist/internal/policy"
	"github.com/pm-assist/pm-assist/internal/profile"
	"github.com/pm-assist/pm-assist/internal/ui"
	"github.com/spf13/cobra"
)

// NewAgentCmd returns the agent command.
func NewAgentCmd(global *app.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "LLM-assisted guidance and setup",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAgentGuide(global)
		},
	}
	cmd.AddCommand(newAgentSetupCmd(global))
	return cmd
}

func newAgentSetupCmd(global *app.GlobalFlags) *cobra.Command {
	var flagProvider string
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Configure LLM provider settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.PrintCommandStart(ui.CommandFrame{
				Title:   "pm-assist agent setup",
				Purpose: "Configure LLM provider settings",
				Writes:  []string{"pm-assist.yaml"},
				Next:    "pm-assist agent",
			})
			success := false
			defer func() {
				ui.PrintCommandEnd(ui.CommandFrame{Title: "pm-assist agent setup", Next: "pm-assist agent"}, success)
			}()
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
			if cfg.Path == "" {
				cfg.Path = filepath.Join(projectPath, "pm-assist.yaml")
			}
			policies := policy.FromConfig(cfg)
			if policies.LLMEnabled != nil && !*policies.LLMEnabled {
				return fmt.Errorf("LLM is disabled by policy")
			}
			provider, err := resolveChoice(flagProvider, "LLM provider", []string{"openai", "anthropic", "gemini", "ollama", "none"}, "none", true)
			if err != nil {
				return err
			}
			if policies.OfflineOnly && provider != "ollama" && provider != "none" {
				return fmt.Errorf("offline-only policy blocks external LLM providers")
			}
			cfg.LLM.Provider = provider
			if provider == "none" {
				cfg.LLM.Model = ""
				cfg.LLM.Endpoint = ""
				cfg.LLM.MaxTokens = 0
				cfg.LLM.TokenBudget = 0
				cfg.LLM.CostCapUSD = 0
			} else {
				modelDefault := defaultModel(provider)
				model, err := resolveString("", "Model", modelDefault, true)
				if err != nil {
					return err
				}
				tokenBudgetText, err := resolveString("", "Per-run token budget", "8000", true)
				if err != nil {
					return err
				}
				tokenBudget, err := strconv.Atoi(tokenBudgetText)
				if err != nil || tokenBudget <= 0 {
					return fmt.Errorf("invalid token budget: %s", tokenBudgetText)
				}
				maxTokensText, err := resolveString("", "Max tokens per call", "1200", true)
				if err != nil {
					return err
				}
				maxTokens, err := strconv.Atoi(maxTokensText)
				if err != nil || maxTokens <= 0 {
					return fmt.Errorf("invalid max tokens: %s", maxTokensText)
				}
				costCapText, err := resolveString("", "Cost cap in USD (0 for no cap)", "0", true)
				if err != nil {
					return err
				}
				costCap, err := strconv.ParseFloat(costCapText, 64)
				if err != nil || costCap < 0 {
					return fmt.Errorf("invalid cost cap: %s", costCapText)
				}
				endpoint := ""
				if provider == "ollama" {
					endpoint, err = resolveString("", "Ollama endpoint", "http://localhost:11434", true)
					if err != nil {
						return err
					}
				}
				cfg.LLM.Model = model
				cfg.LLM.Endpoint = endpoint
				cfg.LLM.MaxTokens = maxTokens
				cfg.LLM.TokenBudget = tokenBudget
				cfg.LLM.CostCapUSD = costCap
			}

			if cfg.Policy.LLMEnabled == nil {
				enabled := provider != "none"
				cfg.Policy.LLMEnabled = &enabled
			}

			if err := cfg.Save(); err != nil {
				return err
			}

			fmt.Printf("[SUCCESS] LLM provider configured: %s\n", provider)
			printProviderNextSteps(provider)
			success = true
			return nil
		},
		Example: "  pm-assist agent setup",
	}
	cmd.Flags().StringVar(&flagProvider, "provider", "", "LLM provider (openai|anthropic|gemini|ollama|none)")
	return cmd
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

func runAgentGuide(global *app.GlobalFlags) error {
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
	promptDepth := "standard"
	if cfg.Profiles.Active != "" {
		if prof, err := profile.Load(projectPath, cfg.Profiles.Active); err == nil && prof.PromptDepth != "" {
			promptDepth = prof.PromptDepth
		}
	}

	goal, err := prompt.AskString("What would you like to achieve?", "", true)
	if err != nil {
		return err
	}
	fmt.Printf("[INFO] Goal captured: %s\n", goal)

	recommended := recommendNextStep(cfg, projectPath)
	if recommended == "" {
		fmt.Println("[INFO] No recommended next step found.")
		return nil
	}
	if promptDepth == "detailed" {
		fmt.Printf("[INFO] Recommended next step based on project state: %s\n", recommended)
		fmt.Println("[INFO] You can override this recommendation at any time.")
	} else {
		fmt.Printf("[INFO] Recommended next step: %s\n", recommended)
	}
	confirm, err := prompt.AskBool(fmt.Sprintf("Show command for %s now?", recommended), true)
	if err != nil {
		return err
	}
	if confirm {
		fmt.Printf("Next command: pm-assist %s\n", recommended)
	}
	return nil
}

func recommendNextStep(cfg *config.Config, projectPath string) string {
	if cfg == nil || cfg.Path == "" {
		return "init"
	}
	if len(cfg.Connectors) == 0 {
		return "connect"
	}
	if cfg.Mapping == nil {
		return "map"
	}
	manifestPath, manifest := findLatestManifest(filepath.Join(projectPath, "outputs"))
	if manifest == nil {
		_ = manifestPath
		return "ingest"
	}
	next := nextRecommendedStep(manifest)
	if next == "" {
		return ""
	}
	return next
}

func findLatestManifest(outputsPath string) (string, *runManifest) {
	pattern := filepath.Join(outputsPath, "*", "run_manifest.json")
	candidates, err := filepath.Glob(pattern)
	if err != nil || len(candidates) == 0 {
		return "", nil
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
		return latest, nil
	}
	var manifest runManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return latest, nil
	}
	return latest, &manifest
}

func nextRecommendedStep(manifest *runManifest) string {
	if manifest == nil {
		return ""
	}
	pipeline := []string{"ingest", "map", "prepare", "mine", "report", "review"}
	statusByStep := make(map[string]string)
	for _, step := range manifest.Steps {
		statusByStep[strings.ToLower(step.Name)] = step.Status
	}
	for _, step := range pipeline {
		status := statusByStep[step]
		if status == "" || status == "failed" || status == "started" {
			return step
		}
	}
	return ""
}

func defaultModel(provider string) string {
	switch provider {
	case "openai":
		return "gpt-4o-mini"
	case "anthropic":
		return "claude-3-5-sonnet-latest"
	case "gemini":
		return "gemini-1.5-pro"
	case "ollama":
		return "llama3"
	default:
		return ""
	}
}

func printProviderNextSteps(provider string) {
	switch provider {
	case "openai":
		fmt.Println("Next: export OPENAI_API_KEY=\"...\"")
	case "anthropic":
		fmt.Println("Next: export ANTHROPIC_API_KEY=\"...\"")
	case "gemini":
		fmt.Println("Next: export GEMINI_API_KEY=\"...\" or GOOGLE_API_KEY=\"...\"")
	case "ollama":
		fmt.Println("Next: ensure Ollama is running and reachable at OLLAMA_HOST.")
	case "none":
		fmt.Println("LLM features are disabled. You can re-enable them with `pm-assist agent setup`.")
	}
}
