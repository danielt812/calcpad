package main

import "github.com/charmbracelet/lipgloss"

var (
	normalStyle     lipgloss.Style
	operStyle       lipgloss.Style
	parenStyle      lipgloss.Style
	aggregateStyle  lipgloss.Style
	constantStyle   lipgloss.Style
	funcStyle       lipgloss.Style
	lineNumStyle    lipgloss.Style
	cursorStyle     = lipgloss.NewStyle().Reverse(true)
	resultColor     string
	resultPrecision int
	autoFormat      bool
)

func initStyles(cfg Config) {
	c := cfg.Colors
	resultPrecision = cfg.Precision
	autoFormat = cfg.AutoFormat
	normalStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(c.Normal))
	operStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(c.Operators))
	parenStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(c.Parens))
	aggregateStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(c.Aggregates))
	constantStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(c.Constants))
	funcStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(c.Functions))
	lineNumStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(c.LineNumbers))
	resultColor = c.Results
}
