package main

import (
	"flag"
	"io"
)

type checkOptions struct {
	syntaxOptions
}

func flagSetCheck(o *checkOptions, output io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.Usage = func() {}
	fs.SetOutput(output)
	fs.StringVar(&o.syntax, "syntax-string", "", syntaxStringUsage)
	fs.StringVar(&o.syntaxFile, "syntax", "", syntaxFileUsage)
	return fs
}

func flagErrorCheck(fs *flag.FlagSet) {
	stderr()
	stderr("Options:")
	fs.PrintDefaults()
}

func helpCheck() {
	stdout(checkUsage)
	stdout()
	stdout("Options:")
	fs := flagSetCheck(&checkOptions{}, wout)
	fs.PrintDefaults()
	stdout()
	stdout(checkExample)
	stdout()
	stdout(docRef)
}

func check(args []string) int {
	if len(args) > 0 && args[0] == "-help" {
		helpCheck()
		return 0
	}

	var options checkOptions
	fs := flagSetCheck(&options, werr)
	if err := fs.Parse(args); err != nil {
		flagErrorCheck(fs)
		return -1
	}

	_, code := open(options.syntaxOptions, fs)
	return code
}
