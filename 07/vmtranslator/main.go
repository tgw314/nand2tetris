package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"vmtranslator/codewriter"
	"vmtranslator/parser"
)

func errorAt(p *parser.Parser, ipath string, err error) {
	ln := strconv.Itoa(p.LineNumber())
	log.Panic(ipath, ":", ln, ": ", err)
}

func main() {
	if len(os.Args) < 2 {
		log.Panic("No file specified")
	}

	ipath := os.Args[1]
	in, err := os.Open(ipath)
	if err != nil {
		log.Panic(err)
	}
	defer in.Close()

	opath := strings.TrimSuffix(ipath, filepath.Ext(ipath)) + ".asm"
	out, err := os.Create(opath)
	if err != nil {
		log.Panic(err)
	}
	cw := codewriter.New(out)
	defer cw.Close()

	p := parser.New(in)

	p.Advance()
	for p.HasMoreLines() {
		cmdTy := p.CommandType()
		switch cmdTy {
		case parser.C_ARITHMETIC:
			cmd, err := p.Arg1()
			if err != nil {
				errorAt(p, ipath, err)
			}
			if err := cw.WriteArithmetic(cmd); err != nil {
				errorAt(p, ipath, err)
			}
		case parser.C_PUSH, parser.C_POP:
			seg, err := p.Arg1()
			if err != nil {
				errorAt(p, ipath, err)
			}
			idx, err := p.Arg2()
			if err != nil {
				errorAt(p, ipath, err)
			}

			if err := cw.WritePushPop(cmdTy, seg, idx); err != nil {
				errorAt(p, ipath, err)
			}
		}
		p.Advance()
	}
}
