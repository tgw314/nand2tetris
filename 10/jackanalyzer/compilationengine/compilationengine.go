package compilationengine

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	token "jackanalyzer/tokenizer"
)

type CompilationEngine struct {
	tok    *token.Tokenizer
	writer bufio.Writer
	depth  int
}

func New(tok *token.Tokenizer, f io.Writer) *CompilationEngine {
	ce := &CompilationEngine{
		tok:    tok,
		writer: *bufio.NewWriter(f),
		depth:  0,
	}

	return ce
}

func (ce *CompilationEngine) write(s string) {
	fmt.Fprint(&ce.writer, s)
}

func (ce *CompilationEngine) indent(s string) {
	ce.write(strings.Repeat(" ", 2*ce.depth) + s)
}

func (ce *CompilationEngine) indentln(s string) {
	ce.indent(s + "\n")
}

func escape(s string) string {
	rep := strings.NewReplacer(
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"&", "&amp;",
	)

	return rep.Replace(s)
}

func (ce *CompilationEngine) writetok(count int) {
	for range count {
		ty := string(ce.tok.TokenType())

		ce.indent("<" + ty + "> ")

		if !ce.tok.HasMoreTokens() {
			err := fmt.Errorf("no more tokens")
			log.Panic(err)
		}
		switch ce.tok.TokenType() {
		case token.KEYWORD:
			ce.write(string(ce.tok.Keyword()))
		case token.SYMBOL:
			ce.write(escape(string(ce.tok.Symbol())))
		case token.INT_CONST:
			n := ce.tok.IntVal()
			ce.write(strconv.Itoa(n))
		case token.STRING_CONST:
			ce.write(escape(ce.tok.StringVal()))
		case token.IDENTIFIER:
			ce.write(ce.tok.Identifier())
		}

		ce.write(" </" + ty + ">\n")
	}
}

func (ce *CompilationEngine) CompileClass() {
	ce.tok.Advance()

	ce.indentln("<class>")
	ce.depth++

	ce.writetok(3)
	for ce.tok.MatchKw(token.STATIC) || ce.tok.MatchKw(token.FIELD) {
		ce.CompileClassVarDec()
	}
	for ce.tok.MatchKw(token.CONSTRUCTOR) ||
		ce.tok.MatchKw(token.FUNCTION) ||
		ce.tok.MatchKw(token.METHOD) {
		ce.CompileSubroutine()
	}
	ce.writetok(1)

	ce.depth--
	ce.indentln("</class>")

	ce.writer.Flush()
}

func (ce *CompilationEngine) CompileClassVarDec() {
	ce.indentln("<classVarDec>")
	ce.depth++

	ce.writetok(3)
	for ce.tok.MatchSym(',') {
		ce.writetok(2)
	}
	ce.writetok(1)

	ce.depth--
	ce.indentln("</classVarDec>")
}

func (ce *CompilationEngine) CompileSubroutine() {
	ce.indentln("<subroutineDec>")
	ce.depth++

	ce.writetok(4)
	ce.CompileParameterList()
	ce.writetok(1)
	ce.CompileSubroutineBody()

	ce.depth--
	ce.indentln("</subroutineDec>")
}

func (ce *CompilationEngine) CompileParameterList() {
	ce.indentln("<parameterList>")
	ce.depth++

	if !ce.tok.MatchSym(')') {
		ce.writetok(2)
		for ce.tok.MatchSym(',') {
			ce.writetok(3)
		}

	}
	ce.depth--
	ce.indentln("</parameterList>")
}

func (ce *CompilationEngine) CompileSubroutineBody() {
	ce.indentln("<subroutineBody>")
	ce.depth++

	ce.writetok(1)
	for ce.tok.MatchKw(token.VAR) {
		ce.CompileVarDec()
	}
	ce.CompileStatements()
	ce.writetok(1)

	ce.depth--
	ce.indentln("</subroutineBody>")
}

func (ce *CompilationEngine) CompileVarDec() {
	ce.indentln("<varDec>")
	ce.depth++

	ce.writetok(3)
	for ce.tok.MatchSym(',') {
		ce.writetok(2)
	}
	ce.writetok(1)

	ce.depth--
	ce.indentln("</varDec>")
}

func (ce *CompilationEngine) CompileStatements() {
	ce.indentln("<statements>")
	ce.depth++

	for ce.tok.MatchKw(token.LET) ||
		ce.tok.MatchKw(token.IF) ||
		ce.tok.MatchKw(token.WHILE) ||
		ce.tok.MatchKw(token.DO) ||
		ce.tok.MatchKw(token.RETURN) {

		if ce.tok.MatchKw(token.LET) {
			ce.CompileLet()
		}

		if ce.tok.MatchKw(token.IF) {
			ce.CompileIf()
		}

		if ce.tok.MatchKw(token.WHILE) {
			ce.CompileWhile()
		}

		if ce.tok.MatchKw(token.DO) {
			ce.CompileDo()
		}

		if ce.tok.MatchKw(token.RETURN) {
			ce.CompileReturn()
		}

	}

	ce.depth--
	ce.indentln("</statements>")
}

func (ce *CompilationEngine) CompileLet() {
	ce.indentln("<letStatement>")
	ce.depth++

	ce.writetok(2)
	if ce.tok.MatchSym('[') {
		ce.writetok(1)
		ce.CompileExpression()
		ce.writetok(1)
	}
	ce.writetok(1)
	ce.CompileExpression()
	ce.writetok(1)

	ce.depth--
	ce.indentln("</letStatement>")
}

func (ce *CompilationEngine) CompileIf() {
	ce.indentln("<ifStatement>")
	ce.depth++

	ce.writetok(2)
	ce.CompileExpression()
	ce.writetok(2)
	ce.CompileStatements()
	ce.writetok(1)
	if ce.tok.MatchKw(token.ELSE) {
		ce.writetok(2)
		ce.CompileStatements()
		ce.writetok(1)
	}

	ce.depth--
	ce.indentln("</ifStatement>")
}

func (ce *CompilationEngine) CompileWhile() {
	ce.indentln("<whileStatement>")
	ce.depth++

	ce.writetok(2)
	ce.CompileExpression()
	ce.writetok(2)
	ce.CompileStatements()
	ce.writetok(1)

	ce.depth--
	ce.indentln("</whileStatement>")
}

func (ce *CompilationEngine) CompileDo() {
	ce.indentln("<doStatement>")
	ce.depth++

	ce.writetok(1)
	ce.term()
	ce.writetok(1)

	ce.depth--
	ce.indentln("</doStatement>")
}

func (ce *CompilationEngine) CompileReturn() {
	ce.indentln("<returnStatement>")
	ce.depth++

	ce.writetok(1)
	if !ce.tok.MatchSym(';') {
		ce.CompileExpression()
	}
	ce.writetok(1)

	ce.depth--
	ce.indentln("</returnStatement>")
}

func (ce *CompilationEngine) CompileExpression() {
	ce.indentln("<expression>")
	ce.depth++

	ce.CompileTerm()
	for strings.ContainsRune("+-*/&|<>=", ce.tok.PeekSym()) {
		ce.writetok(1)
		ce.CompileTerm()
	}

	ce.depth--
	ce.indentln("</expression>")
}

func (ce *CompilationEngine) CompileTerm() {
	ce.indentln("<term>")
	ce.depth++

	ce.term()

	ce.depth--
	ce.indentln("</term>")
}

func (ce *CompilationEngine) term() {
	if ce.tok.MatchSym('(') {
		ce.writetok(1)
		ce.CompileExpression()
		ce.writetok(1)
		return
	}

	if ce.tok.MatchSym('-') || ce.tok.MatchSym('~') {
		ce.writetok(1)
		ce.CompileTerm()
		return
	}

	ce.writetok(1)

	if ce.tok.MatchSym('[') {
		ce.writetok(1)
		ce.CompileExpression()
		ce.writetok(1)
		return
	}

	if ce.tok.MatchSym('(') {
		ce.writetok(1)
		ce.CompileExpressionList()
		ce.writetok(1)
		return
	}

	if ce.tok.MatchSym('.') {
		ce.writetok(3)
		ce.CompileExpressionList()
		ce.writetok(1)
		return
	}
}

func (ce *CompilationEngine) CompileExpressionList() int {
	ce.indentln("<expressionList>")
	ce.depth++

	if !ce.tok.MatchSym(')') {
		ce.CompileExpression()
		for ce.tok.MatchSym(',') {
			ce.writetok(1)
			ce.CompileExpression()
		}
	}

	ce.depth--
	ce.indentln("</expressionList>")
	return 0
}
