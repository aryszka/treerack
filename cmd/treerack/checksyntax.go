package main

type checkOptions struct {
	command *commandOptions
	syntax  *fileOptions
}

func checkSyntax(args []string) int {
	var o checkOptions
	o.command = initOptions(checkSyntaxUsage, checkSyntaxExample, args)
	o.syntax = &fileOptions{flagSet: o.command.flagSet}

	o.command.flagSet.StringVar(&o.syntax.inline, "syntax-string", "", syntaxStringUsage)
	o.command.flagSet.StringVar(&o.syntax.fileName, "syntax", "", syntaxFileUsage)

	if o.command.checkHelp() {
		return 0
	}

	if code := o.command.parseArgs(); code != 0 {
		return code
	}

	o.syntax.positional = o.command.flagSet.Args()
	_, code := openSyntax(o.syntax)
	return code
}
