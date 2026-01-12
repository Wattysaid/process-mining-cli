package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type TextPrompt struct {
	Question string
	Default  string
	Required bool
}

func AskTextInput(prompt TextPrompt) (string, error) {
	m := newTextInputModel(prompt)
	p := tea.NewProgram(m)
	final, err := p.Run()
	if err != nil {
		return "", err
	}
	model := final.(textInputModel)
	if prompt.Required && strings.TrimSpace(model.value) == "" {
		return "", fmt.Errorf("missing required input")
	}
	if strings.TrimSpace(model.value) == "" {
		return prompt.Default, nil
	}
	return model.value, nil
}

func AskTextArea(question string, defaultValue string) (string, error) {
	m := newTextAreaModel(question, defaultValue)
	p := tea.NewProgram(m)
	final, err := p.Run()
	if err != nil {
		return "", err
	}
	model := final.(textAreaModel)
	value := strings.TrimSpace(model.value)
	if value == "" {
		return defaultValue, nil
	}
	return value, nil
}

func AskChoice(question string, options []string, defaultValue string) (string, error) {
	items := make([]list.Item, 0, len(options))
	defaultIndex := 0
	for i, option := range options {
		if strings.EqualFold(option, defaultValue) {
			defaultIndex = i
		}
		items = append(items, listItem{title: option})
	}
	m := newListModel(question, items, defaultIndex)
	p := tea.NewProgram(m)
	final, err := p.Run()
	if err != nil {
		return "", err
	}
	model := final.(listModel)
	if model.choice == "" {
		return defaultValue, nil
	}
	return model.choice, nil
}

func AskConfirm(question string, defaultValue bool) (bool, error) {
	choices := []string{"yes", "no"}
	def := "no"
	if defaultValue {
		def = "yes"
	}
	choice, err := AskChoice(question, choices, def)
	if err != nil {
		return false, err
	}
	return choice == "yes", nil
}

func SelectFile(startDir string) (string, error) {
	m := newFilePickerModel(startDir, false)
	p := tea.NewProgram(m)
	final, err := p.Run()
	if err != nil {
		return "", err
	}
	model := final.(filePickerModel)
	return model.path, nil
}

func SelectDirectory(startDir string) (string, error) {
	m := newFilePickerModel(startDir, true)
	p := tea.NewProgram(m)
	final, err := p.Run()
	if err != nil {
		return "", err
	}
	model := final.(filePickerModel)
	return model.path, nil
}

func RenderTable(headers []string, rows [][]string) error {
	columns := make([]table.Column, 0, len(headers))
	for _, header := range headers {
		columns = append(columns, table.Column{Title: header, Width: len(header) + 2})
	}
	tableRows := make([]table.Row, 0, len(rows))
	for _, row := range rows {
		tableRows = append(tableRows, row)
	}
	t := table.New(table.WithColumns(columns), table.WithRows(tableRows))
	p := tea.NewProgram(tableModel{table: t})
	_, err := p.Run()
	return err
}

func RunWithSpinner(message string, fn func() error) error {
	done := make(chan error, 1)
	go func() {
		done <- fn()
	}()
	m := newSpinnerModel(message, done)
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}

func RenderProgress(value float64) string {
	p := progress.New(progress.WithGradient("#00A0FF", "#00D7AF"))
	return p.ViewAs(value)
}

type listItem struct {
	title string
}

func (i listItem) Title() string       { return i.title }
func (i listItem) Description() string { return "" }
func (i listItem) FilterValue() string { return i.title }

type listModel struct {
	list   list.Model
	choice string
}

func newListModel(title string, items []list.Item, index int) listModel {
	theme := ThemeDefault()
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Foreground(theme.Highlight)
	l := list.New(items, delegate, 0, 8)
	l.Title = title
	l.Select(index)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowPagination(false)
	return listModel{list: l}
}

func (m listModel) Init() tea.Cmd {
	return nil
}

func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if item, ok := m.list.SelectedItem().(listItem); ok {
				m.choice = item.title
				return m, tea.Quit
			}
		case "q", "esc":
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m listModel) View() string {
	return m.list.View()
}

type textInputModel struct {
	input textinput.Model
	value string
}

func newTextInputModel(prompt TextPrompt) textInputModel {
	ti := textinput.New()
	ti.Placeholder = prompt.Default
	ti.Prompt = fmt.Sprintf("%s: ", prompt.Question)
	ti.Focus()
	return textInputModel{input: ti}
}

func (m textInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m textInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.value = m.input.Value()
			return m, tea.Quit
		case "esc":
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m textInputModel) View() string {
	return m.input.View()
}

type textAreaModel struct {
	area  textarea.Model
	value string
}

func newTextAreaModel(question string, defaultValue string) textAreaModel {
	ta := textarea.New()
	ta.Placeholder = defaultValue
	ta.SetValue(defaultValue)
	ta.Focus()
	ta.CharLimit = 0
	ta.Prompt = question + "\n"
	return textAreaModel{area: ta}
}

func (m textAreaModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m textAreaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" && msg.Alt {
			m.value = m.area.Value()
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.area, cmd = m.area.Update(msg)
	return m, cmd
}

func (m textAreaModel) View() string {
	return m.area.View() + "\n(Alt+Enter to submit)"
}

type filePickerModel struct {
	picker filepicker.Model
	path   string
}

func newFilePickerModel(startDir string, directory bool) filePickerModel {
	fp := filepicker.New()
	fp.CurrentDirectory = startDir
	fp.DirAllowed = directory
	fp.FileAllowed = !directory
	fp.ShowHidden = false
	return filePickerModel{picker: fp}
}

func (m filePickerModel) Init() tea.Cmd {
	return m.picker.Init()
}

func (m filePickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.picker, cmd = m.picker.Update(msg)
	if ok, path := m.picker.DidSelectFile(msg); ok {
		m.path = path
		return m, tea.Quit
	}
	return m, cmd
}

func (m filePickerModel) View() string {
	return m.picker.View()
}

type tableModel struct {
	table table.Model
}

func (m tableModel) Init() tea.Cmd {
	return nil
}

func (m tableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "esc" || msg.String() == "enter" {
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m tableModel) View() string {
	return m.table.View() + "\nPress Enter to continue."
}

type spinnerModel struct {
	spinner spinner.Model
	message string
	done    chan error
	err     error
}

func newSpinnerModel(message string, done chan error) spinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Line
	return spinnerModel{spinner: s, message: message, done: done}
}

func (m spinnerModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, waitForDone(m.done))
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case doneMsg:
		m.err = msg.err
		return m, tea.Quit
	}
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m spinnerModel) View() string {
	if m.err != nil {
		return ""
	}
	return fmt.Sprintf("%s %s", m.spinner.View(), m.message)
}

type doneMsg struct {
	err error
}

func waitForDone(done chan error) tea.Cmd {
	return func() tea.Msg {
		return doneMsg{err: <-done}
	}
}
