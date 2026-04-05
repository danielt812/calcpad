package main

import "unicode"

type TokenType int

const (
	NUMBER TokenType = iota
	PLUS
	MINUS
	STAR
	SLASH
	PERCENT
	LPAREN
	RPAREN
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
		return Token{Type: ILLEGAL, Literal: string(char)}
	}
}
