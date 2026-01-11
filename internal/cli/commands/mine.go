package commands

import (
	"fmt"

	"github.com/pm-assist/pm-assist/internal/app"
	"github.com/spf13/cobra"
)

// NewMineCmd returns the mine command.
func NewMineCmd(global *app.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mine",
		Short: "Run process mining analysis",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = global
			fmt.Println("[INFO] Mining pipeline is not implemented yet.")
			return nil
		},
		Example: "  pm-assist mine",
	}
	return cmd
}
