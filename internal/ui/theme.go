package ui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Primary   lipgloss.Color
	Muted     lipgloss.Color
	Success   lipgloss.Color
	Warning   lipgloss.Color
	Error     lipgloss.Color
	Border    lipgloss.Color
	Highlight lipgloss.Color
}

var defaultTheme = Theme{
	Primary:   lipgloss.Color("33"),
	Muted:     lipgloss.Color("245"),
	Success:   lipgloss.Color("35"),
	Warning:   lipgloss.Color("214"),
	Error:     lipgloss.Color("160"),
	Border:    lipgloss.Color("240"),
	Highlight: lipgloss.Color("39"),
}

func ThemeDefault() Theme {
	return defaultTheme
}
