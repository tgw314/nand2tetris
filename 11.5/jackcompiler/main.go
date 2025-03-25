package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"jackcompiler/compilationengine"
	"jackcompiler/tokenizer"
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

		tok, err := tokenizer.NewTokenizer(in)
		if err != nil {
			log.Panic(err)
		}

		ce, err := compilationengine.NewCompilationEngine(
			tok,
			removeExt(srcPath)+".vm",
		)
		if err != nil {
			log.Panic(err)
		}
		ce.CompileClass()
	}
}
