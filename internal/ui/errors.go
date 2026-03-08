package ui

import (
	"fmt"
	"os"

	"charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// ErrorModel represents an error message with styling
type ErrorModel struct {
	err      error
	hint     string
	width    int
	height   int
	quitting bool
}

// NewErrorModel creates a new error model
func NewErrorModel(err error, hint string) ErrorModel {
	return ErrorModel{
		err:    err,
		hint:   hint,
		width:  80,
		height: 24,
	}
}

// Init returns an initial command for the application to run
func (m ErrorModel) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages
func (m ErrorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			m.quitting = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

// View renders the error UI
func (m ErrorModel) View() tea.View {
	if m.quitting {
		return tea.NewView("\n")
	}

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)

	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Italic(true)

	s := "\n"
	s += errorStyle.Render("✗ Error:") + " " + m.err.Error() + "\n\n"

	if m.hint != "" {
		s += hintStyle.Render("Hint: "+m.hint) + "\n\n"
	}

	s += "Press q or ctrl+c to close\n"

	return tea.NewView(s)
}

// ShowError displays an error message with styling
func ShowError(message string) {
	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)

	fmt.Fprintf(os.Stderr, "\n%s %s\n\n",
		errorStyle.Render("✗ Error:"), message)
}

// ShowErrorWithHint displays an error message with a hint
func ShowErrorWithHint(message, hint string) {
	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)

	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Italic(true)

	fmt.Fprintf(os.Stderr, "\n%s %s\n\n",
		errorStyle.Render("✗ Error:"), message)

	if hint != "" {
		fmt.Fprintf(os.Stderr, "%s\n\n",
			hintStyle.Render("Hint: "+hint))
	}
}

// ShowSuccess displays a success message
func ShowSuccess(message string) {
	successStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")).
		Bold(true)

	fmt.Fprintf(os.Stdout, "\n%s %s\n\n",
		successStyle.Render("✓ Success:"), message)
}

// ShowWarning displays a warning message
func ShowWarning(message string) {
	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Bold(true)

	fmt.Fprintf(os.Stderr, "\n%s %s\n\n",
		warningStyle.Render("⚠ Warning:"), message)
}

// ShowInfo displays an info message
func ShowInfo(message string) {
	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("33")).
		Bold(true)

	fmt.Fprintf(os.Stdout, "\n%s %s\n\n",
		infoStyle.Render("ℹ Info:"), message)
}

// FormatError formats an error with helpful hints
func FormatError(err error, hints ...string) string {
	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)

	s := errorStyle.Render("Error: " + err.Error())

	for _, hint := range hints {
		if hint != "" {
			hintStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")).
				Italic(true)
			s += "\n" + hintStyle.Render("Hint: "+hint)
		}
	}

	return s
}

// FormatSuccess formats a success message
func FormatSuccess(message string) string {
	successStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")).
		Bold(true)

	return successStyle.Render("✓ " + message)
}

// FormatWarning formats a warning message
func FormatWarning(message string) string {
	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Bold(true)

	return warningStyle.Render("⚠ " + message)
}

// FormatInfo formats an info message
func FormatInfo(message string) string {
	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("33")).
		Bold(true)

	return infoStyle.Render("ℹ " + message)
}
