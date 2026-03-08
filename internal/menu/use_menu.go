package menu

import (
	"fmt"

	"charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// JDKInfo represents a JDK version with its details
type JDKInfo struct {
	Name     string
	Version  string
	Provider string
	Managed  bool
	Path     string
}

// UseModel represents the JDK selection menu state
type UseModel struct {
	cursor    int
	jdkList   []JDKInfo
	quitting  bool
	statusMsg string
	width     int
	height    int
}

// NewUseModel creates a new use menu model
func NewUseModel(jdkList []JDKInfo, statusMsg string) UseModel {
	return UseModel{
		cursor:    0,
		jdkList:   jdkList,
		quitting:  false,
		statusMsg: statusMsg,
		width:     80,
		height:    24,
	}
}

// Init returns an initial command for the application to run
func (m UseModel) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages for JDK selection
func (m UseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.jdkList)-1 {
				m.cursor++
			}
		case "enter":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

// View renders the JDK selection UI
func (m UseModel) View() tea.View {
	if m.quitting {
		return tea.NewView("\nGoodbye!\n\n")
	}

	s := "\n"

	// Status message
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		MarginBottom(1)
	s += statusStyle.Render(m.statusMsg) + "\n\n"

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Underline(true).
		MarginBottom(1)
	s += titleStyle.Render("Select a JDK") + "\n\n"

	// JDK items
	for i, jdk := range m.jdkList {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		managed := " "
		if jdk.Managed {
			managed = "✓"
		}

		itemStyle := lipgloss.NewStyle().
			PaddingLeft(2).
			MarginRight(2)

		if m.cursor == i {
			itemStyle = itemStyle.
				Foreground(lipgloss.Color("205")).
				Bold(true)
		}

		jdkInfo := fmt.Sprintf("%s (%s)", jdk.Name, jdk.Version)
		if !jdk.Managed {
			jdkInfo += fmt.Sprintf(" [external: %s]", jdk.Path)
		}

		s += fmt.Sprintf("%s [%s] %s\n", cursor, managed, itemStyle.Render(jdkInfo))
	}

	// Help text
	helpStyle := lipgloss.NewStyle().
		MarginTop(2).
		Foreground(lipgloss.Color("240")).
		Italic(true)

	s += "\n" + helpStyle.Render("Use arrow keys to navigate, Enter to select, q to quit") + "\n"

	return tea.NewView(s)
}

// RunUseMenu runs the JDK selection menu and returns the selected JDK
func RunUseMenu(jdkList []JDKInfo, statusMsg string) (*JDKInfo, error) {
	if len(jdkList) == 0 {
		return nil, fmt.Errorf("no JDKs available")
	}

	p := tea.NewProgram(NewUseModel(jdkList, statusMsg))
	model, err := p.Run()
	if err != nil {
		return nil, err
	}

	if m, ok := model.(UseModel); ok {
		if m.quitting && m.cursor >= 0 && m.cursor < len(m.jdkList) {
			return &m.jdkList[m.cursor], nil
		}
	}

	return nil, fmt.Errorf("no JDK selected")
}
