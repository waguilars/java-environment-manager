package menu

import (
	"fmt"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// MenuItem represents a menu option
type MenuItem struct {
	Title    string
	Action   string
	Disabled bool
}

// Model represents the main menu state
type Model struct {
	cursor    int
	items     []MenuItem
	selected  map[int]struct{}
	quitting  bool
	spinner   spinner.Model
	loading   bool
	statusMsg string
	width     int
	height    int
}

// Initial model state
func NewModel() Model {
	s := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("205"))),
	)

	items := []MenuItem{
		{Title: "Setup - Initialize jem configuration", Action: "setup"},
		{Title: "Scan - Scan for JDKs on your system", Action: "scan"},
		{Title: "List - List installed and detected JDKs", Action: "list"},
		{Title: "Current - Show the currently active JDK", Action: "current"},
		{Title: "Use - Switch to a different JDK version", Action: "use"},
		{Title: "Install - Install a new JDK version", Action: "install"},
		{Title: "Exit", Action: "exit"},
	}

	return Model{
		cursor:    0,
		items:     items,
		selected:  make(map[int]struct{}),
		quitting:  false,
		spinner:   s,
		loading:   false,
		statusMsg: "Welcome to jem (Java Environment Manager)",
		width:     80,
		height:    24,
	}
}

// Init returns an initial command for the application to run
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter":
			if m.items[m.cursor].Action == "exit" {
				m.quitting = true
				return m, tea.Quit
			}
			return m, tea.Batch(m.handleSelection(), tea.Quit)
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

// Handle selection and return appropriate command
func (m Model) handleSelection() tea.Cmd {
	action := m.items[m.cursor].Action
	return tea.Batch(
		tea.Printf("action:%s", action),
		tea.Quit,
	)
}

// View renders the menu UI
func (m Model) View() tea.View {
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

	// Menu title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Underline(true).
		MarginBottom(1)
	s += titleStyle.Render("What would you like to do?") + "\n\n"

	// Menu items
	for i, item := range m.items {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		itemStyle := lipgloss.NewStyle().
			PaddingLeft(2).
			MarginRight(2)

		if m.cursor == i {
			itemStyle = itemStyle.
				Foreground(lipgloss.Color("205")).
				Bold(true)
		}

		if item.Disabled {
			itemStyle = itemStyle.Foreground(lipgloss.Color("240"))
		}

		s += fmt.Sprintf("%s %s\n", cursor, itemStyle.Render(item.Title))
	}

	// Help text
	helpStyle := lipgloss.NewStyle().
		MarginTop(2).
		Foreground(lipgloss.Color("240")).
		Italic(true)

	s += "\n" + helpStyle.Render("Use arrow keys to navigate, Enter to select, q to quit") + "\n"

	return tea.NewView(s)
}

// Run the interactive menu
func Run() error {
	p := tea.NewProgram(NewModel())
	_, err := p.Run()
	return err
}

// IsInteractive checks if terminal is interactive
func IsInteractive() bool {
	return true
}
