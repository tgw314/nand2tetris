package compilationengine

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	symbt "jackcompiler/symboltable"
	token "jackcompiler/tokenizer"
)

type CompilationEngine struct {
	tok          *token.Tokenizer
	writer       *bufio.Writer
	depth        int
	classSt      *symbt.SymbolTable
	subroutineSt *symbt.SymbolTable
}

func NewCompilationEngine(tok *token.Tokenizer, f io.Writer) *CompilationEngine {
	ce := &CompilationEngine{
		tok:          tok,
		writer:       bufio.NewWriter(f),
		depth:        0,
		classSt:      symbt.NewSymbolTable(),
		subroutineSt: symbt.NewSymbolTable(),
	}

	return ce
}

func (ce *CompilationEngine) write(s string) {
	fmt.Fprint(ce.writer, s)
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

	ce.tok.Advance()

	s := fmt.Sprintf("<name> %s </name>", ce.tok.Identifier())
	ce.indentln(s)
	s = fmt.Sprintf("<category> %s </category>", "class")
	ce.indentln(s)
	ce.indentln("<usage> false </usage>")

	ce.tok.Advance()

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

	var kind symbt.SymbolKind

	if ce.tok.ConsumeKw(token.STATIC) {
		kind = symbt.STATIC
	} else {
		ce.tok.ConsumeKw(token.FIELD)
		kind = symbt.FIELD
	}
	ty := string(ce.tok.Keyword())
	name := ce.tok.Identifier()

	ce.classSt.Define(name, ty, kind)

	s := fmt.Sprintf("<name> %s </name>", name)
	ce.indentln(s)
	s = fmt.Sprintf("<category> %s </category>", string(kind))
	ce.indentln(s)
	s = fmt.Sprintf("<index> %d </index>", ce.classSt.IndexOf(name))
	ce.indentln(s)
	ce.indentln("<usage> false </usage>")

	for ce.tok.ConsumeSym(',') {
		name = ce.tok.Identifier()
		ce.classSt.Define(name, ty, kind)

		s := fmt.Sprintf("<name> %s </name>", name)
		ce.indentln(s)
		s = fmt.Sprintf("<category> %s </category>", string(kind))
		ce.indentln(s)
		s = fmt.Sprintf("<index> %d </index>", ce.classSt.IndexOf(name))
		ce.indentln(s)
		ce.indentln("<usage> false </usage>")
	}
	ce.tok.Advance()

	ce.depth--
	ce.indentln("</classVarDec>")
}

func (ce *CompilationEngine) CompileSubroutine() {
	ce.indentln("<subroutineDec>")
	ce.depth++

	ce.subroutineSt.Reset()
	ce.writetok(2)

	s := fmt.Sprintf("<name> %s </name>", ce.tok.Identifier())
	ce.indentln(s)
	s = fmt.Sprintf("<category> %s </category>", "subroutine")
	ce.indentln(s)
	ce.indentln("<usage> false </usage>")

	ce.writetok(1)

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
		ty := string(ce.tok.Keyword())
		name := ce.tok.Identifier()
		ce.subroutineSt.Define(name, ty, symbt.ARG)

		for ce.tok.ConsumeSym(',') {
			ty = string(ce.tok.Keyword())
			name = ce.tok.Identifier()
			ce.subroutineSt.Define(name, ty, symbt.ARG)
		}

		s := fmt.Sprintf("<name> %s </name>", name)
		ce.indentln(s)
		ce.indentln("<category> arg </category>")
		s = fmt.Sprintf("<index> %d </index>", ce.subroutineSt.IndexOf(name))
		ce.indentln(s)
		ce.indentln("<usage> false </usage>")
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

	ce.tok.Advance()

	ty := string(ce.tok.Keyword())
	name := ce.tok.Identifier()
	ce.subroutineSt.Define(name, ty, symbt.VAR)

	for ce.tok.ConsumeSym(',') {
		name = ce.tok.Identifier()
		ce.subroutineSt.Define(name, ty, symbt.VAR)
	}

	ce.tok.Advance()

	s := fmt.Sprintf("<name> %s </name>", name)
	ce.indentln(s)
	ce.indentln("<category> var </category>")
	s = fmt.Sprintf("<index> %d </index>", ce.subroutineSt.IndexOf(name))
	ce.indentln(s)
	ce.indentln("<usage> false </usage>")

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

	if ce.tok.TokenType() == token.IDENTIFIER {
		name := ce.tok.Identifier()

		s := fmt.Sprintf("<name> %s </name>", name)
		ce.indentln(s)

		if ce.tok.MatchSym('(') {
			s := fmt.Sprintf("<category> %s </category>",
				"subroutine")
			ce.indentln(s)
			ce.indentln("<usage> true </usage>")

			ce.writetok(1)
			ce.CompileExpressionList()
			ce.writetok(1)
			return
		}

		if ce.tok.MatchSym('.') {
			category := "class"
			if ce.subroutineSt.KindOf(name) != symbt.NONE {
				category = string(ce.subroutineSt.KindOf(name))
				s := fmt.Sprintf("<index> %d </index>", ce.subroutineSt.IndexOf(name))
				ce.indentln(s)
			} else if ce.classSt.KindOf(name) != symbt.NONE {
				category = string(ce.classSt.KindOf(name))
				s := fmt.Sprintf("<index> %d </index>", ce.classSt.IndexOf(name))
				ce.indentln(s)
			}
			s = fmt.Sprintf("<category> %s </category>", category)
			ce.indentln(s)
			ce.indentln("<usage> true </usage>")

			ce.writetok(3)
			ce.CompileExpressionList()
			ce.writetok(1)
			return
		}

		if ce.subroutineSt.KindOf(name) != symbt.NONE {
			s = fmt.Sprintf("<index> %d </index>", ce.subroutineSt.IndexOf(name))
			ce.indentln(s)
			s = fmt.Sprintf("<category> %s </category>",
				string(ce.subroutineSt.KindOf(name)))
			ce.indentln(s)
		} else if ce.classSt.KindOf(name) != symbt.NONE {
			s = fmt.Sprintf("<index> %d </index>", ce.classSt.IndexOf(name))
			ce.indentln(s)
			s = fmt.Sprintf("<category> %s </category>",
				string(ce.classSt.KindOf(name)))
			ce.indentln(s)
		}

		if ce.tok.MatchSym('[') {
			ce.indentln("<usage> true </usage>")
			ce.writetok(1)
			ce.CompileExpression()
			ce.writetok(1)
			return
		}
		ce.indentln("<usage> true </usage>")
	} else {
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
