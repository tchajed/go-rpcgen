package main

import (
	"flag"
	"fmt"
	"go/format"
	"go/token"
	"io"
	"io/ioutil"
	"os"
)

var inputFile = flag.String("i", "", "Input file (.x)")
var outputFile = flag.String("o", "", "Output file (.go)")
var outputPackage = flag.String("p", "main", "Output package name")
var debugFlag = flag.Bool("d", false, "Debug parsing")
var out io.Writer

func main() {
	flag.Parse()

	if *inputFile == "" {
		fmt.Fprintf(os.Stderr, "Must specify input file (-i)\n")
		os.Exit(1)
	}

	if *outputFile == "" {
		fmt.Fprintf(os.Stderr, "Must specify output file (-o)\n")
		os.Exit(1)
	}

	src, err := ioutil.ReadFile(*inputFile)
	if err != nil {
		panic(err)
	}

	fset := token.NewFileSet()
	f := fset.AddFile(*inputFile, -1, len(src))

	var l lexer
	l.s.Init(f, src, nil, 0)

	outTmp := *outputFile + ".tmp"
	outf, err := os.OpenFile(outTmp, os.O_WRONLY|os.O_EXCL|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}

	defer os.Remove(outTmp)

	out = outf
	fmt.Fprintf(out, "package %s\n", *outputPackage)
	fmt.Fprintf(out, "import . \"github.com/zeldovich/go-rpcgen/xdr\"\n")
	xdrParse(&l)
	outf.Close()

	buf, err := ioutil.ReadFile(outTmp)
	if err != nil {
		panic(err)
	}

	buf, err = format.Source(buf)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(*outputFile, buf, 0666)
	if err != nil {
		panic(err)
	}
}
