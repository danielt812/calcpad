package main

import (
	"fmt"
	"math"
	"strconv"
)

type Node interface {
	nodeType()
}

type NumberNode struct {
	Value float64
}

type BinaryNode struct {
	Left     Node
	Operator TokenType
	Right    Node
}

type UnaryNode struct {
	Operator TokenType
	Operand  Node
}

func (n *NumberNode) nodeType() {}
func (n *BinaryNode) nodeType() {}
func (n *UnaryNode) nodeType()  {}

type Parser struct {
	lexer   *Lexer
	current Token
}

func NewParser(input string) *Parser {
	p := &Parser{lexer: NewLexer(input)}
	p.current = p.lexer.NextToken()
	return p
}

func (p *Parser) advance() {
	p.current = p.lexer.NextToken()
}

// parsePrimary: number | '(' expr ')'
func (p *Parser) parsePrimary() (Node, error) {
	if p.current.Type == NUMBER {
		value, err := strconv.ParseFloat(p.current.Literal, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", p.current.Literal)
		}
		p.advance()
		return &NumberNode{Value: value}, nil
	}
	if p.current.Type == LPAREN {
		p.advance()
		node, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if p.current.Type != RPAREN {
			return nil, fmt.Errorf("expected closing parenthesis")
		}
		p.advance()
		return node, nil
	}
	return nil, fmt.Errorf("unexpected token: %q", p.current.Literal)
}

// parseUnary: ['-'] primary
func (p *Parser) parseUnary() (Node, error) {
	if p.current.Type == MINUS {
		p.advance()
		operand, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		return &UnaryNode{Operator: MINUS, Operand: operand}, nil
	}
	return p.parsePrimary()
}

// parseTerm: unary (('*' | '/' | '%') unary)*
func (p *Parser) parseTerm() (Node, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for p.current.Type == STAR || p.current.Type == SLASH || p.current.Type == PERCENT {
		op := p.current.Type
		p.advance()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &BinaryNode{Left: left, Operator: op, Right: right}
	}
	return left, nil
}

// parseExpr: term (('+' | '-') term)*
func (p *Parser) parseExpr() (Node, error) {
	left, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	for p.current.Type == PLUS || p.current.Type == MINUS {
		op := p.current.Type
		p.advance()
		right, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		left = &BinaryNode{Left: left, Operator: op, Right: right}
	}
	return left, nil
}

// Parse returns the AST for the full input, or nil for empty input.
func (p *Parser) Parse() (Node, error) {
	if p.current.Type == EOF {
		return nil, nil
	}
	node, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if p.current.Type != EOF {
		return nil, fmt.Errorf("unexpected token: %q", p.current.Literal)
	}
	return node, nil
}

// Evaluate walks the AST and returns the numeric result.
func Evaluate(node Node) (float64, error) {
	switch n := node.(type) {
	case *NumberNode:
		return n.Value, nil
	case *UnaryNode:
		val, err := Evaluate(n.Operand)
		if err != nil {
			return 0, err
		}
		if n.Operator == MINUS {
			return -val, nil
		}
		return val, nil
	case *BinaryNode:
		left, err := Evaluate(n.Left)
		if err != nil {
			return 0, err
		}
		right, err := Evaluate(n.Right)
		if err != nil {
			return 0, err
		}
		switch n.Operator {
		case PLUS:
			return left + right, nil
		case MINUS:
			return left - right, nil
		case STAR:
			return left * right, nil
		case SLASH:
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return left / right, nil
		case PERCENT:
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return math.Mod(left, right), nil
		}
	}
	return 0, fmt.Errorf("unknown node type")
}

// FormatResult formats a float64 for display, using integer form when exact.
func FormatResult(v float64) string {
	if v == math.Trunc(v) && !math.IsInf(v, 0) {
		return strconv.FormatInt(int64(v), 10)
	}
	return strconv.FormatFloat(v, 'f', 10, 64)
}
