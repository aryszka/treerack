package main

import (
	"bytes"
	"flag"
	"io"
	"os"

	"github.com/aryszka/treerack"
	"golang.org/x/crypto/ssh/terminal"
)

func multipleSyntaxesError(fs *flag.FlagSet) {
	stderr("only one of syntax file or syntax string is allowed")
	stderr()
	stderr("Options:")
	fs.PrintDefaults()
}

func missingSyntaxError(fs *flag.FlagSet) {
	stderr("missing syntax input")
	stderr()
	stderr("Options:")
	fs.PrintDefaults()
}

func getSource(options *syntaxOptions) (hasInput bool, fileName string, syntax string, code int) {
	if len(options.positional) > 1 {
		multipleSyntaxesError(options.flagSet)
		code = -1
		return
	}

	hasPositional := len(options.positional) == 1
	hasFile := options.syntaxFile != ""
	hasSyntax := options.syntax != ""

	var has bool
	for _, h := range []bool{hasPositional, hasFile, hasSyntax} {
		if h && has {
			multipleSyntaxesError(options.flagSet)
			code = -1
			return
		}

		has = h
	}

	switch {
	case hasPositional:
		fileName = options.positional[0]
		return
	case hasFile:
		fileName = options.syntaxFile
		return
	case hasSyntax:
		syntax = options.syntax
		return
	}

	// check input last to allow explicit syntax in non-TTY environments:
	hasInput = isTest && rin != nil || !isTest && !terminal.IsTerminal(0)
	if !hasInput {
		missingSyntaxError(options.flagSet)
		code = -1
		return
	}

	return
}

func openSyntax(options *syntaxOptions) (*treerack.Syntax, int) {
	hasInput, fileName, syntax, code := getSource(options)
	if code != 0 {
		return nil, code
	}

	var input io.Reader
	if hasInput {
		input = rin
	} else if fileName != "" {
		f, err := os.Open(fileName)
		if err != nil {
			stderr(err)
			return nil, -1
		}

		defer f.Close()
		input = f
	} else {
		input = bytes.NewBufferString(syntax)
	}

	s := &treerack.Syntax{}
	if err := s.ReadSyntax(input); err != nil {
		stderr(err)
		return nil, -1
	}

	return s, 0
}
