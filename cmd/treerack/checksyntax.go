package main

type checkSyntaxOptions struct {
	command *commandOptions
	syntax  *fileOptions
}

func checkSyntax(args []string) int {
	var o checkSyntaxOptions
	o.command = initOptions(checkSyntaxUsage, checkSyntaxExample, args)
	o.syntax = &fileOptions{typ: "syntax", flagSet: o.command.flagSet}

	o.command.flagSet.StringVar(&o.syntax.inline, "syntax-string", "", syntaxStringUsage)
	o.command.flagSet.StringVar(&o.syntax.fileName, "syntax", "", syntaxFileUsage)

	if o.command.checkHelp() {
		return 0
	}

	if code := o.command.parseArgs(); code != 0 {
		return code
	}

	o.syntax.positional = o.command.flagSet.Args()
	_, code := o.syntax.openSyntax()
	return code
}
