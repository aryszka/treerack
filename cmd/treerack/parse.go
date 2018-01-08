package main

import (
	"flag"
	"io"
)

type parseOptions struct {
	syntaxOptions
	input     string
	inputFile string
	pretty    bool
	indent    string
}

func flagSetParse(o *parseOptions, output io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.Usage = func() {}
	fs.SetOutput(output)
	fs.StringVar(&o.syntax, "syntax-string", "", syntaxStringUsage)
	fs.StringVar(&o.syntaxFile, "syntax", "", syntaxFileUsage)
	return fs
}

func flagErrorParse(fs *flag.FlagSet) {
	stderr()
	stderr("Options:")
	fs.PrintDefaults()
}

func helpParse() {
	stdout(parseUsage)
	stdout()
	stdout("Options:")
	fs := flagSetParse(&parseOptions{}, wout)
	fs.PrintDefaults()
	stdout()
	stdout(parseExample)
	stdout()
	stdout(docRef)
}

func parse(args []string) int {
	return 0
}
