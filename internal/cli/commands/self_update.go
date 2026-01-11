package commands

import (
	"fmt"

	"github.com/pm-assist/pm-assist/internal/app"
	"github.com/pm-assist/pm-assist/internal/updater"
	"github.com/spf13/cobra"
)

// NewSelfUpdateCmd returns the self-update command.
func NewSelfUpdateCmd(global *app.GlobalFlags) *cobra.Command {
	var version string
	var baseURL string
	cmd := &cobra.Command{
		Use:   "self-update",
		Short: "Download and replace the pm-assist binary",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = global
			fmt.Println("[INFO] Updating pm-assist...")
			if err := updater.Update(updater.Options{BaseURL: baseURL, Version: version}); err != nil {
				return err
			}
			fmt.Println("[SUCCESS] Update complete.")
			return nil
		},
		Example: "  pm-assist self-update --version v0.1.0",
	}
	cmd.Flags().StringVar(&version, "version", "", "Version tag (default: latest)")
	cmd.Flags().StringVar(&baseURL, "base-url", "", "Release base URL (default: GitHub releases)")
	return cmd
}
