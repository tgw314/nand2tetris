package main

import (
	"fmt"
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

func translate(p *parser.Parser, cw *codewriter.CodeWriter) error {
	p.Advance()
	for p.HasMoreLines() {
		cmdTy := p.CommandType()
		switch cmdTy {
		case parser.C_ARITHMETIC:
			{
				cmd, err := p.Arg1()
				if err != nil {
					return err
				}
				if err := cw.WriteArithmetic(cmd); err != nil {
					return err
				}
			}
		case parser.C_PUSH, parser.C_POP:
			{
				seg, err := p.Arg1()
				if err != nil {
					return err
				}
				idx, err := p.Arg2()
				if err != nil {
					return err
				}

				if err := cw.WritePushPop(cmdTy, seg, idx); err != nil {
					return err
				}
			}
		case parser.C_LABEL:
			{
				l, err := p.Arg1()
				if err != nil {
					return err
				}

				cw.WriteLabel(l)
			}
		case parser.C_GOTO:
			{
				l, err := p.Arg1()
				if err != nil {
					return err
				}

				cw.WriteGoto(l)
			}
		case parser.C_IF:
			{
				l, err := p.Arg1()
				if err != nil {
					return err
				}

				if err := cw.WriteIf(l); err != nil {
					return err
				}
			}
		case parser.C_FUNCTION:
			{
				f, err := p.Arg1()
				if err != nil {
					return err
				}
				n, err := p.Arg2()
				if err != nil {
					return err
				}

				cw.WriteFunction(f, n)
			}
		case parser.C_CALL:
			{
				f, err := p.Arg1()
				if err != nil {
					return err
				}
				n, err := p.Arg2()
				if err != nil {
					return err
				}

				cw.WriteCall(f, n)
			}
		case parser.C_RETURN:
			if err := cw.WriteReturn(); err != nil {
				return err
			}

		}
		p.Advance()
	}
	return nil
}

func main() {
	var ipath string
	var err error

	if len(os.Args) < 2 {
		ipath, err = os.Getwd()
	} else {
		ipath, err = filepath.Abs(os.Args[1])
	}
	if err != nil {
		log.Panic(err)
	}

	stat, err := os.Stat(ipath)
	if err != nil {
		log.Panic(err)
	}

	var srcs []string
	if stat.IsDir() {
		srcs, err = filepath.Glob(filepath.Join(ipath, "*.vm"))
		if err != nil {
			log.Panic(err)
		}
	} else {
		if filepath.Ext(ipath) != ".vm" {
			err = fmt.Errorf("invalid file extension")
			log.Panic(err)
		}
		srcs = []string{ipath}
	}

	opath := strings.TrimSuffix(filepath.Base(ipath), filepath.Ext(ipath)) + ".asm"
	out, err := os.Create(opath)
	if err != nil {
		log.Panic(err)
	}
	cw := codewriter.New(out, ipath)
	defer cw.Close()

	for _, src := range srcs {
		in, err := os.Open(src)
		if err != nil {
			log.Panic(err)
		}
		defer in.Close()

		p := parser.New(in)
		cw.SetFileName(src)
		if err := translate(p, cw); err != nil {
			errorAt(p, src, err)
		}
	}
}
