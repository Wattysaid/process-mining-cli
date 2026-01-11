package commands

import (
	"fmt"

	"github.com/pm-assist/pm-assist/internal/app"
	"github.com/spf13/cobra"
)

// NewMapCmd returns the map command.
func NewMapCmd(global *app.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "map",
		Short: "Map columns to process mining schema",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = global
			fmt.Println("[INFO] Column mapping is not implemented yet.")
			return nil
		},
		Example: "  pm-assist map",
	}
	return cmd
}
