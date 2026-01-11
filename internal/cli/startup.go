package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pm-assist/pm-assist/internal/cli/prompt"
	"github.com/spf13/cobra"
)

func runStartup(cmdRoot *cobra.Command) error {
	printBanner()
	printStatus()
	if !hasProjectConfig() {
		confirm, err := prompt.AskBool("It looks like this is your first time using PM Assist. Continue setup?", true)
		if err != nil {
			return err
		}
		if confirm {
			return dispatchCommand(cmdRoot, "init")
		}
	}
	fmt.Println()
	fmt.Println("What would you like to do?")
	fmt.Println()
	fmt.Println("  1) Start a new process mining project")
	fmt.Println("  2) Continue an existing project")
	fmt.Println("  3) Run environment diagnostics (doctor)")
	fmt.Println("  4) Configure LLM integration")
	fmt.Println("  5) Manage user or business profiles")
	fmt.Println("  6) Exit")
	fmt.Println()
	choice, err := prompt.AskChoice("Select an option", []string{"1", "2", "3", "4", "5", "6"}, "1", true)
	if err != nil {
		return err
	}

	switch choice {
	case "1":
		return dispatchCommand(cmdRoot, "init")
	case "2":
		fmt.Println("[INFO] Continue project flow is not implemented yet.")
		return nil
	case "3":
		return dispatchCommand(cmdRoot, "doctor")
	case "4":
		return dispatchCommand(cmdRoot, "agent", "setup")
	case "5":
		manage, err := prompt.AskChoice("Manage which profile", []string{"user", "business"}, "user", true)
		if err != nil {
			return err
		}
		if manage == "business" {
			return dispatchCommand(cmdRoot, "business", "init")
		}
		return dispatchCommand(cmdRoot, "profile", "init")
	case "6":
		fmt.Println("Goodbye.")
		return nil
	default:
		return fmt.Errorf("invalid selection")
	}
}

func dispatchCommand(cmdRoot *cobra.Command, name string, args ...string) error {
	for _, cmd := range cmdRoot.Commands() {
		if cmd.Name() != name {
			continue
		}
		if len(args) > 0 {
			return dispatchSubcommand(cmd, args[0], args[1:]...)
		}
		if run := cmd.RunE; run != nil {
			return run(cmd, args)
		}
		return nil
	}
	return fmt.Errorf("command not found: %s", name)
}

func dispatchSubcommand(parent *cobra.Command, name string, args ...string) error {
	for _, cmd := range parent.Commands() {
		if cmd.Name() != name {
			continue
		}
		if run := cmd.RunE; run != nil {
			return run(cmd, args)
		}
		return nil
	}
	return fmt.Errorf("subcommand not found: %s", name)
}

func printBanner() {
	banner := `            ____  __  __        ___              __
  (\_/)    |  _ \|  \/  |      / _ \ ___ ___ ___/ _\___
  ( •_•)   | |_) | |\/| |_____/ /_)/ __/ __/ _ \ \ / __|
   />[_]   |  __/| |  | |_____/ ___/ (_| (_|  __/\ \__ \
           |_|   |_|  |_|      \/    \___\___\___\__/___/

            PM Assist · Enterprise Process Mining CLI
            -----------------------------------------`
	fmt.Println(banner)
}

func printStatus() {
	pythonStatus := "missing"
	if _, err := lookupPath("python3"); err == nil {
		pythonStatus = "ready"
	} else if _, err := lookupPath("python"); err == nil {
		pythonStatus = "ready"
	}
	llmStatus := "not configured"
	if hasEnv("OPENAI_API_KEY") || hasEnv("ANTHROPIC_API_KEY") || hasEnv("GEMINI_API_KEY") || hasEnv("GOOGLE_API_KEY") || hasEnv("OLLAMA_HOST") {
		llmStatus = "configured"
	}
	graphvizStatus := "missing"
	if _, err := lookupPath("dot"); err == nil {
		graphvizStatus = "ready"
	}
	fmt.Printf("\nVersion: 0.1.0 | Python: %s | LLM: %s | Graphviz: %s\n", pythonStatus, llmStatus, graphvizStatus)

	if hasProjectConfig() {
		fmt.Println("[INFO] Project detected in current directory.")
	}
}

func hasProjectConfig() bool {
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(cwd, "pm-assist.yaml"))
	return err == nil
}

func hasEnv(key string) bool {
	return os.Getenv(key) != ""
}

func lookupPath(binary string) (string, error) {
	return execLookPath(binary)
}

var execLookPath = func(binary string) (string, error) {
	return exec.LookPath(binary)
}
