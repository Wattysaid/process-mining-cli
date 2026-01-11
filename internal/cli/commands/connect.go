package commands

import (
	"fmt"

	"github.com/pm-assist/pm-assist/internal/cli"
	"github.com/spf13/cobra"
)

// NewConnectCmd returns the connect command.
func NewConnectCmd(global *cli.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Register read-only data connectors",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = global
			fmt.Println("[INFO] Connector setup is not implemented yet.")
			fmt.Println("[INFO] This command will prompt for connection details and verify read-only access.")
			return nil
		},
		Example: "  pm-assist connect",
	}
	return cmd
}
