package commands

import (
	"fmt"

	"github.com/pm-assist/pm-assist/internal/app"
	"github.com/pm-assist/pm-assist/internal/buildinfo"
	"github.com/spf13/cobra"
)

// NewVersionCmd returns the version command.
func NewVersionCmd(global *app.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version and build metadata",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = global
			fmt.Printf("pm-assist %s (%s)\n", buildinfo.Version, buildinfo.Commit)
			return nil
		},
		Example: "  pm-assist version",
	}
	return cmd
}
