package interactive

import (
	"fmt"
	"os"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Menu model for Bubble Tea interactive menu
type Menu struct {
	cursor    int
	choices   []string
	selected  map[int]struct{}
	quitting  bool
	spinner   spinner.Model
	loading   bool
	statusMsg string
}

// Initial model state
func initialMenuModel() Menu {
	s := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("205"))),
	)

	return Menu{
		cursor:    0,
		choices:   []string{"Ver versión actual", "Listar JDKs", "Cambiar JDK", "Instalar JDK", "Escanear sistema", "Configurar PATH", "Salir"},
		selected:  make(map[int]struct{}),
		quitting:  false,
		spinner:   s,
		loading:   false,
		statusMsg: "Bienvenido a jem (Java Environment Manager)",
	}
}

// Init returns an initial command for the application to run
func (m Menu) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages
func (m Menu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
			// Handle selection immediately for menu actions
			return m, tea.Batch(m.handleSelection(), tea.Quit)
		}
	}
	return m, nil
}

// Handle selection and return appropriate command
func (m Menu) handleSelection() tea.Cmd {
	switch m.cursor {
	case 0: // Ver versión actual
		return tea.Batch(
			tea.Printf("Ver versión actual"),
			m.runAction("current"),
		)
	case 1: // Listar JDKs
		return tea.Batch(
			tea.Printf("Listar JDKs"),
			m.runAction("list"),
		)
	case 2: // Cambiar JDK
		return tea.Batch(
			tea.Printf("Cambiar JDK"),
			m.runAction("use"),
		)
	case 3: // Instalar JDK
		return tea.Batch(
			tea.Printf("Instalar JDK"),
			m.runAction("install"),
		)
	case 4: // Escanear sistema
		return tea.Batch(
			tea.Printf("Escanear sistema"),
			m.runAction("scan"),
		)
	case 5: // Configurar PATH
		return tea.Batch(
			tea.Printf("Configurar PATH"),
			m.runAction("setup"),
		)
	case 6: // Salir
		return tea.Quit
	}
	return nil
}

// Run action by calling the appropriate command
func (m Menu) runAction(action string) tea.Cmd {
	return func() tea.Msg {
		// In a real implementation, this would call the actual command
		// For now, just return a message with the action
		return actionMessage(action)
	}
}

type actionMessage string

func (a actionMessage) String() string {
	return string(a)
}

// View renders the menu UI
func (m Menu) View() tea.View {
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

	// Menu title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Underline(true).
		MarginBottom(1)
	s += titleStyle.Render("¿Qué quieres hacer?") + "\n\n"

	// Menu items
	for i, choice := range m.choices {
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Selected state
		checked := " "
		if _, ok := m.selected[i]; ok {
			checked = "✓"
		}

		itemStyle := lipgloss.NewStyle().
			PaddingLeft(2).
			MarginRight(2)

		if m.cursor == i {
			itemStyle = itemStyle.
				Foreground(lipgloss.Color("205")).
				Bold(true)
		}

		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, itemStyle.Render(choice))
	}

	// Help text
	helpStyle := lipgloss.NewStyle().
		MarginTop(2).
		Foreground(lipgloss.Color("240")).
		Italic(true)

	s += "\n" + helpStyle.Render("Usa las flechas para navegar, Enter para seleccionar, q para salir") + "\n"

	return tea.NewView(s)
}

// Run the interactive menu
func RunMenu() error {
	p := tea.NewProgram(initialMenuModel())
	_, err := p.Run()
	return err
}

// ShowSpinner starts a spinner for long operations
func ShowSpinner(message string) tea.Cmd {
	return func() tea.Msg {
		// In a real implementation, this would control the spinner
		return spinnerMessage{message: message}
	}
}

type spinnerMessage struct {
	message string
}

func (s spinnerMessage) String() string {
	return s.message
}

// ShowProgress starts a progress bar for downloads
func ShowProgress(total, current int64, message string) tea.Cmd {
	return func() tea.Msg {
		return progressMessage{
			total:   total,
			current: current,
			message: message,
		}
	}
}

type progressMessage struct {
	total   int64
	current int64
	message string
}

func (p progressMessage) Percent() float64 {
	if p.total == 0 {
		return 0
	}
	return float64(p.current) / float64(p.total)
}

func (p progressMessage) String() string {
	return fmt.Sprintf("%s: %.1f%%", p.message, p.Percent()*100)
}

// ShowError displays an error message with styling
func ShowError(message string) {
	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)

	fmt.Fprintf(os.Stderr, "\n%s %s\n\n",
		errorStyle.Render("✗ Error:"), message)
}

// ShowSuccess displays a success message
func ShowSuccess(message string) {
	successStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")).
		Bold(true)

	fmt.Fprintf(os.Stdout, "\n%s %s\n\n",
		successStyle.Render("✓ Éxito:"), message)
}

// ShowWarning displays a warning message
func ShowWarning(message string) {
	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Bold(true)

	fmt.Fprintf(os.Stderr, "\n%s %s\n\n",
		warningStyle.Render("⚠ Advertencia:"), message)
}

// ShowInfo displays an info message
func ShowInfo(message string) {
	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("33")).
		Bold(true)

	fmt.Fprintf(os.Stdout, "\n%s %s\n\n",
		infoStyle.Render("ℹ Información:"), message)
}

// Confirm asks for user confirmation
func Confirm(message string) (bool, error) {
	fmt.Printf("\n%s ", message)
	fmt.Print("[y/N] ")

	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		return false, err
	}

	switch response {
	case "y", "Y", "yes", "Yes":
		return true, nil
	default:
		return false, nil
	}
}

// Select presents a selection prompt
func Select(message string, options []string) (string, error) {
	fmt.Printf("\n%s\n", message)

	for i, option := range options {
		fmt.Printf("  %d. %s\n", i+1, option)
	}

	fmt.Printf("\nSelección [1-%d]: ", len(options))

	var choice int
	_, err := fmt.Scanln(&choice)
	if err != nil {
		return "", err
	}

	if choice < 1 || choice > len(options) {
		return "", fmt.Errorf("opción inválida: %d", choice)
	}

	return options[choice-1], nil
}

// Input prompts for user input
func Input(message string, defaultValue string) (string, error) {
	fmt.Printf("\n%s ", message)
	if defaultValue != "" {
		fmt.Printf("[%s] ", defaultValue)
	}

	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		return "", err
	}

	if response == "" && defaultValue != "" {
		return defaultValue, nil
	}

	return response, nil
}
