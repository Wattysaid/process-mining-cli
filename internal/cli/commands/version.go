package commands

import (
	"fmt"

	"github.com/pm-assist/pm-assist/internal/cli"
	"github.com/spf13/cobra"
)

const (
	version = "0.1.0"
	commit  = "dev"
)

// NewVersionCmd returns the version command.
func NewVersionCmd(global *cli.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version and build metadata",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = global
			fmt.Printf("pm-assist %s (%s)\n", version, commit)
			return nil
		},
		Example: "  pm-assist version",
	}
	return cmd
}
