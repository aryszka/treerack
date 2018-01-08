package main

import (
	"bytes"
	"flag"
	"io"
	"os"

	"github.com/aryszka/treerack"
	"golang.org/x/crypto/ssh/terminal"
)

type syntaxOptions struct {
	syntax     string
	syntaxFile string
}

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

func open(options syntaxOptions, fs *flag.FlagSet) (*treerack.Syntax, int) {
	if options.syntaxFile != "" && options.syntax != "" {
		multipleSyntaxesError(fs)
		return nil, -1
	}

	var hasInput bool
	if options.syntaxFile == "" && options.syntax == "" {
		hasInput = isTest && rin != nil || !isTest && !terminal.IsTerminal(0)
	}

	if !hasInput && options.syntaxFile == "" && options.syntax == "" {
		missingSyntaxError(fs)
		return nil, -1
	}

	var input io.Reader
	if hasInput {
		input = rin
	} else if options.syntaxFile != "" {
		f, err := os.Open(options.syntaxFile)
		if err != nil {
			stderr(err)
			return nil, -1
		}

		defer f.Close()
		input = f
	} else if options.syntax != "" {
		input = bytes.NewBufferString(options.syntax)
	}

	s := &treerack.Syntax{}
	if err := s.ReadSyntax(input); err != nil {
		stderr(err)
		return nil, -1
	}

	return s, 0
}
