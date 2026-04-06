package main

import (
	"strconv"
	"strings"
)

func lineStartsWithOperator(line string) bool {
	if len(line) == 0 {
		return false
	}
	switch line[0] {
	case '+', '-', '*', '/', '%':
		return true
	}
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
		case "SUM":
			s := 0.0
			for _, v := range prevValues {
				s += v
			}
			return FormatResult(s)
		case "AVG":
			s := 0.0
			for _, v := range prevValues {
				s += v
			}
			return FormatResult(s / float64(len(prevValues)))
		case "MIN":
			m := prevValues[0]
			for _, v := range prevValues[1:] {
				if v < m {
					m = v
				}
			}
			return FormatResult(m)
		case "MAX":
			m := prevValues[0]
			for _, v := range prevValues[1:] {
				if v > m {
					m = v
				}
			}
			return FormatResult(m)
		case "CNT":
			return FormatResult(float64(len(prevValues)))
		case "PRD":
			p := 1.0
			for _, v := range prevValues {
				p *= v
			}
			return FormatResult(p)
		case "MED":
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
		if r := aggregate(line); r != "" {
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
