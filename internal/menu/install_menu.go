package menu

import (
	"fmt"

	"charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// InstallableVersion represents an installable JDK version
type InstallableVersion struct {
	Version   string
	Major     int
	LTS       bool
	Available bool
}

// InstallModel represents the install version selection menu state
type InstallModel struct {
	cursor    int
	versions  []InstallableVersion
	filterLTS bool
	quitting  bool
	statusMsg string
	width     int
	height    int
}

// NewInstallModel creates a new install version selection model
func NewInstallModel(versions []InstallableVersion, statusMsg string) InstallModel {
	return InstallModel{
		cursor:    0,
		versions:  versions,
		filterLTS: false,
		quitting:  false,
		statusMsg: statusMsg,
		width:     80,
		height:    24,
	}
}

// Init returns an initial command for the application to run
func (m InstallModel) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages for install version selection
func (m InstallModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.cursor < len(m.versions)-1 {
				m.cursor++
			}
		case "enter":
			return m, tea.Quit
		case "l":
			m.filterLTS = !m.filterLTS
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

// View renders the install version selection UI
func (m InstallModel) View() tea.View {
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
	s += titleStyle.Render("Select a version to install") + "\n\n"

	// Filter indicator
	filterStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	filterText := "Filter: All versions"
	if m.filterLTS {
		filterText = "Filter: LTS only [l to toggle]"
	} else {
		filterText = "Filter: All versions [l to toggle]"
	}
	s += filterStyle.Render(filterText) + "\n\n"

	// Version items
	for i, version := range m.versions {
		// Apply LTS filter
		if m.filterLTS && !version.LTS {
			continue
		}

		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		versionInfo := version.Version
		if version.LTS {
			versionInfo += " [LTS]"
		}

		itemStyle := lipgloss.NewStyle().
			PaddingLeft(2).
			MarginRight(2)

		if m.cursor == i {
			itemStyle = itemStyle.
				Foreground(lipgloss.Color("205")).
				Bold(true)
		}

		s += fmt.Sprintf("%s %s\n", cursor, itemStyle.Render(versionInfo))
	}

	// Help text
	helpStyle := lipgloss.NewStyle().
		MarginTop(2).
		Foreground(lipgloss.Color("240")).
		Italic(true)

	s += "\n" + helpStyle.Render("Use arrow keys to navigate, Enter to select, q to quit, l to toggle LTS filter") + "\n"

	return tea.NewView(s)
}

// RunInstallMenu runs the install version selection menu
func RunInstallMenu(versions []InstallableVersion, statusMsg string) (*InstallableVersion, error) {
	if len(versions) == 0 {
		return nil, fmt.Errorf("no versions available")
	}

	p := tea.NewProgram(NewInstallModel(versions, statusMsg))
	model, err := p.Run()
	if err != nil {
		return nil, err
	}

	if m, ok := model.(InstallModel); ok {
		if m.quitting && m.cursor >= 0 && m.cursor < len(m.versions) {
			return &m.versions[m.cursor], nil
		}
	}

	return nil, fmt.Errorf("no version selected")
}
