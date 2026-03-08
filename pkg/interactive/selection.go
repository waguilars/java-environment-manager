package interactive

import (
	"fmt"

	"charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// JDKSelectionModel for selecting a JDK version
type JDKSelectionModel struct {
	cursor    int
	choices   []JDKInfo
	quitting  bool
	statusMsg string
}

// JDKInfo represents a JDK version with its details
type JDKInfo struct {
	Name     string
	Version  string
	Provider string
	Managed  bool
	Path     string
}

// Initial JDK selection model
func initialJDKSelectionModel(jdkList []JDKInfo, statusMsg string) JDKSelectionModel {
	return JDKSelectionModel{
		cursor:    0,
		choices:   jdkList,
		quitting:  false,
		statusMsg: statusMsg,
	}
}

// Init returns an initial command for the application to run
func (m JDKSelectionModel) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages for JDK selection
func (m JDKSelectionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter":
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the JDK selection UI
func (m JDKSelectionModel) View() tea.View {
	if m.quitting {
		return tea.NewView("\n¡Hasta luego!\n\n")
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
	s += titleStyle.Render("Selecciona un JDK") + "\n\n"

	// JDK items
	for i, jdk := range m.choices {
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
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

	s += "\n" + helpStyle.Render("Usa las flechas para navegar, Enter para seleccionar, q para salir") + "\n"

	return tea.NewView(s)
}

// RunJDKSelectionMenu runs the JDK selection menu and returns the selected JDK
func RunJDKSelectionMenu(jdkList []JDKInfo, statusMsg string) (*JDKInfo, error) {
	if len(jdkList) == 0 {
		return nil, fmt.Errorf("no JDKs available")
	}

	p := tea.NewProgram(initialJDKSelectionModel(jdkList, statusMsg))
	model, err := p.Run()
	if err != nil {
		return nil, err
	}

	if m, ok := model.(JDKSelectionModel); ok {
		if m.quitting && m.cursor >= 0 && m.cursor < len(m.choices) {
			return &m.choices[m.cursor], nil
		}
	}

	return nil, fmt.Errorf("no JDK selected")
}

// InstallVersionModel for selecting an installable version
type InstallVersionModel struct {
	cursor    int
	choices   []InstallableVersion
	quitting  bool
	statusMsg string
}

// InstallableVersion represents an installable JDK version
type InstallableVersion struct {
	Version   string
	Major     int
	LTS       bool
	Available bool
}

// Initial install version selection model
func initialInstallVersionModel(versions []InstallableVersion, statusMsg string) InstallVersionModel {
	return InstallVersionModel{
		cursor:    0,
		choices:   versions,
		quitting:  false,
		statusMsg: statusMsg,
	}
}

// Init returns an initial command for the application to run
func (m InstallVersionModel) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages for install version selection
func (m InstallVersionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter":
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the install version selection UI
func (m InstallVersionModel) View() tea.View {
	if m.quitting {
		return tea.NewView("\n¡Hasta luego!\n\n")
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
	s += titleStyle.Render("Selecciona una versión para instalar") + "\n\n"

	// Version items
	for i, version := range m.choices {
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
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

	s += "\n" + helpStyle.Render("Usa las flechas para navegar, Enter para seleccionar, q para salir") + "\n"

	return tea.NewView(s)
}

// RunInstallVersionMenu runs the install version selection menu
func RunInstallVersionMenu(versions []InstallableVersion, statusMsg string) (*InstallableVersion, error) {
	if len(versions) == 0 {
		return nil, fmt.Errorf("no versions available")
	}

	p := tea.NewProgram(initialInstallVersionModel(versions, statusMsg))
	model, err := p.Run()
	if err != nil {
		return nil, err
	}

	if m, ok := model.(InstallVersionModel); ok {
		if m.quitting && m.cursor >= 0 && m.cursor < len(m.choices) {
			return &m.choices[m.cursor], nil
		}
	}

	return nil, fmt.Errorf("no version selected")
}
