// https://hexdocs.pm/color_palette/color_table.html#content

package main

import (
	os "os"

	textarea "github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
)

type model struct {
	textarea textarea.Model
	width    int
	height   int
}

// https://deepwiki.com/charmbracelet/bubbles/2.2-text-area
func initialModel() model {
	ta := textarea.New()

	// Remove default border
	ta.Prompt = ""

	// Clear default styling of CursorLine
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	// Line Number always grey
	ta.FocusedStyle.LineNumber = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	ta.BlurredStyle.LineNumber = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	ta.FocusedStyle.CursorLineNumber = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	ta.BlurredStyle.CursorLineNumber = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	// Text always white
	ta.FocusedStyle.Text = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	ta.BlurredStyle.Text = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))

	ta.FocusedStyle.Base = lipgloss.NewStyle()
	ta.BlurredStyle.Base = lipgloss.NewStyle()

	ta.Focus()

	return model{
		textarea: ta,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textarea.SetWidth(msg.Width)
		m.textarea.SetHeight(msg.Height - 1)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m model) View() string {
	hint := func(key, desc string) string {
		k := lipgloss.NewStyle().
			Background(lipgloss.Color("235")).
			Foreground(lipgloss.Color("84")).
			PaddingRight(1).
			Render(key)
		d := lipgloss.NewStyle().
			Background(lipgloss.Color("235")).
			Foreground(lipgloss.Color("245")).
			PaddingRight(1).
			Render(desc)
		s := lipgloss.NewStyle().
			Background(lipgloss.Color("235")).
			Foreground(lipgloss.Color("15")).
			PaddingRight(1).
			Render("│")
		return k + d + s
	}

	footer := lipgloss.NewStyle().
		Width(m.width).
		PaddingLeft(2).
		Background(lipgloss.Color("235")).
		Render(hint("^c", "quit") + hint("^r", "reset"))

	return m.textarea.View() + "\n" + footer
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		os.Exit(1)
	}
}
