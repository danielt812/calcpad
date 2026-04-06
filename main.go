// https://hexdocs.pm/color_palette/color_table.html#content

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

// ── Config ────────────────────────────────────────────────────────────────────

type ColorConfig struct {
	Normal      string `yaml:"normal"`
	Operators   string `yaml:"operators"`
	Parens      string `yaml:"parens"`
	Aggregates  string `yaml:"aggregates"`
	Constants   string `yaml:"constants"`
	Functions   string `yaml:"functions"`
	LineNumbers string `yaml:"line_numbers"`
	Results     string `yaml:"results"`
}

type KeyConfig struct {
	Quit   string `yaml:"quit"`
	Reset  string `yaml:"reset"`
	Format string `yaml:"format"`
	Help   string `yaml:"help"`
}

type Config struct {
	Colors     ColorConfig `yaml:"colors"`
	Keys       KeyConfig   `yaml:"keys"`
	Precision  int         `yaml:"precision"`   // decimal places, -1 = full precision
	AutoFormat bool        `yaml:"auto_format"` // format lines as you type
}

func defaultConfig() Config {
	return Config{
		Precision:  -1,
		AutoFormat: false,
		Colors: ColorConfig{
			Normal:      "15",
			Operators:   "14",
			Parens:      "214",
			Aggregates:  "141",
			Constants:   "220",
			Functions:   "213",
			LineNumbers: "240",
			Results:     "84",
		},
		Keys: KeyConfig{
			Quit:   "ctrl+c",
			Reset:  "ctrl+r",
			Format: "ctrl+f",
			Help:   "ctrl+h",
		},
	}
}

func loadConfig() Config {
	cfg := defaultConfig()
	home, err := os.UserHomeDir()
	if err != nil {
		return cfg
	}
	data, err := os.ReadFile(filepath.Join(home, ".config", "calcpad", "config.yaml"))
	if err != nil {
		return cfg
	}
	// Unmarshal into the populated default — only keys present in the file are overwritten.
	yaml.Unmarshal(data, &cfg) //nolint:errcheck
	return cfg
}

// ── Styles ────────────────────────────────────────────────────────────────────

var (
	normalStyle    lipgloss.Style
	operStyle      lipgloss.Style
	parenStyle     lipgloss.Style
	aggregateStyle lipgloss.Style
	constantStyle  lipgloss.Style
	funcStyle      lipgloss.Style
	lineNumStyle   lipgloss.Style
	cursorStyle      = lipgloss.NewStyle().Reverse(true)
	resultColor      string
	resultPrecision  int
	autoFormat       bool
)

func initStyles(cfg Config) {
	c := cfg.Colors
	resultPrecision = cfg.Precision
	autoFormat = cfg.AutoFormat
	normalStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(c.Normal))
	operStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color(c.Operators))
	parenStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color(c.Parens))
	aggregateStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(c.Aggregates))
	constantStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color(c.Constants))
	funcStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color(c.Functions))
	lineNumStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(c.LineNumbers))
	resultColor    = c.Results
}

// Editor is a minimal multi-line editor with syntax highlighting support.
type Editor struct {
	lines  []string
	row    int
	col    int
	offset int // vertical scroll
	Width  int
	Height int
}

func NewEditor() *Editor {
	return &Editor{lines: []string{""}}
}

func (e *Editor) Reset() {
	e.lines = []string{""}
	e.row, e.col, e.offset = 0, 0, 0
}

func (e *Editor) Value() string   { return strings.Join(e.lines, "\n") }
func (e *Editor) Lines() []string { return e.lines }

// formatLine tokenizes s and re-joins tokens with single spaces.

var constantKeywords = map[string]bool{
	"pi": true, "tau": true, "e": true, "phi": true,
}

var funcKeywords = map[string]bool{
	"sqrt": true, "squareroot": true, "root": true,
}

var aggregateKeywords = map[string]bool{
	"sum": true, "avg": true, "average": true,
	"min": true, "max": true,
	"cnt": true, "med": true, "prd": true,
}

var wordOperators = map[string]bool{
	// addition
	"plus": true, "add": true, "added": true,
	// subtraction
	"sub": true, "minus": true, "subtract": true, "subtracted": true, "less": true,
	// multiplication
	"mul": true, "multiply": true, "multiplied": true, "times": true,
	// division
	"div": true, "divide": true, "divided": true, "over": true,
	// modulo
	"mod": true, "modulo": true, "remainder": true,
}

func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func formatLine(s string) string {
	lx := NewLexer(s)
	var tokens []Token
	for {
		tok := lx.NextToken()
		if tok.Type == EOF {
			break
		}
		if tok.Type == ILLEGAL {
			return s
		}
		tokens = append(tokens, tok)
	}
	if len(tokens) == 0 {
		return s
	}
	var sb strings.Builder
	for i, tok := range tokens {
		if i > 0 && tokens[i-1].Type != LPAREN && tok.Type != RPAREN {
			sb.WriteByte(' ')
		}
		sb.WriteString(tok.Literal)
	}
	return sb.String()
}

func adjustCursor(original, formatted string, col int) int {
	count := 0
	for i := 0; i < col && i < len(original); i++ {
		if original[i] != ' ' {
			count++
		}
	}
	seen := 0
	for i := 0; i < len(formatted); i++ {
		if seen == count {
			return i
		}
		if formatted[i] != ' ' {
			seen++
		}
	}
	return len(formatted)
}

func highlightLine(s string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		ch := s[i]
		switch {
		case ch == '+' || ch == '-' || ch == '*' || ch == '/' || ch == '%' || ch == '^':
			b.WriteString(operStyle.Render(string(ch)))
			i++
		case ch == '(' || ch == ')':
			b.WriteString(parenStyle.Render(string(ch)))
			i++
		case isLetter(ch):
			start := i
			for i < len(s) && isLetter(s[i]) {
				i++
			}
			word := s[start:i]
			lower := strings.ToLower(word)
			if wordOperators[lower] {
				b.WriteString(operStyle.Render(word))
			} else if aggregateKeywords[lower] {
				b.WriteString(aggregateStyle.Render(word))
			} else if constantKeywords[lower] {
				b.WriteString(constantStyle.Render(word))
			} else if funcKeywords[lower] {
				b.WriteString(funcStyle.Render(word))
			} else {
				b.WriteString(normalStyle.Render(word))
			}
		default:
			b.WriteString(normalStyle.Render(string(ch)))
			i++
		}
	}
	return b.String()
}

func (e *Editor) Update(msg tea.KeyMsg) {
	switch msg.Type {
	case tea.KeyRunes, tea.KeySpace:
		ins := string(msg.Runes)
		if msg.Type == tea.KeySpace {
			ins = " "
		}
		line := e.lines[e.row]
		e.lines[e.row] = line[:e.col] + ins + line[e.col:]
		e.col += len(ins)

	case tea.KeyBackspace:
		if e.col > 0 {
			line := e.lines[e.row]
			e.lines[e.row] = line[:e.col-1] + line[e.col:]
			e.col--
		} else if e.row > 0 {
			prevLen := len(e.lines[e.row-1])
			e.lines[e.row-1] += e.lines[e.row]
			e.lines = append(e.lines[:e.row], e.lines[e.row+1:]...)
			e.row--
			e.col = prevLen
			if e.row < e.offset {
				e.offset = e.row
			}
		}

	case tea.KeyDelete:
		line := e.lines[e.row]
		if e.col < len(line) {
			e.lines[e.row] = line[:e.col] + line[e.col+1:]
		} else if e.row < len(e.lines)-1 {
			e.lines[e.row] += e.lines[e.row+1]
			e.lines = append(e.lines[:e.row+1], e.lines[e.row+2:]...)
		}

	case tea.KeyEnter:
		line := e.lines[e.row]
		next := make([]string, len(e.lines)+1)
		copy(next, e.lines[:e.row])
		next[e.row] = line[:e.col]
		next[e.row+1] = line[e.col:]
		copy(next[e.row+2:], e.lines[e.row+1:])
		e.lines = next
		e.row++
		e.col = 0
		if e.row >= e.offset+e.Height {
			e.offset++
		}

	case tea.KeyLeft:
		if e.col > 0 {
			e.col--
		} else if e.row > 0 {
			e.row--
			e.col = len(e.lines[e.row])
			if e.row < e.offset {
				e.offset = e.row
			}
		}

	case tea.KeyRight:
		if e.col < len(e.lines[e.row]) {
			e.col++
		} else if e.row < len(e.lines)-1 {
			e.row++
			e.col = 0
			if e.row >= e.offset+e.Height {
				e.offset = e.row - e.Height + 1
			}
		}

	case tea.KeyUp:
		if e.row > 0 {
			e.row--
			if e.col > len(e.lines[e.row]) {
				e.col = len(e.lines[e.row])
			}
			if e.row < e.offset {
				e.offset = e.row
			}
		}

	case tea.KeyDown:
		if e.row < len(e.lines)-1 {
			e.row++
			if e.col > len(e.lines[e.row]) {
				e.col = len(e.lines[e.row])
			}
			if e.row >= e.offset+e.Height {
				e.offset = e.row - e.Height + 1
			}
		}

	case tea.KeyHome, tea.KeyCtrlA:
		e.col = 0

	case tea.KeyEnd, tea.KeyCtrlE:
		e.col = len(e.lines[e.row])
	}

}

func (e *Editor) View() string {
	if e.Height <= 0 {
		return ""
	}
	var sb strings.Builder
	for i := 0; i < e.Height; i++ {
		if i > 0 {
			sb.WriteByte('\n')
		}
		lineIdx := e.offset + i
		if lineIdx >= len(e.lines) {
			sb.WriteString(strings.Repeat(" ", 5))
			continue
		}
		if e.lines[lineIdx] != "" || lineIdx == e.row {
			sb.WriteString(lineNumStyle.Render(fmt.Sprintf("%3d  ", lineIdx+1)))
		} else {
			sb.WriteString(strings.Repeat(" ", 5))
		}
		line := e.lines[lineIdx]
		if lineIdx == e.row {
			sb.WriteString(highlightLine(line[:e.col]))
			if e.col < len(line) {
				sb.WriteString(cursorStyle.Render(string(line[e.col])))
				sb.WriteString(highlightLine(line[e.col+1:]))
			} else {
				sb.WriteString(cursorStyle.Render(" "))
			}
		} else {
			sb.WriteString(highlightLine(line))
		}
	}
	return sb.String()
}

type model struct {
	editor   *Editor
	results  []string
	width    int
	height   int
	showHelp bool
	cfg      Config
}

func lineStartsWithOperator(line string) bool {
	if len(line) == 0 {
		return false
	}
	switch line[0] {
	case '+', '-', '*', '/', '%':
		return true
	}
	// Check for word operator at start
	i := 0
	for i < len(line) && isLetter(line[i]) {
		i++
	}
	if i > 0 {
		return wordOperators[strings.ToLower(line[:i])]
	}
	return false
}

func evalLines(lines []string) []string {
	results := make([]string, len(lines))
	var prevValues []float64

	aggregate := func(keyword string) string {
		if len(prevValues) == 0 {
			return ""
		}
		switch keyword {
		case "sum":
			s := 0.0
			for _, v := range prevValues {
				s += v
			}
			return FormatResult(s)
		case "avg", "average":
			s := 0.0
			for _, v := range prevValues {
				s += v
			}
			return FormatResult(s / float64(len(prevValues)))
		case "min":
			m := prevValues[0]
			for _, v := range prevValues[1:] {
				if v < m {
					m = v
				}
			}
			return FormatResult(m)
		case "max":
			m := prevValues[0]
			for _, v := range prevValues[1:] {
				if v > m {
					m = v
				}
			}
			return FormatResult(m)
		case "cnt":
			return FormatResult(float64(len(prevValues)))
		case "prd":
			p := 1.0
			for _, v := range prevValues {
				p *= v
			}
			return FormatResult(p)
		case "med":
			sorted := make([]float64, len(prevValues))
			copy(sorted, prevValues)
			for i := 1; i < len(sorted); i++ {
				for j := i; j > 0 && sorted[j] < sorted[j-1]; j-- {
					sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
				}
			}
			n := len(sorted)
			if n%2 == 0 {
				return FormatResult((sorted[n/2-1] + sorted[n/2]) / 2)
			}
			return FormatResult(sorted[n/2])
		}
		return ""
	}

	var lastVal float64

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if r := aggregate(strings.ToLower(line)); r != "" {
			results[i] = r
			if v, err := strconv.ParseFloat(r, 64); err == nil {
				prevValues = append(prevValues, v)
				lastVal = v
			}
			continue
		}
		exprLine := line
		if lineStartsWithOperator(line) {
			exprLine = FormatResult(lastVal) + " " + line
		}
		p := NewParser(exprLine)
		node, err := p.Parse()
		if err != nil || node == nil {
			continue
		}
		val, err := Evaluate(node)
		if err != nil {
			continue
		}
		results[i] = FormatResult(val)
		prevValues = append(prevValues, val)
		lastVal = val
	}
	return results
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
		k := msg.String()
		switch {
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
			line := m.editor.lines[row]
			if formatted := formatLine(line); formatted != line {
				m.editor.col = adjustCursor(line, formatted, m.editor.col)
				m.editor.lines[row] = formatted
			}
		}
		m.results = evalLines(m.editor.Lines())
	}

	return m, nil
}

func (m model) helpView() string {
	type row struct{ operator, words string }
	operatorRows := []row{
		{"+", "plus, add, added"},
		{"-", "minus, sub, subtract, subtracted, less"},
		{"*", "times, mul, multiply, multiplied"},
		{"/", "div, divide, divided, over"},
		{"%", "mod, modulo, remainder"},
	}
	aggregateRows := []row{
		{"SUM", "sum of results"},
		{"PRD", "product of results"},
		{"AVG", "average of results"},
		{"CNT", "count of results"},
		{"MIN", "minimum value"},
		{"MAX", "maximum value"},
		{"MED", "median value"},
	}

	grey := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	wordSt := normalStyle.Render

	const colW = 10
	hdrCol := grey.Width(colW)
	opCol := lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Width(colW)
	aggCol := lipgloss.NewStyle().Foreground(lipgloss.Color("141")).Width(colW)
	constCol := lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Width(colW)

	section := func(label, sep, rightLabel string) (string, string) {
		return "  " + hdrCol.Render(label) + grey.Render(rightLabel),
			"  " + hdrCol.Render(strings.Repeat("─", len(label))) + grey.Render(strings.Repeat("─", len(rightLabel)))
	}

	constRows := []row{
		{"PI", "3.14159… (π)"},
		{"TAU", "6.28318… (2π)"},
		{"E", "2.71828… (Euler's number)"},
		{"PHI", "1.61803… (golden ratio)"},
	}

	var lines []string
	h, s := section("operator", "", "words")
	lines = append(lines, h, s)
	for _, r := range operatorRows {
		lines = append(lines, "  "+opCol.Render(r.operator)+wordSt(r.words))
	}
	lines = append(lines, "")
	h, s = section("keyword", "", "description")
	lines = append(lines, h, s)
	for _, r := range aggregateRows {
		lines = append(lines, "  "+aggCol.Render(r.operator)+wordSt(r.words))
	}
	lines = append(lines, "")
	h, s = section("const", "", "value")
	lines = append(lines, h, s)
	for _, r := range constRows {
		lines = append(lines, "  "+constCol.Render(r.operator)+wordSt(r.words))
	}
	lines = append(lines, "")
	lines = append(lines, grey.Render("  esc or ctrl+h to close"))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2).
		Render(strings.Join(lines, "\n"))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
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
	for i := 0; i < displayHeight; i++ {
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
		Render(strings.Join([]string{
			hint(m.cfg.Keys.Quit, "quit"),
			hint(m.cfg.Keys.Reset, "reset"),
			hint(m.cfg.Keys.Format, "format"),
			hint(m.cfg.Keys.Help, "help"),
		}, sep))

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
