package tokenizer

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

type TokenType string

const (
	KEYWORD      = TokenType("keyword")
	SYMBOL       = TokenType("symbol")
	IDENTIFIER   = TokenType("identifier")
	INT_CONST    = TokenType("integerConstant")
	STRING_CONST = TokenType("stringConstant")
)

type Keyword string

const (
	CLASS       = Keyword("class")
	CONSTRUCTOR = Keyword("constructor")
	FUNCTION    = Keyword("function")
	METHOD      = Keyword("method")
	FIELD       = Keyword("field")
	STATIC      = Keyword("static")
	VAR         = Keyword("var")
	INT         = Keyword("int")
	CHAR        = Keyword("char")
	BOOLEAN     = Keyword("boolean")
	VOID        = Keyword("void")
	TRUE        = Keyword("true")
	FALSE       = Keyword("false")
	NULL        = Keyword("null")
	THIS        = Keyword("this")
	LET         = Keyword("let")
	DO          = Keyword("do")
	IF          = Keyword("if")
	ELSE        = Keyword("else")
	WHILE       = Keyword("while")
	RETURN      = Keyword("return")
)

var keywords = []string{
	"class",
	"constructor",
	"function",
	"method",
	"field",
	"static",
	"var",
	"int",
	"char",
	"boolean",
	"void",
	"true",
	"false",
	"null",
	"this",
	"let",
	"do",
	"if",
	"else",
	"while",
	"return",
}

var keywordMap = map[string]Keyword{
	"class":       CLASS,
	"constructor": CONSTRUCTOR,
	"function":    FUNCTION,
	"method":      METHOD,
	"field":       FIELD,
	"static":      STATIC,
	"var":         VAR,
	"int":         INT,
	"char":        CHAR,
	"boolean":     BOOLEAN,
	"void":        VOID,
	"true":        TRUE,
	"false":       FALSE,
	"null":        NULL,
	"this":        THIS,
	"let":         LET,
	"do":          DO,
	"if":          IF,
	"else":        ELSE,
	"while":       WHILE,
	"return":      RETURN,
}

type Tokenizer struct {
	src  string
	r    *bufio.Reader
	pos  int
	body string
	ty   TokenType
}

func isIdentHead(r rune) bool {
	return ('A' <= r && r <= 'Z') || ('a' <= r && r <= 'z') || r == '_'
}

func notIdentRune(r rune) bool {
	return !isIdentHead(r) && !unicode.IsDigit(r)
}

func New(r *os.File) (*Tokenizer, error) {
	src, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	tok := &Tokenizer{
		src: string(src),
		r:   bufio.NewReader(r),
		pos: 0,
	}

	return tok, nil
}

func (t *Tokenizer) HasMoreTokens() bool {
	return len(t.src) > t.pos
}

func (t *Tokenizer) Advance() {
	for t.HasMoreTokens() {
		if unicode.IsSpace(rune(t.src[t.pos])) {
			t.pos++
			continue
		}

		if strings.HasPrefix(t.src[t.pos:], "//") {
			t.pos += 2
			lf := strings.Index(t.src[t.pos:], "\n")
			t.pos += lf + 1
			continue
		}

		if strings.HasPrefix(t.src[t.pos:], "/*") {
			t.pos += 2
			skip := strings.Index(t.src[t.pos:], "*/")
			if skip == -1 {
				log.Panic("unterminated comment")
			}
			t.pos += skip + 2
			continue
		}

		if unicode.IsDigit(rune(t.src[t.pos])) {
			t.ty = INT_CONST
			for i, c := range t.src[t.pos:] {
				if !unicode.IsDigit(c) {
					t.body = t.src[t.pos : t.pos+i]
					t.pos += i
					return
				}
			}
		}

		if strings.IndexAny(t.src[t.pos:], "{}()[].,;+-*/&|<>=~") == 0 {
			t.ty = SYMBOL
			t.body = t.src[t.pos : t.pos+1]

			t.pos++
			return
		}

		if strings.HasPrefix(t.src[t.pos:], "\"") {
			t.ty = STRING_CONST
			t.pos++
			quote := strings.Index(t.src[t.pos:], "\"")
			tail := t.pos + quote
			t.body = t.src[t.pos:tail]

			t.pos = tail + 1
			return
		}

		if isIdentHead(rune(t.src[t.pos])) {
			end := strings.IndexFunc(t.src[t.pos:], notIdentRune)
			tail := t.pos + end

			t.body = t.src[t.pos:tail]
			t.ty = IDENTIFIER
			if slices.Contains(keywords, t.body) {
				t.ty = KEYWORD
			}

			t.pos = tail
			return
		}

		err := fmt.Errorf(
			"unexpected token: `%s`",
			strings.TrimRight(t.src[t.pos:], "\n"),
		)
		log.Panic(err)
	}
}

func (t *Tokenizer) TokenType() TokenType {
	return t.ty
}

func (t *Tokenizer) Keyword() Keyword {
	defer t.Advance()
	return keywordMap[t.body]
}

func (t *Tokenizer) MatchKw(k Keyword) bool {
	return keywordMap[t.body] == k
}

func (t *Tokenizer) Symbol() rune {
	defer t.Advance()
	return rune(t.body[0])
}

func (t *Tokenizer) MatchSym(s rune) bool {
	return rune(t.body[0]) == s
}

func (t *Tokenizer) PeekSym() rune {
	return rune(t.body[0])
}

func (t *Tokenizer) Identifier() string {
	defer t.Advance()
	return t.body
}

func (t *Tokenizer) IntVal() int {
	defer t.Advance()
	n, err := strconv.Atoi(t.body)
	if err != nil {
		log.Panic(err)
	}
	return n
}

func (t *Tokenizer) StringVal() string {
	defer t.Advance()
	return t.body
}
