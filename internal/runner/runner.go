package runner

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Runner handles python environment provisioning and pipeline execution.
type Runner struct {
	ProjectPath string
	VenvPath    string
}

// EnsureVenv creates a project-local venv and installs requirements when needed.
func (r *Runner) EnsureVenv(requirementsPath string) error {
	if r.ProjectPath == "" {
		return errors.New("project path is required")
	}
	if r.VenvPath == "" {
		r.VenvPath = filepath.Join(r.ProjectPath, ".venv")
	}

	if _, err := os.Stat(r.VenvPath); os.IsNotExist(err) {
		python, err := detectPython()
		if err != nil {
			return err
		}
		if err := runCommand(python, []string{"-m", "venv", r.VenvPath}); err != nil {
			return fmt.Errorf("failed to create venv: %w", err)
		}
	}

	if requirementsPath != "" {
		pipPath := filepath.Join(r.VenvPath, "bin", "pip")
		if _, err := os.Stat(pipPath); err != nil {
			return fmt.Errorf("pip not found in venv: %w", err)
		}
		if err := runCommand(pipPath, []string{"install", "-r", requirementsPath}); err != nil {
			return fmt.Errorf("failed to install requirements: %w", err)
		}
	}

	return nil
}

// RunModule runs a python module inside the venv.
func (r *Runner) RunModule(module string, args []string, env map[string]string) error {
	if module == "" {
		return errors.New("module name is required")
	}
	pythonPath := filepath.Join(r.VenvPath, "bin", "python")
	cmdArgs := append([]string{"-m", module}, args...)
	return runCommandWithEnv(pythonPath, cmdArgs, env)
}

func detectPython() (string, error) {
	python, err := exec.LookPath("python3")
	if err == nil {
		return python, nil
	}
	python, err = exec.LookPath("python")
	if err != nil {
		return "", errors.New("python not found")
	}
	return python, nil
}

func runCommand(binary string, args []string) error {
	return runCommandWithEnv(binary, args, nil)
}

func runCommandWithEnv(binary string, args []string, env map[string]string) error {
	cmd := exec.Command(binary, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), formatEnv(env)...)
	}
	return cmd.Run()
}

func formatEnv(env map[string]string) []string {
	formatted := make([]string, 0, len(env))
	for key, value := range env {
		formatted = append(formatted, fmt.Sprintf("%s=%s", key, value))
	}
	return formatted
}
