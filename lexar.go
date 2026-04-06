package main

import (
	"math"
	"strconv"
	"strings"
	"unicode"
)

type TokenType int

const (
	NUMBER TokenType = iota
	PLUS
	MINUS
	STAR
	SLASH
	PERCENT
	CARET
	LPAREN
	RPAREN
	SQRT
	EOF
	ILLEGAL
)

type Token struct {
	Type    TokenType
	Literal string
}

type Lexer struct {
	input string
	pos   int
}

func NewLexer(input string) *Lexer {
	return &Lexer{input: input}
}

func (line *Lexer) NextToken() Token {
	// Skip whitespace
	for line.pos < len(line.input) && line.input[line.pos] == ' ' {
		line.pos++
	}

	// End of input
	if line.pos >= len(line.input) {
		return Token{Type: EOF}
	}

	char := line.input[line.pos]
	line.pos++

	switch char {
	case '+':
		return Token{Type: PLUS, Literal: "+"}
	case '-':
		return Token{Type: MINUS, Literal: "-"}
	case '*', 'x', 'X':
		return Token{Type: STAR, Literal: "*"}
	case '/':
		return Token{Type: SLASH, Literal: "/"}
	case '%':
		return Token{Type: PERCENT, Literal: "%"}
	case '^':
		return Token{Type: CARET, Literal: "^"}
	case '(':
		return Token{Type: LPAREN, Literal: "("}
	case ')':
		return Token{Type: RPAREN, Literal: ")"}
	default:
		if unicode.IsDigit(rune(char)) {
			start := line.pos - 1
			for line.pos < len(line.input) &&
				(unicode.IsDigit(rune(line.input[line.pos])) || line.input[line.pos] == '.') {
				line.pos++
			}
			return Token{Type: NUMBER, Literal: line.input[start:line.pos]}
		}
		if unicode.IsLetter(rune(char)) {
			start := line.pos - 1
			for line.pos < len(line.input) && unicode.IsLetter(rune(line.input[line.pos])) {
				line.pos++
			}
			switch strings.ToLower(line.input[start:line.pos]) {
			case "plus", "add", "added":
				return Token{Type: PLUS, Literal: "+"}
			case "minus", "subtract", "subtracted", "less":
				return Token{Type: MINUS, Literal: "-"}
			case "sub", "times", "multiply", "multiplied", "mul":
				return Token{Type: STAR, Literal: "*"}
			case "div", "divide", "divided", "over":
				return Token{Type: SLASH, Literal: "/"}
			case "mod", "modulo", "remainder":
				return Token{Type: PERCENT, Literal: "%"}
			case "pi":
				return Token{Type: NUMBER, Literal: strconv.FormatFloat(math.Pi, 'f', -1, 64)}
			case "tau":
				return Token{Type: NUMBER, Literal: strconv.FormatFloat(math.Pi*2, 'f', -1, 64)}
			case "e":
				return Token{Type: NUMBER, Literal: strconv.FormatFloat(math.E, 'f', -1, 64)}
			case "phi":
				return Token{Type: NUMBER, Literal: strconv.FormatFloat(math.Phi, 'f', -1, 64)}
			case "sqrt", "squareroot", "root":
				return Token{Type: SQRT, Literal: "sqrt"}
			}
		}
		return Token{Type: ILLEGAL, Literal: string(char)}
	}
}
