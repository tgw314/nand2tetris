package codewriter

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"vmtranslater/parser"
)

type CodeWriter struct {
	file        io.Closer
	fname       string
	writer      bufio.Writer
	sp          int
	labelCount  int
	staticCount int
	staticTable map[int]string
}

func New(f *os.File) *CodeWriter {
	return &CodeWriter{file: f, fname: f.Name(), writer: *bufio.NewWriter(f), staticTable: make(map[int]string)}
}

func (cw *CodeWriter) write(format string, a ...any) {
	s := fmt.Sprintf(format, a...)
	cw.writer.WriteString(s)
}

func (cw *CodeWriter) writeln(format string, a ...any) {
	cw.write(format+"\n", a...)
}

func (cw *CodeWriter) pushArith() {
	cw.write(`    // RAM[SP] = D
    @SP
    A=M
    M=D
    // SP++
    @SP
    M=M+1
`)
	cw.sp++
}

func (cw *CodeWriter) popUnary() error {
	if cw.sp-1 < 0 {
		return fmt.Errorf("stack underflow")
	}

	cw.write(`    // y = D = RAM[--SP]
    @SP
    M=M-1
    A=M
    D=M
`)
	cw.sp--
	return nil
}

func (cw *CodeWriter) popBinary() error {
	if cw.sp-2 < 0 {
		return fmt.Errorf("stack underflow")
	}

	cw.write(`    // y = R13 = RAM[--SP]
    @SP
    M=M-1
    A=M
    D=M
    @R13
    M=D
    // x = D = RAM[--SP]
    @SP
    M=M-1
    A=M
    D=M
`)
	cw.sp -= 2
	return nil
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
			ucmd := strings.ToUpper(cmd)
			trueL := fmt.Sprintf(".%s.true.%03d", ucmd, cw.labelCount)
			endL := fmt.Sprintf(".%s.end.%03d", ucmd, cw.labelCount)
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
			cw.labelCount++
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

	cw.pushArith()
	return nil
}

func (cw *CodeWriter) staticLabel(idx int) string {
	_, fn := filepath.Split(cw.fname)
	bn := strings.TrimSuffix(fn, filepath.Ext(fn))

	if _, ok := cw.staticTable[idx]; !ok {
		cw.staticTable[idx] = fmt.Sprintf("%s.%03d", bn, cw.staticCount)
		cw.staticCount++
	}

	return cw.staticTable[idx]
}

func (cw *CodeWriter) WritePushPop(cmd parser.CommandType, seg string, idx int) error {
	switch cmd {
	case parser.C_PUSH:
		cw.sp++
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

		cw.writeln("    // RAM[SP] = D")
		cw.writeln("    @SP")
		cw.writeln("    A=M")
		cw.writeln("    M=D")
		cw.writeln("    // SP++")
		cw.writeln("    @SP")
		cw.writeln("    M=M+1")
	case parser.C_POP:
		if cw.sp-1 < 0 {
			return fmt.Errorf("stack underflow")
		}
		cw.sp--
		cw.writeln("// pop %s %d", seg, idx)
		cw.writeln("    // D = RAM[--SP]")
		cw.writeln("    @SP")
		cw.writeln("    M=M-1")
		cw.writeln("    A=M")
		cw.writeln("    D=M")

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

func (cw *CodeWriter) Close() error {
	cw.write(`(END)
    @END
    0;JMP`)
	cw.writer.Flush()
	return cw.file.Close()
}
