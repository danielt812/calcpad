// https://hexdocs.pm/color_palette/color_table.html#content

package main

import (
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	editor   *Editor
	results  []string
	width    int
	height   int
	showHelp bool
	cfg      Config
}

func initialModel(cfg Config) model {
	return model{editor: NewEditor(), cfg: cfg}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.editor.Width = msg.Width * 2 / 3
		m.editor.Height = msg.Height - 1

	case tea.KeyMsg:
		switch k := msg.String(); {
		case k == "esc" || k == m.cfg.Keys.Quit:
			if m.showHelp {
				m.showHelp = false
				return m, nil
			}
			return m, tea.Quit

		case k == m.cfg.Keys.Help:
			m.showHelp = !m.showHelp
			return m, nil

		case k == m.cfg.Keys.Format:
			for i, line := range m.editor.lines {
				formatted := formatLine(line)
				if formatted != line {
					if i == m.editor.row {
						m.editor.col = adjustCursor(line, formatted, m.editor.col)
					}
					m.editor.lines[i] = formatted
				}
			}
			m.results = evalLines(m.editor.Lines())
			return m, nil

		case k == m.cfg.Keys.Reset:
			m.editor.Reset()
			m.results = nil
			return m, nil
		}
		if m.showHelp {
			return m, nil
		}
		m.editor.Update(msg)
		if autoFormat {
			row := m.editor.row
			if msg.Type == tea.KeyRunes && strings.ContainsAny(string(msg.Runes), "+-*/%^") {
				m.editor.lines[row], m.editor.col = padOperator(m.editor.lines[row], m.editor.col)
			} else if msg.Type == tea.KeyRunes && isLetter(string(msg.Runes)[0]) {
				// Trigger immediately when a letter completes a word operator
				m.editor.lines[row], m.editor.col = padCompletedWordOp(m.editor.lines[row], m.editor.col)
			} else if msg.Type == tea.KeyRunes || msg.Type == tea.KeySpace {
				// Fallback: non-letter typed after a word operator
				m.editor.lines[row], m.editor.col = padWordOperator(m.editor.lines[row], m.editor.col)
			}
		}
		m.results = evalLines(m.editor.Lines())
	}

	return m, nil
}

func (m model) View() string {
	if m.height <= 0 {
		return ""
	}

	if m.showHelp {
		return m.helpView()
	}

	inputWidth := m.width * 2 / 3
	resultWidth := m.width - inputWidth
	displayHeight := m.height - 1

	resultLines := make([]string, displayHeight)
	for i := range displayHeight {
		lineIdx := m.editor.offset + i
		if lineIdx < len(m.results) && m.results[lineIdx] != "" {
			resultLines[i] = lipgloss.NewStyle().
				Foreground(lipgloss.Color(resultColor)).
				Width(resultWidth).
				Align(lipgloss.Right).
				Render(m.results[lineIdx])
		} else {
			resultLines[i] = lipgloss.NewStyle().Width(resultWidth).Render("")
		}
	}

	hint := func(key, desc string) string {
		display := strings.ReplaceAll(key, "ctrl+", "^")
		k := lipgloss.NewStyle().Background(lipgloss.Color("235")).Foreground(lipgloss.Color(resultColor)).PaddingRight(1).Render(display)
		d := lipgloss.NewStyle().Background(lipgloss.Color("235")).Foreground(lipgloss.Color("245")).PaddingRight(1).Render(desc)
		return k + d
	}
	sep := lipgloss.NewStyle().Background(lipgloss.Color("235")).Foreground(lipgloss.Color("15")).PaddingRight(1).Render("│")

	footer := lipgloss.NewStyle().
		Width(m.width).
		PaddingLeft(1).
		Background(lipgloss.Color("235")).
		Render(func() string {
			hints := []string{
				hint(m.cfg.Keys.Quit, "quit"),
				hint(m.cfg.Keys.Reset, "reset"),
			}
			if !m.cfg.AutoFormat {
				hints = append(hints, hint(m.cfg.Keys.Format, "format"))
			}
			hints = append(hints, hint(m.cfg.Keys.Help, "help"))
			return strings.Join(hints, sep)
		}())

	editorView := lipgloss.NewStyle().Width(inputWidth).Render(m.editor.View())
	return lipgloss.JoinHorizontal(lipgloss.Top, editorView, strings.Join(resultLines, "\n")) + "\n" + footer
}

func main() {
	cfg := loadConfig()
	initStyles(cfg)
	p := tea.NewProgram(initialModel(cfg), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		os.Exit(1)
	}
}
