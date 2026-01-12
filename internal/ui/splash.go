package ui

import (
	"fmt"
	"strings"

	"github.com/pm-assist/pm-assist/internal/buildinfo"
	"github.com/pm-assist/pm-assist/internal/config"
)

type SplashOptions struct {
	CompletedCommand string
	WorkingDir       string
}

func PrintSplash(cfg *config.Config, opts SplashOptions) {
	printBanner()
	printInfoPanel(cfg, opts.WorkingDir)
	printCommandHelp(opts.CompletedCommand)
}

func printBanner() {
	banner := `  ____  __  __        ___              __
 |  _ \|  \/  |      / _ \ ___ ___ ___/ _\___
 | |_) | |\/| |_____/ /_)/ __/ __/ _ \ \ / __|
 |  __/| |  | |_____/ ___/ (_| (_|  __/\ \__ \
 |_|   |_|  |_|      \/    \___\___\___\__/___/

 PM Assist · Enterprise Process Mining CLI
 -----------------------------------------`
	fmt.Println(banner)
}

func printInfoPanel(cfg *config.Config, cwd string) {
	llmProvider := "none"
	llmModel := "n/a"
	if cfg != nil {
		if cfg.LLM.Provider != "" {
			llmProvider = cfg.LLM.Provider
		}
		if cfg.LLM.Model != "" {
			llmModel = cfg.LLM.Model
		}
	}
	lines := []string{
		fmt.Sprintf(">_ PM Assist CLI (%s)", buildinfo.Version),
		fmt.Sprintf("model:     %s / %s", llmProvider, llmModel),
		fmt.Sprintf("directory: %s", cwd),
	}
	printBox(lines)
}

func printCommandHelp(completed string) {
	fmt.Println()
	fmt.Println("To get started, choose one of these commands:")
	fmt.Println()
	commands := []string{
		"init        - create a new project scaffold",
		"connect     - register a data source",
		"ingest      - ingest and normalize data",
		"prepare     - run data quality and cleaning",
		"mine        - run discovery and analysis",
		"report      - generate reports and bundles",
		"review      - run QA checks",
		"agent setup - configure LLM integration",
		"doctor      - check environment readiness",
	}
	for _, cmd := range commands {
		prefix := "[ ]"
		if completed != "" && strings.HasPrefix(cmd, completed) {
			prefix = "[OK]"
		}
		fmt.Printf("  %s %s\n", prefix, cmd)
	}
	fmt.Println()
}

func printBox(lines []string) {
	maxLen := 0
	for _, line := range lines {
		if len(line) > maxLen {
			maxLen = len(line)
		}
	}
	width := maxLen + 2
	border := "+" + strings.Repeat("-", width) + "+"
	fmt.Println(border)
	for _, line := range lines {
		padding := width - len(line)
		fmt.Printf("| %s%s|\n", line, strings.Repeat(" ", padding-1))
	}
	fmt.Println(border)
}
