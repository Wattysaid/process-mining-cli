package runner

import "errors"

// Runner handles python environment provisioning and pipeline execution.
type Runner struct {
	VenvPath string
}

// EnsureVenv is a placeholder for future venv setup.
func (r *Runner) EnsureVenv() error {
	if r.VenvPath == "" {
		return errors.New("venv path is required")
	}
	return nil
}

// RunModule is a placeholder for running python modules.
func (r *Runner) RunModule(module string, args []string, env map[string]string) error {
	_ = module
	_ = args
	_ = env
	return nil
}
