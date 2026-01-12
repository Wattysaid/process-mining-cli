package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pm-assist/pm-assist/internal/app"
	"github.com/pm-assist/pm-assist/internal/ui"
	"github.com/spf13/cobra"
)

// NewStartCmd returns the wizard command.
func NewStartCmd(global *app.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Run the guided end-to-end workflow",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.PrintCommandStart(ui.CommandFrame{
				Title:     "pm-assist start",
				Purpose:   "Guided end-to-end workflow",
				StepIndex: 1,
				StepTotal: 7,
				Next:      "pm-assist status",
			})
			success := false
			defer func() {
				ui.PrintCommandEnd(ui.CommandFrame{Title: "pm-assist start", Next: "pm-assist status"}, success)
			}()

			projectPath := global.ProjectPath
			if projectPath == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				projectPath = cwd
			}
			runID := global.RunID
			if runID == "" {
				runID = defaultRunID()
			}

			steps := []string{"connect", "ingest", "map", "prepare", "mine", "report", "review"}
			for i, step := range steps {
				printStepProgress(i+1, len(steps), fmt.Sprintf("Running %s", step))
				if err := runSubcommand(global, step, runID, projectPath); err != nil {
					return err
				}
			}
			success = true
			return nil
		},
	}
	return cmd
}

func runSubcommand(global *app.GlobalFlags, name string, runID string, projectPath string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	args := []string{name, "--run-id", runID}
	if global.ProjectPath != "" {
		args = append(args, "--project", projectPath)
	}
	if global.ConfigPath != "" {
		args = append(args, "--config", global.ConfigPath)
	} else {
		configPath := filepath.Join(projectPath, "pm-assist.yaml")
		args = append(args, "--config", configPath)
	}
	cmd := exec.Command(exe, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
