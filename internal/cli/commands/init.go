package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pm-assist/pm-assist/internal/app"
	"github.com/pm-assist/pm-assist/internal/business"
	"github.com/pm-assist/pm-assist/internal/config"
	"github.com/pm-assist/pm-assist/internal/paths"
	"github.com/pm-assist/pm-assist/internal/policy"
	"github.com/pm-assist/pm-assist/internal/profile"
	"github.com/pm-assist/pm-assist/internal/runner"
	"github.com/pm-assist/pm-assist/internal/scaffold"
	"github.com/pm-assist/pm-assist/internal/ui"
	"github.com/spf13/cobra"
)

// NewInitCmd returns the init command.
func NewInitCmd(global *app.GlobalFlags) *cobra.Command {
	var (
		flagProjectName    string
		flagUserName       string
		flagRole           string
		flagAptitude       string
		flagPromptDepth    string
		flagTemplate       string
		flagCustomFolders  string
		flagLLMProvider    string
		flagAllowLLM       string
		flagOfflineOnly    string
		flagCreateBusiness string
		flagBusinessName   string
		flagBusinessInd    string
		flagBusinessRegion string
		flagInstallDeps    string
	)
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create a new PM Assist project scaffold",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.PrintCommandStart(ui.CommandFrame{
				Title:     "pm-assist init",
				Purpose:   "Create a new project scaffold and profile",
				StepIndex: 1,
				StepTotal: 7,
				Writes:    []string{"pm-assist.yaml", ".profiles/", ".business/"},
				Asks:      []string{"project info", "profiles", "policy"},
				Next:      "pm-assist connect",
			})
			success := false
			defer func() {
				ui.PrintCommandEnd(ui.CommandFrame{Title: "pm-assist init", Next: "pm-assist connect"}, success)
			}()
			projectPath := global.ProjectPath
			if projectPath == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				projectPath = cwd
			}
			fmt.Println()
			fmt.Println("===============================================")
			fmt.Println(" PM Assist Setup · Guided Project Initialization")
			fmt.Println("===============================================")
			fmt.Println()

			projectName, err := resolveString(flagProjectName, "Project name", filepath.Base(projectPath), true)
			if err != nil {
				return err
			}
			userName, err := resolveString(flagUserName, "Your name", "", true)
			if err != nil {
				return err
			}
			role, err := resolveString(flagRole, "Your role", "", true)
			if err != nil {
				return err
			}
			aptitude, err := resolveChoice(flagAptitude, "Aptitude level", []string{"beginner", "intermediate", "expert"}, "intermediate", true)
			if err != nil {
				return err
			}
			promptDepth, err := resolveChoice(flagPromptDepth, "Prompt depth", []string{"short", "standard", "detailed"}, "standard", true)
			if err != nil {
				return err
			}
			templateChoice, err := resolveChoice(flagTemplate, "Project layout", []string{"standard", "minimal", "consulting", "custom"}, "standard", true)
			if err != nil {
				return err
			}
			customFolders := []string{}
			if templateChoice == "custom" {
				folderInput, err := resolveString(flagCustomFolders, "Folder list (comma-separated)", "data, outputs, docs", true)
				if err != nil {
					return err
				}
				customFolders = scaffold.ParseCustomFolders(folderInput)
			}
			llmProvider, err := resolveChoice(flagLLMProvider, "LLM provider", []string{"openai", "anthropic", "gemini", "ollama", "none"}, "none", true)
			if err != nil {
				return err
			}
			allowLLM, err := resolveBool(flagAllowLLM, "Allow LLM features?", true)
			if err != nil {
				return err
			}
			offlineOnly, err := resolveBool(flagOfflineOnly, "Enable offline-only policy?", false)
			if err != nil {
				return err
			}
			if offlineOnly && llmProvider != "ollama" && llmProvider != "none" {
				fmt.Println("[WARN] Offline-only mode blocks external LLM providers. Setting provider to none.")
				llmProvider = "none"
			}
			createBusiness, err := resolveBool(flagCreateBusiness, "Create a business profile now?", true)
			if err != nil {
				return err
			}
			businessName := ""
			businessIndustry := ""
			businessRegion := ""
			if createBusiness {
				businessName, err = resolveString(flagBusinessName, "Business name", "", true)
				if err != nil {
					return err
				}
				businessIndustry, err = resolveString(flagBusinessInd, "Business industry", "", true)
				if err != nil {
					return err
				}
				businessRegion, err = resolveString(flagBusinessRegion, "Business region", "", true)
				if err != nil {
					return err
				}
			}

			summary := []string{
				fmt.Sprintf("Project: %s", projectName),
				fmt.Sprintf("User: %s (%s, %s)", userName, role, aptitude),
				fmt.Sprintf("Layout: %s", templateChoice),
				fmt.Sprintf("LLM: %s (enabled=%t)", llmProvider, allowLLM),
				fmt.Sprintf("Offline-only: %t", offlineOnly),
			}
			if createBusiness {
				summary = append(summary, fmt.Sprintf("Business: %s (%s, %s)", businessName, businessIndustry, businessRegion))
			}
			confirm, err := confirmSummary("Confirm project setup", summary)
			if err != nil {
				return err
			}
			if !confirm {
				fmt.Println("[INFO] Init canceled by user.")
				return nil
			}

			printWalkthrough("Setup walkthrough", []string{
				"Create project scaffold",
				"Write config and policy defaults",
				"Save user profile",
				"Save business profile (optional)",
				"Provision Python environment",
			})

			step := 1
			totalSteps := 5
			if templateChoice == "custom" {
				template := scaffold.Template{Name: "custom", Folders: customFolders}
				printStepProgress(step, totalSteps, "Creating project scaffold (custom)")
				if err := scaffold.ApplyTemplate(projectPath, template); err != nil {
					return err
				}
			} else {
				template, ok := scaffold.Templates[templateChoice]
				if !ok {
					template = scaffold.StandardTemplate
				}
				printStepProgress(step, totalSteps, fmt.Sprintf("Creating project scaffold (%s)", templateChoice))
				if err := scaffold.ApplyTemplate(projectPath, template); err != nil {
					return err
				}
			}
			step++
			configPath := filepath.Join(projectPath, "pm-assist.yaml")
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				printStepProgress(step, totalSteps, "Writing config and policy defaults")
				llmEnabled := allowLLM
				cfg := config.Config{
					Path:    configPath,
					Version: config.CurrentSchemaVersion,
					Project: config.ProjectConfig{Name: projectName},
					Profiles: config.ProfilesConfig{
						Active: userName,
					},
					Business: config.BusinessConfig{Active: businessName},
					LLM:      config.LLMConfig{Provider: llmProvider},
					Policy: config.PolicyConfig{
						LLMEnabled:  &llmEnabled,
						OfflineOnly: offlineOnly,
					},
				}
				if err := cfg.Save(); err != nil {
					return err
				}
			}
			step++

			printStepProgress(step, totalSteps, "Saving user profile")
			_, err = profile.Save(projectPath, profile.Profile{
				Name:        userName,
				Role:        role,
				Aptitude:    aptitude,
				PromptDepth: promptDepth,
			})
			if err != nil {
				return err
			}
			step++

			if createBusiness {
				printStepProgress(step, totalSteps, "Saving business profile")
				_, err = business.Save(projectPath, business.Profile{
					Name:     businessName,
					Industry: businessIndustry,
					Region:   businessRegion,
				})
				if err != nil {
					return err
				}
			}
			if !createBusiness {
				printStepProgress(step, totalSteps, "Skipping business profile")
			}
			step++

			gitignorePath := filepath.Join(projectPath, ".gitignore")
			_ = scaffold.EnsureGitignore(gitignorePath, []string{"outputs/", ".venv/", ".profiles/", ".business/", ".pm-assist/", "*.pyc"})

			installDeps, err := resolveBool(flagInstallDeps, "Install Python dependencies now? (requires network)", false)
			if err != nil {
				return err
			}
			skillsRoot, err := paths.SkillsRoot(projectPath)
			if err != nil {
				return err
			}
			skillRequirements := paths.SkillPath(skillsRoot, "pm-99-utils-and-standards", "requirements.txt")
			if !installDeps {
				skillRequirements = ""
			}
			options := runner.VenvOptions{}
			if installDeps {
				options, err = resolveVenvOptions(projectPath, policy.Policy{OfflineOnly: offlineOnly})
				if err != nil {
					return err
				}
				options.Quiet = true
				options.LogPath = filepath.Join(projectPath, ".pm-assist", "logs", "pip-install.log")
				fmt.Printf("[INFO] Installing Python dependencies (this may take a few minutes). Logs: %s\n", options.LogPath)
			}
			printStepProgress(step, totalSteps, "Provisioning Python environment")
			venvRunner := &runner.Runner{ProjectPath: projectPath}
			if err := venvRunner.EnsureVenv(skillRequirements, options); err != nil {
				return err
			}

			fmt.Println("[SUCCESS] Project scaffold created")
			cfg, _ := config.Load(configPath)
			ui.PrintSplash(cfg, ui.SplashOptions{
				CompletedCommand: "init",
				WorkingDir:       projectPath,
			})
			success = true
			return nil
		},
		Example: "  pm-assist init",
	}
	cmd.Flags().StringVar(&flagProjectName, "project-name", "", "Project name")
	cmd.Flags().StringVar(&flagUserName, "user-name", "", "User name")
	cmd.Flags().StringVar(&flagRole, "role", "", "User role")
	cmd.Flags().StringVar(&flagAptitude, "aptitude", "", "Aptitude level (beginner|intermediate|expert)")
	cmd.Flags().StringVar(&flagPromptDepth, "prompt-depth", "", "Prompt depth (short|standard|detailed)")
	cmd.Flags().StringVar(&flagTemplate, "template", "", "Project layout (standard|minimal|consulting|custom)")
	cmd.Flags().StringVar(&flagCustomFolders, "custom-folders", "", "Custom folders (comma-separated)")
	cmd.Flags().StringVar(&flagLLMProvider, "llm-provider", "", "LLM provider (openai|anthropic|gemini|ollama|none)")
	cmd.Flags().StringVar(&flagAllowLLM, "allow-llm", "", "Allow LLM features (true|false)")
	cmd.Flags().StringVar(&flagOfflineOnly, "offline-only", "", "Enable offline-only policy (true|false)")
	cmd.Flags().StringVar(&flagCreateBusiness, "create-business", "", "Create a business profile now (true|false)")
	cmd.Flags().StringVar(&flagBusinessName, "business-name", "", "Business name")
	cmd.Flags().StringVar(&flagBusinessInd, "business-industry", "", "Business industry")
	cmd.Flags().StringVar(&flagBusinessRegion, "business-region", "", "Business region")
	cmd.Flags().StringVar(&flagInstallDeps, "install-deps", "", "Install Python dependencies now (true|false)")
	return cmd
}
