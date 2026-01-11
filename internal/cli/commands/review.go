package commands

import (
	"fmt"

	"github.com/pm-assist/pm-assist/internal/app"
	"github.com/spf13/cobra"
)

// NewReviewCmd returns the review command.
func NewReviewCmd(global *app.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review",
		Short: "Run QA checks and summarize issues",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = global
			fmt.Println("[INFO] QA review is not implemented yet.")
			return nil
		},
		Example: "  pm-assist review",
	}
	return cmd
}
