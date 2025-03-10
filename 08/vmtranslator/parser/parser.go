package parser

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type CommandType int

const (
	C_ARITHMETIC CommandType = iota
	C_PUSH
	C_POP
	C_LABEL
	C_GOTO
	C_IF
	C_FUNCTION
	C_RETURN
	C_CALL
)

type Parser struct {
	sc           *bufio.Scanner
	toks         []string
	hasMoreLines bool
	lineNumber   int
}

func New(r io.Reader) *Parser {
	return &Parser{sc: bufio.NewScanner(r), hasMoreLines: true}
}

func (p *Parser) HasMoreLines() bool {
	return p.hasMoreLines
}

func (p *Parser) getLine() string {
	return strings.TrimSpace(p.sc.Text())
}

func (p *Parser) getTokens() []string {
	return p.toks
}

func (p *Parser) isBlankLine() bool {
	return len(p.getLine()) == 0
}

func (p *Parser) isComment() bool {
	return strings.Index(p.getLine(), "//") == 0
}

func (p *Parser) LineNumber() int {
	return p.lineNumber
}

func (p *Parser) Advance() {
	p.hasMoreLines = p.sc.Scan()
	p.lineNumber++
	for p.HasMoreLines() && (p.isBlankLine() || p.isComment()) {
		p.hasMoreLines = p.sc.Scan()
		p.lineNumber++
	}

	p.toks = strings.Fields(p.getLine())
}

func (p *Parser) CommandType() CommandType {
	cmd := p.getTokens()[0]
	switch cmd {
	case "add", "sub", "neg", "eq", "gt", "lt", "and", "or", "not":
		return C_ARITHMETIC
	case "push":
		return C_PUSH
	case "pop":
		return C_POP
	case "label":
		return C_LABEL
	case "goto":
		return C_GOTO
	case "if-goto":
		return C_IF
	case "function":
		return C_FUNCTION
	case "return":
		return C_RETURN
	case "call":
		return C_CALL
	default:
		return -1
	}
}

func (p *Parser) Arg1() (string, error) {
	ty := p.CommandType()

	if ty == C_ARITHMETIC {
		return p.getTokens()[0], nil
	}

	if ty == C_RETURN {
		return "", fmt.Errorf("error getting arg1: `return` command doesn't have arg1")
	}

	return p.getTokens()[1], nil
}

func (p *Parser) Arg2() (int, error) {
	ty := p.CommandType()

	switch ty {
	case C_PUSH, C_POP, C_FUNCTION, C_CALL:
		{
			n, err := strconv.Atoi(p.getTokens()[2])
			if err != nil {
				return 0, fmt.Errorf("error getting arg2: %w", err)
			}
			return n, nil
		}
	default:
		return 0, fmt.Errorf("error getting arg2: `%s` command doesn't have arg2", p.getTokens()[0])
	}
}
