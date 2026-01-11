package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pm-assist/pm-assist/internal/app"
	"github.com/pm-assist/pm-assist/internal/business"
	"github.com/pm-assist/pm-assist/internal/cli/prompt"
	"github.com/spf13/cobra"
)

// NewBusinessCmd returns the business command.
func NewBusinessCmd(global *app.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "business",
		Short: "Manage business profiles",
	}
	cmd.AddCommand(newBusinessInitCmd(global))
	cmd.AddCommand(newBusinessSetCmd(global))
	cmd.AddCommand(newBusinessShowCmd(global))
	return cmd
}

func newBusinessInitCmd(global *app.GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create a new business profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectPath := global.ProjectPath
			if projectPath == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				projectPath = cwd
			}

			name, err := prompt.AskString("Business name", "", true)
			if err != nil {
				return err
			}
			industry, err := prompt.AskString("Industry", "", true)
			if err != nil {
				return err
			}
			region, err := prompt.AskString("Region", "", true)
			if err != nil {
				return err
			}

			path, err := business.Save(projectPath, business.Profile{
				Name:     name,
				Industry: industry,
				Region:   region,
			})
			if err != nil {
				return err
			}
			fmt.Printf("[SUCCESS] Business profile saved: %s\n", path)
			return nil
		},
		Example: "  pm-assist business init",
	}
}

func newBusinessSetCmd(global *app.GlobalFlags) *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set active business name in config",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = global
			if name == "" {
				return fmt.Errorf("business name is required")
			}
			fmt.Printf("[INFO] Requested active business: %s\n", name)
			fmt.Println("[INFO] Business selection is not implemented yet.")
			return nil
		},
		Example: "  pm-assist business set --name acme",
	}
	cmd.Flags().StringVar(&name, "name", "", "Business name")
	return cmd
}

func newBusinessShowCmd(global *app.GlobalFlags) *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show a business profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectPath := global.ProjectPath
			if projectPath == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				projectPath = cwd
			}
			if name == "" {
				return fmt.Errorf("business name is required")
			}
			path := filepath.Join(projectPath, ".business", name+".yaml")
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			fmt.Print(string(content))
			return nil
		},
		Example: "  pm-assist business show --name acme",
	}
	cmd.Flags().StringVar(&name, "name", "", "Business name")
	return cmd
}
