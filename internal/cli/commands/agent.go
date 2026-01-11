package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pm-assist/pm-assist/internal/app"
	"github.com/pm-assist/pm-assist/internal/config"
	"github.com/pm-assist/pm-assist/internal/policy"
	"github.com/spf13/cobra"
)

// NewAgentCmd returns the agent command.
func NewAgentCmd(global *app.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "LLM-assisted guidance and setup",
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
			return nil
		},
		Example: "  pm-assist agent setup",
	}
	cmd.Flags().StringVar(&flagProvider, "provider", "", "LLM provider (openai|anthropic|gemini|ollama|none)")
	return cmd
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
