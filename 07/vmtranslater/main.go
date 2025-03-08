package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"vmtranslater/codewriter"
	"vmtranslater/parser"
)

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
				log.Panic(err)
			}
			if err := cw.WriteArithmetic(cmd); err != nil {
				log.Panic(err)
			}
		case parser.C_PUSH, parser.C_POP:
			seg, err := p.Arg1()
			if err != nil {
				log.Panic(err)
			}
			idx, err := p.Arg2()
			if err != nil {
				log.Panic(err)
			}

			if err := cw.WritePushPop(cmdTy, seg, idx); err != nil {
				log.Panic(err)
			}
		}
		p.Advance()
	}
}
