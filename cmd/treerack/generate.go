package main

import "github.com/aryszka/treerack"

type generateOptions struct {
	command     *commandOptions
	syntax      *fileOptions
	packageName string
	export      bool
}

func generate(args []string) int {
	var o generateOptions
	o.command = initOptions(generateUsage, generateExample, positionalSyntaxUsage, args)
	o.syntax = &fileOptions{typ: "syntax", flagSet: o.command.flagSet, positionalDoc: positionalSyntaxUsage}

	o.command.boolFlag(&o.export, "export", exportUsage)
	o.command.stringFlag(&o.packageName, "package-name", packageNameUsage)
	o.command.stringFlag(&o.syntax.inline, "syntax-string", syntaxStringUsage)
	o.command.stringFlag(&o.syntax.fileName, "syntax", syntaxFileUsage)

	if o.command.help() {
		return 0
	}

	if code := o.command.parseArgs(); code != 0 {
		return code
	}

	o.syntax.positional = o.command.flagSet.Args()
	s, code := o.syntax.openSyntax()
	if code != 0 {
		return code
	}

	var g treerack.GeneratorOptions
	g.PackageName = o.packageName
	g.Export = o.export

	if err := s.Generate(g, wout); err != nil {
		stderr(err)
		return -1
	}

	return 0
}
