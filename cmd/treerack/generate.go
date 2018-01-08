package main

import "github.com/aryszka/treerack"

type generateOptions struct {
	*syntaxOptions
	packageName string
	export      bool
}

func generate(args []string) int {
	var options generateOptions
	options.syntaxOptions = initOptions(generateUsage, generateExample, args)
	options.flagSet.BoolVar(&options.export, "export", false, exportUsage)
	options.flagSet.StringVar(&options.packageName, "package-name", "", packageNameUsage)

	if options.checkHelp() {
		return 0
	}

	if code := options.parse(); code != 0 {
		return code
	}

	var goptions treerack.GeneratorOptions
	goptions.PackageName = options.packageName
	goptions.Export = options.export

	s, code := openSyntax(options.syntaxOptions)
	if code != 0 {
		return code
	}

	if err := s.Generate(goptions, wout); err != nil {
		stderr(err)
		return -1
	}

	return 0
}
