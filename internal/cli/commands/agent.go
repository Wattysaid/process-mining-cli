package commands

import (
	"fmt"

	"github.com/pm-assist/pm-assist/internal/app"
	"github.com/pm-assist/pm-assist/internal/cli/prompt"
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
	return &cobra.Command{
		Use:   "setup",
		Short: "Configure LLM provider settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = global
			provider, err := prompt.AskChoice("LLM provider", []string{"openai", "anthropic", "gemini", "ollama", "none"}, "none", true)
			if err != nil {
				return err
			}
			fmt.Printf("[INFO] Selected provider: %s\n", provider)
			fmt.Println("[INFO] Provider setup is not implemented yet.")
			return nil
		},
		Example: "  pm-assist agent setup",
	}
}
