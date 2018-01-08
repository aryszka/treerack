package main

import (
	"bytes"
	"flag"
	"io"
	"io/ioutil"
	"os"

	"github.com/aryszka/treerack"
	"golang.org/x/crypto/ssh/terminal"
)

type fileOptions struct {
	inline     string
	fileName   string
	positional []string
	flagSet    *flag.FlagSet
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

func getSource(options *fileOptions) (hasInput bool, fileName string, syntax string, code int) {
	if len(options.positional) > 1 {
		multipleSyntaxesError(options.flagSet)
		code = -1
		return
	}

	hasPositional := len(options.positional) == 1
	hasFile := options.fileName != ""
	hasSyntax := options.inline != ""

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
		fileName = options.fileName
		return
	case hasSyntax:
		syntax = options.inline
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

func open(options *fileOptions) (io.ReadCloser, int) {
	hasInput, fileName, syntax, code := getSource(options)
	if code != 0 {
		return nil, code
	}

	var r io.ReadCloser
	if hasInput {
		r = ioutil.NopCloser(rin)
	} else if fileName != "" {
		f, err := os.Open(fileName)
		if err != nil {
			stderr(err)
			return nil, -1
		}

		r = f
	} else {
		r = ioutil.NopCloser(bytes.NewBufferString(syntax))
	}

	return r, 0
}

func openSyntax(options *fileOptions) (*treerack.Syntax, int) {
	input, code := open(options)
	if code != 0 {
		return nil, code
	}

	s := &treerack.Syntax{}
	if err := s.ReadSyntax(input); err != nil {
		stderr(err)
		return nil, -1
	}

	return s, 0
}
