package parser

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"assembler/symboltable"
)

type InstructionType int

const (
	A_INSTRUCTION InstructionType = iota
	C_INSTRUCTION
	L_INSTRUCTION
)

func isInt(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

type Parser struct {
	sc           *bufio.Scanner
	hasMoreLines bool
	lineNum      int
}

func New(r io.Reader, st *symbol.SymbolTable) *Parser {
	return &Parser{sc: bufio.NewScanner(r), hasMoreLines: true, lineNum: -1}
}

func (p *Parser) HasMoreLines() bool {
	return p.hasMoreLines
}

func (p *Parser) getLine() string {
	return strings.TrimSpace(p.sc.Text())
}

func (p *Parser) isBlankLine() bool {
	return len(p.getLine()) == 0
}

func (p *Parser) isComment() bool {
	return strings.Index(p.getLine(), "//") == 0
}

func (p *Parser) Advance() {
	p.hasMoreLines = p.sc.Scan()
	for p.HasMoreLines() && (p.isBlankLine() || p.isComment()) {
		p.hasMoreLines = p.sc.Scan()
	}

	if p.HasMoreLines() && p.InstructionType() != L_INSTRUCTION {
		p.lineNum++
	}
}

func (p *Parser) LineNum() int {
	return p.lineNum
}

func (p *Parser) InstructionType() InstructionType {
	switch p.getLine()[0] {
	case '@':
		return A_INSTRUCTION
	case '(':
		return L_INSTRUCTION
	default:
		return C_INSTRUCTION
	}
}

func (p *Parser) Symbol(st *symboltable.SymbolTable) (string, error) {
	var s string

	switch p.InstructionType() {
	case A_INSTRUCTION:
		fmt.Sscanf(p.getLine(), "@%s", &s)
		if isInt(s[0:1]) {
			if !isInt(s) {
				return "", fmt.Errorf("invalid symbol: %s", s)
			}

			return s, nil
		}

		if !st.Contains(s) {
			st.AddVar(s)
		}

		addr, _ := st.GetAddress(s)
		return strconv.Itoa(addr), nil

	case L_INSTRUCTION:
		re := regexp.MustCompile(`\((.+)\)`)
		s = re.FindStringSubmatch(p.getLine())[1]

		if isInt(s[0:1]) {
			return "", fmt.Errorf("invalid symbol: %s", s)
		}

		return s, nil
	}

	return "", nil
}

func (p *Parser) Dest() string {
	if p.InstructionType() != C_INSTRUCTION {
		return ""
	}

	t := p.getLine()
	tail := strings.Index(t, "=")

	if tail == -1 {
		return ""
	}

	return t[:tail]
}

func (p *Parser) Comp() string {
	if p.InstructionType() != C_INSTRUCTION {
		return ""
	}

	t := p.getLine()
	head := strings.Index(t, "=") + 1
	tail := strings.Index(t, ";")
	if tail == -1 {
		return t[head:]
	}

	return t[head:tail]
}

func (p *Parser) Jump() string {
	if p.InstructionType() != C_INSTRUCTION {
		return ""
	}

	t := p.getLine()
	head := strings.Index(t, ";")

	if head == -1 {
		return ""
	}
	head++

	return t[head:]
}
