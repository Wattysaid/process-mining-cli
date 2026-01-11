package cli

import (
	"fmt"
	"os"

	"github.com/pm-assist/pm-assist/internal/cli/commands"
	"github.com/pm-assist/pm-assist/internal/cli/prompt"
	"github.com/pm-assist/pm-assist/internal/config"
	"github.com/pm-assist/pm-assist/internal/logging"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pm-assist",
	Short: "PM Assist is an enterprise process mining CLI assistant",
	Long:  "PM Assist is an enterprise process mining CLI assistant for guided, reproducible end-to-end workflows.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := logging.Init(Global.LogLevel, Global.JSONOutput); err != nil {
			return err
		}
		cfg, err := config.Load(Global.ConfigPath)
		if err != nil {
			return err
		}
		Global.Config = cfg
		prompt.SetNonInteractive(Global.NonInteractive)
		return nil
	},
}

// GlobalFlags stores CLI flags used across commands.
type GlobalFlags struct {
	ConfigPath     string
	ProjectPath    string
	RunID          string
	NonInteractive bool
	LogLevel       string
	JSONOutput     bool
	Yes            bool
	LLMProvider    string
	ProfileName    string
	Config         *config.Config
}

var Global = &GlobalFlags{}

func init() {
	rootCmd.PersistentFlags().StringVar(&Global.ConfigPath, "config", "", "Path to config file (default: ./pm-assist.yaml if present)")
	rootCmd.PersistentFlags().StringVar(&Global.ProjectPath, "project", "", "Project root (default: current directory)")
	rootCmd.PersistentFlags().StringVar(&Global.RunID, "run-id", "", "Reuse a run folder")
	rootCmd.PersistentFlags().BoolVar(&Global.NonInteractive, "non-interactive", false, "Fail if required inputs are missing")
	rootCmd.PersistentFlags().StringVar(&Global.LogLevel, "log-level", "info", "Log level: debug|info|warn|error")
	rootCmd.PersistentFlags().BoolVar(&Global.JSONOutput, "json", false, "Machine-readable output")
	rootCmd.PersistentFlags().BoolVar(&Global.Yes, "yes", false, "Assume yes for safe prompts")
	rootCmd.PersistentFlags().StringVar(&Global.LLMProvider, "llm-provider", "", "Override configured LLM provider (openai|anthropic|gemini|ollama|none)")
	rootCmd.PersistentFlags().StringVar(&Global.ProfileName, "profile", "", "Use a specific user profile from .profiles/")

	rootCmd.AddCommand(
		commands.NewVersionCmd(Global),
		commands.NewDoctorCmd(Global),
		commands.NewInitCmd(Global),
		commands.NewConnectCmd(Global),
		commands.NewIngestCmd(Global),
		commands.NewMapCmd(Global),
		commands.NewPrepareCmd(Global),
		commands.NewMineCmd(Global),
		commands.NewReportCmd(Global),
		commands.NewReviewCmd(Global),
		commands.NewAgentCmd(Global),
		commands.NewProfileCmd(Global),
	)
}

// Execute runs the root command.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}
