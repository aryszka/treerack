package main

type checkSyntaxOptions struct {
	command *commandOptions
	syntax  *fileOptions
}

func checkSyntax(args []string) int {
	var o checkSyntaxOptions
	o.command = initOptions(checkSyntaxUsage, checkSyntaxExample, positionalSyntaxUsage, args)
	o.syntax = &fileOptions{typ: "syntax", flagSet: o.command.flagSet, positionalDoc: positionalSyntaxUsage}

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

	if err := s.Init(); err != nil {
		stderr(err)
		return -1
	}

	return 0
}
