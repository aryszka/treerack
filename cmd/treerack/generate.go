package main

import (
	"flag"
	"io"

	"github.com/aryszka/treerack"
)

type generateOptions struct {
	syntaxOptions
	packageName string
	export      bool
}

func flagSetGenerate(o *generateOptions, output io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.Usage = func() {}
	fs.SetOutput(output)
	fs.StringVar(&o.syntax, "syntax-string", "", syntaxStringUsage)
	fs.StringVar(&o.syntaxFile, "syntax", "", syntaxFileUsage)
	fs.StringVar(&o.packageName, "package-name", "", packageNameUsage)
	fs.BoolVar(&o.export, "export", false, exportUsage)
	return fs
}

func flagErrorGenerate(fs *flag.FlagSet) {
	stderr()
	stderr("Options:")
	fs.PrintDefaults()
}

func helpGenerate() {
	stdout(generateUsage)
	stdout()
	stdout("Options:")
	fs := flagSetGenerate(&generateOptions{}, wout)
	fs.PrintDefaults()
	stdout()
	stdout(generateExample)
	stdout()
	stdout(docRef)
}

func generate(args []string) int {
	if len(args) > 0 && args[0] == "-help" {
		helpGenerate()
		return 0
	}

	var options generateOptions
	fs := flagSetGenerate(&options, werr)
	if err := fs.Parse(args); err != nil {
		flagErrorGenerate(fs)
		return -1
	}

	s, code := open(options.syntaxOptions, fs)
	if code != 0 {
		return code
	}

	var goptions treerack.GeneratorOptions
	goptions.PackageName = options.packageName
	goptions.Export = options.export

	if err := s.Generate(goptions, wout); err != nil {
		stderr(err)
		return -1
	}

	return 0
}
