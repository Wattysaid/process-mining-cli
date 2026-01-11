package commands

import (
	"fmt"

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
			cfg, err := config.Load(global.ConfigPath)
			if err != nil {
				return err
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
			fmt.Printf("[INFO] Selected provider: %s\n", provider)
			fmt.Println("[INFO] Provider setup is not implemented yet.")
			return nil
		},
		Example: "  pm-assist agent setup",
	}
	cmd.Flags().StringVar(&flagProvider, "provider", "", "LLM provider (openai|anthropic|gemini|ollama|none)")
	return cmd
}
