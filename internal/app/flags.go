package app

import "github.com/pm-assist/pm-assist/internal/config"

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
