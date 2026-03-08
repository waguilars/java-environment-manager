package ui

import (
	"time"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// SpinnerModel wraps the spinner component
type SpinnerModel struct {
	spinner spinner.Model
	message string
}

// NewSpinner creates a new spinner model
func NewSpinner(message string) SpinnerModel {
	s := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("205"))),
	)

	return SpinnerModel{
		spinner: s,
		message: message,
	}
}

// Init returns an initial command for the application to run
func (m SpinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update handles incoming messages
func (m SpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

// View renders the spinner UI
func (m SpinnerModel) View() tea.View {
	s := "\n"
	s += m.spinner.View() + " " + m.message + "\n\n"
	return tea.NewView(s)
}

// RunSpinner runs a spinner for a given duration or until cancelled
func RunSpinner(message string, duration time.Duration) error {
	p := tea.NewProgram(NewSpinner(message))

	if duration > 0 {
		go func() {
			time.Sleep(duration)
			p.Send(tea.Quit)
		}()
	}

	_, err := p.Run()
	return err
}

// SpinnerCmd creates a command that runs a spinner
func SpinnerCmd(message string) tea.Cmd {
	return func() tea.Msg {
		return spinnerMessage{message: message}
	}
}

type spinnerMessage struct {
	message string
}

func (s spinnerMessage) String() string {
	return s.message
}
