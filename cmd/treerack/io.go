package main

import (
	"fmt"
	"io"
	"os"
)

type exitFunc func(int)

var (
	isTest bool
	rin    io.Reader = os.Stdin

	wout io.Writer = os.Stdout
	werr io.Writer = os.Stderr

	exit exitFunc = func(code int) {
		os.Exit(code)
	}
)

func stdout(a ...interface{}) {
	fmt.Fprintln(wout, a...)
}

func stderr(a ...interface{}) {
	fmt.Fprintln(werr, a...)
}
