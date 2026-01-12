package commands

import (
	"fmt"
	"os"

	"github.com/pm-assist/pm-assist/internal/app"
	"github.com/pm-assist/pm-assist/internal/config"
	"github.com/pm-assist/pm-assist/internal/ui"
	"github.com/spf13/cobra"
)

// NewConnectorsCmd returns the connectors command.
func NewConnectorsCmd(global *app.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connectors",
		Short: "Manage connectors",
	}
	cmd.AddCommand(newConnectorsListCmd(global))
	return cmd
}

func newConnectorsListCmd(global *app.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configured connectors",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.PrintCommandStart(ui.CommandFrame{
				Title:   "pm-assist connectors list",
				Purpose: "List configured connectors",
				Next:    "pm-assist connect",
			})
			success := false
			defer func() {
				ui.PrintCommandEnd(ui.CommandFrame{Title: "pm-assist connectors list"}, success)
			}()
			cfg, err := config.Load(global.ConfigPath)
			if err != nil {
				return err
			}
			if len(cfg.Connectors) == 0 {
				fmt.Println("[INFO] No connectors configured.")
				success = true
				return nil
			}
			for _, connector := range cfg.Connectors {
				status := "unknown"
				if connector.Type == "file" && connector.File != nil && len(connector.File.Paths) > 0 {
					if _, err := os.Stat(connector.File.Paths[0]); err == nil {
						status = "ok"
					} else {
						status = "missing"
					}
				}
				fmt.Printf("- %s (%s) [%s]\n", connector.Name, connector.Type, status)
			}
			success = true
			return nil
		},
	}
	return cmd
}
