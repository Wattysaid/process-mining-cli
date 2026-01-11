package commands

import (
	"fmt"
	"os/exec"

	"github.com/pm-assist/pm-assist/internal/cli"
	"github.com/spf13/cobra"
)

// NewDoctorCmd returns the doctor command.
func NewDoctorCmd(global *cli.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check environment readiness",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = global
			pythonPath, pythonErr := exec.LookPath("python3")
			if pythonErr != nil {
				pythonPath, pythonErr = exec.LookPath("python")
			}
			if pythonErr != nil {
				fmt.Println("[ERROR] Python not found")
			} else {
				fmt.Printf("[SUCCESS] Python found: %s\n", pythonPath)
			}

			dotPath, dotErr := exec.LookPath("dot")
			if dotErr != nil {
				fmt.Println("[WARN] Graphviz not found (dot missing)")
			} else {
				fmt.Printf("[SUCCESS] Graphviz found: %s\n", dotPath)
			}

			if pythonErr != nil {
				return fmt.Errorf("environment check failed")
			}
			return nil
		},
		Example: "  pm-assist doctor",
	}
	return cmd
}
