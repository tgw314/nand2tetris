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

func inputPath() (string, error) {
	if len(os.Args) < 2 {
		return os.Getwd()
	}
	return filepath.Abs(os.Args[1])
}

func sourceList(path string) ([]string, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, nil
	}

	if stat.IsDir() {
		return filepath.Glob(filepath.Join(path, "*.vm"))
	} else {
		if filepath.Ext(path) != ".vm" {
			return nil, fmt.Errorf("invalid file extension")
		}
		return []string{path}, nil
	}
}

func removeExt(path string) string {
	return strings.TrimSuffix(path, filepath.Ext(path))
}

func main() {
	ipath, err := inputPath()
	if err != nil {
		log.Panic(err)
	}

	srcs, err := sourceList(ipath)
	if err != nil {
		log.Panic(err)
	}

	opath := removeExt(ipath) + ".asm"
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
