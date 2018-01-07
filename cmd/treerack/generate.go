package main

import (
	"bytes"
	"flag"
	"io"
	"os"

	"github.com/aryszka/treerack"

	"golang.org/x/crypto/ssh/terminal"
)

type generateOptions struct {
	syntax      string
	syntaxFile  string
	packageName string
	export      bool
}

var isTest bool

func flagSet(o *generateOptions, output io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.Usage = func() {}
	fs.SetOutput(output)
	fs.StringVar(&o.syntax, "syntax-string", "", syntaxStringUsage)
	fs.StringVar(&o.syntaxFile, "syntax", "", syntaxFileUsage)
	fs.StringVar(&o.packageName, "package-name", "", packageNameUsage)
	fs.BoolVar(&o.export, "export", false, exportUsage)
	return fs
}

func helpGenerate() {
	stdout(generateUsage)
	stdout()
	stdout("Options:")
	fs := flagSet(&generateOptions{}, os.Stdout)
	fs.PrintDefaults()
	stdout()
	stdout(generateExample)
}

func flagError(fs *flag.FlagSet) {
	stderr()
	stderr("Options:")
	fs.PrintDefaults()
}

func multipleInputsError(fs *flag.FlagSet) {
	stderr("only one of syntax file or syntax string is allowed")
	stderr()
	stderr("Options:")
	fs.PrintDefaults()
}

func noInputError(fs *flag.FlagSet) {
	stderr("missing syntax input")
	stderr()
	stderr("Options:")
	fs.PrintDefaults()
}

func generate(args []string) int {
	if len(args) > 0 && args[0] == "-help" {
		helpGenerate()
		return 0
	}

	var options generateOptions
	fs := flagSet(&options, os.Stderr)
	if err := fs.Parse(args); err != nil {
		flagError(fs)
		return -1
	}

	if options.syntaxFile != "" && options.syntax != "" {
		multipleInputsError(fs)
		return -1
	}

	var hasInput bool
	if options.syntaxFile == "" && options.syntax == "" {
		fdint := int(os.Stdin.Fd())
		hasInput = !isTest && !terminal.IsTerminal(fdint)
	}

	if !hasInput && options.syntaxFile == "" && options.syntax == "" {
		noInputError(fs)
		return -1
	}

	var input io.Reader
	if hasInput {
		input = os.Stdin
	} else if options.syntaxFile != "" {
		f, err := os.Open(options.syntaxFile)
		if err != nil {
			stderr(err)
			return -1
		}

		defer f.Close()
		input = f
	} else if options.syntax != "" {
		input = bytes.NewBufferString(options.syntax)
	}

	s := &treerack.Syntax{}
	if err := s.ReadSyntax(input); err != nil {
		stderr(err)
		return -1
	}

	var goptions treerack.GeneratorOptions
	goptions.PackageName = options.packageName
	goptions.Export = options.export

	if err := s.Generate(goptions, os.Stdout); err != nil {
		stderr(err)
		return -1
	}

	return 0
}
