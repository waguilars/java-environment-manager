package ui

import (
	"fmt"

	"charm.land/bubbletea/v2"
	"github.com/schollz/progressbar/v3"
)

// ProgressBarModel wraps the progress bar
type ProgressBarModel struct {
	bar     *progressbar.ProgressBar
	message string
}

// NewProgressBar creates a new progress bar model
func NewProgressBar(total int64, message string) ProgressBarModel {
	bar := progressbar.NewOptions64(
		total,
		progressbar.OptionSetDescription(message),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(40),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Println()
		}),
	)

	return ProgressBarModel{
		bar:     bar,
		message: message,
	}
}

// Update updates the progress bar with new progress
func (m *ProgressBarModel) Update(current int64) {
	m.bar.Set64(current)
}

// Finish finishes the progress bar
func (m *ProgressBarModel) Finish() {
	m.bar.Finish()
}

// View renders the progress bar UI
func (m ProgressBarModel) View() tea.View {
	s := "\n" + m.message + "\n"
	s += m.bar.String() + "\n\n"
	return tea.NewView(s)
}

// ProgressBarCmd creates a command that updates progress
func ProgressBarCmd(model *ProgressBarModel, current, total int64) tea.Cmd {
	return func() tea.Msg {
		model.Update(current)
		return progressMessage{
			current: current,
			total:   total,
		}
	}
}

type progressMessage struct {
	current int64
	total   int64
}

func (p progressMessage) Percent() float64 {
	if p.total == 0 {
		return 0
	}
	return float64(p.current) / float64(p.total) * 100
}

func (p progressMessage) String() string {
	return fmt.Sprintf("%.1f%%", p.Percent())
}

// RunProgressBar runs a progress bar for a download
func RunProgressBar(total int64, message string) *ProgressBarModel {
	return &ProgressBarModel{
		bar: progressbar.NewOptions64(
			total,
			progressbar.OptionSetDescription(message),
			progressbar.OptionShowBytes(true),
			progressbar.OptionSetWidth(40),
			progressbar.OptionEnableColorCodes(true),
			progressbar.OptionShowCount(),
			progressbar.OptionOnCompletion(func() {
				fmt.Println()
			}),
		),
		message: message,
	}
}
