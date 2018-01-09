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
	typ           string
	inline        string
	fileName      string
	positional    []string
	flagSet       *flag.FlagSet
	positionalDoc string
}

func (o *fileOptions) multipleInputsError() {
	stderr("only one", o.typ, "is allowed")
	stderr()
	stderr("Options:")
	o.flagSet.PrintDefaults()
	stderr()
	stderr(wrapLines(o.positionalDoc))
}

func (o *fileOptions) missingInputError() {
	stderr("missing", o.typ)
	stderr()
	stderr("Options:")
	o.flagSet.PrintDefaults()
	stderr()
	stderr(wrapLines(o.positionalDoc))
}

func (o *fileOptions) getSource() (hasInput bool, fileName string, inline string, code int) {
	if len(o.positional) > 1 {
		o.multipleInputsError()
		code = -1
		return
	}

	hasPositional := len(o.positional) == 1
	hasFile := o.fileName != ""
	hasInline := o.inline != ""

	var has bool
	for _, h := range []bool{hasPositional, hasFile, hasInline} {
		if h && has {
			o.multipleInputsError()
			code = -1
			return
		}

		has = h
	}

	switch {
	case hasPositional:
		fileName = o.positional[0]
		return
	case hasFile:
		fileName = o.fileName
		return
	case hasInline:
		inline = o.inline
		return
	}

	// check input last to allow explicit input in non-TTY environments:
	hasInput = isTest && rin != nil || !isTest && !terminal.IsTerminal(0)
	if !hasInput {
		o.missingInputError()
		code = -1
		return
	}

	return
}

func (o *fileOptions) open() (io.ReadCloser, int) {
	hasInput, fileName, inline, code := o.getSource()
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
		r = ioutil.NopCloser(bytes.NewBufferString(inline))
	}

	return r, 0
}

func (o *fileOptions) openSyntax() (*treerack.Syntax, int) {
	input, code := o.open()
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
