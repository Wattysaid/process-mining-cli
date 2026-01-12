package runner

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Runner handles python environment provisioning and pipeline execution.
type Runner struct {
	ProjectPath string
	VenvPath    string
}

type VenvOptions struct {
	Offline    bool
	WheelsPath string
	Quiet      bool
	LogPath    string
}

// EnsureVenv creates a project-local venv and installs requirements when needed.
func (r *Runner) EnsureVenv(requirementsPath string, options VenvOptions) error {
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
		args := []string{"install"}
		if options.WheelsPath != "" {
			args = append(args, "--find-links", options.WheelsPath)
		}
		if options.Offline {
			if options.WheelsPath == "" {
				return errors.New("offline mode requires bundled wheels")
			}
			args = append(args, "--no-index")
		}
		args = append(args, "-r", requirementsPath)
		if options.Quiet && options.LogPath != "" {
			if err := runCommandToFile(pipPath, args, options.LogPath); err != nil {
				tail := tailLog(options.LogPath, 30)
				if tail != "" {
					return fmt.Errorf("failed to install requirements (see %s): %w\nLast output:\n%s", options.LogPath, err, tail)
				}
				return fmt.Errorf("failed to install requirements (see %s): %w", options.LogPath, err)
			}
		} else if err := runCommand(pipPath, args); err != nil {
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

// RunScript runs a python script inside the venv.
func (r *Runner) RunScript(scriptPath string, args []string, env map[string]string) error {
	if scriptPath == "" {
		return errors.New("script path is required")
	}
	pythonPath := filepath.Join(r.VenvPath, "bin", "python")
	cmdArgs := append([]string{scriptPath}, args...)
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

func runCommandToFile(binary string, args []string, logPath string) error {
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	cmd := exec.Command(binary, args...)
	cmd.Stdout = file
	cmd.Stderr = file
	return cmd.Run()
}

func tailLog(path string, maxLines int) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) <= maxLines {
		return strings.Join(lines, "\n")
	}
	return strings.Join(lines[len(lines)-maxLines:], "\n")
}

func formatEnv(env map[string]string) []string {
	formatted := make([]string, 0, len(env))
	for key, value := range env {
		formatted = append(formatted, fmt.Sprintf("%s=%s", key, value))
	}
	return formatted
}
