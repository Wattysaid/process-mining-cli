package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type CommandFrame struct {
	Title        string
	Purpose      string
	StepIndex    int
	StepTotal    int
	Writes       []string
	Asks         []string
	Next         string
	CompletedCmd string
}

func PrintCommandStart(frame CommandFrame) {
	theme := ThemeDefault()
	header := lipgloss.NewStyle().Bold(true).Foreground(theme.Primary).Render(frame.Title)
	fmt.Println(header)
	if frame.Purpose != "" {
		fmt.Printf("Purpose: %s\n", frame.Purpose)
	}
	if frame.StepTotal > 0 {
		fmt.Printf("Step: %d/%d\n", frame.StepIndex, frame.StepTotal)
	}
	if len(frame.Writes) > 0 {
		fmt.Printf("Writes: %s\n", strings.Join(frame.Writes, ", "))
	}
	if len(frame.Asks) > 0 {
		fmt.Printf("Prompts: %s\n", strings.Join(frame.Asks, ", "))
	}
	fmt.Println(lipgloss.NewStyle().Foreground(theme.Border).Render(strings.Repeat("-", 56)))
}

func PrintCommandEnd(frame CommandFrame, success bool) {
	theme := ThemeDefault()
	status := "DONE"
	if !success {
		status = "FAILED"
	}
	color := theme.Success
	if !success {
		color = theme.Error
	}
	label := lipgloss.NewStyle().Bold(true).Foreground(color).Render(status)
	fmt.Printf("%s: %s\n", label, frame.Title)
	if frame.Next != "" {
		fmt.Printf("Next: %s\n", frame.Next)
	}
	fmt.Println(lipgloss.NewStyle().Foreground(theme.Border).Render(strings.Repeat("-", 56)))
}
