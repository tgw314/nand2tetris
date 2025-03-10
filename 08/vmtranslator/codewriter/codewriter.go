package codewriter

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"vmtranslator/parser"
)

type counter func(bool) int

func cnt() counter {
	var i int
	return func(reset bool) int {
		defer func() {
			if !reset {
				i++
			}
		}()

		if reset {
			i = 0
		}

		return i
	}
}

type CodeWriter struct {
	file        io.Closer
	writer      bufio.Writer
	sp          int
	labelCount  counter
	staticCount counter
	returnCount counter
	staticTable map[int]string
	vmName      string
	labelPrefix string
}

func New(f io.WriteCloser, ipath string) *CodeWriter {
	cw := &CodeWriter{
		file:        f,
		writer:      *bufio.NewWriter(f),
		labelCount:  cnt(),
		staticCount: cnt(),
		returnCount: cnt(),
		staticTable: make(map[int]string),
	}

	cw.SetFileName(ipath)

	cw.writeln("    // SP=256")
	cw.writeln("    @256")
	cw.writeln("    D=A")
	cw.writeln("    @SP")
	cw.writeln("    M=D")
	cw.WriteCall("Sys.init", 0)

	return cw
}

func (cw *CodeWriter) SetFileName(path string) {
	_, fn := filepath.Split(path)
	bn := strings.TrimSuffix(fn, filepath.Ext(fn))
	cw.vmName = bn
}

func (cw *CodeWriter) write(format string, a ...any) {
	s := fmt.Sprintf(format, a...)
	cw.writer.WriteString(s)
}

func (cw *CodeWriter) writeln(format string, a ...any) {
	cw.write(format+"\n", a...)
}

func (cw *CodeWriter) push() {
	cw.writeln("    @SP")
	cw.writeln("    A=M")
	cw.writeln("    M=D")
	cw.writeln("    @SP")
	cw.writeln("    M=M+1")

	cw.sp++
}

func (cw *CodeWriter) popUnary() error {
	if cw.sp-1 < 0 {
		return fmt.Errorf("stack underflow")
	}

	cw.writeln("    // pop D")
	cw.writeln("    @SP")
	cw.writeln("    M=M-1")
	cw.writeln("    A=M")
	cw.writeln("    D=M")

	cw.sp--
	return nil
}

func (cw *CodeWriter) popBinary() error {
	if cw.sp-2 < 0 {
		return fmt.Errorf("stack underflow")
	}

	cw.writeln("    // pop R13 (= y)")
	cw.writeln("    @SP")
	cw.writeln("    M=M-1")
	cw.writeln("    A=M")
	cw.writeln("    D=M")
	cw.writeln("    @R13")
	cw.writeln("    M=D")
	cw.writeln("    // pop D (= x)")
	cw.writeln("    @SP")
	cw.writeln("    M=M-1")
	cw.writeln("    A=M")
	cw.writeln("    D=M")

	cw.sp -= 2
	return nil
}

func (cw *CodeWriter) staticLabel(idx int) string {
	if _, ok := cw.staticTable[idx]; !ok {
		cw.staticTable[idx] = fmt.Sprintf("%s.%03d", cw.vmName, cw.staticCount(false))
	}

	return cw.staticTable[idx]
}

func (cw *CodeWriter) WriteArithmetic(cmd string) error {
	cw.writeln("// %s", cmd)
	switch cmd {
	case "add":
		if err := cw.popBinary(); err != nil {
			return err
		}
		cw.writeln("    // D = D + R13")
		cw.writeln("    @R13")
		cw.writeln("    D=D+M")
	case "sub":
		if err := cw.popBinary(); err != nil {
			return err
		}
		cw.writeln("    // D = D - R13")
		cw.writeln("    @R13")
		cw.writeln("    D=D-M")
	case "neg":
		if err := cw.popUnary(); err != nil {
			return err
		}
		cw.writeln("    // D = -D")
		cw.writeln("    D=-D")
	case "eq", "gt", "lt":
		if err := cw.popBinary(); err != nil {
			return err
		}
		{
			i := cw.labelCount(false)
			ucmd := strings.ToUpper(cmd)
			trueL := fmt.Sprintf(".%s.true.%03d", ucmd, i)
			endL := fmt.Sprintf(".%s.end.%03d", ucmd, i)
			cw.writeln("    // D = D - R13")
			cw.writeln("    @R13")
			cw.writeln("    D=D-M")
			cw.writeln("    // D = D %s 0", cmd)
			cw.writeln("    @%s", trueL)
			cw.writeln("    D;J%s", ucmd)
			cw.writeln("    D=0")
			cw.writeln("    @%s", endL)
			cw.writeln("    0;JMP")
			cw.writeln("(%s) ", trueL)
			cw.writeln("    D=-1")
			cw.writeln("(%s)", endL)
		}
	case "and":
		if err := cw.popBinary(); err != nil {
			return err
		}
		cw.writeln("    // D = D & R13")
		cw.writeln("    @R13")
		cw.writeln("    D=D&M")
	case "or":
		if err := cw.popBinary(); err != nil {
			return err
		}
		cw.writeln("    // D = D | R13")
		cw.writeln("    @R13")
		cw.writeln("    D=D|M")
	case "not":
		if err := cw.popUnary(); err != nil {
			return err
		}
		cw.writeln("    // D = !D")
		cw.writeln("    D=!D")
	}

	cw.push()
	return nil
}

func (cw *CodeWriter) WritePushPop(cmd parser.CommandType, seg string, idx int) error {
	switch cmd {
	case parser.C_PUSH:
		cw.writeln("// push %s %d", seg, idx)

		switch seg {
		case "local", "argument", "this", "that":
			var uSeg string
			switch seg {
			case "local":
				uSeg = "LCL"
			case "argument":
				uSeg = "ARG"
			default:
				uSeg = strings.ToUpper(seg)
			}

			cw.writeln("    // D = %d", idx)
			cw.writeln("    @%d", idx)
			cw.writeln("    D=A")

			cw.writeln("    // D = RAM[%s][D]", uSeg)
			cw.writeln("    @%s", uSeg)
			cw.writeln("    A=M")
			cw.writeln("    A=D+A")
			cw.writeln("    D=M")

		case "pointer":
			if idx != 0 && idx != 1 {
				return fmt.Errorf("invalid pointer index %d", idx)
			}
			cw.writeln("    // D = RAM[%d]", 3+idx)
			cw.writeln("    @%d", 3+idx)
			cw.writeln("    D=M")

		case "temp":
			if idx < 0 || 7 < idx {
				return fmt.Errorf("invalid temp index %d", idx)
			}
			cw.writeln("    // D = RAM[%d]", 5+idx)
			cw.writeln("    @%d", 5+idx)
			cw.writeln("    D=M")

		case "constant":
			cw.writeln("    // D = %d", idx)
			cw.writeln("    @%d", idx)
			cw.writeln("    D=A")

		case "static":
			{
				l := cw.staticLabel(idx)
				cw.writeln("    // D = RAM[%s]", l)
				cw.writeln("    @%s", l)
				cw.writeln("    D=M")
			}
		}

		cw.push()
	case parser.C_POP:
		cw.writeln("// pop %s %d", seg, idx)
		if err := cw.popUnary(); err != nil {
			return err
		}
		switch seg {
		case "local", "argument", "this", "that":
			var uSeg string
			switch seg {
			case "local":
				uSeg = "LCL"
			case "argument":
				uSeg = "ARG"
			default:
				uSeg = strings.ToUpper(seg)
			}

			cw.writeln("    // R13 = D")
			cw.writeln("    @R13")
			cw.writeln("    M=D")

			cw.writeln("    // D = %d", idx)
			cw.writeln("    @%d", idx)
			cw.writeln("    D=A")

			cw.writeln("    // R14 = %s + D", uSeg)
			cw.writeln("    @%s", uSeg)
			cw.writeln("    A=M")
			cw.writeln("    D=D+A")
			cw.writeln("    @R14")
			cw.writeln("    M=D")

			cw.writeln("    // D = R13")
			cw.writeln("    @R13")
			cw.writeln("    D=M")

			cw.writeln("    // A = R14")
			cw.writeln("    @R14")
			cw.writeln("    A=M")

			cw.writeln("    // RAM[%s][%d] = D", uSeg, idx)
			cw.writeln("    M=D")

		case "pointer":
			if idx != 0 && idx != 1 {
				return fmt.Errorf("invalid pointer index %d", idx)
			}
			cw.writeln("    // RAM[%d] = D", 3+idx)
			cw.writeln("    @%d", 3+idx)
			cw.writeln("    M=D")

		case "temp":
			if idx < 0 || 7 < idx {
				return fmt.Errorf("invalid temp index %d", idx)
			}
			cw.writeln("    // RAM[%d] = D", 5+idx)
			cw.writeln("    @%d", 5+idx)
			cw.writeln("    M=D")

		case "constant":
			cw.writeln("    // RAM[%d] = D", idx)
			cw.writeln("    @%d", idx)
			cw.writeln("    M=D")

		case "static":
			{
				l := cw.staticLabel(idx)
				cw.writeln("    // RAM[%s] = D", l)
				cw.writeln("    @%s", l)
				cw.writeln("    M=D")
			}
		}
	}

	return nil
}

func (cw *CodeWriter) WriteLabel(label string) {
	l := fmt.Sprintf("%s$%s", cw.labelPrefix, label)

	cw.writeln("// label %s", l)
	cw.writeln("(%s)", l)
}

func (cw *CodeWriter) WriteGoto(label string) {
	l := fmt.Sprintf("%s$%s", cw.labelPrefix, label)

	cw.writeln("// goto %s", l)
	cw.writeln("    @%s", l)
	cw.writeln("    0;JMP")
}

func (cw *CodeWriter) WriteIf(label string) error {
	l := fmt.Sprintf("%s$%s", cw.labelPrefix, label)

	cw.writeln("// if-goto %s", l)
	if err := cw.popUnary(); err != nil {
		return err
	}
	cw.writeln("    @%s", l)
	cw.writeln("    D;JNE")

	return nil
}

func (cw *CodeWriter) WriteFunction(name string, nVars int) {
	cw.labelPrefix = name
	cw.returnCount(true)

	cw.writeln("// function %s %d", name, nVars)
	cw.writeln("(%s)", cw.labelPrefix)
	for range nVars {
		cw.writeln("    D=0")
		cw.push()
	}
}

func (cw *CodeWriter) WriteCall(name string, nArgs int) {
	ret := fmt.Sprintf("%s$ret.%03d", cw.labelPrefix, cw.returnCount(false))

	cw.writeln("// call %s %d", name, nArgs)
	for _, r := range []string{ret, "LCL", "ARG", "THIS", "THAT"} {
		cw.writeln("    // push %s", r)
		cw.writeln("    @%s", r)
		if r == ret {
			cw.writeln("    D=A")
		} else {
			cw.writeln("    D=M")
		}
		cw.push()
	}
	cw.writeln("    // ARG = SP - 5 - %d", nArgs)
	cw.writeln("    @SP")
	cw.writeln("    D=M")
	cw.writeln("    @5")
	cw.writeln("    D=D-A")
	cw.writeln("    @%d", nArgs)
	cw.writeln("    D=D-A")
	cw.writeln("    @ARG")
	cw.writeln("    M=D")
	cw.writeln("    // LCL = SP")
	cw.writeln("    @SP")
	cw.writeln("    D=M")
	cw.writeln("    @LCL")
	cw.writeln("    M=D")
	cw.writeln("    // goto %s", name)
	cw.writeln("    @%s", name)
	cw.writeln("    0;JMP")

	cw.writeln("(%s)", ret)
}

func (cw *CodeWriter) WriteReturn() error {
	cw.writeln("// return")
	cw.writeln("    // frame = LCL")
	cw.writeln("    @LCL")
	cw.writeln("    D=M")
	cw.writeln("    @frame")
	cw.writeln("    M=D")

	cw.writeln("    // retAddr = RAM[frame - 5]")
	cw.writeln("    @5")
	cw.writeln("    A=D-A")
	cw.writeln("    D=M")
	cw.writeln("    @retAddr")
	cw.writeln("    M=D")

	if err := cw.popUnary(); err != nil {
		return err
	}
	cw.writeln("    // RAM[ARG] = D")
	cw.writeln("    @ARG")
	cw.writeln("    A=M")
	cw.writeln("    M=D")
	cw.writeln("    // SP = ARG + 1")
	cw.writeln("    @ARG")
	cw.writeln("    D=M")
	cw.writeln("    @SP")
	cw.writeln("    M=D+1")

	for i, r := range []string{"THAT", "THIS", "ARG", "LCL"} {
		cw.writeln("    // %s = RAM[frame - %d]", r, i+1)
		cw.writeln("    @frame")
		cw.writeln("    D=M")
		cw.writeln("    @%d", i+1)
		cw.writeln("    A=D-A")
		cw.writeln("    D=M")
		cw.writeln("    @%s", r)
		cw.writeln("    M=D")
	}

	cw.writeln("    // goto retAddr")
	cw.writeln("    @retAddr")
	cw.writeln("    A=M")
	cw.writeln("    0;JMP")
	return nil
}

func (cw *CodeWriter) Close() error {
	cw.write(`(END)
    @END
    0;JMP`)
	cw.writer.Flush()
	return cw.file.Close()
}
