package treerack

import (
	"bytes"
	"io"
	"os"
)

func openSyntaxReader(r io.Reader) (*Syntax, error) {
	b, err := bootSyntax()
	if err != nil {
		return nil, err
	}

	doc, err := b.Parse(r)
	if err != nil {
		return nil, err
	}

	println("starting")
	s := &Syntax{}
	if err := define(s, doc); err != nil {
		return nil, err
	}

	if err := s.Init(); err != nil {
		return nil, err
	}

	return s, nil
}

func openSyntaxFile(file string) (*Syntax, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	return openSyntaxReader(f)
}

func openSyntaxString(syntax string) (*Syntax, error) {
	b := bytes.NewBufferString(syntax)
	return openSyntaxReader(b)
}
