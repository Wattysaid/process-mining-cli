package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pm-assist/pm-assist/internal/app"
	"github.com/pm-assist/pm-assist/internal/cli/prompt"
	"github.com/pm-assist/pm-assist/internal/profile"
	"github.com/spf13/cobra"
)

// NewProfileCmd returns the profile command.
func NewProfileCmd(global *app.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage user profiles",
	}
	cmd.AddCommand(newProfileInitCmd(global))
	cmd.AddCommand(newProfileSetCmd(global))
	cmd.AddCommand(newProfileShowCmd(global))
	return cmd
}

func newProfileInitCmd(global *app.GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create a new user profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectPath := global.ProjectPath
			if projectPath == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				projectPath = cwd
			}

			name, err := prompt.AskString("Your name", "", true)
			if err != nil {
				return err
			}
			role, err := prompt.AskString("Your role", "", true)
			if err != nil {
				return err
			}
			aptitude, err := prompt.AskChoice("Aptitude level", []string{"beginner", "intermediate", "expert"}, "intermediate", true)
			if err != nil {
				return err
			}
			promptDepth, err := prompt.AskChoice("Prompt depth", []string{"short", "standard", "detailed"}, "standard", true)
			if err != nil {
				return err
			}

			path, err := profile.Save(projectPath, profile.Profile{
				Name:        name,
				Role:        role,
				Aptitude:    aptitude,
				PromptDepth: promptDepth,
			})
			if err != nil {
				return err
			}
			fmt.Printf("[SUCCESS] Profile saved: %s\n", path)
			return nil
		},
		Example: "  pm-assist profile init",
	}
}

func newProfileSetCmd(global *app.GlobalFlags) *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set active profile name in config",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = global
			if name == "" {
				return fmt.Errorf("profile name is required")
			}
			fmt.Printf("[INFO] Requested active profile: %s\n", name)
			fmt.Println("[INFO] Profile selection is not implemented yet.")
			return nil
		},
		Example: "  pm-assist profile set --name jane-doe",
	}
	cmd.Flags().StringVar(&name, "name", "", "Profile name")
	return cmd
}

func newProfileShowCmd(global *app.GlobalFlags) *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show a profile",
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
				return fmt.Errorf("profile name is required")
			}
			path := filepath.Join(projectPath, ".profiles", name+".yaml")
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			fmt.Print(string(content))
			return nil
		},
		Example: "  pm-assist profile show --name jane-doe",
	}
	cmd.Flags().StringVar(&name, "name", "", "Profile name")
	return cmd
}
