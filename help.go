package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

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
	constRows := []row{
		{"PI", "3.14159… (π)"},
		{"TAU", "6.28318… (2π)"},
		{"E", "2.71828… (Euler's number)"},
		{"PHI", "1.61803… (golden ratio)"},
	}

	grey := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	wordSt := normalStyle.Render

	const colW = 10
	hdrCol := grey.Width(colW)
	opCol := lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Width(colW)
	aggCol := lipgloss.NewStyle().Foreground(lipgloss.Color("141")).Width(colW)
	constCol := lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Width(colW)

	section := func(label, rightLabel string) (string, string) {
		return "  " + hdrCol.Render(label) + grey.Render(rightLabel),
			"  " + hdrCol.Render(strings.Repeat("─", len(label))) + grey.Render(strings.Repeat("─", len(rightLabel)))
	}

	var lines []string
	h, s := section("operator", "words")
	lines = append(lines, h, s)
	for _, r := range operatorRows {
		lines = append(lines, "  "+opCol.Render(r.operator)+wordSt(r.words))
	}
	lines = append(lines, "")
	h, s = section("keyword", "description")
	lines = append(lines, h, s)
	for _, r := range aggregateRows {
		lines = append(lines, "  "+aggCol.Render(r.operator)+wordSt(r.words))
	}
	lines = append(lines, "")
	h, s = section("const", "value")
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
