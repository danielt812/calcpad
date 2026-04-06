package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

var aggregateKeywords = map[string]bool{
	"SUM": true, "AVG": true,
	"MIN": true, "MAX": true,
	"CNT": true, "MED": true, "PRD": true,
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

var constantKeywords = map[string]bool{
	"pi": true, "tau": true, "e": true, "phi": true,
}

var funcKeywords = map[string]bool{
	"sqrt": true, "squareroot": true, "root": true,
}

func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// padCompletedWordOp triggers when a letter is typed and the word ending at
// the cursor exactly matches a word operator — mirrors the + behaviour.
func padCompletedWordOp(line string, col int) (string, int) {
	if col == 0 {
		return line, col
	}
	end := col - 1
	start := end
	for start > 0 && isLetter(line[start-1]) {
		start--
	}
	word := line[start:col]
	if !wordOperators[strings.ToLower(word)] {
		return line, col
	}
	// Add trailing space so next character lands correctly
	if col >= len(line) || line[col] != ' ' {
		line = line[:col] + " " + line[col:]
		col++
	}
	// Add space before word
	if start > 0 && line[start-1] != ' ' {
		line = line[:start] + " " + line[start:]
		col++
	}
	return line, col
}

// padWordOperator adds spaces around a word operator that was just completed.
// Triggered when a non-letter is typed and the char before it ends a word operator.
func padWordOperator(line string, col int) (string, int) {
	if col < 2 {
		return line, col
	}
	prev := col - 2 // char before what was just typed
	if !isLetter(line[prev]) {
		return line, col
	}
	end := prev
	start := end
	for start > 0 && isLetter(line[start-1]) {
		start--
	}
	word := line[start : end+1]
	if !wordOperators[strings.ToLower(word)] {
		return line, col
	}
	afterWord := end + 1
	if afterWord < len(line) && line[afterWord] != ' ' {
		line = line[:afterWord] + " " + line[afterWord:]
		col++
	}
	if start > 0 && line[start-1] != ' ' {
		line = line[:start] + " " + line[start:]
		col++
	}
	return line, col
}

// padOperator ensures a single space before and after the operator at col-1.
func padOperator(line string, col int) (string, int) {
	if col == 0 {
		return line, col
	}
	opPos := col - 1
	// Add space after operator
	afterPos := opPos + 1
	if afterPos >= len(line) || line[afterPos] != ' ' {
		line = line[:afterPos] + " " + line[afterPos:]
		col++
	}
	// Add space before operator
	if opPos > 0 && line[opPos-1] != ' ' {
		line = line[:opPos] + " " + line[opPos:]
		col++
	}
	return line, col
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
			} else if aggregateKeywords[word] {
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
