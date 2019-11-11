package main

//go:generate goyacc -o xdr.go -p xdr xdr.y

import (
	"fmt"
	"go/scanner"
	"go/token"
	"io/ioutil"
	"strconv"
)

type lexer struct {
	s scanner.Scanner
}

const eof = 0

func (l *lexer) Lex(lval *xdrSymType) int {
	pos, tok, lit := l.s.Scan()
	if tok == token.EOF {
		return eof
	}

	fmt.Printf("pos=%v, tok=%v, lit=%v\n", pos, tok, lit)
	var err error

	switch tok {
	case token.CONST:
		return KWCONST

	case token.STRUCT:
		return KWSTRUCT

	case token.TYPE:
		lval.ident = "type"
		return IDENT

	case token.SWITCH:
		return KWSWITCH

	case token.CASE:
		return KWCASE

	case token.DEFAULT:
		return KWDEFAULT

	case token.IDENT:
		switch lit {
		case "typedef":
			return KWTYPEDEF

		case "enum":
			return KWENUM

		case "union":
			return KWUNION

		case "void":
			return KWVOID

		default:
			lval.ident = lit
			return IDENT
		}

	case token.ASSIGN:
		return '='

	case token.SEMICOLON:
		return ';'

	case token.COLON:
		return ':'

	case token.LSS:
		return '<'

	case token.GTR:
		return '>'

	case token.LBRACK:
		return '['

	case token.RBRACK:
		return ']'

	case token.LBRACE:
		return '{'

	case token.RBRACE:
		return '}'

	case token.COMMA:
		return ','

	case token.LPAREN:
		return '('

	case token.RPAREN:
		return ')'

	case token.MUL:
		return '*'

	case token.INT:
		lval.num, err = strconv.ParseUint(lit, 0, 64)
		if err != nil {
			panic(err)
		}
		return NUM

	default:
		panic("token not handled")
	}
}

func (l *lexer) Error(e string) {
	panic(e)
}

func main() {
	filename := "rpc_nfs3_prot.x"
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	fset := token.NewFileSet()
	f := fset.AddFile(filename, -1, len(src))

	var l lexer
	l.s.Init(f, src, nil, 0)

	xdrParse(&l)
}
