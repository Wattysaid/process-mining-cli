package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pm-assist/pm-assist/internal/app"
	"github.com/pm-assist/pm-assist/internal/business"
	"github.com/pm-assist/pm-assist/internal/config"
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
	var (
		flagName     string
		flagIndustry string
		flagRegion   string
	)
	cmd := &cobra.Command{
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

			name, err := resolveString(flagName, "Business name", "", true)
			if err != nil {
				return err
			}
			industry, err := resolveString(flagIndustry, "Industry", "", true)
			if err != nil {
				return err
			}
			region, err := resolveString(flagRegion, "Region", "", true)
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
	cmd.Flags().StringVar(&flagName, "name", "", "Business name")
	cmd.Flags().StringVar(&flagIndustry, "industry", "", "Industry")
	cmd.Flags().StringVar(&flagRegion, "region", "", "Region")
	return cmd
}

func newBusinessSetCmd(global *app.GlobalFlags) *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set active business name in config",
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
			if _, err := os.Stat(path); err != nil {
				return fmt.Errorf("business not found: %s", name)
			}
			cfg, err := config.Load(global.ConfigPath)
			if err != nil {
				return err
			}
			if cfg.Path == "" {
				cfg.Path = filepath.Join(projectPath, "pm-assist.yaml")
			}
			cfg.Business.Active = name
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Printf("[SUCCESS] Active business set to %s\n", name)
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
