package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"assembler/code"
	"assembler/parser"
	"assembler/symboltable"
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

	opath := strings.TrimSuffix(ipath, filepath.Ext(ipath)) + ".hack"
	out, err := os.Create(opath)
	if err != nil {
		log.Panic(err)
	}
	defer out.Close()

	st := symboltable.New()
	w := bufio.NewWriter(out)

	p1 := parser.New(in)
	p1.Advance()
	for p1.HasMoreLines() {
		if p1.InstructionType() == parser.L_INSTRUCTION {
			sb, err := p1.Symbol(nil)
			if err != nil {
				log.Panic(err)
			}

			if st.Contains(sb) {
				err := fmt.Errorf("duplicate symbol: %s", sb)
				log.Panic(err)
			}

			st.AddEntry(sb, p1.LineNum()+1)
		}
		p1.Advance()
	}
	if _, err := in.Seek(0, 0); err != nil {
		log.Panic(err)
	}

	p2 := parser.New(in)

	p2.Advance()
	for p2.HasMoreLines() {
		switch p2.InstructionType() {
		case parser.A_INSTRUCTION:
			{
				sb, err := p2.Symbol(st)
				if err != nil {
					log.Panic(err)
				}

				n, err := strconv.ParseInt(sb, 10, 64)
				if err != nil {
					log.Panic(err)
				}
				bin := fmt.Sprintf("0%015b\n", n)

				fmt.Print(bin)
				// fmt.Printf("%5d: %s", p.Lineno, bin)
				w.WriteString(bin)
			}

		case parser.C_INSTRUCTION:
			{
				comp, err := code.Comp(p2.Comp())
				if err != nil {
					log.Panic(err)
				}

				dest, err := code.Dest(p2.Dest())
				if err != nil {
					log.Panic(err)
				}

				jump, err := code.Jump(p2.Jump())
				if err != nil {
					log.Panic(err)
				}

				bin := "111" + comp + dest + jump + "\n"

				fmt.Print(bin)
				// fmt.Printf("%5d: %s", p.Lineno, bin)
				w.WriteString(bin)
			}
		}

		p2.Advance()
	}
	w.Flush()
}
