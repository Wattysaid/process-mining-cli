package commands

import (
	"fmt"

	"github.com/pm-assist/pm-assist/internal/cli"
	"github.com/spf13/cobra"
)

// NewReportCmd returns the report command.
func NewReportCmd(global *cli.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate notebooks and reports",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = global
			fmt.Println("[INFO] Report generation is not implemented yet.")
			return nil
		},
		Example: "  pm-assist report",
	}
	return cmd
}
