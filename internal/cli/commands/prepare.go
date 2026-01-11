package commands

import (
	"fmt"

	"github.com/pm-assist/pm-assist/internal/app"
	"github.com/spf13/cobra"
)

// NewPrepareCmd returns the prepare command.
func NewPrepareCmd(global *app.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prepare",
		Short: "Run data preparation pipeline",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = global
			fmt.Println("[INFO] Data preparation pipeline is not implemented yet.")
			return nil
		},
		Example: "  pm-assist prepare",
	}
	return cmd
}
