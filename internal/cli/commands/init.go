package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pm-assist/pm-assist/internal/app"
	"github.com/pm-assist/pm-assist/internal/business"
	"github.com/pm-assist/pm-assist/internal/cli/prompt"
	"github.com/pm-assist/pm-assist/internal/profile"
	"github.com/pm-assist/pm-assist/internal/runner"
	"github.com/pm-assist/pm-assist/internal/scaffold"
	"github.com/spf13/cobra"
)

// NewInitCmd returns the init command.
func NewInitCmd(global *app.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create a new PM Assist project scaffold",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectPath := global.ProjectPath
			if projectPath == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				projectPath = cwd
			}

			projectName, err := prompt.AskString("Project name", filepath.Base(projectPath), true)
			if err != nil {
				return err
			}
			userName, err := prompt.AskString("Your name", "", true)
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
			templateChoice, err := prompt.AskChoice("Project layout", []string{"standard", "custom"}, "standard", true)
			if err != nil {
				return err
			}
			customFolders := []string{}
			if templateChoice == "custom" {
				folderInput, err := prompt.AskString("Folder list (comma-separated)", "data, outputs, docs", true)
				if err != nil {
					return err
				}
				customFolders = scaffold.ParseCustomFolders(folderInput)
			}
			llmProvider, err := prompt.AskChoice("LLM provider", []string{"openai", "anthropic", "gemini", "ollama", "none"}, "none", true)
			if err != nil {
				return err
			}
			createBusiness, err := prompt.AskBool("Create a business profile now?", true)
			if err != nil {
				return err
			}
			businessName := ""
			businessIndustry := ""
			businessRegion := ""
			if createBusiness {
				businessName, err = prompt.AskString("Business name", "", true)
				if err != nil {
					return err
				}
				businessIndustry, err = prompt.AskString("Business industry", "", true)
				if err != nil {
					return err
				}
				businessRegion, err = prompt.AskString("Business region", "", true)
				if err != nil {
					return err
				}
			}

			if templateChoice == "custom" {
				template := scaffold.Template{Name: "custom", Folders: customFolders}
				if err := scaffold.ApplyTemplate(projectPath, template); err != nil {
					return err
				}
			} else {
				if err := scaffold.ApplyTemplate(projectPath, scaffold.StandardTemplate); err != nil {
					return err
				}
			}
			configPath := filepath.Join(projectPath, "pm-assist.yaml")
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				content := fmt.Sprintf("project:\n  name: %s\nprofiles:\n  active: %s\nbusiness:\n  active: %s\nllm:\n  provider: %s\n", projectName, userName, businessName, llmProvider)
				if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
					return err
				}
			}

			_, err = profile.Save(projectPath, profile.Profile{
				Name:        userName,
				Role:        role,
				Aptitude:    aptitude,
				PromptDepth: promptDepth,
			})
			if err != nil {
				return err
			}

			if createBusiness {
				_, err = business.Save(projectPath, business.Profile{
					Name:     businessName,
					Industry: businessIndustry,
					Region:   businessRegion,
				})
				if err != nil {
					return err
				}
			}

			gitignorePath := filepath.Join(projectPath, ".gitignore")
			_ = scaffold.EnsureGitignore(gitignorePath, []string{"outputs/", ".venv/", ".profiles/", ".business/", "*.pyc"})

			installDeps, err := prompt.AskBool("Install Python dependencies now? (requires network)", false)
			if err != nil {
				return err
			}
			skillRequirements := filepath.Join(projectPath, ".codex", "skills", "cli-tool-skills", "pm-99-utils-and-standards", "requirements.txt")
			if !installDeps {
				skillRequirements = ""
			}
			venvRunner := &runner.Runner{ProjectPath: projectPath}
			if err := venvRunner.EnsureVenv(skillRequirements); err != nil {
				return err
			}

			fmt.Println("[SUCCESS] Project scaffold created")
			return nil
		},
		Example: "  pm-assist init",
	}
	return cmd
}
