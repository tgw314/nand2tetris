package vmwriter

import (
	"bufio"
	"fmt"
	"io"
)

type Segment string

const (
	CONST    = Segment("constant")
	ARGUMENT = Segment("argument")
	LOCAL    = Segment("local")
	STATIC   = Segment("static")
	THIS     = Segment("this")
	THAT     = Segment("that")
	POINTER  = Segment("pointer")
	TEMP     = Segment("temp")
)

type Arithmetic string

const (
	ADD = Arithmetic("add")
	SUB = Arithmetic("sub")
	NEG = Arithmetic("neg")
	EQ  = Arithmetic("eq")
	GT  = Arithmetic("gt")
	LT  = Arithmetic("lt")
	AND = Arithmetic("and")
	OR  = Arithmetic("or")
	NOT = Arithmetic("not")
)

type VMWriter struct {
	file   io.Closer
	writer *bufio.Writer
}

func NewVMWriter(f io.WriteCloser) *VMWriter {
	return &VMWriter{
		file:   f,
		writer: bufio.NewWriter(f),
	}
}

func (w *VMWriter) writeln(format string, args ...any) {
	fmt.Fprintf(w.writer, format+"\n", args...)
}

func (w *VMWriter) Push(seg Segment, idx int) {
	w.writeln("    push %s %d", string(seg), idx)
}

func (w *VMWriter) Pop(seg Segment, idx int) {
	w.writeln("    pop %s %d", string(seg), idx)
}

func (w *VMWriter) Arithmetic(cmd Arithmetic) {
	w.writeln("    %s", string(cmd))
}

func (w *VMWriter) Label(l string) {
	w.writeln("label %s", l)
}

func (w *VMWriter) Goto(l string) {
	w.writeln("    goto %s", l)
}

func (w *VMWriter) If(l string) {
	w.writeln("    if-goto %s", l)
}

func (w *VMWriter) Call(name string, nArgs int) {
	w.writeln("    call %s %d", name, nArgs)
}

func (w *VMWriter) Function(name string, nVars int) {
	w.writeln("function %s %d", name, nVars)
}

func (w *VMWriter) Return() {
	w.writeln("    return")
}

func (w *VMWriter) Close() error {
	if err := w.writer.Flush(); err != nil {
		return err
	}
	return w.file.Close()
}
