// https://hexdocs.pm/color_palette/color_table.html#content

package main

import (
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	textarea textarea.Model
	results  []string
	width    int
	height   int
}

func evalLines(lines []string) []string {
	results := make([]string, len(lines))
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		p := NewParser(line)
		node, err := p.Parse()
		if err != nil || node == nil {
			continue
		}
		val, err := Evaluate(node)
		if err != nil {
			continue
		}
		results[i] = FormatResult(val)
	}
	return results
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
		m.textarea.SetWidth(msg.Width * 2 / 3)
		m.textarea.SetHeight(msg.Height - 1)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+r":
			m.textarea.Reset()
			m.results = nil
			return m, nil
		}
	}

	m.textarea, cmd = m.textarea.Update(msg)
	lines := strings.Split(m.textarea.Value(), "\n")
	m.results = evalLines(lines)
	return m, cmd
}

func (m model) View() string {
	inputWidth := m.width * 2 / 3
	resultWidth := m.width - inputWidth

	displayHeight := m.height - 1
	if displayHeight <= 0 {
		return ""
	}
	resultLines := make([]string, displayHeight)
	for i := 0; i < displayHeight; i++ {
		if i < len(m.results) && m.results[i] != "" {
			resultLines[i] = lipgloss.NewStyle().
				Width(resultWidth).
				Align(lipgloss.Right).
				Foreground(lipgloss.Color("84")).
				Render(m.results[i])
		} else {
			resultLines[i] = lipgloss.NewStyle().Width(resultWidth).Render("")
		}
	}
	results := strings.Join(resultLines, "\n")

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

	return lipgloss.JoinHorizontal(lipgloss.Top, m.textarea.View(), results) + "\n" + footer
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		os.Exit(1)
	}
}
