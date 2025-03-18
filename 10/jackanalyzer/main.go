package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"jackanalyzer/tokenizer"
)

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
		return filepath.Glob(filepath.Join(path, "*.jack"))
	}

	if filepath.Ext(path) != ".jack" {
		return nil, fmt.Errorf("invalid file extension")
	}

	return []string{path}, nil
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

func saveTokens(tok *tokenizer.Tokenizer, path string) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	w := bufio.NewWriter(out)
	defer w.Flush()

	w.WriteString("<tokens>\n")

	if err := tok.Advance(); err != nil {
		return err
	}
	for tok.HasMoreTokens() {
		w.WriteString("<" + string(tok.TokenType()) + "> ")

		switch tok.TokenType() {
		case tokenizer.KEYWORD:
			w.WriteString(string(tok.Keyword()))
		case tokenizer.SYMBOL:
			w.WriteString(escape(string(tok.Symbol())))
		case tokenizer.INT_CONST:
			n, err := tok.IntVal()
			if err != nil {
				return err
			}
			w.WriteString(strconv.Itoa(n))
		case tokenizer.STRING_CONST:
			w.WriteString(escape(tok.StringVal()))
		case tokenizer.IDENTIFIER:
			w.WriteString(tok.Identifier())
		}

		w.WriteString(" </" + string(tok.TokenType()) + ">\n")
		if err := tok.Advance(); err != nil {
			return err
		}
	}
	w.WriteString("</tokens>")

	return nil
}

func removeExt(path string) string {
	return strings.TrimSuffix(path, filepath.Ext(path))
}

func main() {
	ipath, err := inputPath()
	if err != nil {
		log.Panic(err)
	}

	srcPaths, err := sourceList(ipath)
	if err != nil {
		log.Panic(err)
	}

	for _, srcPath := range srcPaths {
		in, err := os.Open(srcPath)
		if err != nil {
			log.Panic(err)
		}
		defer in.Close()

		tok, err := tokenizer.New(in)
		if err != nil {
			log.Panic(err)
		}

		if err := saveTokens(tok, removeExt(srcPath)+"T.xml"); err != nil {
			log.Panic(err)
		}

	}
}
