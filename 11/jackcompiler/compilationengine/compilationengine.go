package compilationengine

import (
	"fmt"
	"log"
	"os"
	"strings"

	symbt "jackcompiler/symboltable"
	token "jackcompiler/tokenizer"
	vm "jackcompiler/vmwriter"
)

type counter func() int

func count() counter {
	var i int
	return func() int {
		defer func() {
			i++
		}()
		return i
	}
}

func (c *counter) Reset() {
	*c = count()
}

type CompilationEngine struct {
	tok           *token.Tokenizer
	vm            vm.VMWriter
	class         string
	classSt       *symbt.SymbolTable
	subroutineSt  *symbt.SymbolTable
	labelCount    counter
	isVoid        bool
	isConstructor bool
}

func NewCompilationEngine(tok *token.Tokenizer, opath string) (*CompilationEngine, error) {
	out, err := os.Create(opath)
	if err != nil {
		return nil, err
	}

	ce := &CompilationEngine{
		tok:          tok,
		vm:           *vm.NewVMWriter(out),
		classSt:      symbt.NewSymbolTable(),
		subroutineSt: symbt.NewSymbolTable(),
		labelCount:   count(),
	}

	return ce, nil
}

func (ce *CompilationEngine) CompileClass() {
	defer ce.vm.Close()

	ce.classSt.Reset()
	ce.tok.Advance()

	ce.tok.ExpectKw(token.CLASS)
	ce.class = ce.tok.Identifier()
	ce.tok.ExpectSym('{')

	for ce.tok.MatchKw(token.STATIC) ||
		ce.tok.MatchKw(token.FIELD) {
		ce.compileClassVarDec()
	}

	for ce.tok.MatchKw(token.CONSTRUCTOR) ||
		ce.tok.MatchKw(token.FUNCTION) ||
		ce.tok.MatchKw(token.METHOD) {
		ce.compileSubroutine()
	}

	ce.tok.ExpectSym('}')
}

func (ce *CompilationEngine) compileClassVarDec() {
	kind := symbt.SymbolKind(ce.tok.Keyword())
	ty := string(ce.tok.Keyword())
	name := ce.tok.Identifier()

	ce.classSt.Define(name, ty, kind)

	for ce.tok.ConsumeSym(',') {
		name = ce.tok.Identifier()
		ce.classSt.Define(name, ty, kind)
	}

	ce.tok.ExpectSym(';')
}

func (ce *CompilationEngine) compileSubroutine() {
	ce.subroutineSt.Reset()
	ce.labelCount.Reset()

	ce.isConstructor = false
	isM := false
	ce.isVoid = false

	switch kw := ce.tok.Keyword(); kw {
	case token.CONSTRUCTOR:
		ce.isConstructor = true
		ce.isVoid = ce.tok.Keyword() == token.VOID
	case token.FUNCTION:
		ce.isVoid = ce.tok.Keyword() == token.VOID
	case token.METHOD:
		isM = true
		ce.isVoid = ce.tok.Keyword() == token.VOID
		ce.subroutineSt.Define("this", ce.class, symbt.ARG)
	default:
		log.Panicf("unexpected keyword %s", kw)
	}
	name := ce.tok.Identifier()

	ce.tok.ExpectSym('(')
	ce.compileParameterList()
	ce.tok.ExpectSym(')')
	ce.compileSubroutineBody(name, isM)
}

func (ce *CompilationEngine) compileParameterList() {
	for !ce.tok.MatchSym(')') {
		ty := string(ce.tok.Keyword())
		name := ce.tok.Identifier()
		ce.subroutineSt.Define(name, ty, symbt.ARG)

		if ce.tok.ConsumeSym(',') {
			continue
		}
	}
}

func (ce *CompilationEngine) compileSubroutineBody(name string, isM bool) {
	ce.tok.ExpectSym('{')
	for ce.tok.MatchKw(token.VAR) {
		ce.compileVarDec()
	}
	ce.vm.Function(
		ce.class+"."+name,
		ce.subroutineSt.VarCount(symbt.VAR),
	)

	if ce.isConstructor {
		ce.vm.Push(vm.CONST, ce.classSt.VarCount(symbt.FIELD))
		ce.vm.Call("Memory.alloc", 1)
		ce.vm.Pop(vm.POINTER, 0)
	} else if isM {
		ce.vm.Push(vm.ARGUMENT, 0)
		ce.vm.Push(vm.POINTER, 0)
	}

	ce.compileStatements()

	ce.tok.ExpectSym('}')
}

func (ce *CompilationEngine) compileVarDec() {
	ce.tok.ExpectKw(token.VAR)
	ty := string(ce.tok.Keyword())

	name := ce.tok.Identifier()
	ce.subroutineSt.Define(name, ty, symbt.VAR)

	for ce.tok.ConsumeSym(',') {
		name = ce.tok.Identifier()
		ce.subroutineSt.Define(name, ty, symbt.VAR)
	}
	ce.tok.ExpectSym(';')
}

func (ce *CompilationEngine) compileStatements() {
	for ce.tok.MatchKw(token.LET) ||
		ce.tok.MatchKw(token.IF) ||
		ce.tok.MatchKw(token.WHILE) ||
		ce.tok.MatchKw(token.DO) ||
		ce.tok.MatchKw(token.RETURN) {

		if ce.tok.MatchKw(token.LET) {
			ce.compileLet()
		}

		if ce.tok.MatchKw(token.IF) {
			ce.compileIf()
		}

		if ce.tok.MatchKw(token.WHILE) {
			ce.compileWhile()
		}

		if ce.tok.MatchKw(token.DO) {
			ce.compileDo()
		}

		if ce.tok.MatchKw(token.RETURN) {
			ce.compileReturn()
		}

	}
}

func (ce *CompilationEngine) findVar(name string) (vm.Segment, int) {
	switch ce.subroutineSt.KindOf(name) {
	case symbt.STATIC:
		return vm.STATIC, ce.subroutineSt.IndexOf(name)
	case symbt.FIELD:
		return vm.THIS, ce.subroutineSt.IndexOf(name)
	case symbt.ARG:
		return vm.ARGUMENT, ce.subroutineSt.IndexOf(name)
	case symbt.VAR:
		return vm.LOCAL, ce.subroutineSt.IndexOf(name)
	case symbt.NONE:
	default:
		switch ce.classSt.KindOf(name) {
		case symbt.STATIC:
			return vm.STATIC, ce.classSt.IndexOf(name)
		case symbt.FIELD:
			return vm.THIS, ce.classSt.IndexOf(name)
		case symbt.ARG:
			return vm.ARGUMENT, ce.classSt.IndexOf(name)
		case symbt.VAR:
			return vm.LOCAL, ce.classSt.IndexOf(name)
		}
	}

	return "", -1
}

func (ce *CompilationEngine) compileLet() {
	ce.tok.ExpectKw(token.LET)

	lhs := ce.tok.Identifier()
	seg, idx := ce.findVar(lhs)
	if idx == -1 {
		log.Panicf("undeclared variable %s", lhs)
	}

	if ce.tok.ConsumeSym('[') {
		ce.vm.Push(seg, idx)
		ce.compileExpression()
		ce.vm.Arithmetic(vm.ADD)
		ce.tok.ExpectSym(']')

		ce.tok.ExpectSym('=')

		// rhs
		ce.compileExpression()
		ce.vm.Pop(vm.TEMP, 0)
		ce.vm.Pop(vm.POINTER, 1)
		ce.vm.Push(vm.TEMP, 0)
		ce.vm.Pop(vm.THAT, 0)
	} else {
		ce.tok.ExpectSym('=')

		ce.compileExpression()
		ce.vm.Pop(seg, idx)
	}

	ce.tok.ExpectSym(';')
}

func (ce *CompilationEngine) compileIf() {
	i := ce.labelCount()
	ce.tok.ExpectKw(token.IF)

	ce.tok.ExpectSym('(')
	ce.compileExpression()
	ce.tok.ExpectSym(')')

	ce.vm.Arithmetic(vm.NOT)
	ce.vm.If(
		fmt.Sprintf("L.else.%03d", i),
	)

	ce.tok.ExpectSym('{')
	ce.compileStatements()
	ce.tok.ExpectSym('}')

	ce.vm.Goto(
		fmt.Sprintf("L.end.%03d", i),
	)

	ce.vm.Label(
		fmt.Sprintf("L.else.%03d", i),
	)
	if ce.tok.ConsumeKw(token.ELSE) {
		ce.tok.ExpectSym('{')
		ce.compileStatements()
		ce.tok.ExpectSym('}')
	}
	ce.vm.Label(
		fmt.Sprintf("L.end.%03d", i),
	)
}

func (ce *CompilationEngine) compileWhile() {
	i := ce.labelCount()
	ce.tok.ExpectKw(token.WHILE)

	ce.vm.Label(
		fmt.Sprintf("L.begin.%03d", i),
	)

	ce.tok.ExpectSym('(')
	ce.compileExpression()
	ce.tok.ExpectSym(')')

	ce.vm.Arithmetic(vm.NOT)
	ce.vm.If(
		fmt.Sprintf("L.end.%03d", i),
	)

	ce.tok.ExpectSym('{')
	ce.compileStatements()
	ce.tok.ExpectSym('}')

	ce.vm.Goto(
		fmt.Sprintf("L.begin.%03d", i),
	)

	ce.vm.Label(
		fmt.Sprintf("L.end.%03d", i),
	)
}

func (ce *CompilationEngine) compileDo() {
	ce.tok.ExpectKw(token.DO)

	ce.compileExpression()
	ce.vm.Pop(vm.TEMP, 0) // discard return value

	ce.tok.ExpectSym(';')
}

func (ce *CompilationEngine) compileReturn() {
	ce.tok.ExpectKw(token.RETURN)
	if ce.tok.ConsumeSym(';') {
		return
	}
	ce.compileExpression()
	ce.tok.ExpectSym(';')
	if ce.isVoid {
		ce.vm.Push(vm.CONST, 0)
	} else if ce.isConstructor {
		ce.vm.Push(vm.POINTER, 0)
	}
	ce.vm.Return()
}

func (ce *CompilationEngine) compileExpression() {
	ce.compileTerm()
	for strings.ContainsRune("+-*/&|<>=", ce.tok.PeekSym()) {
		op := ce.tok.Symbol()
		ce.compileTerm()

		switch op {
		case '+':
			ce.vm.Arithmetic(vm.ADD)
		case '-':
			ce.vm.Arithmetic(vm.SUB)
		case '*':
			ce.vm.Call("Math.multiply", 2)
		case '/':
			ce.vm.Call("Math.divide", 2)
		case '&':
			ce.vm.Arithmetic(vm.AND)
		case '|':
			ce.vm.Arithmetic(vm.OR)
		case '<':
			ce.vm.Arithmetic(vm.LT)
		case '>':
			ce.vm.Arithmetic(vm.GT)
		case '=':
			ce.vm.Arithmetic(vm.EQ)
		}
	}
}

func (ce *CompilationEngine) compileTerm() {
	if ce.tok.TokenType() == token.INT_CONST {
		ce.vm.Push(vm.CONST, ce.tok.IntVal())
		return
	}

	if ce.tok.TokenType() == token.STRING_CONST {
		ce.vm.Call("String.new", 0)
		for _, c := range []byte(ce.tok.StringVal()) {
			ce.vm.Push(vm.CONST, int(c))
			ce.vm.Call("String.appendChar", 2)
		}
		return
	}

	if ce.tok.TokenType() == token.KEYWORD {
		if ce.tok.ConsumeKw(token.TRUE) {
			ce.vm.Push(vm.CONST, 1)
			ce.vm.Arithmetic(vm.NEG)
			return
		}

		if ce.tok.ConsumeKw(token.FALSE) {
			ce.vm.Push(vm.CONST, 0)
			return
		}

		if ce.tok.ConsumeKw(token.NULL) {
			ce.vm.Push(vm.CONST, 0)
			return
		}

		ce.tok.ExpectKw(token.THIS)
		ce.vm.Push(vm.POINTER, 0)
		return
	}

	if ce.tok.ConsumeSym('-') {
		ce.compileTerm()
		ce.vm.Arithmetic(vm.NEG)
		return
	}

	if ce.tok.ConsumeSym('~') {
		ce.compileTerm()
		ce.vm.Arithmetic(vm.NOT)
		return
	}

	if ce.tok.ConsumeSym('(') {
		ce.compileExpression()
		ce.tok.ExpectSym(')')
		return
	}

	name := ce.tok.Identifier()
	seg, idx := ce.findVar(name)

	if ce.tok.ConsumeSym('(') {
		ce.vm.Call(ce.class+"."+name, ce.compileExpressionList())
		ce.tok.ExpectSym(')')
		return
	}

	if ce.tok.ConsumeSym('.') {
		fnName := ce.tok.Identifier()
		ce.tok.ExpectSym('(')

		if idx == -1 { // name is a class name
			ce.vm.Call(name+"."+fnName, ce.compileExpressionList())
			ce.tok.ExpectSym(')')
			return
		}

		// name is a variable name
		ce.vm.Push(seg, idx)
		ce.vm.Call(name+"."+fnName, ce.compileExpressionList()+1)
		ce.tok.ExpectSym(')')
		return
	}

	if idx == -1 {
		log.Panicf("undeclared variable %s", name)
	}

	ce.vm.Push(seg, idx)
	if ce.tok.ConsumeSym('[') {
		ce.compileExpression()
		ce.tok.ExpectSym(']')
		ce.vm.Arithmetic(vm.ADD)
		ce.vm.Pop(vm.POINTER, 1)
		ce.vm.Push(vm.THAT, 0)
		return
	}
}

func (ce *CompilationEngine) compileExpressionList() int {
	i := 0
	for !ce.tok.MatchSym(')') {
		ce.compileExpression()
		i++
		if ce.tok.ConsumeSym(',') {
			continue
		}
	}

	return i
}
