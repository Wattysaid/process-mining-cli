package commands

import (
	"fmt"

	"github.com/pm-assist/pm-assist/internal/cli"
	"github.com/spf13/cobra"
)

// NewIngestCmd returns the ingest command.
func NewIngestCmd(global *cli.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ingest",
		Short: "Ingest data into a staging dataset",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = global
			fmt.Println("[INFO] Ingest pipeline is not implemented yet.")
			return nil
		},
		Example: "  pm-assist ingest",
	}
	return cmd
}
